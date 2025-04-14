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

func TestNewPVZService(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPVZRepo := mocks.NewMockTxPVZRepository(ctrl)
	mockTxManager := &MockTxManager{}

	type args struct {
		repo      interfaces.TxPVZRepository
		txManager postgres.TxManager
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "create PVZ service",
			args: args{
				repo:      mockPVZRepo,
				txManager: mockTxManager,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPVZService(tt.args.repo, tt.args.txManager)

			if got == nil {
				t.Errorf("NewPVZService() returned nil")
			}

			if got.repo != tt.args.repo {
				t.Errorf("repo not initialized correctly")
			}
			if got.txManager != tt.args.txManager {
				t.Errorf("txManager not initialized correctly")
			}
		})
	}
}

func TestPVZService_CreatePVZ(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPVZRepo := mocks.NewMockTxPVZRepository(ctrl)

	mockTxManager := &MockTxManager{
		RunTransactionFunc: func(ctx context.Context, fn func(*sql.Tx) error) error {
			return fn(nil)
		},
	}

	ctx := context.Background()
	validCity := models.CityMoscow
	invalidCity := "Novosibirsk"

	pvzID := uuid.New()
	validPVZ := &models.PVZ{
		ID:               pvzID,
		RegistrationDate: time.Now(),
		City:             validCity,
	}

	type fields struct {
		repo      interfaces.TxPVZRepository
		txManager postgres.TxManager
	}
	type args struct {
		ctx  context.Context
		city string
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		setupMocks      func()
		want            *models.PVZ
		wantErr         bool
		expectedErrType error
	}{
		{
			name: "successful PVZ creation",
			fields: fields{
				repo:      mockPVZRepo,
				txManager: mockTxManager,
			},
			args: args{
				ctx:  ctx,
				city: validCity,
			},
			setupMocks: func() {

				mockPVZRepo.EXPECT().WithTx(gomock.Any()).Return(mockPVZRepo)

				mockPVZRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, pvz *models.PVZ) error {

						validPVZ.ID = pvz.ID
						validPVZ.RegistrationDate = pvz.RegistrationDate
						return nil
					})
			},
			want:    validPVZ,
			wantErr: false,
		},
		{
			name: "error: invalid city",
			fields: fields{
				repo:      mockPVZRepo,
				txManager: mockTxManager,
			},
			args: args{
				ctx:  ctx,
				city: invalidCity,
			},
			setupMocks: func() {

			},
			want:            nil,
			wantErr:         true,
			expectedErrType: apperrors.ErrInvalidCity,
		},
		{
			name: "error creating PVZ",
			fields: fields{
				repo:      mockPVZRepo,
				txManager: mockTxManager,
			},
			args: args{
				ctx:  ctx,
				city: validCity,
			},
			setupMocks: func() {
				mockPVZRepo.EXPECT().WithTx(gomock.Any()).Return(mockPVZRepo)

				mockPVZRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error: PVZ already exists",
			fields: fields{
				repo:      mockPVZRepo,
				txManager: mockTxManager,
			},
			args: args{
				ctx:  ctx,
				city: validCity,
			},
			setupMocks: func() {
				mockPVZRepo.EXPECT().WithTx(gomock.Any()).Return(mockPVZRepo)

				mockPVZRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(repoerrors.ErrPVZAlreadyExists)
			},
			want:            nil,
			wantErr:         true,
			expectedErrType: repoerrors.ErrPVZAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			s := &PVZService{
				repo:      tt.fields.repo,
				txManager: tt.fields.txManager,
			}

			got, err := s.CreatePVZ(tt.args.ctx, tt.args.city)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePVZ() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErrType != nil && !errors.Is(err, tt.expectedErrType) {
				t.Errorf("CreatePVZ() expected error type = %v, got = %v", tt.expectedErrType, err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreatePVZ() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPVZService_GetPVZByID(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPVZRepo := mocks.NewMockTxPVZRepository(ctrl)
	mockTxManager := &MockTxManager{}

	ctx := context.Background()
	pvzID := uuid.New()
	nonExistentID := uuid.New()

	validPVZ := &models.PVZ{
		ID:               pvzID,
		RegistrationDate: time.Now(),
		City:             models.CityMoscow,
	}

	type fields struct {
		repo      interfaces.TxPVZRepository
		txManager postgres.TxManager
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
		want            *models.PVZ
		wantErr         bool
		expectedErrType error
	}{
		{
			name: "successful retrieval of PVZ by ID",
			fields: fields{
				repo:      mockPVZRepo,
				txManager: mockTxManager,
			},
			args: args{
				ctx: ctx,
				id:  pvzID,
			},
			setupMocks: func() {
				mockPVZRepo.EXPECT().
					GetByID(gomock.Any(), pvzID).
					Return(validPVZ, nil)
			},
			want:    validPVZ,
			wantErr: false,
		},
		{
			name: "error: PVZ not found",
			fields: fields{
				repo:      mockPVZRepo,
				txManager: mockTxManager,
			},
			args: args{
				ctx: ctx,
				id:  nonExistentID,
			},
			setupMocks: func() {
				mockPVZRepo.EXPECT().
					GetByID(gomock.Any(), nonExistentID).
					Return(nil, repoerrors.ErrPVZNotFound)
			},
			want:            nil,
			wantErr:         true,
			expectedErrType: repoerrors.ErrPVZNotFound,
		},
		{
			name: "database error",
			fields: fields{
				repo:      mockPVZRepo,
				txManager: mockTxManager,
			},
			args: args{
				ctx: ctx,
				id:  pvzID,
			},
			setupMocks: func() {
				mockPVZRepo.EXPECT().
					GetByID(gomock.Any(), pvzID).
					Return(nil, errors.New("database error"))
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

			s := &PVZService{
				repo:      tt.fields.repo,
				txManager: tt.fields.txManager,
			}

			got, err := s.GetPVZByID(tt.args.ctx, tt.args.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPVZByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErrType != nil && !errors.Is(err, tt.expectedErrType) {
				t.Errorf("GetPVZByID() expected error type = %v, got = %v", tt.expectedErrType, err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPVZByID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPVZService_GetAllPVZ(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPVZRepo := mocks.NewMockTxPVZRepository(ctrl)
	mockTxManager := &MockTxManager{}

	ctx := context.Background()
	filter := models.PVZFilter{
		Page:  1,
		Limit: 10,
	}

	pvzList := []*models.PVZ{
		{
			ID:               uuid.New(),
			RegistrationDate: time.Now().Add(-48 * time.Hour),
			City:             models.CityMoscow,
		},
		{
			ID:               uuid.New(),
			RegistrationDate: time.Now().Add(-24 * time.Hour),
			City:             models.CitySaintPete,
		},
	}

	type fields struct {
		repo      interfaces.TxPVZRepository
		txManager postgres.TxManager
	}
	type args struct {
		ctx    context.Context
		filter models.PVZFilter
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		setupMocks      func()
		want            []*models.PVZ
		want1           int
		wantErr         bool
		expectedErrType error
	}{
		{
			name: "successful retrieval of PVZ list",
			fields: fields{
				repo:      mockPVZRepo,
				txManager: mockTxManager,
			},
			args: args{
				ctx:    ctx,
				filter: filter,
			},
			setupMocks: func() {
				mockPVZRepo.EXPECT().
					GetAll(gomock.Any(), filter).
					Return(pvzList, len(pvzList), nil)
			},
			want:    pvzList,
			want1:   len(pvzList),
			wantErr: false,
		},
		{
			name: "empty PVZ list",
			fields: fields{
				repo:      mockPVZRepo,
				txManager: mockTxManager,
			},
			args: args{
				ctx:    ctx,
				filter: filter,
			},
			setupMocks: func() {
				mockPVZRepo.EXPECT().
					GetAll(gomock.Any(), filter).
					Return([]*models.PVZ{}, 0, nil)
			},
			want:    []*models.PVZ{},
			want1:   0,
			wantErr: false,
		},
		{
			name: "error retrieving PVZ list",
			fields: fields{
				repo:      mockPVZRepo,
				txManager: mockTxManager,
			},
			args: args{
				ctx:    ctx,
				filter: filter,
			},
			setupMocks: func() {
				mockPVZRepo.EXPECT().
					GetAll(gomock.Any(), filter).
					Return(nil, 0, errors.New("database error"))
			},
			want:    nil,
			want1:   0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			s := &PVZService{
				repo:      tt.fields.repo,
				txManager: tt.fields.txManager,
			}

			got, got1, err := s.GetAllPVZ(tt.args.ctx, tt.args.filter)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllPVZ() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErrType != nil && !errors.Is(err, tt.expectedErrType) {
				t.Errorf("GetAllPVZ() expected error type = %v, got = %v", tt.expectedErrType, err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllPVZ() got = %v, want %v", got, tt.want)
			}

			if got1 != tt.want1 {
				t.Errorf("GetAllPVZ() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestPVZService_GetAllPVZWithReceptions(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPVZRepo := mocks.NewMockTxPVZRepository(ctrl)
	mockTxManager := &MockTxManager{}

	ctx := context.Background()
	filter := models.PVZFilter{
		Page:  1,
		Limit: 10,
	}

	pvzID := uuid.New()
	receptionID := uuid.New()

	pvz := &models.PVZ{
		ID:               pvzID,
		RegistrationDate: time.Now(),
		City:             models.CityMoscow,
	}

	reception := &models.Reception{
		ID:       receptionID,
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.ReceptionStatusInProgress,
	}

	pvzWithReceptions := []models.PVZWithReceptions{
		{
			PVZ:        pvz,
			Receptions: []*models.Reception{reception},
		},
	}

	type fields struct {
		repo      interfaces.TxPVZRepository
		txManager postgres.TxManager
	}
	type args struct {
		ctx    context.Context
		filter models.PVZFilter
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		setupMocks      func()
		want            []models.PVZWithReceptions
		want1           int
		wantErr         bool
		expectedErrType error
	}{
		{
			name: "successful retrieval of PVZ with receptions",
			fields: fields{
				repo:      mockPVZRepo,
				txManager: mockTxManager,
			},
			args: args{
				ctx:    ctx,
				filter: filter,
			},
			setupMocks: func() {
				mockPVZRepo.EXPECT().
					GetAllWithReceptions(gomock.Any(), filter).
					Return(pvzWithReceptions, 1, nil)
			},
			want:    pvzWithReceptions,
			want1:   1,
			wantErr: false,
		},
		{
			name: "empty PVZ list with receptions",
			fields: fields{
				repo:      mockPVZRepo,
				txManager: mockTxManager,
			},
			args: args{
				ctx:    ctx,
				filter: filter,
			},
			setupMocks: func() {
				mockPVZRepo.EXPECT().
					GetAllWithReceptions(gomock.Any(), filter).
					Return([]models.PVZWithReceptions{}, 0, nil)
			},
			want:    []models.PVZWithReceptions{},
			want1:   0,
			wantErr: false,
		},
		{
			name: "error retrieving PVZ with receptions",
			fields: fields{
				repo:      mockPVZRepo,
				txManager: mockTxManager,
			},
			args: args{
				ctx:    ctx,
				filter: filter,
			},
			setupMocks: func() {
				mockPVZRepo.EXPECT().
					GetAllWithReceptions(gomock.Any(), filter).
					Return(nil, 0, errors.New("database error"))
			},
			want:    nil,
			want1:   0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			s := &PVZService{
				repo:      tt.fields.repo,
				txManager: tt.fields.txManager,
			}

			got, got1, err := s.GetAllPVZWithReceptions(tt.args.ctx, tt.args.filter)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllPVZWithReceptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErrType != nil && !errors.Is(err, tt.expectedErrType) {
				t.Errorf("GetAllPVZWithReceptions() expected error type = %v, got = %v", tt.expectedErrType, err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllPVZWithReceptions() got = %v, want %v", got, tt.want)
			}

			if got1 != tt.want1 {
				t.Errorf("GetAllPVZWithReceptions() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
