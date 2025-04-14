package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func setupTxManagerMock(t *testing.T) (*DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}

	dbWrapper := &DB{
		DB: db,
	}

	return dbWrapper, mock
}

func TestNewTxManager(t *testing.T) {
	dbWrapper, _ := setupTxManagerMock(t)
	defer dbWrapper.Close()

	txManager := NewTxManager(dbWrapper)

	assert.NotNil(t, txManager, "Transaction manager should not be nil")
	assert.Equal(t, dbWrapper, txManager.db, "DB instance should be set correctly")
}

func TestDBTxManager_RunTransaction(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(sqlmock.Sqlmock)
		txFunc      func(*sql.Tx) error
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful transaction",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectCommit()
			},
			txFunc: func(tx *sql.Tx) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "transaction function returns error",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			txFunc: func(tx *sql.Tx) error {
				return errors.New("transaction function error")
			},
			wantErr:     true,
			expectedErr: errors.New("transaction function error"),
		},
		{
			name: "begin transaction fails",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(errors.New("begin error"))
			},
			txFunc: func(tx *sql.Tx) error {
				return nil
			},
			wantErr: true,
		},
		{
			name: "commit fails",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			txFunc: func(tx *sql.Tx) error {
				return nil
			},
			wantErr: true,
		},
		{
			name: "panic in transaction function",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			txFunc: func(tx *sql.Tx) error {
				panic("panic in transaction function")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTxManagerMock(t)
			defer db.Close()

			tt.mockSetup(mock)

			txManager := &DBTxManager{
				db: db,
			}

			var err error
			if tt.name == "panic in transaction function" {
				defer func() {
					if r := recover(); r == nil {
						t.Error("Expected panic was not recovered")
					}
				}()
			}

			err = txManager.RunTransaction(context.Background(), tt.txFunc)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil && err != nil {
					assert.Equal(t, tt.expectedErr.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDBTxManager_RunTransaction_PanicHandling(t *testing.T) {
	dbWrapper, mock := setupTxManagerMock(t)
	defer dbWrapper.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	txManager := &DBTxManager{
		db: dbWrapper,
	}

	txFunc := func(tx *sql.Tx) error {
		panic("test panic")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic was not recovered")
		} else {
			assert.Equal(t, "test panic", r)
		}
	}()

	_ = txManager.RunTransaction(context.Background(), txFunc)

	assert.NoError(t, mock.ExpectationsWereMet())
}
