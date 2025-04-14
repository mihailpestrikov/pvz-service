package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"testing"
	"time"
)

func TestGenerateDummyToken(t *testing.T) {
	type args struct {
		role       string
		secret     string
		expiration time.Duration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Successful token generation with role moderator",
			args: args{
				role:       "moderator",
				secret:     "test-secret",
				expiration: time.Hour,
			},
			wantErr: false,
		},
		{
			name: "Successful token generation with role employee",
			args: args{
				role:       "employee",
				secret:     "test-secret",
				expiration: time.Hour,
			},
			wantErr: false,
		},
		{
			name: "Token generation with empty secret",
			args: args{
				role:       "moderator",
				secret:     "",
				expiration: time.Hour,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateDummyToken(tt.args.role, tt.args.secret, tt.args.expiration)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateDummyToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == "" {
				t.Errorf("GenerateDummyToken() returned empty token")
			}

			if !tt.wantErr {
				claims, err := ValidateToken(got, tt.args.secret)
				if err != nil {
					t.Errorf("GenerateDummyToken() generated token that cannot be validated: %v", err)
				}
				if claims.Role != tt.args.role {
					t.Errorf("GenerateDummyToken() role in claims = %v, want %v", claims.Role, tt.args.role)
				}
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	userID := uuid.New()

	type args struct {
		userID     uuid.UUID
		role       string
		secret     string
		expiration time.Duration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Successful token generation for user",
			args: args{
				userID:     userID,
				role:       "moderator",
				secret:     "test-secret",
				expiration: time.Hour,
			},
			wantErr: false,
		},
		{
			name: "Token generation with empty UUID",
			args: args{
				userID:     uuid.Nil,
				role:       "employee",
				secret:     "test-secret",
				expiration: time.Hour,
			},
			wantErr: false,
		},
		{
			name: "Token generation with negative expiration time",
			args: args{
				userID:     userID,
				role:       "employee",
				secret:     "test-secret",
				expiration: -time.Hour,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateToken(tt.args.userID, tt.args.role, tt.args.secret, tt.args.expiration)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == "" {
				t.Errorf("GenerateToken() returned empty token")
			}

			if !tt.wantErr && tt.args.expiration > 0 {
				claims, err := ValidateToken(got, tt.args.secret)
				if err != nil {
					t.Errorf("GenerateToken() generated token that cannot be validated: %v", err)
				}
				if claims.Role != tt.args.role {
					t.Errorf("GenerateToken() role in claims = %v, want %v", claims.Role, tt.args.role)
				}
				if claims.UserID != tt.args.userID {
					t.Errorf("GenerateToken() userID in claims = %v, want %v", claims.UserID, tt.args.userID)
				}
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	validToken, _ := GenerateToken(userID, "moderator", secret, time.Hour)

	expiredToken, _ := GenerateToken(userID, "employee", secret, -time.Hour)

	tokenWithDifferentSecret, _ := GenerateToken(userID, "moderator", "different-secret", time.Hour)

	type args struct {
		tokenString string
		secret      string
	}
	tests := []struct {
		name    string
		args    args
		want    *Claims
		wantErr bool
	}{
		{
			name: "Valid token",
			args: args{
				tokenString: validToken,
				secret:      secret,
			},
			want: &Claims{
				UserID: userID,
				Role:   "moderator",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
				},
			},
			wantErr: false,
		},
		{
			name: "Expired token",
			args: args{
				tokenString: expiredToken,
				secret:      secret,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Incorrect secret",
			args: args{
				tokenString: validToken,
				secret:      "wrong-secret",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Token signed with a different secret",
			args: args{
				tokenString: tokenWithDifferentSecret,
				secret:      secret,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Empty token",
			args: args{
				tokenString: "",
				secret:      secret,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid token format",
			args: args{
				tokenString: "not-a-jwt-token",
				secret:      secret,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateToken(tt.args.tokenString, tt.args.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want == nil {
				if got != nil {
					t.Errorf("ValidateToken() got = %v, want nil", got)
				}
				return
			}

			if got != nil {
				if got.UserID != tt.want.UserID {
					t.Errorf("ValidateToken() got UserID = %v, want %v", got.UserID, tt.want.UserID)
				}
				if got.Role != tt.want.Role {
					t.Errorf("ValidateToken() got Role = %v, want %v", got.Role, tt.want.Role)
				}
			}
		})
	}
}
