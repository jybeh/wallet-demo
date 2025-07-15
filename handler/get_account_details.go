package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"wallet/storage"
)

type GetAccountDetailRequest struct {
	AccountID string `json:"accountID" binding:"required"`
}

type GetAccountDetailResponse struct {
	AccountID string `json:"accountID"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Currency  string `json:"currency"`
	Balance   int64  `json:"balance"` // in minor unit (e.g. sen/cents)
}

func (p *WalletService) GetAccountDetails(c *gin.Context) {
	var req GetAccountDetailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	account, err := p.accountDAO.FindByAccountID(c.Request.Context(), req.AccountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve account"})
		}
		return
	}

	c.JSON(http.StatusOK, toGetAccountDetailResponse(account))
}

func toGetAccountDetailResponse(a *storage.Account) *GetAccountDetailResponse {
	return &GetAccountDetailResponse{
		AccountID: a.AccountID,
		Name:      a.Name,
		Type:      a.Type,
		Currency:  a.Currency,
		Balance:   a.Balance,
	}
}
