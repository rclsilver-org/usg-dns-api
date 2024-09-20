package db

import "testing"

func Test_validateName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "example", wantErr: false},
		{name: "example.com", wantErr: false},
		{name: "sub.example.com", wantErr: false},
		{name: "sub-01.example.com", wantErr: false},
		{name: "01.example.com", wantErr: false},
		{name: "-abc.example.com", wantErr: true},
		{name: "sub.example.com.", wantErr: true},
		{name: "abcdefghijklmnopqrstuvwxyz-abcdefghijklmnopqrstuvwxyz-123456789", wantErr: false}, // 63 chars
		{name: "abcdefghijklmnopqrstuvwxyz-abcdefghijklmnopqrstuvwxyz-1234567890", wantErr: true}, // 64 chars
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateName(tt.name); (err != nil) != tt.wantErr {
				t.Errorf("validateName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
