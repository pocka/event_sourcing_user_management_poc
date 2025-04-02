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
	"pocka.jp/x/event_sourcing_user_management_poc/gen/projection"
)

func build(events []proto.Message) *projection.InitialAdminCreationPassword {
	var p projection.InitialAdminCreationPassword

	for _, e := range events {
		apply(e, &p)
	}

	return &p
}

func TestReturnsNonNil(t *testing.T) {
	p := build([]proto.Message{
		&event.InitialAdminCreationPasswordCreated{
			PasswordHash: []byte{0, 1, 2},
			Salt:         []byte{3, 4, 5},
		},
	})

	if p.PasswordHash == nil {
		t.Error("Expected found password, got nil")
	}

	if !bytes.Equal([]byte{0, 1, 2}, p.PasswordHash) {
		t.Errorf("Hash does not match to [0,1,2]: %v", p.PasswordHash)
	}

	if !bytes.Equal([]byte{3, 4, 5}, p.Salt) {
		t.Errorf("Salt does not match to [3,4,5]: %v", p.Salt)
	}
}

func TestAdminCreationExpiresOne(t *testing.T) {
	p := build([]proto.Message{
		&event.InitialAdminCreationPasswordCreated{
			PasswordHash: []byte{0, 1, 2},
			Salt:         []byte{3, 4, 5},
		},
		&event.RoleAssigned{
			UserId: proto.String(""),
			Role:   model.Role_ROLE_ADMIN.Enum(),
		},
	})

	if p.PasswordHash != nil {
		t.Errorf("Expected nil, got %v", p)
	}
}

func TestNonAdminCreationShouldNotExpiresOne(t *testing.T) {
	p := build([]proto.Message{
		&event.InitialAdminCreationPasswordCreated{
			PasswordHash: []byte{0, 1, 2},
			Salt:         []byte{3, 4, 5},
		},
		&event.RoleAssigned{
			UserId: proto.String(""),
			Role:   model.Role_ROLE_EDITOR.Enum(),
		},
	})

	if p.PasswordHash == nil {
		t.Errorf("Expected non-nil, got nil")
	}
}
