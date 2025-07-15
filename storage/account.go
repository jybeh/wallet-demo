package storage

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"time"
)

type Account struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	AccountID string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"account_id"`
	Name      string    `gorm:"type:text;not null" json:"name"`
	Type      string    `gorm:"type:varchar(20);not null;default:wallet" json:"type"`
	Currency  string    `gorm:"type:char(3);not null;default:MYR" json:"currency"`
	Balance   int64     `gorm:"not null;default:0" json:"balance"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"updated_at"`
}

// accountDAO handles DB operations for accounts
type accountDAO struct {
	DB *gorm.DB
}

// todo add mockery
type IAccountDAO interface {
	FindByAccountID(context.Context, string) (*Account, error)
	UpdateBalance(context.Context, *Account, int64) error
}

func NewAccountDAO(db *gorm.DB) IAccountDAO {
	return &accountDAO{DB: db}
}

func (dao *accountDAO) FindByAccountID(ctx context.Context, accountID string) (*Account, error) {
	var acc Account
	err := dao.DB.WithContext(ctx).
		Where("account_id = ?", accountID).
		First(&acc).Error
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

func (dao *accountDAO) Create(ctx context.Context, account *Account) (*Account, error) {
	if createErr := dao.DB.WithContext(ctx).Create(account).Error; createErr != nil {
		return nil, createErr
	}
	return dao.FindByAccountID(ctx, account.AccountID)
}

func (dao *accountDAO) UpdateBalance(ctx context.Context, selectedAccount *Account, amountDelta int64) error {
	result := dao.DB.WithContext(ctx).
		Model(&Account{}).
		Where("account_id = ? AND updated_at = ?", selectedAccount.AccountID, selectedAccount.UpdatedAt).
		UpdateColumns(map[string]interface{}{
			"balance":    gorm.Expr("balance + ?", amountDelta),
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected <= 0 {
		return errors.New("concurrent balance update")
	}
	return nil
}
