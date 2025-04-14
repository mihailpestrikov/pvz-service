package services

import (
	"avito-backend-trainee-assignment-spring-2025/internal/domain/apperrors"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/interfaces"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
	"avito-backend-trainee-assignment-spring-2025/internal/repository/postgres"
	"avito-backend-trainee-assignment-spring-2025/internal/repository/repoerrors"
	"avito-backend-trainee-assignment-spring-2025/internal/services/mocks"
	"avito-backend-trainee-assignment-spring-2025/pkg/config"
	"avito-backend-trainee-assignment-spring-2025/pkg/hasher"
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

func TestNewUserService(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockTxUserRepository(ctrl)

	jwtConfig := config.JWTConfig{
		Secret:     "test-secret",
		Expiration: 24 * time.Hour,
	}

	mockTxManager := &MockTxManager{}

	type args struct {
		repo      interfaces.TxUserRepository
		jwtConfig config.JWTConfig
		txManager postgres.TxManager
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "создание сервиса пользователей",
			args: args{
				repo:      mockUserRepo,
				jwtConfig: jwtConfig,
				txManager: mockTxManager,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewUserService(tt.args.repo, tt.args.jwtConfig, tt.args.txManager)

			if got == nil {
				t.Errorf("NewUserService() returned nil")
			}

			if got.repo != tt.args.repo {
				t.Errorf("repo not initialized correctly")
			}
			if got.jwtConfig != tt.args.jwtConfig {
				t.Errorf("jwtConfig not initialized correctly")
			}
			if got.txManager != tt.args.txManager {
				t.Errorf("txManager not initialized correctly")
			}
		})
	}
}

