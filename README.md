
# Go-Redis

A lightweight, concurrent in-memory key-value store built in Go, designed with a custom TCP server, interface-driven storage engines, TTL background expiration sweeps, and an interactive command-line experience.

---

## Overview

Go-Redis provides a thread-safe, Redis-inspired key-value store. It includes an interactive TCP socket interface that works directly with `telnet`, `nc`, or standard socket clients, alongside support for background janitor sweeps and one-off execution modes.

---

## Features

- **Concurrent & Thread-Safe Engine**: Backed by `sync.RWMutex` to ensure safe, concurrent reads and synchronized mutations across active client connections.
- **TTL & Hybrid Eviction Strategy**:
  - **Active Janitor Sweep**: An automated background routine (`sync.WaitGroup` + `context.CancelFunc`) continuously cleans stale keys at a configurable duration.
- **Pluggable Storage Modes via `IStore`**:
  - `TTLStore`: Background janitor-driven storage for ephemeral cache items.
  - `Store`: Persistent in-memory storage (with `--no-ttl`).
- **Interactive Terminal REPL**: Automatic terminal greeting, per-connection prompt (`127.0.0.1:6379> `), and graceful session exits.

---

## Project Structure

```text
go-redis/
├── stores/
│   ├── store_types.go     # IStore interface and TTL payload definitions
│   ├── store.go           # Thread-safe persistent in-memory store
│   ├── ttl_store.go       # TTL-aware store with background janitor cleaner
│   └── store_test.go      # Storage layer unit tests
├── cmd.go                 # Core command handlers (GET, SET, DELETE, etc.)
├── main.go                # TCP listener, REPL router, and CLI flag parser
└── go.mod                 # Go module definition

```

---

## Getting Started

### Prerequisites

Ensure you have **Go 1.20+** installed on your system.

### Installation

Clone the repository and enter the directory:

```bash
git clone https://github.com/NooRMaseR/go-redis.git
cd go-redis
```

---

## Usage

### 1. Starting the Server

Run the server with the default configuration (binds to port `6379` with a `1s` janitor interval):

```bash
go run .

```

Run in persistent mode (TTL janitor disabled):

```bash
go run . --no-ttl

```

Customize the janitor interval and port:

```bash
go run . --duration 500ms --port 6380

```

#### Available Flags

| Flag | Type | Default | Description |
| --- | --- | --- | --- |
| `-port` | string | `6379` | TCP port to listen on or connect to |
| `-no-ttl` | bool | `false` | Disable TTL background sweeps and run in persistent mode |
| `-duration` | duration | `1s` | Janitor interval for expired key collection (e.g., `250ms`, `2s`) |

---

### 2. Interactive Terminal Connection

Connect via `telnet` or `nc` from any terminal:

```bash
telnet 127.0.0.1 6379

```

Once connected, execute commands directly inside the interactive session:

```text
Connected to Go-Redis Server [127.0.0.1:6379]
Type 'HELP' or run commands directly.
127.0.0.1:6379> PING
PONG
127.0.0.1:6379> SET user "admin"
OK
127.0.0.1:6379> SET session "active" 5000
OK
127.0.0.1:6379> GET user
admin
127.0.0.1:6379> EXIST user
true
127.0.0.1:6379> LEN
2
127.0.0.1:6379> EXIT
Bye!

```

---

## Supported Commands

| Command | Syntax | Description |
| --- | --- | --- |
| **`SET`** | `SET <key> <value> [ttl_ms]` | Stores a key-value pair, with optional expiration in milliseconds. |
| **`GET`** | `GET <key...>` | Returns value for a single key or a formatted list for multiple keys. |
| **`DELETE`** | `DELETE <key...>` | Removes one or more keys from the store. |
| **`EXIST`** | `EXIST <key...>` | Checks existence of keys (`true` / `false`). |
| **`POP`** | `POP <key...>` | Atomically fetches and deletes key(s). |
| **`RENAME`** | `RENAME <old> <new>` | Renames an existing key, overwriting destination if present. |
| **`RENAMENX`** | `RENAMENX <old> <new>` | Renames an existing key only if destination key does not exist. |
| **`KEYS`** | `KEYS` | Returns all active keys in the database. |
| **`LEN`** | `LEN` | Returns the total count of active keys. |
| **`CLEAR`** | `CLEAR` | Flushes all entries from memory. |
| **`HELP`** | `HELP` | Prints available command signatures and descriptions. |
| **`EXIT`** | `EXIT` | Closes client connection gracefully. |

---

## Architecture & Concurrency Model

* **Connection Isolation**: Every incoming client connection accepted via `net.Listen` is spawned into an independent goroutine (`go handleConnection(conn, store)`).
* **Sequential In-Connection Processing**: Individual multi-key operations process keys sequentially within the connection thread, eliminating redundant goroutine allocation overhead and avoiding lock thrashing across memory reads.
* **Interface-Driven Decoupling**: Storage implementations (`Store` and `TTLStore`) satisfy the unified `stores.IStore` contract, keeping TCP and command-line parsing code decoupled from storage logic.
