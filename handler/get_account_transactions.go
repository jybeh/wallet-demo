package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type GetAccountTransactionsPathParam struct {
	AccountID string `uri:"accountID" validate:"required"`
}

type PaginationQuery struct {
	Limit     int    `form:"limit" validate:"omitempty,min=1,max=100"`
	NextToken string `form:"nextToken" validate:"omitempty,uuid4"`
}

func (p *WalletService) GetAccountTransactions(c *gin.Context) {
	//accountID := c.Param("accountID")
	//
	//// Query params binding
	//var query TransactionQuery
	//if err := c.ShouldBindQuery(&query); err != nil {
	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
	//	return
	//}
	//
	//if err := validate.Struct(query); err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"validation_error": err.Error()})
	//	return
	//}
}
