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
)

func TestWalletService_CreateTransfer(t *testing.T) {
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
			name: "happy path - successful transfer",
			args: func() args {
				c, w := newMockGinContext(t, http.MethodPost, "/v1/payment/transfers", &dto.CreateTransferRequest{
					IdempotencyKey: "idempotency-key",
					Amount:         1000,
					Currency:       "MYR",
					SourceAccount: dto.CreateTransferRequestAccountDetail{
						Number: "source-account",
					},
					DestinationAccount: dto.CreateTransferRequestAccountDetail{
						Number: "destination-account",
					},
					Note: "Test transfer",
				})
				return args{c: c, w: w}
			}(),
			setupMocks: func(fields *fields) {
				fields.validator = newMockValidator()
				fields.accountDAO = newMockAccountDAO(t)
				fields.transferDAO = newMockTransferDAO(t)
				fields.transactionDAO = newMockTransactionDAO(t)

				// Create a mock transfer logic with expectations
				mockTransferLogic := transfermock.NewMockITransferLogic(t)
				mockTransferLogic.On("CreateTransfer",
					mock.Anything,
					mock.MatchedBy(func(req *dto.CreateTransferRequest) bool {
						return req.IdempotencyKey == "idempotency-key" &&
							req.Amount == 1000 &&
							req.Currency == "MYR" &&
							req.SourceAccount.Number == "source-account" &&
							req.DestinationAccount.Number == "destination-account" &&
							req.Note == "Test transfer"
					}),
					mock.MatchedBy(func(opts *transfer.CreateTransferOpts) bool {
						return opts.TxType == transfer.TxTypeP2PTransfer
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

				fields.transferLogic = mockTransferLogic
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
				req, _ := http.NewRequest(http.MethodPost, "/v1/payment/transfers", bytes.NewBufferString("{invalid json}"))
				req.Header.Set("Content-Type", "application/json")
				c.Request = req

				return args{c: c, w: w}
			}(),
			setupMocks: func(fields *fields) {
				fields.validator = newMockValidator()
				fields.accountDAO = newMockAccountDAO(t)
				fields.transferDAO = newMockTransferDAO(t)
				fields.transactionDAO = newMockTransactionDAO(t)
				fields.transferLogic = transfermock.NewMockITransferLogic(t)
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
				c, w := newMockGinContext(t, http.MethodPost, "/v1/payment/transfers", &dto.CreateTransferRequest{
					IdempotencyKey: "idempotency-key",
					Amount:         1000,
					Currency:       "MYR",
					SourceAccount: dto.CreateTransferRequestAccountDetail{
						Number: "source-account",
					},
					DestinationAccount: dto.CreateTransferRequestAccountDetail{
						Number: "destination-account",
					},
					Note: "Test transfer",
				})
				return args{c: c, w: w}
			}(),
			setupMocks: func(fields *fields) {
				fields.validator = newMockValidator()
				fields.accountDAO = newMockAccountDAO(t)
				fields.transferDAO = newMockTransferDAO(t)
				fields.transactionDAO = newMockTransactionDAO(t)

				// Create a mock transfer logic with expectations to return an error
				mockTransferLogic := transfermock.NewMockITransferLogic(t)
				mockTransferLogic.On("CreateTransfer",
					mock.Anything,
					mock.MatchedBy(func(req *dto.CreateTransferRequest) bool {
						return req.IdempotencyKey == "idempotency-key" &&
							req.Amount == 1000 &&
							req.Currency == "MYR" &&
							req.SourceAccount.Number == "source-account" &&
							req.DestinationAccount.Number == "destination-account" &&
							req.Note == "Test transfer"
					}),
					mock.MatchedBy(func(opts *transfer.CreateTransferOpts) bool {
						return opts.TxType == transfer.TxTypeP2PTransfer
					}),
				).
					Return(nil, errors.New("transfer error")).
					Once()

				fields.transferLogic = mockTransferLogic
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
			p.CreateTransfer(tt.args.c)

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
