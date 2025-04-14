package postgres

import (
	"avito-backend-trainee-assignment-spring-2025/internal/domain/apperrors"
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

func setupReceptionRepoMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *ReceptionRepository) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}

	repo := &ReceptionRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	return db, mock, repo
}

func TestNewReceptionRepository(t *testing.T) {
	db, _, _ := setupReceptionRepoMock(t)
	defer db.Close()

	repo := NewReceptionRepository(db)
	assert.NotNil(t, repo, "Repository should not be nil")
	assert.Implements(t, (*interfaces.TxReceptionRepository)(nil), repo)
}

func TestReceptionRepository_Create(t *testing.T) {
	receptionID := uuid.New()
	pvzID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		reception   *models.Reception
		mockSetup   func(sqlmock.Sqlmock)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful creation",
			reception: &models.Reception{
				ID:       receptionID,
				DateTime: now,
				PVZID:    pvzID,
				Status:   models.ReceptionStatusInProgress,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO reception (id,date_time,pvz_id,status) VALUES ($1,$2,$3,$4)`).
					WithArgs(receptionID, now, pvzID, models.ReceptionStatusInProgress).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "database error",
			reception: &models.Reception{
				ID:       receptionID,
				DateTime: now,
				PVZID:    pvzID,
				Status:   models.ReceptionStatusInProgress,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO reception (id,date_time,pvz_id,status) VALUES ($1,$2,$3,$4)`).
					WithArgs(receptionID, now, pvzID, models.ReceptionStatusInProgress).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupReceptionRepoMock(t)
			defer db.Close()

			tt.mockSetup(mock)

			err := repo.Create(context.Background(), tt.reception)

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

func TestReceptionRepository_GetByID(t *testing.T) {
	receptionID := uuid.New()
	pvzID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		id          uuid.UUID
		mockSetup   func(sqlmock.Sqlmock)
		want        *models.Reception
		wantErr     bool
		expectedErr error
	}{
		{
			name: "reception found",
			id:   receptionID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
					AddRow(receptionID, now, pvzID, models.ReceptionStatusInProgress)

				mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM reception WHERE id = $1`).
					WithArgs(receptionID).
					WillReturnRows(rows)

				mock.ExpectQuery(`SELECT id, date_time, type, reception_id FROM product WHERE reception_id = $1 ORDER BY date_time ASC`).
					WithArgs(receptionID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id"}))
			},
			want: &models.Reception{
				ID:       receptionID,
				DateTime: now,
				PVZID:    pvzID,
				Status:   models.ReceptionStatusInProgress,
				Products: []models.Product{},
			},
			wantErr: false,
		},
		{
			name: "reception not found",
			id:   receptionID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM reception WHERE id = $1`).
					WithArgs(receptionID).
					WillReturnError(sql.ErrNoRows)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: repoerrors.ErrReceptionNotFound,
		},
		{
			name: "database error",
			id:   receptionID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM reception WHERE id = $1`).
					WithArgs(receptionID).
					WillReturnError(errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupReceptionRepoMock(t)
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
				assert.Equal(t, tt.want.PVZID, got.PVZID)
				assert.Equal(t, tt.want.Status, got.Status)
				assert.WithinDuration(t, tt.want.DateTime, got.DateTime, time.Second)
				assert.Equal(t, len(tt.want.Products), len(got.Products))
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReceptionRepository_GetLastActiveByPVZID(t *testing.T) {
	receptionID := uuid.New()
	pvzID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		pvzID       uuid.UUID
		mockSetup   func(sqlmock.Sqlmock)
		want        *models.Reception
		wantErr     bool
		expectedErr error
	}{
		{
			name:  "active reception found",
			pvzID: pvzID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
					AddRow(receptionID, now, pvzID, models.ReceptionStatusInProgress)

				mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM reception WHERE pvz_id = $1 AND status = $2 ORDER BY date_time DESC LIMIT 1`).
					WithArgs(pvzID, models.ReceptionStatusInProgress).
					WillReturnRows(rows)

				mock.ExpectQuery(`SELECT id, date_time, type, reception_id FROM product WHERE reception_id = $1 ORDER BY date_time ASC`).
					WithArgs(receptionID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id"}))
			},
			want: &models.Reception{
				ID:       receptionID,
				DateTime: now,
				PVZID:    pvzID,
				Status:   models.ReceptionStatusInProgress,
				Products: []models.Product{},
			},
			wantErr: false,
		},
		{
			name:  "no active reception",
			pvzID: pvzID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM reception WHERE pvz_id = $1 AND status = $2 ORDER BY date_time DESC LIMIT 1`).
					WithArgs(pvzID, models.ReceptionStatusInProgress).
					WillReturnError(sql.ErrNoRows)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: apperrors.ErrNoActiveReception,
		},
		{
			name:  "database error",
			pvzID: pvzID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM reception WHERE pvz_id = $1 AND status = $2 ORDER BY date_time DESC LIMIT 1`).
					WithArgs(pvzID, models.ReceptionStatusInProgress).
					WillReturnError(errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupReceptionRepoMock(t)
			defer db.Close()

			tt.mockSetup(mock)

			got, err := repo.GetLastActiveByPVZID(context.Background(), tt.pvzID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.PVZID, got.PVZID)
				assert.Equal(t, tt.want.Status, got.Status)
				assert.WithinDuration(t, tt.want.DateTime, got.DateTime, time.Second)
				assert.Equal(t, len(tt.want.Products), len(got.Products))
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReceptionRepository_GetLastReceptionByPVZID(t *testing.T) {
	receptionID := uuid.New()
	pvzID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		pvzID       uuid.UUID
		mockSetup   func(sqlmock.Sqlmock)
		want        *models.Reception
		wantErr     bool
		expectedErr error
	}{
		{
			name:  "last reception found",
			pvzID: pvzID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
					AddRow(receptionID, now, pvzID, models.ReceptionStatusClosed)

				mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM reception WHERE pvz_id = $1 ORDER BY date_time DESC LIMIT 1`).
					WithArgs(pvzID).
					WillReturnRows(rows)

				mock.ExpectQuery(`SELECT id, date_time, type, reception_id FROM product WHERE reception_id = $1 ORDER BY date_time ASC`).
					WithArgs(receptionID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id"}))
			},
			want: &models.Reception{
				ID:       receptionID,
				DateTime: now,
				PVZID:    pvzID,
				Status:   models.ReceptionStatusClosed,
				Products: []models.Product{},
			},
			wantErr: false,
		},
		{
			name:  "no reception found",
			pvzID: pvzID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM reception WHERE pvz_id = $1 ORDER BY date_time DESC LIMIT 1`).
					WithArgs(pvzID).
					WillReturnError(sql.ErrNoRows)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: repoerrors.ErrReceptionNotFound,
		},
		{
			name:  "database error",
			pvzID: pvzID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM reception WHERE pvz_id = $1 ORDER BY date_time DESC LIMIT 1`).
					WithArgs(pvzID).
					WillReturnError(errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupReceptionRepoMock(t)
			defer db.Close()

			tt.mockSetup(mock)

			got, err := repo.GetLastReceptionByPVZID(context.Background(), tt.pvzID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.PVZID, got.PVZID)
				assert.Equal(t, tt.want.Status, got.Status)
				assert.WithinDuration(t, tt.want.DateTime, got.DateTime, time.Second)
				assert.Equal(t, len(tt.want.Products), len(got.Products))
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReceptionRepository_CloseReception(t *testing.T) {
	receptionID := uuid.New()
	pvzID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		id          uuid.UUID
		mockSetup   func(sqlmock.Sqlmock)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful close",
			id:   receptionID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
					AddRow(receptionID, now, pvzID, models.ReceptionStatusInProgress)

				mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM reception WHERE id = $1`).
					WithArgs(receptionID).
					WillReturnRows(rows)

				mock.ExpectQuery(`SELECT id, date_time, type, reception_id FROM product WHERE reception_id = $1 ORDER BY date_time ASC`).
					WithArgs(receptionID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id"}))

				mock.ExpectExec(`UPDATE reception SET status = $1 WHERE id = $2`).
					WithArgs(models.ReceptionStatusClosed, receptionID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "reception already closed",
			id:   receptionID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
					AddRow(receptionID, now, pvzID, models.ReceptionStatusClosed)

				mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM reception WHERE id = $1`).
					WithArgs(receptionID).
					WillReturnRows(rows)

				mock.ExpectQuery(`SELECT id, date_time, type, reception_id FROM product WHERE reception_id = $1 ORDER BY date_time ASC`).
					WithArgs(receptionID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id"}))
			},
			wantErr:     true,
			expectedErr: apperrors.ErrReceptionAlreadyClosed,
		},
		{
			name: "reception not found",
			id:   receptionID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM reception WHERE id = $1`).
					WithArgs(receptionID).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr:     true,
			expectedErr: repoerrors.ErrReceptionNotFound,
		},
		{
			name: "update error",
			id:   receptionID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
					AddRow(receptionID, now, pvzID, models.ReceptionStatusInProgress)

				mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM reception WHERE id = $1`).
					WithArgs(receptionID).
					WillReturnRows(rows)

				mock.ExpectQuery(`SELECT id, date_time, type, reception_id FROM product WHERE reception_id = $1 ORDER BY date_time ASC`).
					WithArgs(receptionID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id"}))

				mock.ExpectExec(`UPDATE reception SET status = $1 WHERE id = $2`).
					WithArgs(models.ReceptionStatusClosed, receptionID).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupReceptionRepoMock(t)
			defer db.Close()

			tt.mockSetup(mock)

			err := repo.CloseReception(context.Background(), tt.id)

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

func TestReceptionRepository_getProductsForReception(t *testing.T) {
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
		expectedErr error
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
			db, mock, repo := setupReceptionRepoMock(t)
			defer db.Close()

			tt.mockSetup(mock)

			got, err := repo.getProductsForReception(context.Background(), tt.receptionID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
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
