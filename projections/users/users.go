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
	"google.golang.org/protobuf/proto"

	"pocka.jp/x/event_sourcing_user_management_poc/gen/event"
	"pocka.jp/x/event_sourcing_user_management_poc/gen/model"
)

type passwordLogin struct {
	Hash []byte
	Salt []byte
}

type user struct {
	ID          string
	DisplayName string
	Email       string

	PasswordLogin *passwordLogin
	Role          *model.Role
}

func ListFromUserEvents(events []proto.Message) []user {
	users := make(map[string]*user)

	for _, e := range events {
		switch v := e.(type) {
		case *event.UserCreated:
			users[*v.Id] = &user{
				ID:          *v.Id,
				DisplayName: *v.DisplayName,
				Email:       *v.Email,
			}
		case *event.PasswordLoginConfigured:
			if v.UserId == nil {
				break
			}

			found := users[*v.UserId]
			if found != nil {
				found.PasswordLogin = &passwordLogin{
					Hash: v.PasswordHash,
					Salt: v.Salt,
				}
			}
		case *event.RoleAssigned:
			if v.UserId == nil {
				break
			}

			found := users[*v.UserId]
			if found != nil {
				found.Role = v.Role
			}
		}
	}

	ret := make([]user, 0, len(users))

	for _, u := range users {
		ret = append(ret, *u)
	}

	return ret
}
