package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sqlc "github.com/Jibin-Joseph7/user-api/db/sqlc"
)

var ErrNotFound = errors.New("user not found")

type UserRepository struct {
	q *sqlc.Queries
}

func New(q *sqlc.Queries) *UserRepository {
	return &UserRepository{q: q}
}

func (r *UserRepository) Create(ctx context.Context, name string, dob time.Time) (sqlc.User, error) {
	res, err := r.q.CreateUser(ctx, sqlc.CreateUserParams{Name: name, Dob: dob})
	if err != nil {
		return sqlc.User{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return sqlc.User{}, err
	}
	return r.Get(ctx, int32(id))
}

func (r *UserRepository) Get(ctx context.Context, id int32) (sqlc.User, error) {
	u, err := r.q.GetUser(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return sqlc.User{}, ErrNotFound
	}
	return u, err
}

func (r *UserRepository) List(ctx context.Context, limit, offset int32) ([]sqlc.User, int64, error) {
	users, err := r.q.ListUsers(ctx, sqlc.ListUsersParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, 0, err
	}
	total, err := r.q.CountUsers(ctx)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *UserRepository) Update(ctx context.Context, id int32, name string, dob time.Time) (sqlc.User, error) {
	if err := r.q.UpdateUser(ctx, sqlc.UpdateUserParams{ID: id, Name: name, Dob: dob}); err != nil {
		return sqlc.User{}, err
	}
	return r.Get(ctx, id) // returns ErrNotFound if the id doesn't exist
}

func (r *UserRepository) Delete(ctx context.Context, id int32) error {
	res, err := r.q.DeleteUser(ctx, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}