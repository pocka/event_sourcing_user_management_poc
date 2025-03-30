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

	"golang.org/x/crypto/argon2"
	"google.golang.org/protobuf/proto"

	"pocka.jp/x/event_sourcing_user_management_poc/gen/event"
)

// InitAdminCreationPassword inserts InitialAdminCreationPasswordCreated event then
// returns the generated password. As the database resets every server starts, this
// function does not check whether there are events in the stream. This would be
// inefficient in real-world use cases.
func InitAdminCreationPassword(db *sql.DB) (string, error) {
	password := rand.Text()

	salt := make([]byte, 32)
	// rand.Read never returns an error.
	// https://pkg.go.dev/crypto/rand@go1.24.1#Read
	rand.Read(salt)

	// Parameters recommended in RFC (according to Go docs)
	passwordHash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	ev := &event.InitialAdminCreationPasswordCreated{
		PasswordHash: passwordHash,
		Salt:         salt,
	}

	data, err := proto.Marshal(ev)
	if err != nil {
		return "", fmt.Errorf("Failed to encode InitialAdminCreationPasswordCreated message: %s", err)
	}

	stmt, err := db.Prepare("INSERT OR ABORT INTO user_events (payload) VALUES (?)")
	if err != nil {
		return "", fmt.Errorf("Failed to prepare INSERT query: %s", err)
	}

	if _, err := stmt.Exec(data); err != nil {
		return "", fmt.Errorf("Failed to INSERT InitialAdminCreationPasswordCreated: %s", err)
	}

	return password, nil
}
