package notmuch

import (
	"github.com/emersion/go-imap/backend"
	"golang.org/x/crypto/bcrypt"
	"reflect"
	"testing"
)

func TestBackend_Login(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)
	if err != nil {
		t.Errorf(err.Error())
	}
	be := Backend{
		user: &User{
			username: "test",
			password: string(hash),
		},
	}

	tests := []struct {
		name     string
		username string
		password string
		want     backend.User
		wantErr  bool
	}{
		{
			name:     "correct username/password",
			username: "test",
			password: "test",
			want:     be.user,
		},
		{
			name:     "incorrect username",
			username: "incorrect",
			password: "test",
			wantErr:  true,
		},
		{
			name:     "incorrect password",
			username: "test",
			password: "incorrect",
			wantErr:  true,
		},
		{
			name:     "incorrect username & password",
			username: "incorrect",
			password: "incorrect",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := be.Login(nil, tt.username, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Login() got = %v, want %v", got, tt.want)
			}
		})
	}
}
