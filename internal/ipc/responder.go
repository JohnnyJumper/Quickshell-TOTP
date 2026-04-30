package ipc

import (
	"encoding/json"
	"fmt"
	"net"
)

type Request struct {
	Type      request `json:"type"`
	Payload   string  `json:"payload,omitempty"`
	AccountID string  `json:"accountId,omitempty"`
}

type Response struct {
	OK       bool      `json:"ok"`
	Message  string    `json:"message,omitempty"`
	Error    string    `json:"error,omitempty"`
	Accounts []Account `json:"accounts,omitempty"`
}

type Account struct {
	ID        string `json:"id"`
	Issuer    string `json:"issuer"`
	Label     string `json:"label"`
	Algorithm string `json:"algorithm"`
	Digits    int    `json:"digits"`
	Period    int    `json:"period"`
}

func Success(message string) Response {
	return Response{
		OK:      true,
		Message: message,
	}
}

func Failure(err error) Response {
	return Response{
		OK:    false,
		Error: err.Error(),
	}
}

func AccountsImported(accounts []Account) Response {
	return Response{
		OK:       true,
		Message:  fmt.Sprintf("imported %d account(s)", len(accounts)),
		Accounts: accounts,
	}
}

func AccountsListed(accounts []Account) Response {
	return Response{
		OK:       true,
		Accounts: accounts,
	}
}

func WriteResponse(conn net.Conn, response Response) error {
	if err := json.NewEncoder(conn).Encode(response); err != nil {
		return fmt.Errorf("Failed to write response: %w", err)
	}

	return nil
}

func ReadRequest(conn net.Conn) (Request, error) {
	var request Request
	if err := json.NewDecoder(conn).Decode(&request); err != nil {
		return Request{}, fmt.Errorf("Failed to read request: %w", err)
	}

	return request, nil
}
