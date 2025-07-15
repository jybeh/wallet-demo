package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"wallet/dto"
	"wallet/logic/transfer"
)

func (p *WalletService) CreateWithdrawal(c *gin.Context) {
	var req dto.CreateWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	res, createErr := p.transferLogic.CreateTransfer(c.Request.Context(), createWithdrawalRequestToCreateTransferRequest(&req), &transfer.CreateTransferOpts{
		TxType: transfer.TxTypeWithdrawal,
	})
	if createErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create transfer",
			"details": createErr.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func createWithdrawalRequestToCreateTransferRequest(req *dto.CreateWithdrawalRequest) *dto.CreateTransferRequest {
	return &dto.CreateTransferRequest{
		Currency: "MYR",
		Amount:   req.Amount,
		SourceAccount: dto.CreateTransferRequestAccountDetail{
			Number: req.AccountID,
		},
		Note:           req.Note,
		IdempotencyKey: req.IdempotencyKey,
	}
}
