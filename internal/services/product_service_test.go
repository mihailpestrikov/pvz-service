package services

import (
	"avito-backend-trainee-assignment-spring-2025/internal/domain/apperrors"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/interfaces"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
	"avito-backend-trainee-assignment-spring-2025/internal/repository/postgres"
	"avito-backend-trainee-assignment-spring-2025/internal/repository/repoerrors"
	"avito-backend-trainee-assignment-spring-2025/internal/services/mocks"
	"context"
	"database/sql"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"reflect"
	"testing"
	"time"
)

type MockTxManager struct {
	RunTransactionFunc func(ctx context.Context, fn func(*sql.Tx) error) error
}

func (m *MockTxManager) RunTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	return m.RunTransactionFunc(ctx, fn)
}

func TestProductService_AddProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := mocks.NewMockTxProductRepository(ctrl)
	mockReceptionRepo := mocks.NewMockTxReceptionRepository(ctrl)

	mockTxManager := &MockTxManager{
		RunTransactionFunc: func(ctx context.Context, fn func(*sql.Tx) error) error {
			return fn(nil)
		},
	}

	ctx := context.Background()
	pvzID := uuid.New()
	receptionID := uuid.New()

	validProduct := &models.Product{
		ID:          uuid.New(),
		DateTime:    time.Now(),
		Type:        models.ProductTypeElectronics,
		ReceptionID: receptionID,
	}

	activeReception := &models.Reception{
		ID:       receptionID,
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.ReceptionStatusInProgress,
	}

	closedReception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.ReceptionStatusClosed,
	}

	type fields struct {
		productRepo   interfaces.TxProductRepository
		receptionRepo interfaces.TxReceptionRepository
		txManager     postgres.TxManager
	}
	type args struct {
		ctx         context.Context
		productType string
		pvzID       uuid.UUID
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		setupMocks      func()
		want            *models.Product
		wantErr         bool
		expectedErrType error
	}{
		{
			name: "successful product addition",
			fields: fields{
				productRepo:   mockProductRepo,
				receptionRepo: mockReceptionRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:         ctx,
				productType: models.ProductTypeElectronics,
				pvzID:       pvzID,
			},
			setupMocks: func() {

				mockProductRepo.EXPECT().WithTx(gomock.Any()).Return(mockProductRepo)
				mockReceptionRepo.EXPECT().WithTx(gomock.Any()).Return(mockReceptionRepo)

				mockReceptionRepo.EXPECT().
					GetLastActiveByPVZID(gomock.Any(), pvzID).
					Return(activeReception, nil)

				mockProductRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, product *models.Product) error {

						validProduct.ID = product.ID
						validProduct.DateTime = product.DateTime
						validProduct.ReceptionID = product.ReceptionID
						return nil
					})
			},
			want:    validProduct,
			wantErr: false,
		},
		{
			name: "error: invalid product type",
			fields: fields{
				productRepo:   mockProductRepo,
				receptionRepo: mockReceptionRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:         ctx,
				productType: "wrong_type",
				pvzID:       pvzID,
			},
			setupMocks: func() {

			},
			want:            nil,
			wantErr:         true,
			expectedErrType: apperrors.ErrInvalidProductType,
		},
		{
			name: "error: reception closed",
			fields: fields{
				productRepo:   mockProductRepo,
				receptionRepo: mockReceptionRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:         ctx,
				productType: models.ProductTypeElectronics,
				pvzID:       pvzID,
			},
			setupMocks: func() {
				mockProductRepo.EXPECT().WithTx(gomock.Any()).Return(mockProductRepo)
				mockReceptionRepo.EXPECT().WithTx(gomock.Any()).Return(mockReceptionRepo)

				mockReceptionRepo.EXPECT().
					GetLastActiveByPVZID(gomock.Any(), pvzID).
					Return(closedReception, nil)
			},
			want:            nil,
			wantErr:         true,
			expectedErrType: apperrors.ErrReceptionCannotBeModified,
		},
		{
			name: "error: no active reception",
			fields: fields{
				productRepo:   mockProductRepo,
				receptionRepo: mockReceptionRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:         ctx,
				productType: models.ProductTypeElectronics,
				pvzID:       pvzID,
			},
			setupMocks: func() {
				mockProductRepo.EXPECT().WithTx(gomock.Any()).Return(mockProductRepo)
				mockReceptionRepo.EXPECT().WithTx(gomock.Any()).Return(mockReceptionRepo)

				mockReceptionRepo.EXPECT().
					GetLastActiveByPVZID(gomock.Any(), pvzID).
					Return(nil, apperrors.ErrNoActiveReception)
			},
			want:            nil,
			wantErr:         true,
			expectedErrType: apperrors.ErrNoActiveReception,
		},
		{
			name: "error creating product",
			fields: fields{
				productRepo:   mockProductRepo,
				receptionRepo: mockReceptionRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:         ctx,
				productType: models.ProductTypeElectronics,
				pvzID:       pvzID,
			},
			setupMocks: func() {
				mockProductRepo.EXPECT().WithTx(gomock.Any()).Return(mockProductRepo)
				mockReceptionRepo.EXPECT().WithTx(gomock.Any()).Return(mockReceptionRepo)

				mockReceptionRepo.EXPECT().
					GetLastActiveByPVZID(gomock.Any(), pvzID).
					Return(activeReception, nil)

				mockProductRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.New("db error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			s := &ProductService{
				productRepo:   tt.fields.productRepo,
				receptionRepo: tt.fields.receptionRepo,
				txManager:     tt.fields.txManager,
			}

			got, err := s.AddProduct(tt.args.ctx, tt.args.productType, tt.args.pvzID)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddProduct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErrType != nil && !errors.Is(err, tt.expectedErrType) {
				t.Errorf("AddProduct() expected error type = %v, got = %v", tt.expectedErrType, err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddProduct() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProductService_GetProductByID(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := mocks.NewMockTxProductRepository(ctrl)
	mockReceptionRepo := mocks.NewMockTxReceptionRepository(ctrl)
	mockTxManager := &MockTxManager{}

	ctx := context.Background()
	productID := uuid.New()

	validProduct := &models.Product{
		ID:          productID,
		DateTime:    time.Now(),
		Type:        models.ProductTypeElectronics,
		ReceptionID: uuid.New(),
	}

	type fields struct {
		productRepo   interfaces.TxProductRepository
		receptionRepo interfaces.TxReceptionRepository
		txManager     postgres.TxManager
	}
	type args struct {
		ctx context.Context
		id  uuid.UUID
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		setupMocks      func()
		want            *models.Product
		wantErr         bool
		expectedErrType error
	}{
		{
			name: "successful receipt of product by ID",
			fields: fields{
				productRepo:   mockProductRepo,
				receptionRepo: mockReceptionRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx: ctx,
				id:  productID,
			},
			setupMocks: func() {
				mockProductRepo.EXPECT().
					GetByID(gomock.Any(), productID).
					Return(validProduct, nil)
			},
			want:    validProduct,
			wantErr: false,
		},
		{
			name: "error: product not found",
			fields: fields{
				productRepo:   mockProductRepo,
				receptionRepo: mockReceptionRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx: ctx,
				id:  uuid.New(),
			},
			setupMocks: func() {
				mockProductRepo.EXPECT().
					GetByID(gomock.Any(), gomock.Any()).
					Return(nil, repoerrors.ErrProductNotFound)
			},
			want:            nil,
			wantErr:         true,
			expectedErrType: repoerrors.ErrProductNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			s := &ProductService{
				productRepo:   tt.fields.productRepo,
				receptionRepo: tt.fields.receptionRepo,
				txManager:     tt.fields.txManager,
			}

			got, err := s.GetProductByID(tt.args.ctx, tt.args.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetProductByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErrType != nil && !errors.Is(err, tt.expectedErrType) {
				t.Errorf("GetProductByID() expected error type = %v, got = %v", tt.expectedErrType, err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetProductByID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProductService_GetProductsByReceptionID(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := mocks.NewMockTxProductRepository(ctrl)
	mockReceptionRepo := mocks.NewMockTxReceptionRepository(ctrl)
	mockTxManager := &MockTxManager{}

	ctx := context.Background()
	receptionID := uuid.New()

	products := []models.Product{
		{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			Type:        models.ProductTypeElectronics,
			ReceptionID: receptionID,
		},
		{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			Type:        models.ProductTypeClothes,
			ReceptionID: receptionID,
		},
	}

	type fields struct {
		productRepo   interfaces.TxProductRepository
		receptionRepo interfaces.TxReceptionRepository
		txManager     postgres.TxManager
	}
	type args struct {
		ctx         context.Context
		receptionID uuid.UUID
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		setupMocks      func()
		want            []models.Product
		wantErr         bool
		expectedErrType error
	}{
		{
			name: "successful receipt of products by receipt ID",
			fields: fields{
				productRepo:   mockProductRepo,
				receptionRepo: mockReceptionRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:         ctx,
				receptionID: receptionID,
			},
			setupMocks: func() {
				mockProductRepo.EXPECT().
					GetByReceptionID(gomock.Any(), receptionID).
					Return(products, nil)
			},
			want:    products,
			wantErr: false,
		},
		{
			name: "empty product list",
			fields: fields{
				productRepo:   mockProductRepo,
				receptionRepo: mockReceptionRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:         ctx,
				receptionID: uuid.New(),
			},
			setupMocks: func() {
				mockProductRepo.EXPECT().
					GetByReceptionID(gomock.Any(), gomock.Any()).
					Return([]models.Product{}, nil)
			},
			want:    []models.Product{},
			wantErr: false,
		},
		{
			name: "error while receiving products",
			fields: fields{
				productRepo:   mockProductRepo,
				receptionRepo: mockReceptionRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:         ctx,
				receptionID: receptionID,
			},
			setupMocks: func() {
				mockProductRepo.EXPECT().
					GetByReceptionID(gomock.Any(), receptionID).
					Return(nil, errors.New("db error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			s := &ProductService{
				productRepo:   tt.fields.productRepo,
				receptionRepo: tt.fields.receptionRepo,
				txManager:     tt.fields.txManager,
			}

			got, err := s.GetProductsByReceptionID(tt.args.ctx, tt.args.receptionID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetProductsByReceptionID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErrType != nil && !errors.Is(err, tt.expectedErrType) {
				t.Errorf("GetProductsByReceptionID() expected error type = %v, got = %v", tt.expectedErrType, err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetProductsByReceptionID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProductService_DeleteLastProduct(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := mocks.NewMockTxProductRepository(ctrl)
	mockReceptionRepo := mocks.NewMockTxReceptionRepository(ctrl)

	mockTxManager := &MockTxManager{
		RunTransactionFunc: func(ctx context.Context, fn func(*sql.Tx) error) error {
			return fn(nil)
		},
	}

	ctx := context.Background()
	pvzID := uuid.New()
	receptionID := uuid.New()

	activeReception := &models.Reception{
		ID:       receptionID,
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.ReceptionStatusInProgress,
	}

	closedReception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.ReceptionStatusClosed,
	}

	products := []models.Product{
		{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			Type:        models.ProductTypeElectronics,
			ReceptionID: receptionID,
		},
	}

	type fields struct {
		productRepo   interfaces.TxProductRepository
		receptionRepo interfaces.TxReceptionRepository
		txManager     postgres.TxManager
	}
	type args struct {
		ctx   context.Context
		pvzID uuid.UUID
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		setupMocks      func()
		wantErr         bool
		expectedErrType error
	}{
		{
			name: "successful removal of last product",
			fields: fields{
				productRepo:   mockProductRepo,
				receptionRepo: mockReceptionRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:   ctx,
				pvzID: pvzID,
			},
			setupMocks: func() {

				mockProductRepo.EXPECT().WithTx(gomock.Any()).Return(mockProductRepo)
				mockReceptionRepo.EXPECT().WithTx(gomock.Any()).Return(mockReceptionRepo)

				mockReceptionRepo.EXPECT().
					GetLastActiveByPVZID(gomock.Any(), pvzID).
					Return(activeReception, nil)

				mockProductRepo.EXPECT().
					GetByReceptionID(gomock.Any(), receptionID).
					Return(products, nil)

				mockProductRepo.EXPECT().
					DeleteLastFromReception(gomock.Any(), receptionID).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error: reception closed",
			fields: fields{
				productRepo:   mockProductRepo,
				receptionRepo: mockReceptionRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:   ctx,
				pvzID: pvzID,
			},
			setupMocks: func() {
				mockProductRepo.EXPECT().WithTx(gomock.Any()).Return(mockProductRepo)
				mockReceptionRepo.EXPECT().WithTx(gomock.Any()).Return(mockReceptionRepo)

				mockReceptionRepo.EXPECT().
					GetLastActiveByPVZID(gomock.Any(), pvzID).
					Return(closedReception, nil)
			},
			wantErr:         true,
			expectedErrType: apperrors.ErrReceptionCannotBeModified,
		},
		{
			name: "error: no products to remove",
			fields: fields{
				productRepo:   mockProductRepo,
				receptionRepo: mockReceptionRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:   ctx,
				pvzID: pvzID,
			},
			setupMocks: func() {
				mockProductRepo.EXPECT().WithTx(gomock.Any()).Return(mockProductRepo)
				mockReceptionRepo.EXPECT().WithTx(gomock.Any()).Return(mockReceptionRepo)

				mockReceptionRepo.EXPECT().
					GetLastActiveByPVZID(gomock.Any(), pvzID).
					Return(activeReception, nil)

				mockProductRepo.EXPECT().
					GetByReceptionID(gomock.Any(), receptionID).
					Return([]models.Product{}, nil)
			},
			wantErr:         true,
			expectedErrType: apperrors.ErrNoProductsToDelete,
		},
		{
			name: "Error deleting product",
			fields: fields{
				productRepo:   mockProductRepo,
				receptionRepo: mockReceptionRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:   ctx,
				pvzID: pvzID,
			},
			setupMocks: func() {
				mockProductRepo.EXPECT().WithTx(gomock.Any()).Return(mockProductRepo)
				mockReceptionRepo.EXPECT().WithTx(gomock.Any()).Return(mockReceptionRepo)

				mockReceptionRepo.EXPECT().
					GetLastActiveByPVZID(gomock.Any(), pvzID).
					Return(activeReception, nil)

				mockProductRepo.EXPECT().
					GetByReceptionID(gomock.Any(), receptionID).
					Return(products, nil)

				mockProductRepo.EXPECT().
					DeleteLastFromReception(gomock.Any(), receptionID).
					Return(errors.New("delete error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			s := &ProductService{
				productRepo:   tt.fields.productRepo,
				receptionRepo: tt.fields.receptionRepo,
				txManager:     tt.fields.txManager,
			}

			err := s.DeleteLastProduct(tt.args.ctx, tt.args.pvzID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteLastProduct() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.expectedErrType != nil && !errors.Is(err, tt.expectedErrType) {
				t.Errorf("DeleteLastProduct() expected error type = %v, got = %v", tt.expectedErrType, err)
			}
		})
	}
}

func TestNewProductService(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := mocks.NewMockTxProductRepository(ctrl)
	mockReceptionRepo := mocks.NewMockTxReceptionRepository(ctrl)
	mockTxManager := &MockTxManager{}

	type args struct {
		productRepo   interfaces.TxProductRepository
		receptionRepo interfaces.TxReceptionRepository
		txManager     postgres.TxManager
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "creation of product service",
			args: args{
				productRepo:   mockProductRepo,
				receptionRepo: mockReceptionRepo,
				txManager:     mockTxManager,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewProductService(tt.args.productRepo, tt.args.receptionRepo, tt.args.txManager)

			if got == nil {
				t.Errorf("NewProductService() returned nil")
			}

			if got.productRepo != tt.args.productRepo {
				t.Errorf("productRepo not initialized correctly")
			}
			if got.receptionRepo != tt.args.receptionRepo {
				t.Errorf("receptionRepo not initialized correctly")
			}
			if got.txManager != tt.args.txManager {
				t.Errorf("txManager not initialized correctly")
			}
		})
	}
}
