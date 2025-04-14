package postgres

import (
	"avito-backend-trainee-assignment-spring-2025/internal/domain/interfaces"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
	"avito-backend-trainee-assignment-spring-2025/internal/repository/repoerrors"
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func setupUserRepoMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *UserRepository) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}

	repo := &UserRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	return db, mock, repo
}

func TestNewUserRepository(t *testing.T) {
	db, _, _ := setupUserRepoMock(t)
	defer db.Close()

	repo := NewUserRepository(db)
	assert.NotNil(t, repo, "Repository should not be nil")
	assert.Implements(t, (*interfaces.TxUserRepository)(nil), repo)
}

func TestUserRepository_Create(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	email := "test@example.com"
	passwordHash := "hashed_password"

	tests := []struct {
		name        string
		user        *models.User
		mockSetup   func(sqlmock.Sqlmock)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful creation",
			user: &models.User{
				ID:           userID,
				Email:        email,
				PasswordHash: passwordHash,
				Role:         models.RoleEmployee,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users (id,email,password_hash,role) VALUES ($1,$2,$3,$4)").
					WithArgs(userID, email, passwordHash, models.RoleEmployee).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "duplicate email error",
			user: &models.User{
				ID:           userID,
				Email:        email,
				PasswordHash: passwordHash,
				Role:         models.RoleEmployee,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users (id,email,password_hash,role) VALUES ($1,$2,$3,$4)").
					WithArgs(userID, email, passwordHash, models.RoleEmployee).
					WillReturnError(errors.New(`pq: duplicate key value violates unique constraint "users_email_key"`))
			},
			wantErr:     true,
			expectedErr: repoerrors.ErrUserAlreadyExists,
		},
		{
			name: "database error",
			user: &models.User{
				ID:           userID,
				Email:        email,
				PasswordHash: passwordHash,
				Role:         models.RoleEmployee,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users (id,email,password_hash,role) VALUES ($1,$2,$3,$4)").
					WithArgs(userID, email, passwordHash, models.RoleEmployee).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupUserRepoMock(t)
			defer db.Close()

			tt.mockSetup(mock)

			err := repo.Create(context.Background(), tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	passwordHash := "hashed_password"
	role := models.RoleEmployee

	tests := []struct {
		name        string
		id          uuid.UUID
		mockSetup   func(sqlmock.Sqlmock)
		want        *models.User
		wantErr     bool
		expectedErr error
	}{
		{
			name: "user found",
			id:   userID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "role"}).
					AddRow(userID, email, passwordHash, role)

				mock.ExpectQuery(`SELECT id, email, password_hash, role FROM users WHERE id = $1`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			want: &models.User{
				ID:           userID,
				Email:        email,
				PasswordHash: passwordHash,
				Role:         role,
			},
			wantErr: false,
		},
		{
			name: "user not found",
			id:   userID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, email, password_hash, role FROM users WHERE id = $1`).
					WithArgs(userID).
					WillReturnError(sql.ErrNoRows)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: repoerrors.ErrUserNotFound,
		},
		{
			name: "database error",
			id:   userID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, email, password_hash, role FROM users WHERE id = $1`).
					WithArgs(userID).
					WillReturnError(errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupUserRepoMock(t)
			defer db.Close()

			tt.mockSetup(mock)

			got, err := repo.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.Email, got.Email)
				assert.Equal(t, tt.want.PasswordHash, got.PasswordHash)
				assert.Equal(t, tt.want.Role, got.Role)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	passwordHash := "hashed_password"
	role := models.RoleEmployee

	tests := []struct {
		name        string
		email       string
		mockSetup   func(sqlmock.Sqlmock)
		want        *models.User
		wantErr     bool
		expectedErr error
	}{
		{
			name:  "user found",
			email: email,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "role"}).
					AddRow(userID, email, passwordHash, role)

				mock.ExpectQuery(`SELECT id, email, password_hash, role FROM users WHERE email = $1`).
					WithArgs(email).
					WillReturnRows(rows)
			},
			want: &models.User{
				ID:           userID,
				Email:        email,
				PasswordHash: passwordHash,
				Role:         role,
			},
			wantErr: false,
		},
		{
			name:  "user not found",
			email: email,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, email, password_hash, role FROM users WHERE email = $1`).
					WithArgs(email).
					WillReturnError(sql.ErrNoRows)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: repoerrors.ErrUserNotFound,
		},
		{
			name:  "database error",
			email: email,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, email, password_hash, role FROM users WHERE email = $1`).
					WithArgs(email).
					WillReturnError(errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupUserRepoMock(t)
			defer db.Close()

			tt.mockSetup(mock)

			got, err := repo.GetByEmail(context.Background(), tt.email)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.Email, got.Email)
				assert.Equal(t, tt.want.PasswordHash, got.PasswordHash)
				assert.Equal(t, tt.want.Role, got.Role)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_Delete(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name        string
		id          uuid.UUID
		mockSetup   func(sqlmock.Sqlmock)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful deletion",
			id:   userID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM users WHERE id = $1`).
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupUserRepoMock(t)
			defer db.Close()

			tt.mockSetup(mock)

			err := repo.Delete(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_isDuplicateKeyError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "duplicate key error",
			err:  errors.New(`pq: duplicate key value violates unique constraint "users_email_key"`),
			want: true,
		},
		{
			name: "other error",
			err:  errors.New("database error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDuplicateKeyError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}
