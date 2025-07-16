package transfer

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"wallet/dto"
	"wallet/storage"
	storagemock "wallet/storage/mocks"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func Test_logicImpl_CreateTransfer(t *testing.T) {
	type fields struct {
		TransferDAO      storage.ITransferDAO
		AccountDAO       storage.IAccountDAO
		TransactionDAO   storage.ITransactionDAO
		holdingAccountID string
	}
	type args struct {
		ctx  context.Context
		req  *dto.CreateTransferRequest
		opts *CreateTransferOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *dto.CreateTransferResponse
		wantErr bool
	}{
		{
			name: "happy path - record exist",
			fields: fields{
				TransferDAO: func() storage.ITransferDAO {
					mc := &storagemock.MockITransferDAO{}
					mc.On("FindByReferenceID", context.Background(), "idempotency-key").Return(&storage.Transfer{
						Status:      "COMPLETED",
						ReferenceID: "idempotency-key",
						Amount:      1000,
					}, nil).Once()
					return mc
				}(),
			},
			args: args{
				ctx: context.Background(),
				req: &dto.CreateTransferRequest{
					IdempotencyKey: "idempotency-key",
				},
				opts: &CreateTransferOpts{
					TxType: TxTypeP2PTransfer,
				},
			},
			want: &dto.CreateTransferResponse{
				Status:         "COMPLETED",
				IdempotencyKey: "idempotency-key",
				Amount:         1000,
			},
			wantErr: false,
		},
		{
			name: "error - FindByReferenceID returns error",
			fields: fields{
				TransferDAO: func() storage.ITransferDAO {
					mc := &storagemock.MockITransferDAO{}
					mc.On("FindByReferenceID", context.Background(), "idempotency-key").Return(nil, errors.New("database error")).Once()
					return mc
				}(),
			},
			args: args{
				ctx: context.Background(),
				req: &dto.CreateTransferRequest{
					IdempotencyKey: "idempotency-key",
				},
				opts: &CreateTransferOpts{
					TxType: TxTypeP2PTransfer,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error - insufficient balance",
			fields: fields{
				TransferDAO: func() storage.ITransferDAO {
					mc := &storagemock.MockITransferDAO{}
					mc.On("FindByReferenceID", context.Background(), "idempotency-key").Return(nil, gorm.ErrRecordNotFound).Once()
					return mc
				}(),
				AccountDAO: func() storage.IAccountDAO {
					mc := &storagemock.MockIAccountDAO{}
					mc.On("FindByAccountID", context.Background(), "source-account").Return(&storage.Account{
						AccountID: "source-account",
						Balance:   500, // Less than the requested amount
					}, nil).Once()
					return mc
				}(),
			},
			args: args{
				ctx: context.Background(),
				req: &dto.CreateTransferRequest{
					IdempotencyKey: "idempotency-key",
					Amount:         1000,
					SourceAccount: dto.CreateTransferRequestAccountDetail{
						Number: "source-account",
					},
					DestinationAccount: dto.CreateTransferRequestAccountDetail{
						Number: "destination-account",
					},
				},
				opts: &CreateTransferOpts{
					TxType: TxTypeP2PTransfer,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "happy path - P2P transfer successful",
			fields: fields{
				TransferDAO: func() storage.ITransferDAO {
					mc := &storagemock.MockITransferDAO{}
					mc.On("FindByReferenceID", context.Background(), "idempotency-key").Return(nil, gorm.ErrRecordNotFound).Once()
					mc.On("RunInTransaction", mock.AnythingOfType("storage.TxFn")).Return(nil).Once()
					mc.On("FindByReferenceID", context.Background(), "idempotency-key").Return(&storage.Transfer{
						Status:      "COMPLETED",
						ReferenceID: "idempotency-key",
						Amount:      1000,
					}, nil).Once()
					return mc
				}(),
				AccountDAO: func() storage.IAccountDAO {
					mc := &storagemock.MockIAccountDAO{}
					mc.On("FindByAccountID", context.Background(), "source-account").Return(&storage.Account{
						AccountID: "source-account",
						Balance:   2000,
					}, nil).Once()
					mc.On("FindByAccountID", context.Background(), "destination-account").Return(&storage.Account{
						AccountID: "destination-account",
						Balance:   1000,
					}, nil).Once()
					mc.On("UpdateBalance", context.Background(), mock.AnythingOfType("*storage.Account"), int64(-1000)).Return(nil).Once()
					mc.On("UpdateBalance", context.Background(), mock.AnythingOfType("*storage.Account"), int64(1000)).Return(nil).Once()
					return mc
				}(),
			},
			args: args{
				ctx: context.Background(),
				req: &dto.CreateTransferRequest{
					IdempotencyKey: "idempotency-key",
					Amount:         1000,
					Currency:       "MYR",
					SourceAccount: dto.CreateTransferRequestAccountDetail{
						Number: "source-account",
					},
					DestinationAccount: dto.CreateTransferRequestAccountDetail{
						Number: "destination-account",
					},
				},
				opts: &CreateTransferOpts{
					TxType: TxTypeP2PTransfer,
				},
			},
			want: &dto.CreateTransferResponse{
				IdempotencyKey: "idempotency-key",
				Amount:         1000,
				Status:         "COMPLETED",
			},
			wantErr: false,
		},
		{
			name: "happy path - Withdrawal successful",
			fields: fields{
				TransferDAO: func() storage.ITransferDAO {
					mc := &storagemock.MockITransferDAO{}
					mc.On("FindByReferenceID", context.Background(), "idempotency-key").Return(nil, gorm.ErrRecordNotFound).Once()
					mc.On("RunInTransaction", mock.AnythingOfType("storage.TxFn")).Return(nil).Once()
					mc.On("FindByReferenceID", context.Background(), "idempotency-key").Return(&storage.Transfer{
						Status:      "COMPLETED",
						ReferenceID: "idempotency-key",
						Amount:      1000,
					}, nil).Once()
					return mc
				}(),
				AccountDAO: func() storage.IAccountDAO {
					mc := &storagemock.MockIAccountDAO{}
					mc.On("FindByAccountID", context.Background(), "source-account").Return(&storage.Account{
						AccountID: "source-account",
						Balance:   2000,
					}, nil).Once()
					mc.On("FindByAccountID", context.Background(), "1000000001").Return(&storage.Account{
						AccountID: "1000000001",
						Balance:   10000,
					}, nil).Once()
					mc.On("UpdateBalance", context.Background(), mock.AnythingOfType("*storage.Account"), int64(-1000)).Return(nil).Once()
					mc.On("UpdateBalance", context.Background(), mock.AnythingOfType("*storage.Account"), int64(1000)).Return(nil).Once()
					return mc
				}(),
				holdingAccountID: "1000000001",
			},
			args: args{
				ctx: context.Background(),
				req: &dto.CreateTransferRequest{
					IdempotencyKey: "idempotency-key",
					Amount:         1000,
					Currency:       "MYR",
					SourceAccount: dto.CreateTransferRequestAccountDetail{
						Number: "source-account",
					},
					DestinationAccount: dto.CreateTransferRequestAccountDetail{
						Number: "destination-account", // This will be overridden with holdingAccountID
					},
				},
				opts: &CreateTransferOpts{
					TxType: TxTypeWithdrawal,
				},
			},
			want: &dto.CreateTransferResponse{
				IdempotencyKey: "idempotency-key",
				Amount:         1000,
				Status:         "COMPLETED",
			},
			wantErr: false,
		},
		{
			name: "happy path - Deposit successful",
			fields: fields{
				TransferDAO: func() storage.ITransferDAO {
					mc := &storagemock.MockITransferDAO{}
					mc.On("FindByReferenceID", context.Background(), "idempotency-key").Return(nil, gorm.ErrRecordNotFound).Once()
					mc.On("RunInTransaction", mock.AnythingOfType("storage.TxFn")).Return(nil).Once()
					mc.On("FindByReferenceID", context.Background(), "idempotency-key").Return(&storage.Transfer{
						Status:      "COMPLETED",
						ReferenceID: "idempotency-key",
						Amount:      1000,
					}, nil).Once()
					return mc
				}(),
				AccountDAO: func() storage.IAccountDAO {
					mc := &storagemock.MockIAccountDAO{}
					mc.On("FindByAccountID", context.Background(), "1000000001").Return(&storage.Account{
						AccountID: "1000000001",
						Balance:   10000,
					}, nil).Once()
					mc.On("FindByAccountID", context.Background(), "destination-account").Return(&storage.Account{
						AccountID: "destination-account",
						Balance:   1000,
					}, nil).Once()
					mc.On("UpdateBalance", context.Background(), mock.AnythingOfType("*storage.Account"), int64(-1000)).Return(nil).Once()
					mc.On("UpdateBalance", context.Background(), mock.AnythingOfType("*storage.Account"), int64(1000)).Return(nil).Once()
					return mc
				}(),
				holdingAccountID: "1000000001",
			},
			args: args{
				ctx: context.Background(),
				req: &dto.CreateTransferRequest{
					IdempotencyKey: "idempotency-key",
					Amount:         1000,
					Currency:       "MYR",
					SourceAccount: dto.CreateTransferRequestAccountDetail{
						Number: "source-account", // This will be overridden with holdingAccountID
					},
					DestinationAccount: dto.CreateTransferRequestAccountDetail{
						Number: "destination-account",
					},
				},
				opts: &CreateTransferOpts{
					TxType: TxTypeDeposit,
				},
			},
			want: &dto.CreateTransferResponse{
				IdempotencyKey: "idempotency-key",
				Amount:         1000,
				Status:         "COMPLETED",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &logicImpl{
				TransferDAO:      tt.fields.TransferDAO,
				AccountDAO:       tt.fields.AccountDAO,
				TransactionDAO:   tt.fields.TransactionDAO,
				holdingAccountID: tt.fields.holdingAccountID,
			}
			got, err := l.CreateTransfer(tt.args.ctx, tt.args.req, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTransfer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateTransfer() got = %v, want %v", got, tt.want)
			}
		})
	}
}
