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
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestNewReceptionService(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReceptionRepo := mocks.NewMockTxReceptionRepository(ctrl)
	mockPVZRepo := mocks.NewMockTxPVZRepository(ctrl)
	mockTxManager := &MockTxManager{}

	type args struct {
		receptionRepo interfaces.TxReceptionRepository
		pvzRepo       interfaces.TxPVZRepository
		txManager     postgres.TxManager
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "создание сервиса приемок",
			args: args{
				receptionRepo: mockReceptionRepo,
				pvzRepo:       mockPVZRepo,
				txManager:     mockTxManager,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewReceptionService(tt.args.receptionRepo, tt.args.pvzRepo, tt.args.txManager)

			if got == nil {
				t.Errorf("NewReceptionService() returned nil")
			}

			if got.receptionRepo != tt.args.receptionRepo {
				t.Errorf("receptionRepo not initialized correctly")
			}
			if got.pvzRepo != tt.args.pvzRepo {
				t.Errorf("pvzRepo not initialized correctly")
			}
			if got.txManager != tt.args.txManager {
				t.Errorf("txManager not initialized correctly")
			}
		})
	}
}

func TestReceptionService_CreateReception(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReceptionRepo := mocks.NewMockTxReceptionRepository(ctrl)
	mockPVZRepo := mocks.NewMockTxPVZRepository(ctrl)

	mockTxManager := &MockTxManager{
		RunTransactionFunc: func(ctx context.Context, fn func(*sql.Tx) error) error {
			return fn(nil)
		},
	}

	ctx := context.Background()
	pvzID := uuid.New()
	receptionID := uuid.New()

	pvz := &models.PVZ{
		ID:               pvzID,
		RegistrationDate: time.Now(),
		City:             models.CityMoscow,
	}

	newReception := &models.Reception{
		ID:       receptionID,
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.ReceptionStatusInProgress,
	}

	type fields struct {
		receptionRepo interfaces.TxReceptionRepository
		pvzRepo       interfaces.TxPVZRepository
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
		want            *models.Reception
		wantErr         bool
		expectedErrType error
	}{
		{
			name: "успешное создание приемки",
			fields: fields{
				receptionRepo: mockReceptionRepo,
				pvzRepo:       mockPVZRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:   ctx,
				pvzID: pvzID,
			},
			setupMocks: func() {

				mockReceptionRepo.EXPECT().WithTx(gomock.Any()).Return(mockReceptionRepo)
				mockPVZRepo.EXPECT().WithTx(gomock.Any()).Return(mockPVZRepo)

				mockPVZRepo.EXPECT().
					GetByID(gomock.Any(), pvzID).
					Return(pvz, nil)

				mockReceptionRepo.EXPECT().
					GetLastActiveByPVZID(gomock.Any(), pvzID).
					Return(nil, apperrors.ErrNoActiveReception)

				mockReceptionRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, reception *models.Reception) error {

						newReception.ID = reception.ID
						newReception.DateTime = reception.DateTime
						return nil
					})
			},
			want:    newReception,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			s := &ReceptionService{
				receptionRepo: tt.fields.receptionRepo,
				pvzRepo:       tt.fields.pvzRepo,
				txManager:     tt.fields.txManager,
			}

			got, err := s.CreateReception(tt.args.ctx, tt.args.pvzID)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateReception() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErrType != nil && !errors.Is(err, tt.expectedErrType) {
				t.Errorf("CreateReception() expected error type = %v, got = %v", tt.expectedErrType, err)
			}

			if tt.want.Products == nil {
				tt.want.Products = []models.Product{}
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReceptionService_CloseReception(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReceptionRepo := mocks.NewMockTxReceptionRepository(ctrl)
	mockPVZRepo := mocks.NewMockTxPVZRepository(ctrl)

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
		ID:       receptionID,
		DateTime: activeReception.DateTime,
		PVZID:    pvzID,
		Status:   models.ReceptionStatusClosed,
	}

	type fields struct {
		receptionRepo interfaces.TxReceptionRepository
		pvzRepo       interfaces.TxPVZRepository
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
		want            *models.Reception
		wantErr         bool
		expectedErrType error
	}{
		{
			name: "успешное закрытие приемки",
			fields: fields{
				receptionRepo: mockReceptionRepo,
				pvzRepo:       mockPVZRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:   ctx,
				pvzID: pvzID,
			},
			setupMocks: func() {

				mockReceptionRepo.EXPECT().WithTx(gomock.Any()).Return(mockReceptionRepo)

				mockReceptionRepo.EXPECT().
					GetLastActiveByPVZID(gomock.Any(), pvzID).
					Return(activeReception, nil)

				mockReceptionRepo.EXPECT().
					CloseReception(gomock.Any(), receptionID).
					Return(nil)

				mockReceptionRepo.EXPECT().
					GetByID(gomock.Any(), receptionID).
					Return(closedReception, nil)
			},
			want:    closedReception,
			wantErr: false,
		},
		{
			name: "ошибка: нет активной приемки",
			fields: fields{
				receptionRepo: mockReceptionRepo,
				pvzRepo:       mockPVZRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:   ctx,
				pvzID: uuid.New(),
			},
			setupMocks: func() {
				mockReceptionRepo.EXPECT().WithTx(gomock.Any()).Return(mockReceptionRepo)

				mockReceptionRepo.EXPECT().
					GetLastActiveByPVZID(gomock.Any(), gomock.Any()).
					Return(nil, apperrors.ErrNoActiveReception)
			},
			want:            nil,
			wantErr:         true,
			expectedErrType: apperrors.ErrNoActiveReception,
		},
		{
			name: "ошибка при закрытии приемки",
			fields: fields{
				receptionRepo: mockReceptionRepo,
				pvzRepo:       mockPVZRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:   ctx,
				pvzID: pvzID,
			},
			setupMocks: func() {
				mockReceptionRepo.EXPECT().WithTx(gomock.Any()).Return(mockReceptionRepo)

				mockReceptionRepo.EXPECT().
					GetLastActiveByPVZID(gomock.Any(), pvzID).
					Return(activeReception, nil)

				mockReceptionRepo.EXPECT().
					CloseReception(gomock.Any(), receptionID).
					Return(errors.New("ошибка базы данных"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ошибка при получении закрытой приемки",
			fields: fields{
				receptionRepo: mockReceptionRepo,
				pvzRepo:       mockPVZRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:   ctx,
				pvzID: pvzID,
			},
			setupMocks: func() {
				mockReceptionRepo.EXPECT().WithTx(gomock.Any()).Return(mockReceptionRepo)

				mockReceptionRepo.EXPECT().
					GetLastActiveByPVZID(gomock.Any(), pvzID).
					Return(activeReception, nil)

				mockReceptionRepo.EXPECT().
					CloseReception(gomock.Any(), receptionID).
					Return(nil)

				mockReceptionRepo.EXPECT().
					GetByID(gomock.Any(), receptionID).
					Return(nil, errors.New("ошибка базы данных"))
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

			s := &ReceptionService{
				receptionRepo: tt.fields.receptionRepo,
				pvzRepo:       tt.fields.pvzRepo,
				txManager:     tt.fields.txManager,
			}

			got, err := s.CloseReception(tt.args.ctx, tt.args.pvzID)

			if (err != nil) != tt.wantErr {
				t.Errorf("CloseReception() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErrType != nil && !errors.Is(err, tt.expectedErrType) {
				t.Errorf("CloseReception() expected error type = %v, got = %v", tt.expectedErrType, err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CloseReception() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReceptionService_GetReceptionByID(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReceptionRepo := mocks.NewMockTxReceptionRepository(ctrl)
	mockPVZRepo := mocks.NewMockTxPVZRepository(ctrl)
	mockTxManager := &MockTxManager{}

	ctx := context.Background()
	receptionID := uuid.New()
	nonExistentID := uuid.New()

	reception := &models.Reception{
		ID:       receptionID,
		DateTime: time.Now(),
		PVZID:    uuid.New(),
		Status:   models.ReceptionStatusInProgress,
	}

	type fields struct {
		receptionRepo interfaces.TxReceptionRepository
		pvzRepo       interfaces.TxPVZRepository
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
		want            *models.Reception
		wantErr         bool
		expectedErrType error
	}{
		{
			name: "успешное получение приемки по ID",
			fields: fields{
				receptionRepo: mockReceptionRepo,
				pvzRepo:       mockPVZRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx: ctx,
				id:  receptionID,
			},
			setupMocks: func() {
				mockReceptionRepo.EXPECT().
					GetByID(gomock.Any(), receptionID).
					Return(reception, nil)
			},
			want:    reception,
			wantErr: false,
		},
		{
			name: "ошибка: приемка не найдена",
			fields: fields{
				receptionRepo: mockReceptionRepo,
				pvzRepo:       mockPVZRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx: ctx,
				id:  nonExistentID,
			},
			setupMocks: func() {
				mockReceptionRepo.EXPECT().
					GetByID(gomock.Any(), nonExistentID).
					Return(nil, repoerrors.ErrReceptionNotFound)
			},
			want:            nil,
			wantErr:         true,
			expectedErrType: repoerrors.ErrReceptionNotFound,
		},
		{
			name: "ошибка базы данных",
			fields: fields{
				receptionRepo: mockReceptionRepo,
				pvzRepo:       mockPVZRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx: ctx,
				id:  receptionID,
			},
			setupMocks: func() {
				mockReceptionRepo.EXPECT().
					GetByID(gomock.Any(), receptionID).
					Return(nil, errors.New("ошибка базы данных"))
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

			s := &ReceptionService{
				receptionRepo: tt.fields.receptionRepo,
				pvzRepo:       tt.fields.pvzRepo,
				txManager:     tt.fields.txManager,
			}

			got, err := s.GetReceptionByID(tt.args.ctx, tt.args.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetReceptionByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErrType != nil && !errors.Is(err, tt.expectedErrType) {
				t.Errorf("GetReceptionByID() expected error type = %v, got = %v", tt.expectedErrType, err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetReceptionByID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReceptionService_GetLastActiveReception(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReceptionRepo := mocks.NewMockTxReceptionRepository(ctrl)
	mockPVZRepo := mocks.NewMockTxPVZRepository(ctrl)
	mockTxManager := &MockTxManager{}

	ctx := context.Background()
	pvzID := uuid.New()
	nonExistentPVZID := uuid.New()

	activeReception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.ReceptionStatusInProgress,
	}

	type fields struct {
		receptionRepo interfaces.TxReceptionRepository
		pvzRepo       interfaces.TxPVZRepository
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
		want            *models.Reception
		wantErr         bool
		expectedErrType error
	}{
		{
			name: "успешное получение активной приемки",
			fields: fields{
				receptionRepo: mockReceptionRepo,
				pvzRepo:       mockPVZRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:   ctx,
				pvzID: pvzID,
			},
			setupMocks: func() {
				mockReceptionRepo.EXPECT().
					GetLastActiveByPVZID(gomock.Any(), pvzID).
					Return(activeReception, nil)
			},
			want:    activeReception,
			wantErr: false,
		},
		{
			name: "ошибка: нет активной приемки",
			fields: fields{
				receptionRepo: mockReceptionRepo,
				pvzRepo:       mockPVZRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:   ctx,
				pvzID: nonExistentPVZID,
			},
			setupMocks: func() {
				mockReceptionRepo.EXPECT().
					GetLastActiveByPVZID(gomock.Any(), nonExistentPVZID).
					Return(nil, apperrors.ErrNoActiveReception)
			},
			want:            nil,
			wantErr:         true,
			expectedErrType: apperrors.ErrNoActiveReception,
		},
		{
			name: "ошибка базы данных",
			fields: fields{
				receptionRepo: mockReceptionRepo,
				pvzRepo:       mockPVZRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:   ctx,
				pvzID: pvzID,
			},
			setupMocks: func() {
				mockReceptionRepo.EXPECT().
					GetLastActiveByPVZID(gomock.Any(), pvzID).
					Return(nil, errors.New("ошибка базы данных"))
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

			s := &ReceptionService{
				receptionRepo: tt.fields.receptionRepo,
				pvzRepo:       tt.fields.pvzRepo,
				txManager:     tt.fields.txManager,
			}

			got, err := s.GetLastActiveReception(tt.args.ctx, tt.args.pvzID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLastActiveReception() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErrType != nil && !errors.Is(err, tt.expectedErrType) {
				t.Errorf("GetLastActiveReception() expected error type = %v, got = %v", tt.expectedErrType, err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLastActiveReception() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReceptionService_GetLastReception(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReceptionRepo := mocks.NewMockTxReceptionRepository(ctrl)
	mockPVZRepo := mocks.NewMockTxPVZRepository(ctrl)
	mockTxManager := &MockTxManager{}

	ctx := context.Background()
	pvzID := uuid.New()
	nonExistentPVZID := uuid.New()

	lastReception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.ReceptionStatusClosed,
	}

	successCall := mockReceptionRepo.EXPECT().
		GetLastReceptionByPVZID(gomock.Any(), pvzID).
		Return(lastReception, nil)

	notFoundCall := mockReceptionRepo.EXPECT().
		GetLastReceptionByPVZID(gomock.Any(), nonExistentPVZID).
		Return(nil, repoerrors.ErrReceptionNotFound)

	errorCall := mockReceptionRepo.EXPECT().
		GetLastReceptionByPVZID(gomock.Any(), pvzID).
		Return(nil, errors.New("ошибка базы данных"))

	type fields struct {
		receptionRepo interfaces.TxReceptionRepository
		pvzRepo       interfaces.TxPVZRepository
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
		mockCall        *gomock.Call
		want            *models.Reception
		wantErr         bool
		expectedErrType error
	}{
		{
			name: "успешное получение последней приемки",
			fields: fields{
				receptionRepo: mockReceptionRepo,
				pvzRepo:       mockPVZRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:   ctx,
				pvzID: pvzID,
			},
			mockCall: successCall,
			want:     lastReception,
			wantErr:  false,
		},
		{
			name: "ошибка: приемки не найдены",
			fields: fields{
				receptionRepo: mockReceptionRepo,
				pvzRepo:       mockPVZRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:   ctx,
				pvzID: nonExistentPVZID,
			},
			mockCall:        notFoundCall,
			want:            nil,
			wantErr:         true,
			expectedErrType: repoerrors.ErrReceptionNotFound,
		},
		{
			name: "ошибка базы данных",
			fields: fields{
				receptionRepo: mockReceptionRepo,
				pvzRepo:       mockPVZRepo,
				txManager:     mockTxManager,
			},
			args: args{
				ctx:   ctx,
				pvzID: pvzID,
			},
			mockCall: errorCall,
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ReceptionService{
				receptionRepo: tt.fields.receptionRepo,
				pvzRepo:       tt.fields.pvzRepo,
				txManager:     tt.fields.txManager,
			}

			got, err := s.GetLastReception(tt.args.ctx, tt.args.pvzID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLastReception() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErrType != nil && !errors.Is(err, tt.expectedErrType) {
				t.Errorf("GetLastReception() expected error type = %v, got = %v", tt.expectedErrType, err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLastReception() got = %v, want %v", got, tt.want)
			}
		})
	}
}
