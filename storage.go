package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) (*Account, error)
	UpdateBalance(*Account, int64) (*Account, error)
	GetAccountByNumber(int64) (*Account, error)
	GetAccountByID(int) (*Account, error)
	GetAccounts() ([]*Account, error)
}

type PGStore struct {
	db *sql.DB
}

func NewPGStore() (*PGStore, error) {
	connStr := "user=postgres dbname=postgres password=root sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PGStore{
		db: db,
	}, nil
}

func (s *PGStore) Init() error {
	return s.createAccountTable()
}

func (s *PGStore) createAccountTable() error {
	query := `CREATE TABLE IF NOT EXISTS accounts (
		id serial primary key,
		first_name varchar(50),
		last_name varchar(50),
		number serial,
		balance serial,
		created_at timestamp,
		updated_at timestamp
	)`

	_, err := s.db.Exec(query)

	return err
}

func (s *PGStore) CreateAccount(acc *Account) error {
	stmt := `
	INSERT INTO accounts
		(first_name, last_name, number, balance, created_at, updated_at)
	VALUES
		($1, $2, $3, $4, $5, $6)
	`

	_, err := s.db.Exec(
		stmt,
		acc.FirstName,
		acc.LastName,
		acc.Number,
		acc.Balance,
		acc.CreatedAt,
		acc.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *PGStore) DeleteAccount(id int) error {
	_, err := s.db.Query("DELETE FROM accounts WHERE id = $1", id)
	return err
}

func (s *PGStore) UpdateBalance(account *Account, amount int64) (*Account, error) {
	stmt := `
	UPDATE accounts
	SET balance = $1
	WHERE id = $2
	`

	res, err := s.db.Exec(stmt, amount, account.ID)
	if err != nil {
		return nil, err
	}

	if _, err := res.RowsAffected(); err != nil {
		return nil, err
	}

	return account, nil
}

func (s *PGStore) UpdateAccount(account *Account) (*Account, error) {
	stmt := `
	UPDATE accounts
	SET first_name = $1, last_name = $2
	WHERE id = $3
	`

	res, err := s.db.Exec(
		stmt,
		account.FirstName,
		account.LastName,
		account.ID,
	)

	if err != nil {
		return nil, err
	}

	if _, err := res.RowsAffected(); err != nil {
		return nil, err
	}

	return account, nil
}

func (s *PGStore) GetAccountByNumber(number int64) (*Account, error) {
	rows, err := s.db.Query("SELECT * FROM accounts WHERE number = $1", number)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("account with account number %d not found", number)
}

func (s *PGStore) GetAccountByID(id int) (*Account, error) {
	rows, err := s.db.Query("SELECT * FROM accounts WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("account with id %d not found", id)
}

func (s *PGStore) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query("SELECT * FROM accounts")

	if err != nil {
		return nil, err
	}

	accounts := []*Account{}

	for rows.Next() {
		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)

	err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.Balance,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	return account, err
}
