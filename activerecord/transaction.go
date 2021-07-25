package activerecord

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype/pgxtype"
	"github.com/shopspring/decimal"
)

var feeRate = decimal.NewFromFloat32(0.2)

func newTransactionFactory(db pgxtype.Querier) TransactionFactory {
	return TransactionFactory{db: db}
}

type TransactionFactory struct {
	db pgxtype.Querier
}

func (ts TransactionFactory) FindAllWithWallet(ctx context.Context, wallet string) ([]*Transaction, error) {
	rows, err := ts.db.Query(ctx, `SELECT id,currency,to_wallet,from_wallet,amount,fee,timestamp FROM transactions
							WHERE to_wallet=$1 OR from_wallet=$1`, wallet)
	if err != nil {
		return nil, err
	}

	var txs []*Transaction
	for rows.Next() {
		t := &Transaction{
			db: ts.db,
		}
		err := rows.Scan(&t.id, &t.currency, &t.to, &t.from, &t.amount, &t.fee, &t.timestamp)
		if err != nil {
			return nil, err
		}

		txs = append(txs, t)
	}

	return txs, nil
}

func newTransaction(db pgxtype.Querier, currency, from, to string, amount decimal.Decimal) (*Transaction, error) {
	if currency == "" {
		return nil, invalidCurrency
	}

	if from == "" || to == "" {
		return nil, invalidAddress
	}

	if amount.Equal(decimal.Zero) || amount.IsNegative() {
		return nil, invalidAmount
	}

	t := &Transaction{
		id: uuid.New(),
		db: db,
		currency: currency,
		from: from,
		to: to,
		amount: amount,
		timestamp: time.Now().UTC(),
	}

	t.fee = t.CalculateFee()

	return t, nil
}

type Transaction struct {
	db pgxtype.Querier
	id uuid.UUID
	currency string
	from string
	to string
	amount decimal.Decimal
	fee decimal.Decimal
	timestamp time.Time
}

func (t *Transaction) CalculateFee() decimal.Decimal {
	return t.amount.Mul(feeRate)
}

func (t *Transaction) Save(ctx context.Context) error {
	_, err := t.db.Exec(ctx, `INSERT INTO transactions(id, currency, from_wallet, to_wallet, amount, fee, timestamp)
									VALUES($1, $2, $3, $4, $5, $6, $7)`,
						t.id, t.currency, t.from, t.to, t.amount.String(), t.fee.String(), t.timestamp)
	return err
}

func (t *Transaction) Amount() decimal.Decimal {
	return t.amount
}

func (t *Transaction) FullAmount() decimal.Decimal {
	return t.amount.Add(t.CalculateFee())
}
