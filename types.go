package main

import (
	"math/rand"
	"time"
)

type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type UpdateAccountRequest struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type UpdateBalanceRequest struct {
	Amount int64 `json:"amount"`
}

type TransferRequest struct {
	Number int64 `json:"number"`
	Amount int64 `json:"amount"`
}

type Account struct {
	ID        int       `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Number    int64     `json:"number"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func MapAccount(account *Account, firstName, lastName string, amount int64) (*Account, error) {
	return &Account{
		ID:        account.ID,
		FirstName: firstName,
		LastName:  lastName,
		Number:    account.Number,
		Balance:   amount,
		CreatedAt: account.CreatedAt,
		UpdatedAt: time.Now().UTC(),
	}, nil
}

func NewAccount(firstName, lastName string) (*Account, error) {
	return &Account{
		FirstName: firstName,
		LastName:  lastName,
		Number:    int64(rand.Intn(10000000)),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}, nil
}
