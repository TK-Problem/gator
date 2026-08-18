package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/TK-Problem/gator/internal/database"
	"github.com/google/uuid"
)

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	name := cmd.Args[0]

	_, err := s.db.GetUser(context.Background(), name)
	if err == nil {
		return fmt.Errorf("user %s already exists", name)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("couldn't look up user %s: %w", name, err)
	}

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
	})
	if err != nil {
		return fmt.Errorf("couldn't create user: %w", err)
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Printf("User created successfully: %s\n", user.Name)
	fmt.Printf("%+v\n", user)
	return nil
}
