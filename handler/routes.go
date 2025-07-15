package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"wallet/logic/transfer"
	"wallet/storage"
)

type WalletService struct {
	validator *validator.Validate

	accountDAO     storage.IAccountDAO
	transferDAO    storage.ITransferDAO
	transactionDAO storage.ITransactionDAO

	transferLogic transfer.ITransferLogic
}

func NewWalletService(
	AccountDAO storage.IAccountDAO,
	TransactionDAO storage.ITransactionDAO,
	TransferDAO storage.ITransferDAO,
) *WalletService {
	return &WalletService{
		validator:      validator.New(),
		accountDAO:     AccountDAO,
		transferDAO:    TransferDAO,
		transactionDAO: TransactionDAO,
		transferLogic:  transfer.NewTransferLogic(TransferDAO, AccountDAO, TransactionDAO),
	}
}

func (p *WalletService) RegisterRoutes(ge *gin.Engine) {
	v1 := ge.Group("/v1")

	v1accounts := v1.Group("/accounts")
	{
		v1accounts.GET("/:accountID/transactions", p.GetAccountTransactions)
		v1accounts.POST("/query", p.GetAccountDetails)
		v1accounts.POST("/withdrawals", p.CreateWithdrawal)
		v1accounts.POST("/deposits", p.CreateDeposit)
	}

	v1transfers := v1.Group("/payment")
	{
		v1transfers.POST("/transfers", p.CreateTransfer)
	}
}
