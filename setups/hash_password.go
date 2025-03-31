// Copyright 2025 Shota FUJI
//
// This source code is licensed under Zero-Clause BSD License.
// You can find a copy of the Zero-Clause BSD License at LICENSES/0BSD.txt
// You may also obtain a copy of the Zero-Clause BSD License at
// <https://opensource.org/license/0bsd>
//
// SPDX-License-Identifier: 0BSD

package setups

import (
	"crypto/rand"

	"golang.org/x/crypto/argon2"
)

// hashPassword returns hash of the password and salt used for the hash.
func hashPassword(password string) ([]byte, []byte) {
	salt := make([]byte, 32)

	// rand.Read never returns an error.
	// https://pkg.go.dev/crypto/rand@go1.24.1#Read
	rand.Read(salt)

	// Parameters recommended in RFC (according to Go docs)
	return argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32), salt
}
