package main

import (
	"context"
	"fmt"

	"github.com/TK-Problem/gator/internal/database"
)

// middlewareLoggedIn adapts a handler that needs the current user into a plain
// handler, resolving the user from the config once on the way through.
func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("couldn't find current user %s: %w", s.cfg.CurrentUserName, err)
		}
		return handler(s, cmd, user)
	}
}
