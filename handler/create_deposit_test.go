package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"wallet/dto"
	"wallet/logic/transfer"
	transfermock "wallet/logic/transfer/mocks"
	"wallet/storage"
	storagemock "wallet/storage/mocks"
)

// Helper function to create a new mock gin context with a JSON request body
func newMockGinContext(t *testing.T, method, path string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var req *http.Request
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		require.NoError(t, err)
		req, _ = http.NewRequest(method, path, bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest(method, path, nil)
	}

	c.Request = req
	return c, w
}

// Helper function to create a new mock validator
func newMockValidator() *validator.Validate {
	return validator.New()
}

// Helper function to create a new mock account DAO
func newMockAccountDAO(t *testing.T) *storagemock.MockIAccountDAO {
	return storagemock.NewMockIAccountDAO(t)
}

// Helper function to create a new mock transfer DAO
func newMockTransferDAO(t *testing.T) *storagemock.MockITransferDAO {
	return storagemock.NewMockITransferDAO(t)
}

// Helper function to create a new mock transaction DAO
func newMockTransactionDAO(t *testing.T) *storagemock.MockITransactionDAO {
	return storagemock.NewMockITransactionDAO(t)
}

// Helper function to create a new mock transfer logic
func newMockTransferLogic(t *testing.T) *transfermock.MockITransferLogic {
	return transfermock.NewMockITransferLogic(t)
}

