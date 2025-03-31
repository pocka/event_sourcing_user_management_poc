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
	"google.golang.org/protobuf/proto"

	"pocka.jp/x/event_sourcing_user_management_poc/gen/event"
	"pocka.jp/x/event_sourcing_user_management_poc/gen/model"
)

type initialAdminCreationPassword struct {
	Hash []byte
	Salt []byte
}

func GetFromUserEvents(events []proto.Message) *initialAdminCreationPassword {
	var password *initialAdminCreationPassword

	for _, e := range events {
		switch v := e.(type) {
		case *event.InitialAdminCreationPasswordCreated:
			password = &initialAdminCreationPassword{
				Hash: v.PasswordHash,
				Salt: v.Salt,
			}
		case *event.RoleAssigned:
			if *v.Role == model.Role_ROLE_ADMIN {
				password = nil
			}
		}
	}

	return password
}
