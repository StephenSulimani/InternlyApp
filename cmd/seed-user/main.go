package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/stephensulimani/internlyapp/internal/db"
	"github.com/stephensulimani/internlyapp/internal/service"
)

func main() {
	_ = godotenv.Load()

	firstName := flag.String("first-name", "", "First name (required)")
	lastName := flag.String("last-name", "", "Last name (required)")
	email := flag.String("email", "", "Email (required)")
	password := flag.String("password", "", "Password, min 8 characters (required)")
	flag.Parse()

	if *firstName == "" || *lastName == "" || *email == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "Usage: go run ./cmd/seed-user -first-name <name> -last-name <name> -email <email> -password <password>")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Creates an active admin user directly in the database.")
		fmt.Fprintln(os.Stderr, "Loads POSTGRES_* from .env (same as the API).")
		os.Exit(2)
	}

	ctx := context.Background()
	pool, err := db.OpenPool(ctx, db.ConfigFromEnv())
	if err != nil {
		fmt.Fprintf(os.Stderr, "database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	users := service.NewUserService(db.New(pool), nil)
	user, err := users.SeedAdmin(ctx, service.RegisterInput{
		FirstName: *firstName,
		LastName:  *lastName,
		Email:     *email,
		Password:  *password,
	})
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			fmt.Fprintf(os.Stderr, "user already exists: %s\n", *email)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "seed user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created admin user %s <%s> (id=%s)\n", user.FirstName+" "+user.LastName, user.Email, user.ID)
}
