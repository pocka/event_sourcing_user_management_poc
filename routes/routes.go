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
	"html/template"
	"net/http"
	"time"

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

//go:embed initial_admin_creation.html
var initialAdminCreationHtml string

//go:embed logged_in.html.tmpl
var loggedInHTMLTmpl string

//go:embed login.html
var loginHTML string

type loggedInAdminPipeline struct {
	DisplayName string
	Role        string
}

func Handler(db *sql.DB, logger *log.Logger) (http.Handler, error) {
	mux := http.NewServeMux()

	loggedInAdminHtml, err := template.New("loggedInAdminHtml").Parse(loggedInHTMLTmpl)
	if err != nil {
		return nil, err
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		initialAdminPass, _, err := initial_admin_creation_password.GetProjection(db)
		if err != nil {
			logger.Errorf("Error loading initial admin creation password: %s", err)
			w.Header().Add("Content-Type", "text/html;charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, loginHTML)
			return
		}

		if initialAdminPass.PasswordHash != nil {
			fmt.Fprint(w, initialAdminCreationHtml)
			return
		}

		id, err := r.Cookie("id")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, loginHTML)
			return
		}

		p, _, err := users.GetProjection(db)
		if err != nil {
			w.Header().Add("Content-Type", "text/html;charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, loginHTML)
			return
		}

		for _, user := range p.Users {
			// No real auth. No security.
			if *user.Id == id.Value {
				loggedInAdminHtml.Execute(w, loggedInAdminPipeline{
					DisplayName: *user.DisplayName,
					Role:        user.Role.String(),
				})
				return
			}
		}

		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, loginHTML)
		return
	})

	mux.HandleFunc("/initial-admin", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Not found", http.StatusMethodNotAllowed)
			return
		}

		initialAdminPass, _, err := initial_admin_creation_password.GetProjection(db)
		if err != nil {
			logger.Errorf("Error loading initial admin creation password: %s", err)
			w.Header().Add("Content-Type", "text/html;charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, loginHTML)
			return
		}

		if initialAdminPass.PasswordHash == nil {
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
		if !bytes.Equal(initialAdminPass.PasswordHash, initPwHash) {
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
				UserId:       proto.String(id),
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

		go func() {
			logger.Debug("Creating initial admin creation password snapshot (trigger=initial admin creation)")

			if err := initial_admin_creation_password.SaveSnapshot(db); err != nil {
				logger.Warnf("Failed to update initial admin creation password snapshot: %s", err)
			} else {
				logger.Debug("Created initial admin creation password snapshot (trigger=initial admin creation)")
			}

			logger.Debug("Creating snapshot (trigger=initial admin creation)")

			if err := users.SaveSnapshot(db); err != nil {
				logger.Warnf("Failed to create user snapshot: %s", err)
			} else {
				logger.Debug("Created snapshot (trigger=initial admin creation)")
			}
		}()

		// This project is PoC for event sourcing. UI and security is completely out-of-scope.
		http.SetCookie(w, &http.Cookie{
			Name:  "id",
			Value: id,
		})

		http.Redirect(w, r, "/", http.StatusFound)
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		p, _, err := users.GetProjection(db)
		if err != nil {
			w.Header().Add("Content-Type", "text/html;charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, loginHTML)
			return
		}

		email := r.PostForm.Get("email")
		password := r.PostForm.Get("password")

		if email == "" || password == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, loginHTML)
			return
		}

		for _, user := range p.Users {
			// No real auth. No security.
			if *user.Email == email && user.PasswordLogin != nil {
				hash := auth.HashPassword(password, user.PasswordLogin.Salt)
				if bytes.Equal(user.PasswordLogin.Hash, hash) {
					// This project is PoC for event sourcing. UI and security is completely out-of-scope.
					http.SetCookie(w, &http.Cookie{
						Name:  "id",
						Value: *user.Id,
					})

					http.Redirect(w, r, "/", http.StatusFound)
					return
				}
			}
		}

		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, loginHTML)
		return
	})

	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:    "id",
			Value:   "",
			Expires: time.Now(),
		})

		http.Redirect(w, r, "/", http.StatusFound)
	})

	return mux, nil
}
