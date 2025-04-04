<!--
Copyright 2025 Shota FUJI

This source code is licensed under Zero-Clause BSD License.
You can find a copy of the Zero-Clause BSD License at LICENSES/0BSD.txt
You may also obtain a copy of the Zero-Clause BSD License at
<https://opensource.org/license/0bsd>

SPDX-License-Identifier: 0BSD
-->

# Event Sourcing User Management PoC

This is a demo web server written in Go, using simple [Event Sourcing](https://martinfowler.com/eaaDev/EventSourcing.html).

The goals of this demo are,

- get hands-on experience on Event Sourcing
- figure out whether this architecture fits my ongoing personal project
- learning something new by using tech stacks I don't use normally (HTTP server in Go)

This application is not intended to be used as reference or be used in real-world.
Because of the goals previously stated, security and performance are not in consideration.
Also, I avoided using third-party libraries as much as possible, in order to easily transfer this experience onto Zig source code, which I use for my ongoing personal project.

You will be in trouble if you use part or whole of this application in a real product.

## Architecture

The simplest description is "Event Sourcing HTTP server using SQLite3 as an event store".
Everytime HTTP handler needs the current state of something (user list, one-time password for initial admin creation), the application loads latest projection snapshot and events newer than the snapshot, then builds state from those.

Events are stored in SQLite3 table named `user_events` with dead simple schema:

| Column       | Data type |
| ------------ | --------- |
| `seq`        | `INTEGER` |
| `event_name` | `TEXT`    |
| `payload`    | `BLOB`    |

`seq` is auto incrementing sequential number.
`payload` is binary data in Protobuf wire format.
`event_name` is schema name of the `payload`, telling which Protobuf message schema to use for decoding.
As this is demo application, `event_name` does not contain package name.
To reduce the number of events loaded, the latest projection is saved as a snapshot in `users_snapshot` table, which has the following schema:

| Column      | Data type |
| ----------- | --------- |
| `event_seq` | `INTEGER` |
| `payload`   | `BLOB`    |

`event_seq` is a `seq` column of the event this snapshot was created at.
`payload` is binary data in Protobuf wire format, same as `user_events`.
Schema for snapshots is under `proto/projection` directory.

Projections, unlike events, has dedicated table for each projection types.
Therefore `*_snapshot` table does not have a column to store Protobuf schema name.

Table schema is defined in `init.sql` file.
Protobuf message schemas for events payload are under `proto/event/` directory and ones for projections/snapshots are under `proto/projection/` directory.
Events listing and insertion functions are in `events/` directory.
Functions to build projection from events and a snapshot are inside `projections/` directory and they have unit tests.

Creation of demo users and one-time password is defined in `setups/` directory.
This directory is good candidate of unit testing but I'm lazy so there's none.

The rest is joke. I mean, I don't care. It's low-quality simple-as-tutorial `net/http` server.

## Development

This project requires Go toolchain, Protobuf compiler and dprint (source code formatter frontend).
Use [asdf](https://asdf-vm.com/) or [mise-en-place](https://mise.jdx.dev/) to easily install the required tools.
See `.tool-versions` file for specific versions.

Once you setup every required tool, run the following to generate Protobuf binding code.

```sh
buf generate
# Go bindings will be generated under "gen/" directory.
```

### Start HTTP server

```sh
# This flag generates and prints one-time password to create the first admin user.
# Copy that password and paste it to " Initial user password" form field in browser.
go run . -init-admin-creation-password

# Create demo users.
# Alice is an admin user and Bob is viewer user.
# When -create-* flag is set, the server prints login email and password to terminal.
go run . -create-alice -create-bob

# For available options, run with -help flag.
```

### Run unit tests

```sh
# ./... means "this directory and its fucking subdirectories"
go test ./...
```

### Format source code

```sh
# This rewrites files in-place.
dprint fmt
```

### Check copyright and license annotation

This command requires [REUSE tool](https://github.com/fsfe/reuse-tool).

```sh
reuse lint
```