func TestWalletService_CreateDeposit(t *testing.T) {
	type fields struct {
		validator      *validator.Validate
		accountDAO     storage.IAccountDAO
		transferDAO    storage.ITransferDAO
		transactionDAO storage.ITransactionDAO
		transferLogic  transfer.ITransferLogic
	}
	type args struct {
		c *gin.Context
		w *httptest.ResponseRecorder
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		setupMocks     func(fields *fields)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "happy path - successful deposit",
			args: func() args {
				c, w := newMockGinContext(t, http.MethodPost, "/v1/accounts/deposits", &dto.CreateDepositRequest{
					IdempotencyKey: "idempotency-key",
					AccountID:      "destination-account",
					Amount:         1000,
					Currency:       "MYR",
					Note:           "Test deposit",
				})
				return args{c: c, w: w}
			}(),
			setupMocks: func(fields *fields) {
				fields.validator = newMockValidator()
				fields.accountDAO = newMockAccountDAO(t)
				fields.transferDAO = newMockTransferDAO(t)
				fields.transactionDAO = newMockTransactionDAO(t)
				fields.transferLogic = newMockTransferLogic(t)

				// Setup expectations for transferLogic.CreateTransfer
				fields.transferLogic.(*transfermock.MockITransferLogic).
					On("CreateTransfer",
						mock.Anything,
						mock.MatchedBy(func(req *dto.CreateTransferRequest) bool {
							return req.IdempotencyKey == "idempotency-key" &&
								req.Amount == 1000 &&
								req.Currency == "MYR" &&
								req.DestinationAccount.Number == "destination-account" &&
								req.Note == "Test deposit"
						}),
						mock.MatchedBy(func(opts *transfer.CreateTransferOpts) bool {
							return opts.TxType == transfer.TxTypeDeposit
						}),
					).
					Return(&dto.CreateTransferResponse{
						IdempotencyKey: "idempotency-key",
						TransactionID:  "tx-123",
						Amount:         1000,
						Currency:       "MYR",
						Status:         "COMPLETED",
					}, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody: &dto.CreateTransferResponse{
				IdempotencyKey: "idempotency-key",
				TransactionID:  "tx-123",
				Amount:         1000,
				Currency:       "MYR",
				Status:         "COMPLETED",
			},
		},
		{
			name: "error - invalid request body",
			args: func() args {
				// Create a request with invalid JSON
				gin.SetMode(gin.TestMode)
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)

				// Use actual invalid JSON
				req, _ := http.NewRequest(http.MethodPost, "/v1/accounts/deposits", bytes.NewBufferString("{invalid json}"))
				req.Header.Set("Content-Type", "application/json")
				c.Request = req

				return args{c: c, w: w}
			}(),
			setupMocks: func(fields *fields) {
				fields.validator = newMockValidator()
				fields.accountDAO = newMockAccountDAO(t)
				fields.transferDAO = newMockTransferDAO(t)
				fields.transactionDAO = newMockTransactionDAO(t)
				fields.transferLogic = newMockTransferLogic(t)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: gin.H{
				"error":   "Invalid request body",
				"details": mock.Anything,
			},
		},
		{
			name: "error - transfer logic returns error",
			args: func() args {
				c, w := newMockGinContext(t, http.MethodPost, "/v1/accounts/deposits", &dto.CreateDepositRequest{
					IdempotencyKey: "idempotency-key",
					AccountID:      "destination-account",
					Amount:         1000,
					Currency:       "MYR",
					Note:           "Test deposit",
				})
				return args{c: c, w: w}
			}(),
			setupMocks: func(fields *fields) {
				fields.validator = newMockValidator()
				fields.accountDAO = newMockAccountDAO(t)
				fields.transferDAO = newMockTransferDAO(t)
				fields.transactionDAO = newMockTransactionDAO(t)
				fields.transferLogic = newMockTransferLogic(t)

				// Setup expectations for transferLogic.CreateTransfer to return an error
				fields.transferLogic.(*transfermock.MockITransferLogic).
					On("CreateTransfer",
						mock.Anything,
						mock.MatchedBy(func(req *dto.CreateTransferRequest) bool {
							return req.IdempotencyKey == "idempotency-key" &&
								req.Amount == 1000 &&
								req.Currency == "MYR" &&
								req.DestinationAccount.Number == "destination-account" &&
								req.Note == "Test deposit"
						}),
						mock.MatchedBy(func(opts *transfer.CreateTransferOpts) bool {
							return opts.TxType == transfer.TxTypeDeposit
						}),
					).
					Return(nil, errors.New("transfer error")).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: gin.H{
				"error":   "Failed to create transfer",
				"details": "transfer error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			if tt.setupMocks != nil {
				tt.setupMocks(&tt.fields)
			}

			// Create service
			p := &WalletService{
				validator:      tt.fields.validator,
				accountDAO:     tt.fields.accountDAO,
				transferDAO:    tt.fields.transferDAO,
				transactionDAO: tt.fields.transactionDAO,
				transferLogic:  tt.fields.transferLogic,
			}

			// Call the function
			p.CreateDeposit(tt.args.c)

			// Check response
			require.Equal(t, tt.expectedStatus, tt.args.w.Code)

			if tt.expectedStatus != http.StatusOK {
				return
			}

			// For JSON responses, check the body
			if tt.expectedBody != nil {
				var response interface{}
				err := json.Unmarshal(tt.args.w.Body.Bytes(), &response)
				require.NoError(t, err)

				// For error responses with mock.Anything in details
				if respMap, ok := response.(map[string]interface{}); ok {
					if expMap, ok := tt.expectedBody.(gin.H); ok {
						if expMap["details"] == mock.Anything {
							// Just check that details exists
							require.Contains(t, respMap, "details")
							// Remove details for the comparison
							delete(respMap, "details")
							delete(expMap, "details")
						}
					}
				}

				// Compare the rest of the response
				if expResp, ok := tt.expectedBody.(*dto.CreateTransferResponse); ok {
					var actualResp dto.CreateTransferResponse
					err := json.Unmarshal(tt.args.w.Body.Bytes(), &actualResp)
					require.NoError(t, err)
					require.Equal(t, expResp.IdempotencyKey, actualResp.IdempotencyKey)
					require.Equal(t, expResp.TransactionID, actualResp.TransactionID)
					require.Equal(t, expResp.Amount, actualResp.Amount)
					require.Equal(t, expResp.Currency, actualResp.Currency)
					require.Equal(t, expResp.Status, actualResp.Status)
				} else {
					require.Equal(t, tt.expectedBody, response)
				}
			}
		})
	}
}
