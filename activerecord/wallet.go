package activerecord

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype/pgxtype"
	"github.com/shopspring/decimal"
)

var addressesRandomGenerator = rand.New(rand.NewSource(time.Now().Unix()))

func newWalletFactory(db pgxtype.Querier) WalletFactory {
	return WalletFactory{
		db: db,
	}
}

type WalletFactory struct {
	db pgxtype.Querier
}

func (wf WalletFactory) FindByUserID(ctx context.Context, id uuid.UUID) ([]*Wallet, error) {
	rows, err := wf.db.Query(ctx, `SELECT wallet,currency FROM user_wallets WHERE user_id=$1`, id)
	if err != nil {
		return nil, err
	}

	var wallets []*Wallet
	for rows.Next() {
		w := &Wallet{
			db:     wf.db,
			userID: id,
		}
		err := rows.Scan(&w.address, &w.currency)
		if err != nil {
			return nil, err
		}

		wallets = append(wallets, w)
	}

	return wallets, nil
}

func (wf WalletFactory) New(owner uuid.UUID, currency, address string) (*Wallet, error) {
	return newWallet(wf.db, owner, currency, address)
}

func newWalletWithAddress(db pgxtype.Querier, owner uuid.UUID, currency string) (*Wallet, error) {
	randomNum := strconv.FormatInt(addressesRandomGenerator.Int63(), 10)
	hash := sha256.Sum256([]byte(randomNum))
	addr := hex.EncodeToString(hash[:])
	return newWallet(db, owner, currency, addr)
}

func newWallet(db pgxtype.Querier, owner uuid.UUID, currency, address string) (*Wallet, error) {
	if len(currency) == 0 {
		return nil, invalidCurrency
	}

	if len(address) == 0 {
		return nil, invalidAddress
	}

	return &Wallet{
		db: db,
		userID: owner,
		currency: currency,
		address: address,
	}, nil
}

type Wallet struct {
	db     pgxtype.Querier
	userID uuid.UUID
	currency string
	address string
	transactions []*Transaction
}

func (w *Wallet) Save(ctx context.Context) error {
	_, err := w.db.Exec(ctx, `INSERT INTO user_wallets(user_id,wallet,currency) VALUES($1,$2,$3) ON CONFLICT DO NOTHING`,
									w.userID, w.address, w.currency)
	if err != nil {
		return err
	}

	for _, tx := range w.transactions {
		err := tx.Save(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *Wallet) AcceptTransaction(from *Wallet, amount decimal.Decimal) (*Transaction, error) {
	if w.currency != from.currency {
		return nil, walletCurrencyMismatch
	}

	tx, err := newTransaction(w.db, w.currency, from.address, w.address, amount)
	if err != nil {
		return nil, err
	}

	w.transactions = append(w.transactions, tx)
	return tx, nil
}

func (w *Wallet) LoadTransactions(ctx context.Context) ([]*Transaction, error) {
	txs, err := newTransactionFactory(w.db).FindAllWithWallet(ctx, w.address)
	if err != nil {
		return nil, err
	}

	w.transactions = txs
	return txs, nil
}

func (w *Wallet) Balance() decimal.Decimal {
	s := decimal.Zero
	for _, t := range w.transactions {
		if t.to == w.address {
			s = s.Add(t.Amount())
		} else if t.from == w.address {
			s = s.Sub(t.FullAmount())
		}
	}

	return s
}

func (w *Wallet) UserID() uuid.UUID {
	return w.userID
}

func (w *Wallet) Currency() string {
	return w.currency
}

func (w *Wallet) Address() string {
	return w.address
}
