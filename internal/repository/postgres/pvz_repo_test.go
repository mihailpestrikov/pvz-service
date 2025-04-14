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

func setupPVZRepoMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *PVZRepository) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}

	repo := &PVZRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	return db, mock, repo
}

func TestNewPVZRepository(t *testing.T) {
	db, _, _ := setupPVZRepoMock(t)
	defer db.Close()

	repo := NewPVZRepository(db)
	assert.NotNil(t, repo, "Repository should not be nil")
	assert.Implements(t, (*interfaces.TxPVZRepository)(nil), repo)
}

func TestPVZRepository_Create(t *testing.T) {
	pvzID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		pvz         *models.PVZ
		mockSetup   func(sqlmock.Sqlmock)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful creation",
			pvz: &models.PVZ{
				ID:               pvzID,
				RegistrationDate: now,
				City:             models.CityMoscow,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO pvz (id,registration_date,city) VALUES ($1,$2,$3)`).
					WithArgs(pvzID, now, models.CityMoscow).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "duplicate key error",
			pvz: &models.PVZ{
				ID:               pvzID,
				RegistrationDate: now,
				City:             models.CityMoscow,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO pvz (id,registration_date,city) VALUES ($1,$2,$3)`).
					WithArgs(pvzID, now, models.CityMoscow).
					WillReturnError(errors.New(`pq: duplicate key value violates unique constraint "users_email_key"`))
			},
			wantErr:     true,
			expectedErr: repoerrors.ErrPVZAlreadyExists,
		},
		{
			name: "database error",
			pvz: &models.PVZ{
				ID:               pvzID,
				RegistrationDate: now,
				City:             models.CityMoscow,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO pvz (id,registration_date,city) VALUES ($1,$2,$3)`).
					WithArgs(pvzID, now, models.CityMoscow).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupPVZRepoMock(t)
			defer db.Close()

			tt.mockSetup(mock)

			err := repo.Create(context.Background(), tt.pvz)

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

func TestPVZRepository_GetByID(t *testing.T) {
	pvzID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		id          uuid.UUID
		mockSetup   func(sqlmock.Sqlmock)
		want        *models.PVZ
		wantErr     bool
		expectedErr error
	}{
		{
			name: "pvz found",
			id:   pvzID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "registration_date", "city"}).
					AddRow(pvzID, now, models.CityMoscow)

				mock.ExpectQuery(`SELECT id, registration_date, city FROM pvz WHERE id = $1`).
					WithArgs(pvzID).
					WillReturnRows(rows)
			},
			want: &models.PVZ{
				ID:               pvzID,
				RegistrationDate: now,
				City:             models.CityMoscow,
			},
			wantErr: false,
		},
		{
			name: "pvz not found",
			id:   pvzID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, registration_date, city FROM pvz WHERE id = $1`).
					WithArgs(pvzID).
					WillReturnError(sql.ErrNoRows)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: repoerrors.ErrPVZNotFound,
		},
		{
			name: "database error",
			id:   pvzID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, registration_date, city FROM pvz WHERE id = $1`).
					WithArgs(pvzID).
					WillReturnError(errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupPVZRepoMock(t)
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
				assert.Equal(t, tt.want.City, got.City)
				assert.WithinDuration(t, tt.want.RegistrationDate, got.RegistrationDate, time.Second)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPVZRepository_GetAll(t *testing.T) {
	pvzID1 := uuid.New()
	pvzID2 := uuid.New()
	now := time.Now()
	startDate := now.Add(-24 * time.Hour)
	endDate := now

	tests := []struct {
		name        string
		filter      models.PVZFilter
		mockSetup   func(sqlmock.Sqlmock)
		want        []*models.PVZ
		wantTotal   int
		wantErr     bool
		expectedErr error
	}{
		{
			name: "get all pvz without filters",
			filter: models.PVZFilter{
				Page:  1,
				Limit: 10,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery(`SELECT COUNT(*) FROM pvz`).
					WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{"id", "registration_date", "city"}).
					AddRow(pvzID1, now, models.CityMoscow).
					AddRow(pvzID2, now.Add(time.Hour), models.CitySaintPete)

				mock.ExpectQuery(`SELECT id, registration_date, city FROM pvz LIMIT 10 OFFSET 0`).
					WillReturnRows(rows)
			},
			want: []*models.PVZ{
				{
					ID:               pvzID1,
					RegistrationDate: now,
					City:             models.CityMoscow,
				},
				{
					ID:               pvzID2,
					RegistrationDate: now.Add(time.Hour),
					City:             models.CitySaintPete,
				},
			},
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name: "get pvz with date filter",
			filter: models.PVZFilter{
				StartDate: &startDate,
				EndDate:   &endDate,
				Page:      1,
				Limit:     10,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery(`SELECT COUNT(*) FROM pvz WHERE EXISTS ( SELECT 1 FROM reception WHERE reception.pvz_id = pvz.id AND reception.date_time BETWEEN $1 AND $2 )`).
					WithArgs(&startDate, &endDate).
					WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{"id", "registration_date", "city"}).
					AddRow(pvzID1, now, models.CityMoscow)

				mock.ExpectQuery(`SELECT id, registration_date, city FROM pvz WHERE EXISTS ( SELECT 1 FROM reception WHERE reception.pvz_id = pvz.id AND reception.date_time BETWEEN $1 AND $2 ) LIMIT 10 OFFSET 0`).
					WithArgs(&startDate, &endDate).
					WillReturnRows(rows)
			},
			want: []*models.PVZ{
				{
					ID:               pvzID1,
					RegistrationDate: now,
					City:             models.CityMoscow,
				},
			},
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name: "pagination test",
			filter: models.PVZFilter{
				Page:  2,
				Limit: 1,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery(`SELECT COUNT(*) FROM pvz`).
					WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{"id", "registration_date", "city"}).
					AddRow(pvzID2, now.Add(time.Hour), models.CitySaintPete)

				mock.ExpectQuery(`SELECT id, registration_date, city FROM pvz LIMIT 1 OFFSET 1`).
					WillReturnRows(rows)
			},
			want: []*models.PVZ{
				{
					ID:               pvzID2,
					RegistrationDate: now.Add(time.Hour),
					City:             models.CitySaintPete,
				},
			},
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name: "database error on count",
			filter: models.PVZFilter{
				Page:  1,
				Limit: 10,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT(*) FROM pvz`).
					WillReturnError(errors.New("database error"))
			},
			want:      nil,
			wantTotal: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupPVZRepoMock(t)
			defer db.Close()

			tt.mockSetup(mock)

			got, total, err := repo.GetAll(context.Background(), tt.filter)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTotal, total)
				assert.Len(t, got, len(tt.want))

				for i, pvz := range tt.want {
					assert.Equal(t, pvz.ID, got[i].ID)
					assert.Equal(t, pvz.City, got[i].City)
					assert.WithinDuration(t, pvz.RegistrationDate, got[i].RegistrationDate, time.Second)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPVZRepository_GetAllWithReceptions(t *testing.T) {
	pvzID1 := uuid.New()
	receptionID1 := uuid.New()
	productID1 := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		filter      models.PVZFilter
		mockSetup   func(sqlmock.Sqlmock)
		want        []models.PVZWithReceptions
		wantTotal   int
		wantErr     bool
		expectedErr error
	}{
		{
			name: "get pvz with receptions and products",
			filter: models.PVZFilter{
				Page:  1,
				Limit: 10,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery(`SELECT COUNT(*) FROM pvz`).
					WillReturnRows(countRows)

				pvzRows := sqlmock.NewRows([]string{"id", "registration_date", "city"}).
					AddRow(pvzID1, now, models.CityMoscow)

				mock.ExpectQuery(`SELECT id, registration_date, city FROM pvz LIMIT 10 OFFSET 0`).
					WillReturnRows(pvzRows)

				receptionRows := sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
					AddRow(receptionID1, now, pvzID1, models.ReceptionStatusInProgress)

				mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM reception WHERE pvz_id IN ($1)`).
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(receptionRows)

				productRows := sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id"}).
					AddRow(productID1, now, models.ProductTypeElectronics, receptionID1)

				mock.ExpectQuery(`SELECT id, date_time, type, reception_id FROM product WHERE reception_id IN ($1)`).
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(productRows)
			},
			want: []models.PVZWithReceptions{
				{
					PVZ: &models.PVZ{
						ID:               pvzID1,
						RegistrationDate: now,
						City:             models.CityMoscow,
					},
					Receptions: []*models.Reception{
						{
							ID:       receptionID1,
							DateTime: now,
							PVZID:    pvzID1,
							Status:   models.ReceptionStatusInProgress,
							Products: []models.Product{
								{
									ID:          productID1,
									DateTime:    now,
									Type:        models.ProductTypeElectronics,
									ReceptionID: receptionID1,
								},
							},
						},
					},
				},
			},
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name: "get pvz with no receptions",
			filter: models.PVZFilter{
				Page:  1,
				Limit: 10,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery(`SELECT COUNT(*) FROM pvz`).
					WillReturnRows(countRows)

				pvzRows := sqlmock.NewRows([]string{"id", "registration_date", "city"}).
					AddRow(pvzID1, now, models.CityMoscow)

				mock.ExpectQuery(`SELECT id, registration_date, city FROM pvz LIMIT 10 OFFSET 0`).
					WillReturnRows(pvzRows)

				receptionRows := sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"})

				mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM reception WHERE pvz_id IN ($1)`).
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(receptionRows)
			},
			want: []models.PVZWithReceptions{
				{
					PVZ: &models.PVZ{
						ID:               pvzID1,
						RegistrationDate: now,
						City:             models.CityMoscow,
					},
					Receptions: []*models.Reception{},
				},
			},
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name: "error fetching pvz",
			filter: models.PVZFilter{
				Page:  1,
				Limit: 10,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT(*) FROM pvz`).
					WillReturnError(errors.New("database error"))
			},
			want:      nil,
			wantTotal: 0,
			wantErr:   true,
		},
		{
			name: "error fetching receptions",
			filter: models.PVZFilter{
				Page:  1,
				Limit: 10,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery(`SELECT COUNT(*) FROM pvz`).
					WillReturnRows(countRows)

				pvzRows := sqlmock.NewRows([]string{"id", "registration_date", "city"}).
					AddRow(pvzID1, now, models.CityMoscow)

				mock.ExpectQuery(`SELECT id, registration_date, city FROM pvz LIMIT 10 OFFSET 0`).
					WillReturnRows(pvzRows)

				mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM reception WHERE pvz_id IN ($1)`).
					WithArgs(sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			want:      nil,
			wantTotal: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupPVZRepoMock(t)
			defer db.Close()

			tt.mockSetup(mock)

			got, total, err := repo.GetAllWithReceptions(context.Background(), tt.filter)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTotal, total)
				assert.Len(t, got, len(tt.want))

				if len(tt.want) > 0 {
					for i, pvzWithReceptions := range tt.want {
						assert.Equal(t, pvzWithReceptions.PVZ.ID, got[i].PVZ.ID)
						assert.Equal(t, pvzWithReceptions.PVZ.City, got[i].PVZ.City)
						assert.WithinDuration(t, pvzWithReceptions.PVZ.RegistrationDate, got[i].PVZ.RegistrationDate, time.Second)

						assert.Len(t, got[i].Receptions, len(pvzWithReceptions.Receptions))

						for j, reception := range pvzWithReceptions.Receptions {
							if len(pvzWithReceptions.Receptions) > 0 {
								assert.Equal(t, reception.ID, got[i].Receptions[j].ID)
								assert.Equal(t, reception.PVZID, got[i].Receptions[j].PVZID)
								assert.Equal(t, reception.Status, got[i].Receptions[j].Status)
								assert.WithinDuration(t, reception.DateTime, got[i].Receptions[j].DateTime, time.Second)

								assert.Len(t, got[i].Receptions[j].Products, len(reception.Products))

								for k, product := range reception.Products {
									assert.Equal(t, product.ID, got[i].Receptions[j].Products[k].ID)
									assert.Equal(t, product.Type, got[i].Receptions[j].Products[k].Type)
									assert.Equal(t, product.ReceptionID, got[i].Receptions[j].Products[k].ReceptionID)
									assert.WithinDuration(t, product.DateTime, got[i].Receptions[j].Products[k].DateTime, time.Second)
								}
							}
						}
					}
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
