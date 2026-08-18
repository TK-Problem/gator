package main

import "github.com/TK-Problem/gator/internal/config"

// state carries the application-wide dependencies that command handlers need.
type state struct {
	cfg *config.Config
}
