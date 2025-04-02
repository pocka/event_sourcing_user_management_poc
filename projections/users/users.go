// Copyright 2025 Shota FUJI
//
// This source code is licensed under Zero-Clause BSD License.
// You can find a copy of the Zero-Clause BSD License at LICENSES/0BSD.txt
// You may also obtain a copy of the Zero-Clause BSD License at
// <https://opensource.org/license/0bsd>
//
// SPDX-License-Identifier: 0BSD

package users

import (
	"context"
	"database/sql"
	"fmt"

	"google.golang.org/protobuf/proto"

	"pocka.jp/x/event_sourcing_user_management_poc/events"
	"pocka.jp/x/event_sourcing_user_management_poc/gen/event"
	"pocka.jp/x/event_sourcing_user_management_poc/gen/projection"
)

func GetProjection(db *sql.DB) (*projection.UsersProjection, int, error) {
	ctx := context.Background()

	var p projection.UsersProjection

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to begin transaction for UsersProjection: %s", err)
	}
	defer tx.Rollback()

	var eventSeq int
	var payload []byte

	err = tx.QueryRow("SELECT event_seq, payload FROM users_snapshots ORDER BY event_seq DESC LIMIT 1").Scan(&eventSeq, &payload)
	if err == sql.ErrNoRows {
		p = projection.UsersProjection{
			Users: []*projection.User{},
		}
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

func apply(ev proto.Message, p *projection.UsersProjection) {
	switch v := ev.(type) {
	case *event.UserCreated:
		p.Users = append(p.Users, &projection.User{
			Id:          v.Id,
			DisplayName: v.DisplayName,
			Email:       v.Email,
		})
		return
	case *event.PasswordLoginConfigured:
		if v.UserId == nil {
			return
		}

		for _, user := range p.Users {
			if *user.Id != *v.UserId {
				continue
			}

			user.PasswordLogin = &projection.User_PasswordLogin{
				Hash: v.PasswordHash,
				Salt: v.Salt,
			}
			return
		}
		return
	case *event.RoleAssigned:
		if v.UserId == nil {
			return
		}

		for _, user := range p.Users {
			if *user.Id != *v.UserId {
				continue
			}

			user.Role = v.Role
			return
		}
		return
	}
}

func SaveSnapshot(db *sql.DB) error {
	p, seq, err := GetProjection(db)
	if err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT OR ABORT INTO users_snapshots (event_seq, payload) VALUES (?, ?)")
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
