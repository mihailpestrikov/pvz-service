package models

import (
	"github.com/google/uuid"
	"testing"
	"time"
)

func TestIsValidCity(t *testing.T) {
	type args struct {
		city string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Valid city - Moscow",
			args: args{city: "Москва"},
			want: true,
		},
		{
			name: "Valid city - Saint Petersburg",
			args: args{city: "Санкт-Петербург"},
			want: true,
		},
		{
			name: "Valid city - Kazan",
			args: args{city: "Казань"},
			want: true,
		},
		{
			name: "Invalid city",
			args: args{city: "Новосибирск"},
			want: false,
		},
		{
			name: "Empty city",
			args: args{city: ""},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidCity(tt.args.city); got != tt.want {
				t.Errorf("IsValidCity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidProductType(t *testing.T) {
	type args struct {
		productType string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Valid type - electronics",
			args: args{productType: "электроника"},
			want: true,
		},
		{
			name: "Valid type - clothes",
			args: args{productType: "одежда"},
			want: true,
		},
		{
			name: "Valid type - shoes",
			args: args{productType: "обувь"},
			want: true,
		},
		{
			name: "Invalid type",
			args: args{productType: "мебель"},
			want: false,
		},
		{
			name: "Empty type",
			args: args{productType: ""},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidProductType(tt.args.productType); got != tt.want {
				t.Errorf("IsValidProductType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewPVZ(t *testing.T) {
	type args struct {
		city string
	}
	tests := []struct {
		name    string
		args    args
		want    *PVZ
		wantErr bool
	}{
		{
			name: "Valid PVZ creation",
			args: args{city: "Москва"},
			want: &PVZ{
				City: "Москва",
			},
			wantErr: false,
		},
		{
			name:    "Invalid city",
			args:    args{city: "Новосибирск"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Empty city",
			args:    args{city: ""},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPVZ(tt.args.city)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPVZ() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && got != nil {
				if got.City != tt.want.City {
					t.Errorf("NewPVZ().City = %v, want %v", got.City, tt.want.City)
				}
				if got.RegistrationDate.IsZero() {
					t.Errorf("NewPVZ().RegistrationDate should not be zero time")
				}
				if got.ID == uuid.Nil {
					t.Errorf("NewPVZ().ID should not be nil UUID")
				}
			}
		})
	}
}

func TestNewProduct(t *testing.T) {
	receptionID := uuid.New()

	type args struct {
		productType string
		receptionID uuid.UUID
	}
	tests := []struct {
		name    string
		args    args
		want    *Product
		wantErr bool
	}{
		{
			name: "Valid product creation",
			args: args{
				productType: "электроника",
				receptionID: receptionID,
			},
			want: &Product{
				Type:        "электроника",
				ReceptionID: receptionID,
			},
			wantErr: false,
		},
		{
			name: "Invalid product type",
			args: args{
				productType: "мебель",
				receptionID: receptionID,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Empty product type",
			args: args{
				productType: "",
				receptionID: receptionID,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Nil reception ID",
			args: args{
				productType: "электроника",
				receptionID: uuid.Nil,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewProduct(tt.args.productType, tt.args.receptionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProduct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && got != nil {
				if got.Type != tt.want.Type {
					t.Errorf("NewProduct().Type = %v, want %v", got.Type, tt.want.Type)
				}
				if got.DateTime.IsZero() {
					t.Errorf("NewProduct().DateTime should not be zero time")
				}
				if got.ReceptionID != tt.want.ReceptionID {
					t.Errorf("NewProduct().ReceptionID = %v, want %v", got.ReceptionID, tt.want.ReceptionID)
				}
				if got.ID == uuid.Nil {
					t.Errorf("NewProduct().ID should not be nil UUID")
				}
			}
		})
	}
}

func TestNewReception(t *testing.T) {
	pvzID := uuid.New()

	type args struct {
		pvzID uuid.UUID
	}
	tests := []struct {
		name    string
		args    args
		want    *Reception
		wantErr bool
	}{
		{
			name: "Valid reception creation",
			args: args{pvzID: pvzID},
			want: &Reception{
				PVZID:    pvzID,
				Status:   ReceptionStatusInProgress,
				Products: []Product{},
			},
			wantErr: false,
		},
		{
			name:    "Nil PVZ ID",
			args:    args{pvzID: uuid.Nil},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewReception(tt.args.pvzID)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewReception() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && got != nil {
				if got.PVZID != tt.want.PVZID {
					t.Errorf("NewReception().PVZID = %v, want %v", got.PVZID, tt.want.PVZID)
				}
				if got.DateTime.IsZero() {
					t.Errorf("NewReception().DateTime should not be zero time")
				}
				if got.Status != tt.want.Status {
					t.Errorf("NewReception().Status = %v, want %v", got.Status, tt.want.Status)
				}
				if len(got.Products) != len(tt.want.Products) {
					t.Errorf("NewReception().Products length = %v, want %v", len(got.Products), len(tt.want.Products))
				}
				if got.ID == uuid.Nil {
					t.Errorf("NewReception().ID should not be nil UUID")
				}
			}
		})
	}
}

func TestNewUser(t *testing.T) {
	type args struct {
		email    string
		password string
		role     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid user creation - employee",
			args: args{
				email:    "user@example.com",
				password: "password123",
				role:     RoleEmployee,
			},
			wantErr: false,
		},
		{
			name: "Valid user creation - moderator",
			args: args{
				email:    "moderator@example.com",
				password: "password123",
				role:     RoleModerator,
			},
			wantErr: false,
		},
		{
			name: "Invalid email",
			args: args{
				email:    "invalid-email",
				password: "password123",
				role:     RoleEmployee,
			},
			wantErr: true,
		},
		{
			name: "Empty email",
			args: args{
				email:    "",
				password: "password123",
				role:     RoleEmployee,
			},
			wantErr: true,
		},
		{
			name: "Short password",
			args: args{
				email:    "user@example.com",
				password: "123",
				role:     RoleEmployee,
			},
			wantErr: true,
		},
		{
			name: "Empty password",
			args: args{
				email:    "user@example.com",
				password: "",
				role:     RoleEmployee,
			},
			wantErr: true,
		},
		{
			name: "Invalid role",
			args: args{
				email:    "user@example.com",
				password: "password123",
				role:     "admin",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewUser(tt.args.email, tt.args.password, tt.args.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != nil {
				if got.Email != tt.args.email {
					t.Errorf("NewUser().Email = %v, want %v", got.Email, tt.args.email)
				}
				if got.Role != tt.args.role {
					t.Errorf("NewUser().Role = %v, want %v", got.Role, tt.args.role)
				}
				if got.PasswordHash == "" {
					t.Errorf("NewUser().PasswordHash should not be empty")
				}
				if got.PasswordHash == tt.args.password {
					t.Errorf("NewUser().PasswordHash should not equal original password")
				}
				if got.ID == uuid.Nil {
					t.Errorf("NewUser().ID should not be nil UUID")
				}
			}
		})
	}
}

func TestReception_IsClosed(t *testing.T) {
	type fields struct {
		ID       uuid.UUID
		DateTime time.Time
		PVZID    uuid.UUID
		Status   string
		Products []Product
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Reception is closed",
			fields: fields{
				ID:       uuid.New(),
				DateTime: time.Now(),
				PVZID:    uuid.New(),
				Status:   ReceptionStatusClosed,
				Products: []Product{},
			},
			want: true,
		},
		{
			name: "Reception is in progress",
			fields: fields{
				ID:       uuid.New(),
				DateTime: time.Now(),
				PVZID:    uuid.New(),
				Status:   ReceptionStatusInProgress,
				Products: []Product{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reception{
				ID:       tt.fields.ID,
				DateTime: tt.fields.DateTime,
				PVZID:    tt.fields.PVZID,
				Status:   tt.fields.Status,
				Products: tt.fields.Products,
			}
			if got := r.IsClosed(); got != tt.want {
				t.Errorf("IsClosed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReception_IsInProgress(t *testing.T) {
	type fields struct {
		ID       uuid.UUID
		DateTime time.Time
		PVZID    uuid.UUID
		Status   string
		Products []Product
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Reception is in progress",
			fields: fields{
				ID:       uuid.New(),
				DateTime: time.Now(),
				PVZID:    uuid.New(),
				Status:   ReceptionStatusInProgress,
				Products: []Product{},
			},
			want: true,
		},
		{
			name: "Reception is closed",
			fields: fields{
				ID:       uuid.New(),
				DateTime: time.Now(),
				PVZID:    uuid.New(),
				Status:   ReceptionStatusClosed,
				Products: []Product{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reception{
				ID:       tt.fields.ID,
				DateTime: tt.fields.DateTime,
				PVZID:    tt.fields.PVZID,
				Status:   tt.fields.Status,
				Products: tt.fields.Products,
			}
			if got := r.IsInProgress(); got != tt.want {
				t.Errorf("IsInProgress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReception_ProductCount(t *testing.T) {
	product1 := Product{ID: uuid.New(), Type: "электроника"}
	product2 := Product{ID: uuid.New(), Type: "одежда"}

	type fields struct {
		ID       uuid.UUID
		DateTime time.Time
		PVZID    uuid.UUID
		Status   string
		Products []Product
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "Reception with no products",
			fields: fields{
				ID:       uuid.New(),
				DateTime: time.Now(),
				PVZID:    uuid.New(),
				Status:   ReceptionStatusInProgress,
				Products: []Product{},
			},
			want: 0,
		},
		{
			name: "Reception with one product",
			fields: fields{
				ID:       uuid.New(),
				DateTime: time.Now(),
				PVZID:    uuid.New(),
				Status:   ReceptionStatusInProgress,
				Products: []Product{product1},
			},
			want: 1,
		},
		{
			name: "Reception with multiple products",
			fields: fields{
				ID:       uuid.New(),
				DateTime: time.Now(),
				PVZID:    uuid.New(),
				Status:   ReceptionStatusInProgress,
				Products: []Product{product1, product2},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reception{
				ID:       tt.fields.ID,
				DateTime: tt.fields.DateTime,
				PVZID:    tt.fields.PVZID,
				Status:   tt.fields.Status,
				Products: tt.fields.Products,
			}
			if got := r.ProductCount(); got != tt.want {
				t.Errorf("ProductCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_IsEmployee(t *testing.T) {
	type fields struct {
		ID           uuid.UUID
		Email        string
		PasswordHash string
		Role         string
		CreatedAt    time.Time
		UpdatedAt    time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "User is employee",
			fields: fields{
				ID:           uuid.New(),
				Email:        "employee@example.com",
				PasswordHash: "hash",
				Role:         RoleEmployee,
			},
			want: true,
		},
		{
			name: "User is moderator",
			fields: fields{
				ID:           uuid.New(),
				Email:        "moderator@example.com",
				PasswordHash: "hash",
				Role:         RoleModerator,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				ID:           tt.fields.ID,
				Email:        tt.fields.Email,
				PasswordHash: tt.fields.PasswordHash,
				Role:         tt.fields.Role,
				CreatedAt:    tt.fields.CreatedAt,
				UpdatedAt:    tt.fields.UpdatedAt,
			}
			if got := u.IsEmployee(); got != tt.want {
				t.Errorf("IsEmployee() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_IsModerator(t *testing.T) {
	type fields struct {
		ID           uuid.UUID
		Email        string
		PasswordHash string
		Role         string
		CreatedAt    time.Time
		UpdatedAt    time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "User is moderator",
			fields: fields{
				ID:           uuid.New(),
				Email:        "moderator@example.com",
				PasswordHash: "hash",
				Role:         RoleModerator,
			},
			want: true,
		},
		{
			name: "User is employee",
			fields: fields{
				ID:           uuid.New(),
				Email:        "employee@example.com",
				PasswordHash: "hash",
				Role:         RoleEmployee,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				ID:           tt.fields.ID,
				Email:        tt.fields.Email,
				PasswordHash: tt.fields.PasswordHash,
				Role:         tt.fields.Role,
				CreatedAt:    tt.fields.CreatedAt,
				UpdatedAt:    tt.fields.UpdatedAt,
			}
			if got := u.IsModerator(); got != tt.want {
				t.Errorf("IsModerator() = %v, want %v", got, tt.want)
			}
		})
	}
}
