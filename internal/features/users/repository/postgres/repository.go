package users_postgres_repository

import (
	postgres_pool "github.com/horizoonn/todoapp/internal/core/repository/postgres/pool"
)

type UsersRepository struct {
	pool postgres_pool.Pool
}

func NewUsersRepository(pool postgres_pool.Pool) *UsersRepository {
	return &UsersRepository{
		pool: pool,
	}
}
