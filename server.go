// Copyright 2025 Shota FUJI
//
// This source code is licensed under Zero-Clause BSD License.
// You can find a copy of the Zero-Clause BSD License at LICENSES/0BSD.txt
// You may also obtain a copy of the Zero-Clause BSD License at
// <https://opensource.org/license/0bsd>
//
// SPDX-License-Identifier: 0BSD

package main

import (
	"database/sql"
	_ "embed"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	_ "modernc.org/sqlite"

	"pocka.jp/x/event_sourcing_user_management_poc/routes"
	"pocka.jp/x/event_sourcing_user_management_poc/setups"
)

//go:embed init.sql
var initSQL string

var noVerbose = flag.Bool("noverbose", false, "Suppress debug logs")

var port = flag.Uint("port", 8080, "TCP port a web server listens to")

var host = flag.String("host", "localhost", "Hostname to bind a web server to")

var shouldCreateInitAdminCreationPassword = flag.Bool(
	"init-admin-creation-password", false, "Whether generate a password for initial admin user creation",
)

var shouldCreateAlice = flag.Bool(
	"create-alice", false, "Create an admin user \"alice@example.com/Alice's password\"?",
)

var shouldCreateBob = flag.Bool(
	"create-bob", false, "Create a viewer user \"bob@example.com/Bob's password\"?",
)

// charmbracelet/log uses 256-color for default styles.
// In other words, they ignore common terminal emulator's palette and uses
// semi-hard-coded color. Unsafe defaults.
func createLogger() *log.Logger {
	styles := log.DefaultStyles()
	styles.Levels = map[log.Level]lipgloss.Style{
		log.DebugLevel: lipgloss.NewStyle().
			SetString(strings.ToUpper(log.DebugLevel.String())).
			Bold(true).
			Width(5).
			Foreground(lipgloss.Color("7")),
		log.InfoLevel: lipgloss.NewStyle().
			SetString(strings.ToUpper(log.InfoLevel.String())).
			Bold(true).
			Width(5).
			Foreground(lipgloss.Color("4")),
		log.WarnLevel: lipgloss.NewStyle().
			SetString(strings.ToUpper(log.WarnLevel.String())).
			Bold(true).
			Width(5).
			Foreground(lipgloss.Color("3")),
		log.ErrorLevel: lipgloss.NewStyle().
			SetString(strings.ToUpper(log.ErrorLevel.String())).
			Bold(true).
			Width(5).
			Foreground(lipgloss.Color("1")),
		log.FatalLevel: lipgloss.NewStyle().
			SetString(strings.ToUpper(log.FatalLevel.String())).
			Bold(true).
			Width(5).
			Foreground(lipgloss.Color("1")),
	}

	log.SetStyles(styles)
	logger := log.New(os.Stderr)
	logger.SetStyles(styles)

	return logger
}

func main() {
	logger := createLogger()

	flag.Parse()
	if !*noVerbose {
		logger.SetLevel(log.DebugLevel)
	}

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		logger.Fatalf("Opening in-memory database failed: %s\n", err)
	}

	if _, err := db.Exec(initSQL); err != nil {
		logger.Fatalf("Initialization SQL failed: %s\n", err)
	}

	if *shouldCreateInitAdminCreationPassword {
		logger.Debug("Inserting InitialAdminCreationPasswordCreated event...")

		password, err := setups.InitAdminCreationPassword(db)
		if err != nil {
			logger.Fatal(err)
		}

		logger.Infof("Use this password to create initial user: %s", password)
	}

	if *shouldCreateAlice {
		logger.Debug("Creating admin user Alice...")

		id, err := setups.CreateAlice(db, logger)
		if err != nil {
			logger.Fatal(err)
		}

		logger.Infof("Created admin user Alice. ID=%s", id)
	}

	if *shouldCreateBob {
		logger.Debug("Creating viewer user Bob...")

		id, err := setups.CreateBob(db, logger)
		if err != nil {
			logger.Fatal(err)
		}

		logger.Infof("Created viewer user Bob. ID=%s", id)
	}

	addr := fmt.Sprintf("%s:%d", *host, *port)

	logger.Infof("Starting HTTP server at http://%s", addr)

	handler, err := routes.Handler(db, logger)
	if err != nil {
		logger.Fatal(err)
	}

	http.Handle("/", handler)

	logger.Fatal(http.ListenAndServe(addr, nil))
}
