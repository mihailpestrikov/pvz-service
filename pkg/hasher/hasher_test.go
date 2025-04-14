package hasher

import "testing"

func TestHash(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Successful password hashing",
			args:    args{password: "password123"},
			wantErr: false,
		},
		{
			name:    "Hashing empty password",
			args:    args{password: ""},
			wantErr: false,
		},
		{
			name:    "Hashing long password",
			args:    args{password: "thisIsAVeryLongPasswordThatShouldStillBeHashedProperly123456789"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Hash(tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got == "" {
				t.Errorf("Hash() returned empty string")
			}
			if got == tt.args.password {
				t.Errorf("Hash() returned original password")
			}

			if !Verify(got, tt.args.password) {
				t.Errorf("Hash() produced hash that cannot be verified")
			}
		})
	}
}

func TestVerify(t *testing.T) {
	hash1, _ := Hash("password123")
	hash2, _ := Hash("anotherPassword")
	hash3, _ := Hash("")

	type args struct {
		passwordHash string
		password     string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Correct password",
			args: args{
				passwordHash: hash1,
				password:     "password123",
			},
			want: true,
		},
		{
			name: "Wrong password",
			args: args{
				passwordHash: hash1,
				password:     "wrongPassword",
			},
			want: false,
		},
		{
			name: "Empty password, but correct hash",
			args: args{
				passwordHash: hash3,
				password:     "",
			},
			want: true,
		},
		{
			name: "Empty hash",
			args: args{
				passwordHash: "",
				password:     "password123",
			},
			want: false,
		},
		{
			name: "Incorrect hash format",
			args: args{
				passwordHash: "not-a-hash",
				password:     "password123",
			},
			want: false,
		},
		{
			name: "Hash from another password",
			args: args{
				passwordHash: hash2,
				password:     "password123",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Verify(tt.args.passwordHash, tt.args.password); got != tt.want {
				t.Errorf("Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}
