package activerecord

import "errors"

type ValidationError struct {
	error
}

type ConflictError struct {
	error
}

var (
	invalidPasswordError   = ValidationError{errors.New("invalid password")}
	invalidEmailError      = ValidationError{errors.New("invalid email")}
	invalidFirstNameError  = ValidationError{errors.New("invalid first name")}
	invalidLastNameError   = ValidationError{errors.New("invalid last name")}
	invalidOwner           = ValidationError{errors.New("invalid owner")}
	invalidCurrency        = ValidationError{errors.New("invalid currency")}
	invalidAddress         = ValidationError{errors.New("invalid address")}
	invalidAmount          = ValidationError{errors.New("invalid amount")}
	emailConflictError     = ConflictError{errors.New("user with such email already exists")}
	walletCurrencyMismatch = ConflictError{errors.New("wallet currency mismatch")}
)
