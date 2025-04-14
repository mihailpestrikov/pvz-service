package postgres

import (
	"avito-backend-trainee-assignment-spring-2025/internal/domain/interfaces"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
	"avito-backend-trainee-assignment-spring-2025/internal/repository/repoerrors"
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupProductRepoMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *ProductRepository) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}

	repo := &ProductRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	return db, mock, repo
}

func TestNewProductRepository(t *testing.T) {
	db, _, _ := setupProductRepoMock(t)
	defer db.Close()

	repo := NewProductRepository(db)
	assert.NotNil(t, repo, "Repository should not be nil")
	assert.Implements(t, (*interfaces.TxProductRepository)(nil), repo)
}

func TestProductRepository_Create(t *testing.T) {
	productId := uuid.New()
	receptionId := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		product     *models.Product
		mockSetup   func(sqlmock.Sqlmock)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful creation",
			product: &models.Product{
				ID:          productId,
				DateTime:    now,
				Type:        models.ProductTypeElectronics,
				ReceptionID: receptionId,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO product (id,date_time,type,reception_id) VALUES ($1,$2,$3,$4)`).
					WithArgs(productId, now, models.ProductTypeElectronics, receptionId).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "database error",
			product: &models.Product{
				ID:          productId,
				DateTime:    now,
				Type:        models.ProductTypeElectronics,
				ReceptionID: receptionId,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO product (id,date_time,type,reception_id) VALUES ($1,$2,$3,$4)`).
					WithArgs(productId, now, models.ProductTypeElectronics, receptionId).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupProductRepoMock(t)
			defer db.Close()

			tt.mockSetup(mock)

			err := repo.Create(context.Background(), tt.product)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestProductRepository_GetByID(t *testing.T) {
	productID := uuid.New()
	receptionID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		id          uuid.UUID
		mockSetup   func(sqlmock.Sqlmock)
		want        *models.Product
		wantErr     bool
		expectedErr error
	}{
		{
			name: "product found",
			id:   productID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id"}).
					AddRow(productID, now, models.ProductTypeElectronics, receptionID)

				mock.ExpectQuery(`SELECT id, date_time, type, reception_id FROM product WHERE id = $1`).
					WithArgs(productID).
					WillReturnRows(rows)
			},
			want: &models.Product{
				ID:          productID,
				DateTime:    now,
				Type:        models.ProductTypeElectronics,
				ReceptionID: receptionID,
			},
			wantErr: false,
		},
		{
			name: "product not found",
			id:   productID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, date_time, type, reception_id FROM product WHERE id = $1`).
					WithArgs(productID).
					WillReturnError(sql.ErrNoRows)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: repoerrors.ErrProductNotFound,
		},
		{
			name: "database error",
			id:   productID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, date_time, type, reception_id FROM product WHERE id = $1`).
					WithArgs(productID).
					WillReturnError(errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupProductRepoMock(t)
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
				assert.Equal(t, tt.want.Type, got.Type)
				assert.Equal(t, tt.want.ReceptionID, got.ReceptionID)
				assert.WithinDuration(t, tt.want.DateTime, got.DateTime, time.Second)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestProductRepository_GetByReceptionID(t *testing.T) {
	receptionID := uuid.New()
	productID1 := uuid.New()
	productID2 := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		receptionID uuid.UUID
		mockSetup   func(sqlmock.Sqlmock)
		want        []models.Product
		wantErr     bool
	}{
		{
			name:        "products found",
			receptionID: receptionID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id"}).
					AddRow(productID1, now, models.ProductTypeElectronics, receptionID).
					AddRow(productID2, now.Add(time.Hour), models.ProductTypeClothes, receptionID)

				mock.ExpectQuery(`SELECT id, date_time, type, reception_id FROM product WHERE reception_id = $1 ORDER BY date_time ASC`).
					WithArgs(receptionID).
					WillReturnRows(rows)
			},
			want: []models.Product{
				{
					ID:          productID1,
					DateTime:    now,
					Type:        models.ProductTypeElectronics,
					ReceptionID: receptionID,
				},
				{
					ID:          productID2,
					DateTime:    now.Add(time.Hour),
					Type:        models.ProductTypeClothes,
					ReceptionID: receptionID,
				},
			},
			wantErr: false,
		},
		{
			name:        "no products found",
			receptionID: receptionID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id"})

				mock.ExpectQuery(`SELECT id, date_time, type, reception_id FROM product WHERE reception_id = $1 ORDER BY date_time ASC`).
					WithArgs(receptionID).
					WillReturnRows(rows)
			},
			want:    []models.Product{},
			wantErr: false,
		},
		{
			name:        "database error",
			receptionID: receptionID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, date_time, type, reception_id FROM product WHERE reception_id = $1 ORDER BY date_time ASC`).
					WithArgs(receptionID).
					WillReturnError(errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupProductRepoMock(t)
			defer db.Close()

			tt.mockSetup(mock)

			got, err := repo.GetByReceptionID(context.Background(), tt.receptionID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, len(tt.want))

				if len(tt.want) > 0 {
					for i, product := range tt.want {
						assert.Equal(t, product.ID, got[i].ID)
						assert.Equal(t, product.Type, got[i].Type)
						assert.Equal(t, product.ReceptionID, got[i].ReceptionID)
						assert.WithinDuration(t, product.DateTime, got[i].DateTime, time.Second)
					}
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestProductRepository_DeleteLastFromReception(t *testing.T) {
	receptionID := uuid.New()
	productID := uuid.New()

	tests := []struct {
		name        string
		receptionID uuid.UUID
		mockSetup   func(sqlmock.Sqlmock)
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "successful deletion",
			receptionID: receptionID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(productID)
				mock.ExpectQuery(`SELECT id FROM product WHERE reception_id = $1 ORDER BY date_time DESC LIMIT 1`).
					WithArgs(receptionID).
					WillReturnRows(rows)

				mock.ExpectExec(`DELETE FROM product WHERE id = $1`).
					WithArgs(productID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:        "no products found",
			receptionID: receptionID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id FROM product WHERE reception_id = $1 ORDER BY date_time DESC LIMIT 1`).
					WithArgs(receptionID).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr:     true,
			expectedErr: repoerrors.ErrProductNotFound,
		},
		{
			name:        "error finding last product",
			receptionID: receptionID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id FROM product WHERE reception_id = $1 ORDER BY date_time DESC LIMIT 1`).
					WithArgs(receptionID).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name:        "error deleting product",
			receptionID: receptionID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(productID)
				mock.ExpectQuery(`SELECT id FROM product WHERE reception_id = $1 ORDER BY date_time DESC LIMIT 1`).
					WithArgs(receptionID).
					WillReturnRows(rows)

				mock.ExpectExec(`DELETE FROM product WHERE id = $1`).
					WithArgs(productID).
					WillReturnError(errors.New("deletion error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupProductRepoMock(t)
			defer db.Close()

			tt.mockSetup(mock)

			err := repo.DeleteLastFromReception(context.Background(), tt.receptionID)

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
