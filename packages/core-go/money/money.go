package money

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidCurrency = errors.New("money: currency must be a 3-letter ISO-style code")
	ErrNegativeAmount  = errors.New("money: amount must not be negative")
	ErrCurrencyMismatch = errors.New("money: currency mismatch")
)

// Amount stores money in minor units (for example VND = 1, USD = cents).
// It deliberately avoids floating point arithmetic.
type Amount struct {
	Currency string `json:"currency"`
	Minor    int64  `json:"minor"`
}

func New(currency string, minor int64) (Amount, error) {
	a := Amount{Currency: strings.ToUpper(strings.TrimSpace(currency)), Minor: minor}
	if err := a.Validate(); err != nil {
		return Amount{}, err
	}
	return a, nil
}

func Must(currency string, minor int64) Amount {
	a, err := New(currency, minor)
	if err != nil {
		panic(err)
	}
	return a
}

func (a Amount) Validate() error {
	if len(a.Currency) != 3 {
		return ErrInvalidCurrency
	}
	for _, r := range a.Currency {
		if r < 'A' || r > 'Z' {
			return ErrInvalidCurrency
		}
	}
	if a.Minor < 0 {
		return ErrNegativeAmount
	}
	return nil
}

func (a Amount) Add(b Amount) (Amount, error) {
	if err := a.Validate(); err != nil {
		return Amount{}, err
	}
	if err := b.Validate(); err != nil {
		return Amount{}, err
	}
	if a.Currency != b.Currency {
		return Amount{}, ErrCurrencyMismatch
	}
	return Amount{Currency: a.Currency, Minor: a.Minor + b.Minor}, nil
}

func (a Amount) String() string {
	return fmt.Sprintf("%s %d", a.Currency, a.Minor)
}
