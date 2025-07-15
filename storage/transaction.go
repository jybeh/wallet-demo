package storage

import (
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type Transaction struct {
	ID         int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	AccountID  string          `gorm:"type:varchar(64);not null" json:"account_id"`
	Type       string          `gorm:"type:varchar(10);not null;check:type IN ('credit','debit')" json:"type"`
	Amount     int64           `gorm:"not null;check:amount >= 0" json:"amount"`
	Currency   string          `gorm:"type:char(3);not null;default:'MYR'" json:"currency"`
	Timestamp  time.Time       `gorm:"not null;default:now()" json:"timestamp"`
	ValuedAt   time.Time       `gorm:"not null;default:now()" json:"valued_at"`
	UpdatedAt  time.Time       `gorm:"not null;default:now()" json:"updated_at"`
	CreatedAt  time.Time       `gorm:"not null;default:now()" json:"created_at"`
	Note       string          `json:"note,omitempty"`
	Properties json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"properties"`
}

// TransactionDAO handles DB operations for transactions
type TransactionDAO struct {
	DB *gorm.DB
}

// todo add mockery
type ITransactionDAO interface {
}

func NewTransactionDAO(db *gorm.DB) ITransactionDAO {
	return &TransactionDAO{DB: db}
}
