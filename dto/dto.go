package dto

import "time"

type CreateTransferRequest struct {
	Currency           string                             `json:"currency" binding:"required,oneof=MYR"`
	Amount             int64                              `json:"amount" binding:"required,gt=0"`        // must be positive, in minor unit
	SourceAccount      CreateTransferRequestAccountDetail `json:"sourceAccount" binding:"required"`      // required nested struct
	DestinationAccount CreateTransferRequestAccountDetail `json:"destinationAccount" binding:"required"` // required nested struct
	Properties         map[string]interface{}             `json:"properties"`                            // flexible metadata
	Note               string                             `json:"note"`                                  // optional
	IdempotencyKey     string                             `json:"idempotencyKey" binding:"required"`     // must be present for idempotency
}

type CreateTransferRequestAccountDetail struct {
	Number string `json:"number" binding:"required"`
}

type CreateTransferResponse struct {
	IdempotencyKey string `json:"idempotencyKey"`
	TransactionID  string `json:"transactionID"`
	Amount         int64  `json:"amount"`
	Currency       string `json:"currency"`
	Status         string `json:"status"`
}

type CreateDepositRequest struct {
	IdempotencyKey string `json:"idempotencyKey" binding:"required"`            // target account
	AccountID      string `json:"accountID" binding:"required"`                 // target account
	Amount         int64  `json:"amount" binding:"required,gt=0,lt=9999999999"` // must be positive, in minor units
	Currency       string `json:"currency" binding:"required,oneof=MYR"`
	Note           string `json:"note"` // optional
}

type CreateDepositResponse struct {
	TransactionID string `json:"transactionID"`
	Status        string `json:"status"`
}

type CreateWithdrawalRequest struct {
	IdempotencyKey string `json:"idempotencyKey" binding:"required"` // target account
	AccountID      string `json:"accountID" binding:"required"`      // target account
	Amount         int64  `json:"amount" binding:"required,gt=0"`    // must be positive, in minor units
	Currency       string `json:"currency" binding:"required,oneof=MYR"`
	Note           string `json:"note"` // optional
}

type CreateWithdrawalResponse struct {
	TransactionID string `json:"transactionID"`
	Status        string `json:"status"`
}

type GetAccountTransactionsRequest struct {
	AccountID string `json:"accountID" binding:"required"`
	Limit     int    `json:"limit" binding:"omitempty,min=1,max=100"`
	NextToken string `json:"nextToken" binding:"omitempty"`
}

type TransactionResponse struct {
	TransactionID string    `json:"transactionID"`
	TxType        string    `json:"txType"`
	Status        string    `json:"status"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"createdAt"`
	Note          string    `json:"note"`
}

type GetAccountTransactionsResponse struct {
	Data      []*TransactionResponse `json:"data"`
	NextToken string                 `json:"nextToken,omitempty"`
}
