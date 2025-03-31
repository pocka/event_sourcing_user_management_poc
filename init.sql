-- Copyright 2025 Shota FUJI
--
-- This source code is licensed under Zero-Clause BSD License.
-- You can find a copy of the Zero-Clause BSD License at LICENSES/0BSD.txt
-- You may also obtain a copy of the Zero-Clause BSD License at
-- <https://opensource.org/license/0bsd>
--
-- SPDX-License-Identifier: 0BSD

CREATE TABLE user_events (
	seq INTEGER PRIMARY KEY ON CONFLICT ROLLBACK AUTOINCREMENT,
	event_name TEXT NOT NULL ON CONFLICT ROLLBACK,
	payload BLOB
);
