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
	"reflect"

	"google.golang.org/protobuf/proto"
)

func Insert(db *sql.DB, events []proto.Message) error {
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Failed to begin transaction for events insertion: %s", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT OR ABORT INTO user_events (payload, event_name) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("Failed to prepare INSERT statement for event insertion: %s", err)
	}

	for _, event := range events {
		// event is proto.Message, which is interface type. So we have to
		// get the type of the one it points to. Without "Indirect()", we get
		// something like "*event.UserCreated"
		eventName := reflect.Indirect(reflect.ValueOf(event)).Type().Name()

		data, err := proto.Marshal(event)
		if err != nil {
			return fmt.Errorf("Serializing of %s failed: %s", eventName, err)
		}

		if _, err := stmt.Exec(data, eventName); err != nil {
			return fmt.Errorf("Failed to INSERT %s: %s", eventName, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("Failed to commit transaction for events insertion: %s", err)
	}

	return nil
}
