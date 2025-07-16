package transfer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
	"wallet/dto"
	"wallet/storage"
	"wallet/util"
)

const (
	TxStatusPROCESSING = "PROCESSING"
	TxStatusCOMPLETED  = "COMPLETED"
	TypeDebit          = "debit"
	TypeCredit         = "credit"
)

var (
	InsufficientBalanceErr       = errors.New("insufficient balance")
	InvalidAmountErr             = errors.New("invalid amount")
	InvalidCurrencyErr           = errors.New("invalid currency")
	InvalidSourceAccountErr      = errors.New("invalid source account")
	InvalidDestinationAccountErr = errors.New("invalid destination account")
	PossibleErrors               = []error{
		InsufficientBalanceErr,
		InvalidAmountErr,
		InvalidCurrencyErr,
		InvalidSourceAccountErr,
		InvalidDestinationAccountErr,
	}
)

func IsOneOfTransferErrors(err error) bool {
	for _, e := range PossibleErrors {
		if errors.Is(err, e) {
			return true
		}
	}
	return false
}

type logicImpl struct {
	TransferDAO    storage.ITransferDAO
	AccountDAO     storage.IAccountDAO
	TransactionDAO storage.ITransactionDAO

	holdingAccountID string
}

type CreateTransferOpts struct {
	TxType TxType
}

type TxType string

const (
	TxTypeWithdrawal  TxType = "WITHDRAWAL"
	TxTypeP2PTransfer TxType = "TRANSFER"
	TxTypeDeposit     TxType = "DEPOSIT"
)

type ITransferLogic interface {
	CreateTransfer(context.Context, *dto.CreateTransferRequest, *CreateTransferOpts) (*dto.CreateTransferResponse, error)
}

func NewTransferLogic(
	td storage.ITransferDAO,
	ad storage.IAccountDAO,
	txd storage.ITransactionDAO) ITransferLogic {
	return &logicImpl{
		TransferDAO:      td,
		AccountDAO:       ad,
		TransactionDAO:   txd,
		holdingAccountID: "1000000001",
	}
}

