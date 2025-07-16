package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"slices"
	"time"
	"wallet/dto"
	"wallet/logic/transfer"
	"wallet/storage"
	"wallet/util"
)

func (p *WalletService) GetAccountTransactions(c *gin.Context) {
	var req dto.GetAccountTransactionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// default limit if empty
	if req.Limit <= 0 {
		req.Limit = 20
	}

	transfersData, nextToken, listErr := p.listAccountTransfers(c.Request.Context(), req.AccountID, req.Limit, req.NextToken)
	if errors.Is(listErr, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusOK, &dto.GetAccountTransactionsResponse{})
		return
	}
	if listErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch transactions"})
		return
	}

	var resp []*dto.TransactionResponse
	for _, tx := range transfersData {
		amt := tx.Amount
		if tx.SourceAccountID == req.AccountID && slices.Contains([]string{
			string(transfer.TxTypeWithdrawal),
			string(transfer.TxTypeP2PTransfer)}, tx.TxType) {
			amt = -amt
		}
		resp = append(resp, &dto.TransactionResponse{
			TransactionID: tx.TransactionID,
			TxType:        tx.TxType,
			Amount:        amt,
			Currency:      tx.Currency,
			CreatedAt:     tx.CreatedAt,
			Note:          tx.Note,
			Status:        tx.Status,
		})
	}

	c.JSON(http.StatusOK, &dto.GetAccountTransactionsResponse{
		Data:      resp,
		NextToken: nextToken,
	})
}

func (p *WalletService) listAccountTransfers(
	ctx context.Context,
	accountID string,
	limit int,
	nextToken string,
) ([]*storage.Transfer, string, error) {

	// decode cursor
	var beforeTimestamp *time.Time
	if nextToken != "" {
		cursor, err := util.DecodeNextToken(nextToken)
		if err != nil {
			return nil, "", fmt.Errorf("invalid nextToken: %w", err)
		}
		beforeTimestamp = &cursor.LastTimestamp
	}

	// call DAO
	txs, err := p.transferDAO.FindByAccountIDWithCursor(ctx, accountID, limit, beforeTimestamp)
	if err != nil {
		return nil, "", err
	}

	// create new cursor if we still have more
	var newNextToken string
	if len(txs) == limit {
		last := txs[len(txs)-1]
		newNextToken = util.EncodeNextToken(util.DataCursor{
			LastTimestamp: last.CreatedAt,
			LastID:        last.TransactionID,
		})
	}

	return txs, newNextToken, nil
}
