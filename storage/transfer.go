package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type Transfer struct {
	ID                      int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	Type                    string          `gorm:"type:varchar(36);not null;default:''" json:"type"`
	TxType                  string          `gorm:"type:varchar(36);not null;default:''" json:"tx_type"`
	TransactionID           string          `gorm:"type:varchar(36);not null;uniqueIndex:uk_transaction_id" json:"transaction_id"`
	ReferenceID             string          `gorm:"type:varchar(36);not null;uniqueIndex:uk_reference_id" json:"reference_id"`
	Status                  string          `gorm:"type:varchar(36);not null;default:''" json:"status"`
	Amount                  int64           `gorm:"not null" json:"amount"`
	Currency                string          `gorm:"type:varchar(3);not null;default:''" json:"currency"`
	SourceAccountID         string          `gorm:"type:varchar(36)" json:"source_account_id,omitempty"`
	SourceAccount           json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"source_account"`
	DestinationAccountID    string          `gorm:"type:varchar(36)" json:"destination_account_id,omitempty"`
	DestinationAccount      json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"destination_account"`
	StatusReason            string          `gorm:"type:varchar(255);default:''" json:"status_reason"`
	StatusReasonDescription string          `gorm:"type:varchar(500);default:''" json:"status_reason_description"`
	Note                    string          `gorm:"type:varchar(255);not null;default:''" json:"note"`
	Properties              json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"properties"`
	CreatedAt               time.Time       `gorm:"not null;default:now()" json:"created_at"`
	ValuedAt                *time.Time      `json:"valued_at,omitempty"`
	UpdatedAt               time.Time       `gorm:"not null;default:now()" json:"updated_at"`
}

type TxFn func(tx *gorm.DB) error

// TransferDAO handles DB operations for transfer
type transferDAO struct {
	DB *gorm.DB
}

type ITransferDAO interface {
	FindByReferenceID(ctx context.Context, referenceID string) (*Transfer, error)
	FindByAccountIDWithCursor(ctx context.Context, accountID string, limit int, beforeTimestamp *time.Time) ([]*Transfer, error)
	RunInTransaction(fn TxFn, opts ...*sql.TxOptions) error
}

func NewTransferDAO(db *gorm.DB) ITransferDAO {
	return &transferDAO{DB: db}
}

func (t *transferDAO) FindByReferenceID(ctx context.Context, referenceID string) (*Transfer, error) {
	var transfer Transfer
	err := t.DB.WithContext(ctx).
		Where("reference_id = ?", referenceID).
		First(&transfer).Error
	if err != nil {
		return nil, err
	}
	return &transfer, nil
}

func (t *transferDAO) RunInTransaction(fn TxFn, opts ...*sql.TxOptions) error {
	return t.DB.Transaction(fn, opts...)
}

func (t *transferDAO) FindByAccountIDWithCursor(
	ctx context.Context,
	accountID string,
	limit int,
	beforeTimestamp *time.Time,
) ([]*Transfer, error) {

	query := t.DB.WithContext(ctx).
		Where("source_account_id = ?", accountID).
		Or("destination_account_id = ?", accountID)

	// Only fetch older transactions if cursor exists
	if beforeTimestamp != nil {
		query = query.Where("created_at < ?", *beforeTimestamp)
	}

	var txs []*Transfer
	err := query.
		Order("created_at DESC, id DESC"). // ensure deterministic ordering
		Limit(limit).
		Find(&txs).Error

	if err != nil {
		return nil, err
	}
	return txs, nil
}
