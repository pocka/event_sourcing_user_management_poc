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
	"bytes"
	"testing"

	"google.golang.org/protobuf/proto"

	"pocka.jp/x/event_sourcing_user_management_poc/gen/event"
	"pocka.jp/x/event_sourcing_user_management_poc/gen/model"
)

func TestReturnsNonNil(t *testing.T) {
	password := GetFromUserEvents([]proto.Message{
		&event.InitialAdminCreationPasswordCreated{
			PasswordHash: []byte{0, 1, 2},
			Salt:         []byte{3, 4, 5},
		},
	})

	if password == nil {
		t.Error("Expected found password, got nil")
	}

	if !bytes.Equal([]byte{0, 1, 2}, password.Hash) {
		t.Errorf("Hash does not match to [0,1,2]: %v", password.Hash)
	}

	if !bytes.Equal([]byte{3, 4, 5}, password.Salt) {
		t.Errorf("Salt does not match to [3,4,5]: %v", password.Salt)
	}
}

func TestAdminCreationExpiresOne(t *testing.T) {
	password := GetFromUserEvents([]proto.Message{
		&event.InitialAdminCreationPasswordCreated{
			PasswordHash: []byte{0, 1, 2},
			Salt:         []byte{3, 4, 5},
		},
		&event.RoleAssigned{
			UserId: proto.String(""),
			Role:   model.Role_ROLE_ADMIN.Enum(),
		},
	})

	if password != nil {
		t.Errorf("Expected nil, got %v", password)
	}
}

func TestNonAdminCreationShouldNotExpiresOne(t *testing.T) {
	password := GetFromUserEvents([]proto.Message{
		&event.InitialAdminCreationPasswordCreated{
			PasswordHash: []byte{0, 1, 2},
			Salt:         []byte{3, 4, 5},
		},
		&event.RoleAssigned{
			UserId: proto.String(""),
			Role:   model.Role_ROLE_EDITOR.Enum(),
		},
	})

	if password == nil {
		t.Errorf("Expected non-nil, got nil")
	}
}
