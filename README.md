# gator

A command-line RSS feed aggregator. Register users, subscribe them to feeds, and
let a background worker collect posts into PostgreSQL so you can read them later
from the terminal.

Built as part of the [Boot.dev](https://boot.dev) Go course.

## Prerequisites

- **Go 1.26 or newer** — the CLI is installed with `go install`.
- **PostgreSQL 15 or newer** — running locally, or anywhere you can reach with a
  connection string.
- **[goose](https://github.com/pressly/goose)** — needed once, to create the
  database tables:
  ```bash
  go install github.com/pressly/goose/v3/cmd/goose@latest
  ```

## Installation

```bash
go install github.com/TK-Problem/gator@latest
```

The binary lands in `$(go env GOPATH)/bin`, which needs to be on your `PATH`.
Check it worked:

```bash
gator
# Usage: cli <command> [args...]
```

## Database setup

Create the database:

```bash
createdb gator
# or: psql -c 'CREATE DATABASE gator;'
```

Then apply the migrations. Clone the repo for this — the SQL files aren't part of
the installed binary:

```bash
git clone https://github.com/TK-Problem/gator.git
cd gator/sql/schema
goose postgres "postgres://user:password@localhost:5432/gator" up
```

You should end up with five tables plus goose's own bookkeeping: `users`,
`feeds`, `feed_follows`, `posts`, and `goose_db_version`.

> **Watch the connection string.** goose takes it *without* a query string, but
> the config file below needs `?sslmode=disable` appended. The Go driver
> negotiates SSL by default and a local Postgres usually isn't set up for it, so
> omitting it there gives you `pq: SSL is not enabled on the server`.

## Configuration

gator reads a JSON config from your home directory. Create
`~/.gatorconfig.json` by hand with a single key:

```json
{
  "db_url": "postgres://user:password@localhost:5432/gator?sslmode=disable"
}
```

Don't add `current_user_name` — the program writes that itself whenever you
`register` or `login`.

Because this file holds your database password, gator writes it back with `0600`
permissions. If you created it yourself, `chmod 600 ~/.gatorconfig.json` is worth
running, and keep it out of any repo.

## Usage

```
gator <command> [args...]
```

| Command | Arguments | Description |
|---|---|---|
| `register` | `<name>` | Create a user and log in as them. |
| `login` | `<name>` | Switch to an existing user. Fails if they don't exist. |
| `users` | — | List all users; `(current)` marks the active one. |
| `addfeed` | `<name> <url>` | Add a feed and follow it. Requires a logged-in user. |
| `feeds` | — | List every feed with the user who added it. |
| `follow` | `<url>` | Follow an existing feed. |
| `unfollow` | `<url>` | Stop following a feed. |
| `following` | — | List the feeds the current user follows. |
| `agg` | `<time_between_reqs>` | Run the collector in a loop. Duration string: `30s`, `1m`, `1h`. |
| `browse` | `[limit]` | Show recent posts from followed feeds. Defaults to 2. |
| `reset` | — | Delete all users. Cascades to feeds, follows, and posts. |

Commands that act on behalf of someone (`addfeed`, `follow`, `unfollow`,
`following`, `browse`) fail if the config points at a user who doesn't exist.

## Example session

```bash
# Create a user
gator register alice

# Add a couple of feeds (adding also follows)
gator addfeed "Hacker News" https://news.ycombinator.com/rss
gator addfeed "TechCrunch" https://techcrunch.com/feed/

gator following
# * Hacker News
# * TechCrunch
```

Start the collector and leave it running in its own terminal:

```bash
gator agg 1m
# Collecting feeds every 1m0s
# Feed Hacker News collected, 30 posts found
# Feed TechCrunch collected, 20 posts found
```

Then read what it gathered from another terminal:

```bash
gator browse 5
# * Some interesting post title
#   Hacker News | Thu Aug 20 20:34
#   https://example.com/the-post
```

## About the `agg` command

`agg` runs until you stop it — `Ctrl+C` exits cleanly.

It fetches **one feed per tick**, so the request rate is exactly
`1 / time_between_reqs`. Please be considerate of the servers you're fetching
from: `1m` is a reasonable interval, and something like `100ms` would hammer
them. Each fetch prints a line, so if you see output scrolling faster than you
expected, stop it.

A feed that fails (server down, timeout, malformed XML) is logged and skipped;
the loop keeps going rather than dying on one bad feed.

## How it works

- **Migrations** are plain SQL in `sql/schema`, applied with goose. Each has an
  `-- +goose Up` and `-- +goose Down` half, so they roll back cleanly.
- **Queries** live in `sql/queries` and are compiled into type-safe Go by
  [sqlc](https://sqlc.dev) (`sqlc generate`) into `internal/database`. That
  package is generated — edit the SQL, not the Go.
- **Feed rotation** uses a nullable `last_fetched_at` column on `feeds`, selected
  with `ORDER BY last_fetched_at ASC NULLS FIRST LIMIT 1`. Feeds never fetched
  sort first, then the stalest one, so repeated ticks cycle through everything
  without tracking a cursor.
- **Duplicate posts** are handled in the database with `ON CONFLICT (url) DO
  NOTHING`, so re-scraping a feed is free and silent.
- **Publish dates** are parsed against several RSS date layouts; a date that
  can't be parsed is stored as `NULL` rather than dropping the post.

## Development

```bash
go build ./...
go vet ./...

# After editing anything in sql/
sqlc generate
```
