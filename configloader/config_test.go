package configloader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// envVarNames lists all environment variables that configloader reads.
var envVarNames = []string{
	"CONTROLLER_HOST", "CONTROLLER_LOGIN", "CONTROLLER_PASSWORD", "CONTROLLER_INSECURE",
	"NOPERATOR_DEV_MODE", "NOPERATOR_REQUEUE_INTERVAL", "NOPERATOR_CALICO_ASN_RANGE",
	"NOPERATOR_L4LB_TENANT", "NOPERATOR_VPC_ID",
}

// quarantineEnv saves the current values of all config-related environment
// variables, clears them for a clean test, and returns a function that restores
// the original values.
func quarantineEnv(t *testing.T) func() {
	t.Helper()
	saved := make(map[string]string)
	wasSet := make(map[string]bool)

	// Save current values
	for _, name := range envVarNames {
		if val, ok := os.LookupEnv(name); ok {
			saved[name] = val
			wasSet[name] = true
		}
	}

	// Clear all config env vars
	for _, name := range envVarNames {
		os.Unsetenv(name)
	}

	// Return restore function
	return func() {
		for _, name := range envVarNames {
			if wasSet[name] {
				os.Setenv(name, saved[name])
			} else {
				os.Unsetenv(name)
			}
		}
	}
}

// writeConfigFile creates a config.yml file with the given content and returns its path.
func writeConfigFile(t *testing.T, content string) string {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config.yml: %v", err)
	}
	return configPath
}

func TestLoad_Success(t *testing.T) {
	defer quarantineEnv(t)()

	configPath := writeConfigFile(t, `controller:
  host: http://test-controller.example.com
  login: testuser
  password: testpass
  insecure: true
logdevmode: true
requeueinterval: 30
calicoasnrange: 4230000000-4239999999
l4lbtenant: test-tenant
vpcid: 42
`)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.Controller.Host != "http://test-controller.example.com" {
		t.Errorf("Controller.Host = %q, want %q", cfg.Controller.Host, "http://test-controller.example.com")
	}
	if cfg.Controller.Login != "testuser" {
		t.Errorf("Controller.Login = %q, want %q", cfg.Controller.Login, "testuser")
	}
	if cfg.Controller.Password != "testpass" {
		t.Errorf("Controller.Password = %q, want %q", cfg.Controller.Password, "testpass")
	}
	if !cfg.Controller.Insecure {
		t.Error("Controller.Insecure = false, want true")
	}
	if !cfg.LogDevMode {
		t.Error("LogDevMode = false, want true")
	}
	if cfg.RequeueInterval != 30 {
		t.Errorf("RequeueInterval = %d, want %d", cfg.RequeueInterval, 30)
	}
	if cfg.CalicoASNRange != "4230000000-4239999999" {
		t.Errorf("CalicoASNRange = %q, want %q", cfg.CalicoASNRange, "4230000000-4239999999")
	}
	if cfg.L4lbTenant != "test-tenant" {
		t.Errorf("L4lbTenant = %q, want %q", cfg.L4lbTenant, "test-tenant")
	}
	if cfg.VPCID != 42 {
		t.Errorf("VPCID = %d, want %d", cfg.VPCID, 42)
	}
}

func TestLoad_MissingControllerHost(t *testing.T) {
	defer quarantineEnv(t)()

	configPath := writeConfigFile(t, `controller:
  login: testuser
  password: testpass
`)

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load() returned nil error, want error for missing controller host")
	}
}

func TestReadConfig_EnvironmentOverridesYAML(t *testing.T) {
	defer quarantineEnv(t)()

	// Set environment variables to override YAML values
	os.Setenv("CONTROLLER_HOST", "http://env-controller.example.com")
	os.Setenv("CONTROLLER_LOGIN", "envuser")
	os.Setenv("NOPERATOR_REQUEUE_INTERVAL", "60")

	cfg, err := readConfig(strings.NewReader(`controller:
  host: http://yaml-controller.example.com
  login: yamluser
requeueinterval: 10
`))
	if err != nil {
		t.Fatalf("readConfig() returned unexpected error: %v", err)
	}

	if cfg.Controller.Host != "http://env-controller.example.com" {
		t.Errorf("Controller.Host = %q, want %q (env should override YAML)", cfg.Controller.Host, "http://env-controller.example.com")
	}
	if cfg.Controller.Login != "envuser" {
		t.Errorf("Controller.Login = %q, want %q (env should override YAML)", cfg.Controller.Login, "envuser")
	}
	if cfg.RequeueInterval != 60 {
		t.Errorf("RequeueInterval = %d, want %d (env should override YAML)", cfg.RequeueInterval, 60)
	}
}
