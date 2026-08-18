package main

import (
	"github.com/TK-Problem/gator/internal/config"
	"github.com/TK-Problem/gator/internal/database"
)

// state carries the application-wide dependencies that command handlers need.
type state struct {
	db  *database.Queries
	cfg *config.Config
}
