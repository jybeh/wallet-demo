package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"wallet/dto"
	"wallet/logic/transfer"
)

func (p *WalletService) CreateTransfer(c *gin.Context) {
	var req dto.CreateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	res, createErr := p.transferLogic.CreateTransfer(c.Request.Context(), &req, &transfer.CreateTransferOpts{
		TxType: transfer.TxTypeP2PTransfer,
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
