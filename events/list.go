// Copyright 2025 Shota FUJI
//
// This source code is licensed under Zero-Clause BSD License.
// You can find a copy of the Zero-Clause BSD License at LICENSES/0BSD.txt
// You may also obtain a copy of the Zero-Clause BSD License at
// <https://opensource.org/license/0bsd>
//
// SPDX-License-Identifier: 0BSD

package events

import (
	"context"
	"database/sql"
	"fmt"

	"google.golang.org/protobuf/proto"

	"pocka.jp/x/event_sourcing_user_management_poc/gen/event"
)

func List(db *sql.DB) ([]proto.Message, error) {
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to begin transaction for listing user_events: %s", err)
	}
	defer tx.Rollback()

	var rowCount int
	if err := tx.QueryRow("SELECT count(*) FROM user_events").Scan(&rowCount); err != nil {
		return nil, fmt.Errorf("Failed to count user_events: %s", err)
	}

	if rowCount == 0 {
		return []proto.Message{}, nil
	}

	rows, err := tx.Query("SELECT event_name, payload FROM user_events ORDER BY seq ASC")
	if err != nil {
		return nil, fmt.Errorf("Failed to SELECT user_events: %s", err)
	}

	events := make([]proto.Message, rowCount)
	for i := range events {
		if !rows.Next() {
			return nil, fmt.Errorf("Number of events is less than rowCount")
		}

		var eventName string
		var payload []byte
		if err := rows.Scan(&eventName, &payload); err != nil {
			return nil, fmt.Errorf("Failed to scan user event: %s", err)
		}

		switch eventName {
		case "InitialAdminCreationPasswordCreated":
			var event event.InitialAdminCreationPasswordCreated
			if err := proto.Unmarshal(payload, &event); err != nil {
				return nil, fmt.Errorf("Illegal InitialAdminCreationPasswordCreated event: %s", err)
			}
			events[i] = &event
		case "UserCreated":
			var event event.UserCreated
			if err := proto.Unmarshal(payload, &event); err != nil {
				return nil, fmt.Errorf("Illegal UserCreated event: %s", err)
			}
			events[i] = &event
		case "PasswordLoginConfigured":
			var event event.PasswordLoginConfigured
			if err := proto.Unmarshal(payload, &event); err != nil {
				return nil, fmt.Errorf("Illegal PasswordLoginConfigured event: %s", err)
			}
			events[i] = &event
		case "RoleAssigned":
			var event event.RoleAssigned
			if err := proto.Unmarshal(payload, &event); err != nil {
				return nil, fmt.Errorf("Illegal RoleAssigned event: %s", err)
			}
			events[i] = &event
		default:
			return nil, fmt.Errorf("Unknown event in user_events: name=%s", eventName)
		}
	}

	return events, nil
}
