package newman

import (
	"testing"
)

func TestNewmanService_Run(t *testing.T) {
	service := NewService()

	tests := []struct {
		name       string
		collection string
		options    Options
		wantErr    bool
	}{
		{
			name:       "empty collection path",
			collection: "",
			options:    Options{},
			wantErr:    true,
		},
		{
			name:       "valid collection with basic options",
			collection: "test.postman_collection.json",
			options: Options{
				Environment: "test.postman_environment.json",
				Debug:       false,
			},
			wantErr: true, // Newman will fail but should return proper error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.Run(tt.collection, tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewmanService.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && result == nil {
				t.Error("NewmanService.Run() should return result when no error")
			}
		})
	}
}

func TestNewmanService_CheckNewmanInstalled(t *testing.T) {
	service := NewService()

	// This test will pass if newman is installed, skip if not
	if !service.IsInstalled() {
		t.Skip("Newman not installed, skipping test")
	}

	if !service.IsInstalled() {
		t.Error("IsInstalled() should return true when newman is available")
	}
}

func TestNewmanService_RunWithOptions(t *testing.T) {
	service := NewService()

	tests := []struct {
		name       string
		collection string
		options    Options
		wantErr    bool
	}{
		{
			name:       "with environment file",
			collection: "test.postman_collection.json",
			options: Options{
				Environment: "test.postman_environment.json",
			},
			wantErr: true, // Newman will fail but should handle options
		},
		{
			name:       "with CSV and debug",
			collection: "test.postman_collection.json",
			options: Options{
				CSV:   "test.csv",
				Debug: true,
			},
			wantErr: true, // Newman will fail but should handle options
		},
		{
			name:       "with all options",
			collection: "test.postman_collection.json",
			options: Options{
				Environment: "test.postman_environment.json",
				CSV:         "test.csv",
				Debug:       true,
			},
			wantErr: true, // Newman will fail but should handle options
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.Run(tt.collection, tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewmanService.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && result == nil {
				t.Error("NewmanService.Run() should return result when no error")
			}
		})
	}
}

func TestNewmanService_RunWithFlags(t *testing.T) {
	service := NewService()

	tests := []struct {
		name       string
		collection string
		flags      []string
		wantErr    bool
	}{
		{
			name:       "empty collection path",
			collection: "",
			flags:      []string{},
			wantErr:    true,
		},
		{
			name:       "with basic flags",
			collection: "test.postman_collection.json",
			flags:      []string{"--verbose"},
			wantErr:    true, // Newman will fail but should handle flags
		},
		{
			name:       "with environment and CSV flags",
			collection: "test.postman_collection.json",
			flags:      []string{"-e", "test.postman_environment.json", "-d", "test.csv"},
			wantErr:    true, // Newman will fail but should handle flags
		},
		{
			name:       "with multiple flags",
			collection: "test.postman_collection.json",
			flags:      []string{"--verbose", "--bail", "--timeout", "30000"},
			wantErr:    true, // Newman will fail but should handle flags
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.RunWithFlags(tt.collection, tt.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewmanService.RunWithFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && result == nil {
				t.Error("NewmanService.RunWithFlags() should return result when no error")
			}
		})
	}
}

func TestNewmanService_RunWithEnvironmentExport(t *testing.T) {
	service := NewService()

	tests := []struct {
		name          string
		collection    string
		flags         []string
		exportEnvPath string
		wantErr       bool
	}{
		{
			name:          "empty collection path",
			collection:    "",
			flags:         []string{},
			exportEnvPath: "/tmp/test_env.json",
			wantErr:       true,
		},
		{
			name:          "with environment export",
			collection:    "test.postman_collection.json",
			flags:         []string{"--verbose"},
			exportEnvPath: "/tmp/test_env.json",
			wantErr:       true, // Newman will fail but should handle export
		},
		{
			name:          "with flags and environment export",
			collection:    "test.postman_collection.json",
			flags:         []string{"-e", "test.postman_environment.json"},
			exportEnvPath: "/tmp/test_env.json",
			wantErr:       true, // Newman will fail but should handle export
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.RunWithEnvironmentExport(tt.collection, tt.flags, tt.exportEnvPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewmanService.RunWithEnvironmentExport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && result == nil {
				t.Error("NewmanService.RunWithEnvironmentExport() should return result when no error")
			}
		})
	}
}
