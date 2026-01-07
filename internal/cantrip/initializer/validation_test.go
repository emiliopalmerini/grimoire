package initializer

import (
	"strings"
	"testing"
)

func TestProjectOptionsValidate(t *testing.T) {
	tests := []struct {
		name    string
		opts    ProjectOptions
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid options",
			opts: ProjectOptions{
				Name:       "myapp",
				ModulePath: "github.com/user/myapp",
				GoVersion:  "1.25",
				Type:       "api",
				Transports: []string{"http"},
			},
			wantErr: false,
		},
		{
			name: "valid with hyphen in name",
			opts: ProjectOptions{
				Name:       "my-app",
				ModulePath: "github.com/user/my-app",
				GoVersion:  "1.25",
				Type:       "web",
				Transports: []string{"http"},
			},
			wantErr: false,
		},
		{
			name: "valid with underscore in name",
			opts: ProjectOptions{
				Name:       "my_app",
				ModulePath: "github.com/user/my_app",
				GoVersion:  "1.25",
				Type:       "grpc",
				Transports: []string{"grpc"},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			opts: ProjectOptions{
				Name:       "",
				ModulePath: "github.com/user/app",
				Type:       "api",
				Transports: []string{"http"},
			},
			wantErr: true,
			errMsg:  "project name is required",
		},
		{
			name: "invalid name - starts with number",
			opts: ProjectOptions{
				Name:       "1app",
				ModulePath: "github.com/user/1app",
				Type:       "api",
				Transports: []string{"http"},
			},
			wantErr: true,
			errMsg:  "invalid project name",
		},
		{
			name: "invalid name - uppercase",
			opts: ProjectOptions{
				Name:       "MyApp",
				ModulePath: "github.com/user/MyApp",
				Type:       "api",
				Transports: []string{"http"},
			},
			wantErr: true,
			errMsg:  "invalid project name",
		},
		{
			name: "invalid project type",
			opts: ProjectOptions{
				Name:       "app",
				ModulePath: "github.com/user/app",
				Type:       "cli",
				Transports: []string{"http"},
			},
			wantErr: true,
			errMsg:  "invalid project type",
		},
		{
			name: "invalid transport",
			opts: ProjectOptions{
				Name:       "app",
				ModulePath: "github.com/user/app",
				Type:       "api",
				Transports: []string{"websocket"},
			},
			wantErr: true,
			errMsg:  "invalid transport",
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
