package activerecord

import (
	"context"
	"fmt"
	"net"
	"net/mail"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
	"golang.org/x/crypto/bcrypt"
)

func newUserFactory(db pgxtype.Querier) UserFactory {
	return UserFactory{
		db: db,
	}
}

type UserFactory struct {
	db pgxtype.Querier
}

func (uf UserFactory) New(email, password, firstName, lastName string) (*User, error) {
	return newUser(uf.db, email, password, firstName, lastName)
}

func (uf UserFactory) FindByEmail(ctx context.Context, email string) (*User, error) {
	return uf.findOne(ctx, `WHERE email=$1`, email)
}

func (uf UserFactory) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return uf.findOne(ctx, `WHERE id=$1`, id)
}

func (uf UserFactory) findOne(ctx context.Context, where string, whereParams ...interface{}) (*User, error) {
	user := &User{
		db:        uf.db,
		id:        uuid.UUID{},
		email:     "",
		password:  "",
		firstName: "",
		lastName:  "",
		wallets:   nil,
	}

	q := fmt.Sprintf(`SELECT id,email,password,first_name,last_name FROM users %s`, where)
	err := uf.db.QueryRow(ctx, q, whereParams...).
		Scan(&user.id, &user.email, &user.password, &user.firstName, &user.lastName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, notFoundError
		}

		return nil, err
	}

	return user, nil
}

func newUser(db pgxtype.Querier, email, password, firstName, lastName string) (*User, error) {
	if invalidPassword(password) {
		return nil, invalidPasswordError
	}

	if invalidName(firstName) {
		return nil, invalidFirstNameError
	}

	if invalidName(lastName) {
		return nil, invalidLastNameError
	}

	if invalidEmail(email) {
		return nil, invalidEmailError
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		id: uuid.New(),
		db: db,
		email:    email,
		password: string(passHash),
		firstName: firstName,
		lastName: lastName,
	}, nil
}

type User struct {
	db       pgxtype.Querier
	id       uuid.UUID
	email    string
	password string
	firstName string
	lastName string
	wallets []*Wallet
}

func (u *User) Save(ctx context.Context) error {
	return u.create(ctx)
}

func (u *User) create(ctx context.Context) error {
	_, err := u.db.Exec(ctx, `INSERT INTO users(id, email, password, first_name, last_name) VALUES($1,$2,$3,$4,$5)`,
		u.id, u.email, u.password, u.firstName, u.lastName)
	if err != nil {
		if e, ok := err.(*pgconn.PgError); ok && e.Code == uniqueConstraintViolation {
			return emailConflictError
		}

		return err
	}

	for _, w := range u.wallets {
		err := w.Save(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *User) CreateWallets(currencies ...string) ([]*Wallet, error) {
	currs := unique(currencies)

	wallets := make([]*Wallet, len(currs))
	for i, c := range currs {
		w, err := newWalletWithAddress(u.db, u.id, c)
		if err != nil {
			return nil, err
		}

		wallets[i] = w
	}

	u.wallets = wallets
	return wallets, nil
}

func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) Email() string {
	return u.email
}

func (u *User) FirstName() string {
	return u.firstName
}

func (u *User) LastName() string {
	return u.lastName
}

func (u *User) Wallets() []*Wallet {
	return u.wallets
}

func (u *User) Password() string {
	return u.password
}

func (u *User) LoadWallets(ctx context.Context) ([]*Wallet, error) {
	wallets, err := newWalletFactory(u.db).FindByUserID(ctx, u.id)
	if err != nil {
		return nil, err
	}

	u.wallets = wallets
	return wallets, nil
}

func invalidPassword(password string) bool {
	l := len(password)
	return l < 8 || l > 50
}

func invalidEmail(email string) bool {
	if strings.ContainsAny(email, " <>") {
		return true
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return true
	}

	parts := strings.Split(email, "@")
	domain := parts[1]

	_, err = net.LookupIP(domain)
	return err != nil
}

var nameRegexp = regexp.MustCompile(`(?i)^[\p{L}'][ \p{L}'-]*[\p{L}]$`)
func invalidName(name string) bool {
	return !nameRegexp.MatchString(name)
}

func unique(strs []string) []string {
	u := make(map[string]struct{})
	var res []string
	for _, s := range strs {
		if s == "" {
			continue
		}

		if _, ok := u[s]; !ok {
			u[s] = struct{}{}
			res = append(res, s)
		}
	}

	return res
}
