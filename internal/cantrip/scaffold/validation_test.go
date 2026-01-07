package scaffold

import (
	"strings"
	"testing"
)

func TestModuleOptionsValidate(t *testing.T) {
	tests := []struct {
		name    string
		opts    ModuleOptions
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid options",
			opts: ModuleOptions{
				Name:        "user",
				Transports:  []string{"http"},
				APIType:     "json",
				Persistence: "",
			},
			wantErr: false,
		},
		{
			name: "valid with all fields",
			opts: ModuleOptions{
				Name:        "order_item",
				Transports:  []string{"http", "grpc", "amqp"},
				APIType:     "html",
				Persistence: "postgres",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			opts: ModuleOptions{
				Name:       "",
				Transports: []string{"http"},
				APIType:    "json",
			},
			wantErr: true,
			errMsg:  "module name is required",
		},
		{
			name: "invalid name - starts with number",
			opts: ModuleOptions{
				Name:       "1user",
				Transports: []string{"http"},
				APIType:    "json",
			},
			wantErr: true,
			errMsg:  "invalid module name",
		},
		{
			name: "invalid name - uppercase",
			opts: ModuleOptions{
				Name:       "User",
				Transports: []string{"http"},
				APIType:    "json",
			},
			wantErr: true,
			errMsg:  "invalid module name",
		},
		{
			name: "invalid name - special chars",
			opts: ModuleOptions{
				Name:       "user-name",
				Transports: []string{"http"},
				APIType:    "json",
			},
			wantErr: true,
			errMsg:  "invalid module name",
		},
		{
			name: "invalid transport",
			opts: ModuleOptions{
				Name:       "user",
				Transports: []string{"ftp"},
				APIType:    "json",
			},
			wantErr: true,
			errMsg:  "invalid transport",
		},
		{
			name: "invalid API type",
			opts: ModuleOptions{
				Name:       "user",
				Transports: []string{"http"},
				APIType:    "xml",
			},
			wantErr: true,
			errMsg:  "invalid API type",
		},
		{
			name: "invalid persistence",
			opts: ModuleOptions{
				Name:        "user",
				Transports:  []string{"http"},
				APIType:     "json",
				Persistence: "mysql",
			},
			wantErr: true,
			errMsg:  "invalid persistence",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}
