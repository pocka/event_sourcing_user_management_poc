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
	"database/sql"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"pocka.jp/x/event_sourcing_user_management_poc/auth"
	"pocka.jp/x/event_sourcing_user_management_poc/events"
	"pocka.jp/x/event_sourcing_user_management_poc/gen/event"
	"pocka.jp/x/event_sourcing_user_management_poc/gen/model"
	"pocka.jp/x/event_sourcing_user_management_poc/projections/initial_admin_creation_password"
	"pocka.jp/x/event_sourcing_user_management_poc/projections/users"
)

// CreateAlice creates a new admin user named "Alice" with demo password of
// "Alice's password".
// CreateAlice returns an ID of the created user on success.
func CreateAlice(db *sql.DB, logger *log.Logger) (string, error) {
	id := uuid.New().String()

	passwordHash, salt := auth.HashPasswordWithRandomSalt("Alice's password")

	if err := events.Insert(db, []proto.Message{
		&event.UserCreated{
			Id:          proto.String(id),
			DisplayName: proto.String("Alice"),
			Email:       proto.String("alice@example.com"),
		},
		&event.PasswordLoginConfigured{
			UserId:       proto.String(id),
			PasswordHash: passwordHash,
			Salt:         salt,
		},
		&event.RoleAssigned{
			UserId: proto.String(id),
			Role:   model.Role.Enum(model.Role_ROLE_ADMIN),
		},
	}); err != nil {
		return "", fmt.Errorf("Unable to create Alice: %s", err)
	}

	go func() {
		logger.Debug("Creating initial admin creation password snapshot (trigger=create alice)")

		if err := initial_admin_creation_password.SaveSnapshot(db); err != nil {
			logger.Warnf("Failed to update initial admin creation password snapshot: %s", err)
		} else {
			logger.Debug("Created initial admin creation password snapshot (trigger=create alice)")
		}

		logger.Debug("Creating snapshot (trigger=create alice)")

		if err := users.SaveSnapshot(db); err != nil {
			logger.Warnf("Failed to create user snapshot: %s", err)
		} else {
			logger.Debug("Created snapshot (trigger=create alice)")
		}
	}()

	return id, nil
}
