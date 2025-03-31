// Copyright 2025 Shota FUJI
//
// This source code is licensed under Zero-Clause BSD License.
// You can find a copy of the Zero-Clause BSD License at LICENSES/0BSD.txt
// You may also obtain a copy of the Zero-Clause BSD License at
// <https://opensource.org/license/0bsd>
//
// SPDX-License-Identifier: 0BSD

package setups

import (
	"crypto/rand"
	"database/sql"
	"fmt"

	"google.golang.org/protobuf/proto"

	"pocka.jp/x/event_sourcing_user_management_poc/auth"
	"pocka.jp/x/event_sourcing_user_management_poc/events"
	"pocka.jp/x/event_sourcing_user_management_poc/gen/event"
)

// InitAdminCreationPassword inserts InitialAdminCreationPasswordCreated event then
// returns the generated password. As the database resets every server starts, this
// function does not check whether there are events in the stream. This would be
// inefficient in real-world use cases.
func InitAdminCreationPassword(db *sql.DB) (string, error) {
	password := rand.Text()

	passwordHash, salt := auth.HashPasswordWithRandomSalt(password)

	if err := events.Insert(db, []proto.Message{
		&event.InitialAdminCreationPasswordCreated{
			PasswordHash: passwordHash,
			Salt:         salt,
		},
	}); err != nil {
		return "", fmt.Errorf("Unable to create initial admin creation password: %s", err)
	}

	return password, nil
}
