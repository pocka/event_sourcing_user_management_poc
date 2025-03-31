// Copyright 2025 Shota FUJI
//
// This source code is licensed under Zero-Clause BSD License.
// You can find a copy of the Zero-Clause BSD License at LICENSES/0BSD.txt
// You may also obtain a copy of the Zero-Clause BSD License at
// <https://opensource.org/license/0bsd>
//
// SPDX-License-Identifier: 0BSD

package routes

import (
	"bytes"
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"pocka.jp/x/event_sourcing_user_management_poc/auth"
	"pocka.jp/x/event_sourcing_user_management_poc/events"
	"pocka.jp/x/event_sourcing_user_management_poc/gen/event"
	"pocka.jp/x/event_sourcing_user_management_poc/gen/model"
	"pocka.jp/x/event_sourcing_user_management_poc/projections/initial_admin_creation_password"
)

//go:embed index.html
var initialAdminCreationHtml string

func Handler(db *sql.DB, logger *log.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		events, err := events.List(db)
		if err != nil {
			logger.Error(err)
			http.Error(w, "Server error: event loading failure", http.StatusInternalServerError)
			return
		}

		initialAdminPass := initial_admin_creation_password.GetFromUserEvents(events)
		if initialAdminPass != nil {
			fmt.Fprint(w, initialAdminCreationHtml)
			return
		}

		fmt.Fprintf(w, "TODO")
	})

	mux.HandleFunc("/initial-admin", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Not found", http.StatusMethodNotAllowed)
			return
		}

		evs, err := events.List(db)
		if err != nil {
			logger.Error(err)
			http.Error(w, "Server error: event loading failure", http.StatusInternalServerError)
			return
		}

		initialAdminPass := initial_admin_creation_password.GetFromUserEvents(evs)
		if initialAdminPass == nil {
			logger.Debug("Found no active initial admin creation password at POST /initial-admin, redirecting")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		r.ParseForm()

		username := r.PostForm.Get("username")
		email := r.PostForm.Get("email")
		password := r.PostForm.Get("password")
		initPassword := r.PostForm.Get("init_password")

		if username == "" || email == "" || password == "" || initPassword == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, initialAdminCreationHtml)
			return
		}

		initPwHash := auth.HashPassword(initPassword, initialAdminPass.Salt)
		if !bytes.Equal(initialAdminPass.Hash, initPwHash) {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, initialAdminCreationHtml)
			return
		}

		pwHash, salt := auth.HashPasswordWithRandomSalt(password)

		id := uuid.New().String()

		if err := events.Insert(db, []proto.Message{
			&event.UserCreated{
				Id:          proto.String(id),
				DisplayName: proto.String(username),
				Email:       proto.String(email),
			},
			&event.PasswordLoginConfigured{
				PasswordHash: pwHash,
				Salt:         salt,
			},
			&event.RoleAssigned{
				UserId: proto.String(id),
				Role:   model.Role.Enum(model.Role_ROLE_ADMIN),
			},
		}); err != nil {
			logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, initialAdminCreationHtml)
			return
		}

		// This project is PoC for event sourcing. UI and security is completely out-of-scope.
		http.SetCookie(w, &http.Cookie{
			Name:  "id",
			Value: id,
		})

		http.Redirect(w, r, "/", http.StatusFound)
	})

	return mux
}
