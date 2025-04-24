package userservice

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moonmoon1919/go-api-reference/pkg/users"
)

const (
	notFoundMsg = "NOT FOUND"
	errKey      = "ERR"
)

var notFoundError = errors.New("user not found")

// MARK: Interface
type Storer interface {
	Add(ctx context.Context, item users.User) (users.User, error)
	Delete(ctx context.Context, id string) error
}

// MARK: Memory
type userMemoryRepository struct {
	items map[string]users.User
}

func NewInMemoryUserRepository() *userMemoryRepository {
	return &userMemoryRepository{
		items: make(map[string]users.User),
	}
}

func (u *userMemoryRepository) Add(ctx context.Context, item users.User) (users.User, error) {
	u.items[item.Id] = item

	return item, nil
}

func (u *userMemoryRepository) Delete(ctx context.Context, i string) error {
	delete(u.items, i)

	return nil
}

// MARK: SQL
type userSQLRepository struct {
	pool *pgxpool.Pool
}

func NewSQLRepository(pool *pgxpool.Pool) *userSQLRepository {
	return &userSQLRepository{
		pool: pool,
	}
}

func (u *userSQLRepository) Add(ctx context.Context, item users.User) (users.User, error) {
	_, err := u.pool.Exec(ctx, "INSERT INTO users (id) VALUES ($1) ON CONFLICT (id) DO NOTHING", item.Id)

	if err != nil {
		return users.Nil(), err
	}

	return item, nil
}

/*
Delete a user record - PERFORMS A CASCADING DELETE
*/
func (u *userSQLRepository) Delete(ctx context.Context, i string) error {
	var id string
	err := u.pool.QueryRow(ctx, "DELETE FROM users where id=$1 RETURNING id", i).Scan(&id)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return notFoundError
		default:
			return err
		}
	}

	return nil
}