func (l *logicImpl) CreateTransfer(ctx context.Context, req *dto.CreateTransferRequest, opts *CreateTransferOpts) (*dto.CreateTransferResponse, error) {
	// Idempotency check
	existing, err := l.TransferDAO.FindByReferenceID(ctx, req.IdempotencyKey)
	if existing != nil {
		return mapTransferStorageToResponse(existing), nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	transactionID := uuid.New().String()

	transferRecord := mapCreateTransferRequestToTransfer(req, transactionID, opts)

	if doErr := util.Retry(func() error {
		return l.doTransfer(ctx, transferRecord, opts)
	}, 3, 100*time.Millisecond,
		InvalidAmountErr,
		InvalidCurrencyErr,
		InvalidSourceAccountErr,
		InvalidDestinationAccountErr,
		InsufficientBalanceErr,
	); doErr != nil {
		return nil, doErr
	}

	// find data and return
	finalTx, err := l.TransferDAO.FindByReferenceID(ctx, req.IdempotencyKey)
	if err != nil {
		return nil, err
	}

	return mapTransferStorageToResponse(finalTx), nil
}

func (l *logicImpl) doTransfer(ctx context.Context, req *storage.Transfer, opts *CreateTransferOpts) error {
	var sourceAcc *storage.Account
	var destAcc *storage.Account
	var findErr error

	if opts == nil {
		return errors.New("invalid options")
	}

	// Load source and destination accounts
	switch opts.TxType {
	case TxTypeWithdrawal:
		req.DestinationAccountID = l.holdingAccountID
		sourceAcc, findErr = l.AccountDAO.FindByAccountID(ctx, req.SourceAccountID)
		if findErr != nil {
			return InvalidSourceAccountErr
		}
		if sourceAcc.Balance < req.Amount {
			return InsufficientBalanceErr
		}
		destAcc, findErr = l.AccountDAO.FindByAccountID(ctx, l.holdingAccountID)
		if findErr != nil {
			return InvalidDestinationAccountErr
		}
		req.DestinationAccount = toAccountInfo(sourceAcc)
	case TxTypeP2PTransfer:
		sourceAcc, findErr = l.AccountDAO.FindByAccountID(ctx, req.SourceAccountID)
		if findErr != nil {
			return InvalidSourceAccountErr
		}
		if sourceAcc.Balance < req.Amount {
			return InsufficientBalanceErr
		}
		destAcc, findErr = l.AccountDAO.FindByAccountID(ctx, req.DestinationAccountID)
		if findErr != nil {
			return InvalidDestinationAccountErr
		}
		req.SourceAccount = toAccountInfo(sourceAcc)
		req.DestinationAccount = toAccountInfo(destAcc)
	case TxTypeDeposit:
		req.SourceAccountID = l.holdingAccountID
		sourceAcc, findErr = l.AccountDAO.FindByAccountID(ctx, l.holdingAccountID)
		if findErr != nil {
			return InvalidSourceAccountErr
		}
		destAcc, findErr = l.AccountDAO.FindByAccountID(ctx, req.DestinationAccountID)
		if findErr != nil {
			return InvalidDestinationAccountErr
		}
		req.SourceAccount = toAccountInfo(sourceAcc)
	}

	createTransferErr := l.TransferDAO.RunInTransaction(func(tx *gorm.DB) error {
		// Create Source Transaction (debit)
		if srcTxErr := tx.Create(&storage.Transaction{
			AccountID: req.SourceAccountID,
			Type:      TypeDebit,
			Amount:    req.Amount,
			Currency:  req.Currency,
			Note:      fmt.Sprintf("Transfer to %s", req.DestinationAccountID),
			Timestamp: req.CreatedAt,
			ValuedAt:  req.CreatedAt,
			CreatedAt: req.CreatedAt,
			UpdatedAt: req.CreatedAt,
		}).Error; srcTxErr != nil {
			return srcTxErr
		}

		// Create Destination Transaction (credit)
		if dstTxErr := tx.Create(&storage.Transaction{
			AccountID: req.DestinationAccountID,
			Type:      TypeCredit,
			Amount:    req.Amount,
			Currency:  req.Currency,
			Note:      fmt.Sprintf("Transfer from %s", req.SourceAccountID),
			Timestamp: req.CreatedAt,
			ValuedAt:  req.CreatedAt,
			CreatedAt: req.CreatedAt,
			UpdatedAt: req.CreatedAt,
		}).Error; dstTxErr != nil {
			return dstTxErr
		}

		if updateBalanceErr := l.AccountDAO.UpdateBalance(ctx, sourceAcc, -req.Amount); updateBalanceErr != nil {
			return fmt.Errorf("source account update failed: %w", updateBalanceErr)
		}
		if updateBalanceErr := l.AccountDAO.UpdateBalance(ctx, destAcc, req.Amount); updateBalanceErr != nil {
			return fmt.Errorf("destination account update failed: %w", updateBalanceErr)
		}
		// Update transfer status to success
		req.Status = TxStatusCOMPLETED
		req.UpdatedAt = time.Now()
		if saveErr := tx.Save(req).Error; saveErr != nil {
			return saveErr
		}
		return nil
	})
	return createTransferErr
}

func mapCreateTransferRequestToTransfer(req *dto.CreateTransferRequest, transactionID string, opts *CreateTransferOpts) *storage.Transfer {
	now := time.Now()
	trf := &storage.Transfer{
		TxType:               string(opts.TxType),
		Type:                 "INTRA",
		TransactionID:        transactionID,
		ReferenceID:          req.IdempotencyKey,
		Status:               TxStatusPROCESSING,
		Amount:               req.Amount,
		Currency:             req.Currency,
		SourceAccountID:      req.SourceAccount.Number,
		DestinationAccountID: req.DestinationAccount.Number,
		Note:                 req.Note,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	marshalled, err := json.Marshal(req.Properties)
	if err != nil {
		return trf
	}
	trf.Properties = marshalled
	return trf
}

func toAccountInfo(a *storage.Account) json.RawMessage {
	marshalled, err := json.Marshal(map[string]interface{}{"number": a.AccountID})
	if err != nil {
		return nil
	}
	return marshalled
}

func mapTransferStorageToResponse(res *storage.Transfer) *dto.CreateTransferResponse {
	return &dto.CreateTransferResponse{
		IdempotencyKey: res.ReferenceID,
		TransactionID:  res.TransactionID,
		Amount:         res.Amount,
		Currency:       res.Currency,
		Status:         res.Status,
	}
}
