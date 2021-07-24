package service

import (
	"github.com/google/uuid"
	"github.com/merisho/binaryx-test/activerecord"
)

type currency = string

const (
	FakeBTC currency = "fBTC"
	FakeETH currency = "fETH"
)

// NewWallets creates ephemeral wallets for service purposes
func NewWallets(activeRecords activerecord.Facade) (*Wallets, error) {
	fbtc, err := activeRecords.Wallet().New(uuid.UUID{}, FakeBTC, "0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		return nil, err
	}

	feth, err := activeRecords.Wallet().New(uuid.UUID{}, FakeETH, "1111111111111111111111111111111111111111111111111111111111111111")
	if err != nil {
		return nil, err
	}

	return &Wallets{
		wallets: map[currency]*activerecord.Wallet{
			FakeBTC: fbtc,
			FakeETH: feth,
		},
	}, nil
}

type Wallets struct {
	wallets map[currency]*activerecord.Wallet
}

func (w *Wallets) Get(c currency) *activerecord.Wallet {
	return w.wallets[c]
}