func TestUserService_Register(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockTxUserRepository(ctrl)

	jwtConfig := config.JWTConfig{
		Secret:     "test-secret",
		Expiration: 24 * time.Hour,
	}

	mockTxManager := &MockTxManager{
		RunTransactionFunc: func(ctx context.Context, fn func(*sql.Tx) error) error {
			return fn(nil)
		},
	}

	ctx := context.Background()
	validEmail := "test@example.com"
	existingEmail := "existing@example.com"
	validPassword := "password123"
	validRole := models.RoleEmployee

	userID := uuid.New()
	passwordHash, _ := hasher.Hash(validPassword)

	validUser := &models.User{
		ID:           userID,
		Email:        validEmail,
		PasswordHash: passwordHash,
		Role:         validRole,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	existingUser := &models.User{
		ID:           uuid.New(),
		Email:        existingEmail,
		PasswordHash: passwordHash,
		Role:         validRole,
	}

	type fields struct {
		repo      interfaces.TxUserRepository
		jwtConfig config.JWTConfig
		txManager postgres.TxManager
	}
	type args struct {
		ctx      context.Context
		email    string
		password string
		role     string
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		setupMocks      func()
		want            *models.User
		wantErr         bool
		expectedErrType error
	}{
		{
			name: "успешная регистрация пользователя",
			fields: fields{
				repo:      mockUserRepo,
				jwtConfig: jwtConfig,
				txManager: mockTxManager,
			},
			args: args{
				ctx:      ctx,
				email:    validEmail,
				password: validPassword,
				role:     validRole,
			},
			setupMocks: func() {

				mockUserRepo.EXPECT().WithTx(gomock.Any()).Return(mockUserRepo)

				mockUserRepo.EXPECT().
					GetByEmail(gomock.Any(), validEmail).
					Return(nil, repoerrors.ErrUserNotFound)

				mockUserRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, user *models.User) error {
						user.ID = validUser.ID
						user.CreatedAt = validUser.CreatedAt
						user.UpdatedAt = validUser.UpdatedAt
						return nil
					})
			},
			want: &models.User{
				ID:           validUser.ID,
				Email:        validEmail,
				PasswordHash: "",
				Role:         validRole,
				CreatedAt:    validUser.CreatedAt,
				UpdatedAt:    validUser.UpdatedAt,
			},
			wantErr: false,
		},
		{
			name: "ошибка: пользователь уже существует",
			fields: fields{
				repo:      mockUserRepo,
				jwtConfig: jwtConfig,
				txManager: mockTxManager,
			},
			args: args{
				ctx:      ctx,
				email:    existingEmail,
				password: validPassword,
				role:     validRole,
			},
			setupMocks: func() {
				mockUserRepo.EXPECT().WithTx(gomock.Any()).Return(mockUserRepo)

				mockUserRepo.EXPECT().
					GetByEmail(gomock.Any(), existingEmail).
					Return(existingUser, nil)
			},
			want:            nil,
			wantErr:         true,
			expectedErrType: repoerrors.ErrUserAlreadyExists,
		},
		{
			name: "ошибка при создании пользователя",
			fields: fields{
				repo:      mockUserRepo,
				jwtConfig: jwtConfig,
				txManager: mockTxManager,
			},
			args: args{
				ctx:      ctx,
				email:    validEmail,
				password: validPassword,
				role:     validRole,
			},
			setupMocks: func() {
				mockUserRepo.EXPECT().WithTx(gomock.Any()).Return(mockUserRepo)

				mockUserRepo.EXPECT().
					GetByEmail(gomock.Any(), validEmail).
					Return(nil, repoerrors.ErrUserNotFound)

				mockUserRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.New("ошибка базы данных"))
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

			s := &UserService{
				repo:      tt.fields.repo,
				jwtConfig: tt.fields.jwtConfig,
				txManager: tt.fields.txManager,
			}

			got, err := s.Register(tt.args.ctx, tt.args.email, tt.args.password, tt.args.role)

			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErrType != nil && !errors.Is(err, tt.expectedErrType) {
				t.Errorf("Register() expected error type = %v, got = %v", tt.expectedErrType, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserService_Login(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockTxUserRepository(ctrl)

	jwtConfig := config.JWTConfig{
		Secret:     "test-secret",
		Expiration: 24 * time.Hour,
	}

	mockTxManager := &MockTxManager{}

	ctx := context.Background()
	validEmail := "test@example.com"
	validPassword := "password123"
	wrongPassword := "wrong123"
	nonexistentEmail := "nonexistent@example.com"

	userID := uuid.New()
	passwordHash, _ := hasher.Hash(validPassword)

	user := &models.User{
		ID:           userID,
		Email:        validEmail,
		PasswordHash: passwordHash,
		Role:         models.RoleEmployee,
	}

	tokenStub := "jwt-token-stub"

	type fields struct {
		repo      interfaces.TxUserRepository
		jwtConfig config.JWTConfig
		txManager postgres.TxManager
	}
	type args struct {
		ctx      context.Context
		email    string
		password string
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		setupMocks      func()
		want            string
		wantErr         bool
		expectedErrType error
	}{
		{
			name: "успешный вход",
			fields: fields{
				repo:      mockUserRepo,
				jwtConfig: jwtConfig,
				txManager: mockTxManager,
			},
			args: args{
				ctx:      ctx,
				email:    validEmail,
				password: validPassword,
			},
			setupMocks: func() {

				mockUserRepo.EXPECT().
					GetByEmail(gomock.Any(), validEmail).
					Return(user, nil)

			},
			want:    tokenStub,
			wantErr: false,
		},
		{
			name: "ошибка: пользователь не найден",
			fields: fields{
				repo:      mockUserRepo,
				jwtConfig: jwtConfig,
				txManager: mockTxManager,
			},
			args: args{
				ctx:      ctx,
				email:    nonexistentEmail,
				password: validPassword,
			},
			setupMocks: func() {
				mockUserRepo.EXPECT().
					GetByEmail(gomock.Any(), nonexistentEmail).
					Return(nil, repoerrors.ErrUserNotFound)
			},
			want:            "",
			wantErr:         true,
			expectedErrType: apperrors.ErrInvalidCredentials,
		},
		{
			name: "ошибка: неверный пароль",
			fields: fields{
				repo:      mockUserRepo,
				jwtConfig: jwtConfig,
				txManager: mockTxManager,
			},
			args: args{
				ctx:      ctx,
				email:    validEmail,
				password: wrongPassword,
			},
			setupMocks: func() {
				mockUserRepo.EXPECT().
					GetByEmail(gomock.Any(), validEmail).
					Return(user, nil)
			},
			want:            "",
			wantErr:         true,
			expectedErrType: apperrors.ErrInvalidCredentials,
		},
		{
			name: "ошибка базы данных",
			fields: fields{
				repo:      mockUserRepo,
				jwtConfig: jwtConfig,
				txManager: mockTxManager,
			},
			args: args{
				ctx:      ctx,
				email:    validEmail,
				password: validPassword,
			},
			setupMocks: func() {
				mockUserRepo.EXPECT().
					GetByEmail(gomock.Any(), validEmail).
					Return(nil, errors.New("ошибка базы данных"))
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			s := &UserService{
				repo:      tt.fields.repo,
				jwtConfig: tt.fields.jwtConfig,
				txManager: tt.fields.txManager,
			}

			got, err := s.Login(tt.args.ctx, tt.args.email, tt.args.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErrType != nil && !errors.Is(err, tt.expectedErrType) {
				t.Errorf("Login() expected error type = %v, got = %v", tt.expectedErrType, err)
			}

			if tt.name == "успешный вход" {

				if got == "" {
					t.Errorf("Login() вернул пустую строку, ожидался непустой JWT-токен")
				}
			} else if got != tt.want {

				t.Errorf("Login() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserService_DummyLogin(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockTxUserRepository(ctrl)

	jwtConfig := config.JWTConfig{
		Secret:     "test-secret",
		Expiration: 24 * time.Hour,
	}

	mockTxManager := &MockTxManager{}

	validRole := models.RoleEmployee
	invalidRole := "invalid-role"

	tokenStub := "dummy-jwt-token-stub"

	type fields struct {
		repo      interfaces.TxUserRepository
		jwtConfig config.JWTConfig
		txManager postgres.TxManager
	}
	type args struct {
		role string
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		want            string
		wantErr         bool
		expectedErrType error
	}{
		{
			name: "успешное создание dummy-токена",
			fields: fields{
				repo:      mockUserRepo,
				jwtConfig: jwtConfig,
				txManager: mockTxManager,
			},
			args: args{
				role: validRole,
			},
			want:    tokenStub,
			wantErr: false,
		},
		{
			name: "ошибка: неверная роль",
			fields: fields{
				repo:      mockUserRepo,
				jwtConfig: jwtConfig,
				txManager: mockTxManager,
			},
			args: args{
				role: invalidRole,
			},
			want:            "",
			wantErr:         true,
			expectedErrType: apperrors.ErrInvalidRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &UserService{
				repo:      tt.fields.repo,
				jwtConfig: tt.fields.jwtConfig,
				txManager: tt.fields.txManager,
			}

			got, err := s.DummyLogin(tt.args.role)

			if (err != nil) != tt.wantErr {
				t.Errorf("DummyLogin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErrType != nil && !errors.Is(err, tt.expectedErrType) {
				t.Errorf("DummyLogin() expected error type = %v, got = %v", tt.expectedErrType, err)
			}

			if tt.name == "успешное создание dummy-токена" {

				if got == "" {
					t.Errorf("DummyLogin() вернул пустую строку, ожидался непустой JWT-токен")
				}
			} else if got != tt.want {

				t.Errorf("DummyLogin() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserService_GetUserByID(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockTxUserRepository(ctrl)

	jwtConfig := config.JWTConfig{
		Secret:     "test-secret",
		Expiration: 24 * time.Hour,
	}

	mockTxManager := &MockTxManager{}

	ctx := context.Background()
	userID := uuid.New()
	nonExistentID := uuid.New()

	user := &models.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleEmployee,
	}

	userWithoutPassword := &models.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: "",
		Role:         models.RoleEmployee,
	}

	type fields struct {
		repo      interfaces.TxUserRepository
		jwtConfig config.JWTConfig
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
		want            *models.User
		wantErr         bool
		expectedErrType error
	}{
		{
			name: "успешное получение пользователя по ID",
			fields: fields{
				repo:      mockUserRepo,
				jwtConfig: jwtConfig,
				txManager: mockTxManager,
			},
			args: args{
				ctx: ctx,
				id:  userID,
			},
			setupMocks: func() {
				mockUserRepo.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(user, nil)
			},
			want:    userWithoutPassword,
			wantErr: false,
		},
		{
			name: "ошибка: пользователь не найден",
			fields: fields{
				repo:      mockUserRepo,
				jwtConfig: jwtConfig,
				txManager: mockTxManager,
			},
			args: args{
				ctx: ctx,
				id:  nonExistentID,
			},
			setupMocks: func() {
				mockUserRepo.EXPECT().
					GetByID(gomock.Any(), nonExistentID).
					Return(nil, repoerrors.ErrUserNotFound)
			},
			want:            nil,
			wantErr:         true,
			expectedErrType: repoerrors.ErrUserNotFound,
		},
		{
			name: "ошибка базы данных",
			fields: fields{
				repo:      mockUserRepo,
				jwtConfig: jwtConfig,
				txManager: mockTxManager,
			},
			args: args{
				ctx: ctx,
				id:  userID,
			},
			setupMocks: func() {
				mockUserRepo.EXPECT().
					GetByID(gomock.Any(), userID).
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

			s := &UserService{
				repo:      tt.fields.repo,
				jwtConfig: tt.fields.jwtConfig,
				txManager: tt.fields.txManager,
			}

			got, err := s.GetUserByID(tt.args.ctx, tt.args.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErrType != nil && !errors.Is(err, tt.expectedErrType) {
				t.Errorf("GetUserByID() expected error type = %v, got = %v", tt.expectedErrType, err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserByID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserService_DeleteUser(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockTxUserRepository(ctrl)

	jwtConfig := config.JWTConfig{
		Secret:     "test-secret",
		Expiration: 24 * time.Hour,
	}

	mockTxManager := &MockTxManager{
		RunTransactionFunc: func(ctx context.Context, fn func(*sql.Tx) error) error {
			return fn(nil)
		},
	}

	ctx := context.Background()
	userID := uuid.New()
	nonExistentID := uuid.New()

	user := &models.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleEmployee,
	}

	type fields struct {
		repo      interfaces.TxUserRepository
		jwtConfig config.JWTConfig
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
		wantErr         bool
		expectedErrType error
	}{
		{
			name: "успешное удаление пользователя",
			fields: fields{
				repo:      mockUserRepo,
				jwtConfig: jwtConfig,
				txManager: mockTxManager,
			},
			args: args{
				ctx: ctx,
				id:  userID,
			},
			setupMocks: func() {

				mockUserRepo.EXPECT().WithTx(gomock.Any()).Return(mockUserRepo)

				mockUserRepo.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(user, nil)

				mockUserRepo.EXPECT().
					Delete(gomock.Any(), userID).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ошибка: пользователь не найден",
			fields: fields{
				repo:      mockUserRepo,
				jwtConfig: jwtConfig,
				txManager: mockTxManager,
			},
			args: args{
				ctx: ctx,
				id:  nonExistentID,
			},
			setupMocks: func() {
				mockUserRepo.EXPECT().WithTx(gomock.Any()).Return(mockUserRepo)

				mockUserRepo.EXPECT().
					GetByID(gomock.Any(), nonExistentID).
					Return(nil, repoerrors.ErrUserNotFound)
			},
			wantErr:         true,
			expectedErrType: repoerrors.ErrUserNotFound,
		},
		{
			name: "ошибка при удалении пользователя",
			fields: fields{
				repo:      mockUserRepo,
				jwtConfig: jwtConfig,
				txManager: mockTxManager,
			},
			args: args{
				ctx: ctx,
				id:  userID,
			},
			setupMocks: func() {
				mockUserRepo.EXPECT().WithTx(gomock.Any()).Return(mockUserRepo)

				mockUserRepo.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(user, nil)

				mockUserRepo.EXPECT().
					Delete(gomock.Any(), userID).
					Return(errors.New("ошибка базы данных"))
			},
			wantErr: true,
		},
		{
			name: "ошибка транзакции",
			fields: fields{
				repo:      mockUserRepo,
				jwtConfig: jwtConfig,
				txManager: &MockTxManager{
					RunTransactionFunc: func(ctx context.Context, fn func(*sql.Tx) error) error {
						return errors.New("ошибка транзакции")
					},
				},
			},
			args: args{
				ctx: ctx,
				id:  userID,
			},
			setupMocks: func() {

			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			s := &UserService{
				repo:      tt.fields.repo,
				jwtConfig: tt.fields.jwtConfig,
				txManager: tt.fields.txManager,
			}

			err := s.DeleteUser(tt.args.ctx, tt.args.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErrType != nil && !errors.Is(err, tt.expectedErrType) {
				t.Errorf("DeleteUser() expected error type = %v, got = %v", tt.expectedErrType, err)
			}
		})
	}
}
