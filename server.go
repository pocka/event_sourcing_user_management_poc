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
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	_ "modernc.org/sqlite"

	"pocka.jp/x/event_sourcing_user_management_poc/setups"
)

//go:embed init.sql
var initSQL string

var noVerbose = flag.Bool("noverbose", false, "Suppress debug logs")

var shouldCreateInitAdminCreationPassword = flag.Bool(
	"init-admin-creation-password", false, "Whether generate a password for initial admin user creation",
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
}
