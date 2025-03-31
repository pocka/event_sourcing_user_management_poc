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
	"bytes"
	"testing"

	"google.golang.org/protobuf/proto"

	"pocka.jp/x/event_sourcing_user_management_poc/gen/event"
	"pocka.jp/x/event_sourcing_user_management_poc/gen/model"
)

func TestIdentityOnly(t *testing.T) {
	users := ListFromUserEvents([]proto.Message{
		&event.UserCreated{
			Id:          proto.String("foo"),
			DisplayName: proto.String("Foo"),
			Email:       proto.String("foo@example.com"),
		},
		&event.RoleAssigned{
			UserId: proto.String("bar"),
			Role:   model.Role_ROLE_EDITOR.Enum(),
		},
	})

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}

	if users[0].ID != "foo" {
		t.Errorf("Expected ID \"foo\", got \"%s\"", users[0].ID)
	}

	if users[0].DisplayName != "Foo" {
		t.Errorf("Expected DisplayName \"Foo\", got \"%s\"", users[0].DisplayName)
	}

	if users[0].Email != "foo@example.com" {
		t.Errorf("Expected Email \"foo@example.com\", got \"%s\"", users[0].Email)
	}

	if users[0].Role != nil {
		t.Errorf("Expected Role to be nil, got %v", users[0].Role)
	}

	if users[0].PasswordLogin != nil {
		t.Errorf("Expected Role to be nil, got %v", users[0].Role)
	}
}

func TestWithRole(t *testing.T) {
	users := ListFromUserEvents([]proto.Message{
		&event.UserCreated{
			Id:          proto.String("foo"),
			DisplayName: proto.String("Foo"),
			Email:       proto.String("foo@example.com"),
		},
		&event.RoleAssigned{
			UserId: proto.String("foo"),
			Role:   model.Role_ROLE_ADMIN.Enum(),
		},
	})

	if *users[0].Role != model.Role_ROLE_ADMIN {
		t.Errorf("Expected Role_ROLE_ADMIN, got %v", users[0].Role.String())
	}
}

func TestWithPWLogin(t *testing.T) {
	users := ListFromUserEvents([]proto.Message{
		&event.UserCreated{
			Id:          proto.String("foo"),
			DisplayName: proto.String("Foo"),
			Email:       proto.String("foo@example.com"),
		},
		&event.PasswordLoginConfigured{
			UserId:       proto.String("foo"),
			PasswordHash: []byte{0, 1, 2},
			Salt:         []byte{3, 4, 5},
		},
	})

	if !bytes.Equal(users[0].PasswordLogin.Hash, []byte{0, 1, 2}) {
		t.Errorf("Expected [0,1,2], got %v", users[0].PasswordLogin.Hash)
	}

	if !bytes.Equal(users[0].PasswordLogin.Salt, []byte{3, 4, 5}) {
		t.Errorf("Expected [3,4,5], got %v", users[0].PasswordLogin.Salt)
	}
}
