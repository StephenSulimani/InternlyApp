package routes

import (
	"context"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stephensulimani/internlyapp/internal/auth"
	"github.com/stephensulimani/internlyapp/internal/db"
	"github.com/stephensulimani/internlyapp/internal/service"
	"github.com/stephensulimani/internlyapp/internal/testutil"
)

func TestRegisterRoute_DB(t *testing.T) {
	pool, cleanup := testutil.SetupPostgres(t)
	defer cleanup()

	t.Run("creates first user with bootstrap privileges", func(t *testing.T) {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx, queries *db.Queries) {
			testutil.TruncateUsersTx(t, ctx, tx)
			handler := testRegisterHandlerWithService(service.NewUserService(queries, nil))

			rec := postRegister(t, handler, registerBody{
				FirstName: "Ada",
				LastName:  "Lovelace",
				Email:     "ada@example.com",
				Password:  "secure-password",
			})
			assertStatus(t, rec, http.StatusCreated)

			user, err := queries.GetUserByEmail(ctx, "ada@example.com")
			if err != nil {
				t.Fatal(err)
			}
			if !user.IsAdmin || !user.IsActive || !user.IsPremium {
				t.Fatal("expected bootstrap privileges")
			}
			if !auth.CheckPassword("secure-password", user.Password) {
				t.Fatal("expected password to be stored as a bcrypt hash")
			}
		})
	})

	t.Run("creates subsequent user without bootstrap privileges", func(t *testing.T) {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx, queries *db.Queries) {
			testutil.TruncateUsersTx(t, ctx, tx)
			users := service.NewUserService(queries, nil)
			handler := testRegisterHandlerWithService(users)

			if err := users.Register(ctx, service.RegisterInput{
				FirstName: "Bootstrap",
				LastName:  "User",
				Email:     "bootstrap@example.com",
				Password:  "secure-password",
			}); err != nil {
				t.Fatal(err)
			}

			rec := postRegister(t, handler, registerBody{
				FirstName: "Grace",
				LastName:  "Hopper",
				Email:     "grace@example.com",
				Password:  "another-password",
			})
			assertStatus(t, rec, http.StatusCreated)

			user, err := queries.GetUserByEmail(ctx, "grace@example.com")
			if err != nil {
				t.Fatal(err)
			}
			if user.IsAdmin || user.IsActive || user.IsPremium {
				t.Fatal("expected non-first user to have default privilege flags")
			}
		})
	})

	t.Run("rejects duplicate email", func(t *testing.T) {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx, queries *db.Queries) {
			testutil.TruncateUsersTx(t, ctx, tx)
			handler := testRegisterHandlerWithService(service.NewUserService(queries, nil))
			body := registerBody{
				FirstName: "Ada",
				LastName:  "Lovelace",
				Email:     "duplicate@example.com",
				Password:  "secure-password",
			}

			rec := postRegister(t, handler, body)
			assertStatus(t, rec, http.StatusCreated)

			rec = postRegister(t, handler, body)
			assertAPIError(t, rec, http.StatusConflict, "User already exists")
		})
	})
}
