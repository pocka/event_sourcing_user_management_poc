// Copyright 2025 Shota FUJI
//
// This source code is licensed under Zero-Clause BSD License.
// You can find a copy of the Zero-Clause BSD License at LICENSES/0BSD.txt
// You may also obtain a copy of the Zero-Clause BSD License at
// <https://opensource.org/license/0bsd>
//
// SPDX-License-Identifier: 0BSD

package initial_admin_creation_password

import (
	"context"
	"database/sql"
	"fmt"

	"google.golang.org/protobuf/proto"

	"pocka.jp/x/event_sourcing_user_management_poc/events"
	"pocka.jp/x/event_sourcing_user_management_poc/gen/event"
	"pocka.jp/x/event_sourcing_user_management_poc/gen/model"
	"pocka.jp/x/event_sourcing_user_management_poc/gen/projection"
)

func GetProjection(db *sql.DB) (*projection.InitialAdminCreationPassword, int, error) {
	ctx := context.Background()

	var p projection.InitialAdminCreationPassword

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to begin transaction for InitialAdminCreationPasswordProjection: %s", err)
	}
	defer tx.Rollback()

	var eventSeq int
	var payload []byte

	err = tx.QueryRow("SELECT event_seq, payload FROM initial_admin_creation_password_snapshots ORDER BY event_seq DESC LIMIT 1").Scan(&eventSeq, &payload)
	if err == sql.ErrNoRows {
		p = projection.InitialAdminCreationPassword{}
		eventSeq = -1
	} else if err != nil {
		return nil, 0, fmt.Errorf("Failed to get latest snapshot: %s", err)
	} else {
		if err := proto.Unmarshal(payload, &p); err != nil {
			return nil, 0, fmt.Errorf("Failed to decode latest snapshot: %s", err)
		}
	}

	stmt, err := tx.Prepare("SELECT seq, event_name, payload FROM user_events WHERE seq > ? ORDER BY seq ASC")
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to prepare event fetching query: %s", err)
	}

	maxSeq := -1
	rows, err := stmt.Query(eventSeq)
	for rows.Next() {
		ev, seq, err := events.ScanEvent(rows)
		if err != nil {
			return nil, 0, err
		}

		maxSeq = max(maxSeq, seq)

		apply(ev, &p)
	}

	return &p, maxSeq, nil
}

func apply(e proto.Message, p *projection.InitialAdminCreationPassword) {
	switch v := e.(type) {
	case *event.InitialAdminCreationPasswordCreated:
		p.PasswordHash = v.PasswordHash
		p.Salt = v.Salt
	case *event.RoleAssigned:
		if *v.Role == model.Role_ROLE_ADMIN {
			p.PasswordHash = nil
			p.Salt = nil
		}
	}
}

func SaveSnapshot(db *sql.DB) error {
	p, seq, err := GetProjection(db)
	if err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT OR ABORT INTO initial_admin_creation_password_snapshots (event_seq, payload) VALUES (?, ?)")
	if err != nil {
		return err
	}

	payload, err := proto.Marshal(p)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(seq, payload)

	return err
}
