package activerecord

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	uniqueConstraintViolation = "23505"
)

type DBTransaction interface {
	Commit(context.Context) error
	Rollback(context.Context) error
}

type Facade interface {
	Tx(ctx context.Context) (DBTransaction, error)
	User() UserFactory
	Wallet() WalletFactory
}

func New(db *pgxpool.Pool) Facade {
	return facade{db}
}

type facade struct {
	db *pgxpool.Pool
}

func (f facade) User() UserFactory {
	return UserFactory{
		db: f.db,
	}
}

func (f facade) Wallet() WalletFactory {
	return WalletFactory{
		db: f.db,
	}
}

func (f facade) Tx(ctx context.Context) (DBTransaction, error) {
	tx, err := f.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}
