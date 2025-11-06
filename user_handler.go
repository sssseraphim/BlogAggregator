package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sssseraphim/gator/internal/database"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("the login handler expects a single argument, the username")
	}
	name := cmd.args[0]
	_, err := s.db.GetUser(context.Background(), sql.NullString{String: name, Valid: true})
	if err != nil {
		return fmt.Errorf("couldn't find a user: %v", err)
	}
	err = s.cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("couldn't set a user: %v", err)
	}
	fmt.Print("user have been set")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("the register command expects a single argument, the username")
	}

	name := cmd.args[0]
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      sql.NullString{String: name, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("couldn't create a user: %v", err)
	}
	err = s.cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("couldn't set a user: %v", err)
	}
	fmt.Printf("User created! %v %v %v %v \n", user.ID, user.CreatedAt, user.UpdatedAt, user.Name)
	return nil

}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("coudn't reset the database: %v", err)
	}
	fmt.Printf("Database reset!")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.ListUsers(context.Background())
	if err != nil {
		return fmt.Errorf("coudn't list users: %v", err)
	}
	for _, user := range users {
		if user.Name.String == s.cfg.Current_user_name {
			fmt.Printf("* %v (current)\n", user.Name.String)
		} else {
			fmt.Printf("* %v\n", user.Name.String)
		}
	}
	return nil

}
