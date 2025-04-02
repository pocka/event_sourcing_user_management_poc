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
	"pocka.jp/x/event_sourcing_user_management_poc/gen/projection"
)

func build(events []proto.Message) *projection.UsersProjection {
	var p projection.UsersProjection

	for _, e := range events {
		apply(e, &p)
	}

	return &p
}

func TestIdentityOnly(t *testing.T) {
	p := build([]proto.Message{
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

	if len(p.Users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(p.Users))
	}

	if *p.Users[0].Id != "foo" {
		t.Errorf("Expected ID \"foo\", got \"%s\"", *p.Users[0].Id)
	}

	if *p.Users[0].DisplayName != "Foo" {
		t.Errorf("Expected DisplayName \"Foo\", got \"%s\"", *p.Users[0].DisplayName)
	}

	if *p.Users[0].Email != "foo@example.com" {
		t.Errorf("Expected Email \"foo@example.com\", got \"%s\"", *p.Users[0].Email)
	}

	if p.Users[0].Role != nil {
		t.Errorf("Expected Role to be nil, got %v", p.Users[0].Role)
	}

	if p.Users[0].PasswordLogin != nil {
		t.Errorf("Expected Role to be nil, got %v", p.Users[0].Role)
	}
}

func TestWithRole(t *testing.T) {
	p := build([]proto.Message{
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

	if *p.Users[0].Role != model.Role_ROLE_ADMIN {
		t.Errorf("Expected Role_ROLE_ADMIN, got %v", p.Users[0].Role.String())
	}
}

func TestWithPWLogin(t *testing.T) {
	p := build([]proto.Message{
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

	if !bytes.Equal(p.Users[0].PasswordLogin.Hash, []byte{0, 1, 2}) {
		t.Errorf("Expected [0,1,2], got %v", p.Users[0].PasswordLogin.Hash)
	}

	if !bytes.Equal(p.Users[0].PasswordLogin.Salt, []byte{3, 4, 5}) {
		t.Errorf("Expected [3,4,5], got %v", p.Users[0].PasswordLogin.Salt)
	}
}
