package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type apiFunc func(http.ResponseWriter, *http.Request) error

type APIServer struct {
	listenAddr string
	store      Storage
}

type MetaResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, &MetaResponse{
				Status:  http.StatusBadRequest,
				Message: err.Error(),
				Data:    nil,
			})
		}
	}
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/accounts", makeHTTPHandleFunc(s.handleAccount))
	router.HandleFunc("/accounts/{id}", makeHTTPHandleFunc(s.handleGetAccountByID))
	router.HandleFunc("/topup/{number}", makeHTTPHandleFunc(s.handleTopUp))
	router.HandleFunc("/transfer/{number}", makeHTTPHandleFunc(s.handleTransfer))

	log.Println("Server started on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccount(w, r)
	}

	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	return fmt.Errorf("method %s are not allowed", r.Method)
}

func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		id, err := getID(r)
		if err != nil {
			return err
		}

		account, err := s.store.GetAccountByID(id)
		if err != nil {
			return err
		}

		return WriteJSON(w, http.StatusOK, &MetaResponse{
			Status:  http.StatusOK,
			Message: "Account",
			Data:    account,
		})
	}

	if r.Method == "PATCH" {
		return s.handleUpdateAccount(w, r)
	}

	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}

	return fmt.Errorf("method %s are not allowed", r.Method)
}

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, &MetaResponse{
		Status:  http.StatusOK,
		Message: "Accounts",
		Data:    accounts,
	})
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	req := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}

	account, err := NewAccount(req.FirstName, req.LastName)
	if err != nil {
		return err
	}
	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusCreated, &MetaResponse{
		Status:  http.StatusCreated,
		Message: "Account Created",
		Data:    account,
	})
}

func (s *APIServer) handleUpdateAccount(w http.ResponseWriter, r *http.Request) error {
	req := new(UpdateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}

	id, err := getID(r)
	if err != nil {
		return err
	}

	old, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}

	mapped, err := MapAccount(old, req.FirstName, req.LastName, old.Balance)
	if err != nil {
		return err
	}

	account, err := s.store.UpdateAccount(mapped)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, &MetaResponse{
		Status:  http.StatusOK,
		Message: "Account Updated",
		Data:    account,
	})
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}

	if err := s.store.DeleteAccount(id); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, &MetaResponse{
		Status:  http.StatusOK,
		Message: "Account Deleted",
		Data:    id,
	})
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method %s are not allowed", r.Method)
	}

	number, err := getNumber(r)
	if err != nil {
		return err
	}

	req := new(TransferRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}

	balanceSender, err := s.store.GetAccountByNumber(number)
	if err != nil {
		return err
	}

	recipient, err := s.store.GetAccountByNumber(req.Number)
	if err != nil {
		return err
	}

	deduct := balanceSender.Balance - req.Amount

	if deduct < 0 {
		return fmt.Errorf("insufficient balance")
	}

	mapped, err := MapAccount(balanceSender, balanceSender.FirstName, balanceSender.LastName, deduct)
	if err != nil {
		return err
	}

	if _, err := s.store.UpdateBalance(recipient, recipient.Balance+req.Amount); err != nil {
		return err
	}

	sender, err := s.store.UpdateBalance(mapped, deduct)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, &MetaResponse{
		Status:  http.StatusOK,
		Message: "Transfer success",
		Data:    sender,
	})
}

func (s *APIServer) handleTopUp(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method %s are not allowed", r.Method)
	}

	number, err := getNumber(r)
	if err != nil {
		return err
	}

	req := new(UpdateBalanceRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}

	old, err := s.store.GetAccountByNumber(number)
	if err != nil {
		return err
	}

	mapped, err := MapAccount(old, old.FirstName, old.LastName, old.Balance+req.Amount)
	if err != nil {
		return err
	}

	account, err := s.store.UpdateBalance(mapped, old.Balance+req.Amount)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, &MetaResponse{
		Status:  http.StatusOK,
		Message: "Topup success",
		Data:    account,
	})
}

func getNumber(r *http.Request) (int64, error) {
	accountNumber := mux.Vars(r)["number"]
	number, err := strconv.ParseInt(accountNumber, 10, 64)
	if err != nil {
		return number, fmt.Errorf("given id %s invalid", accountNumber)
	}

	return number, nil
}

func getID(r *http.Request) (int, error) {
	strID := mux.Vars(r)["id"]
	id, err := strconv.Atoi(strID)
	if err != nil {
		return id, fmt.Errorf("given number %s is invalid", strID)
	}

	return id, nil
}
