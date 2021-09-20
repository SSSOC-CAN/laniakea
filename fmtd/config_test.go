package fmtd

import (
	"fmt"
	"os"
	"testing"
)

// TestInitConfigNoYAML ensures that if no .yaml is found, a default config is produced
func TestInitConfigNoYAML(t *testing.T) {
	home_dir, err := os.UserHomeDir() // this should be OS agnostic
	if err != nil {
		t.Errorf("%s", err)
	}
	if _, err = os.Stat(home_dir+"/.fmtd/config.yaml"); err == nil {
		err = os.Remove(home_dir+"/.fmtd/config.yaml")
		if err != nil {
			t.Errorf("%s", err)
		}
	}
	config, err := InitConfig()
	if err != nil {
		t.Errorf("%s", err)
	}
	if config != default_config() {
		t.Errorf("InitConfig did not produce a default config when config.yaml was not present")
	}
}

// TestInitConfigFromYAML ensures that a InitConfig properly reads config files
func TestInitConfigFromYAML(t *testing.T) {
	// first check if config.yaml exists
	home_dir, err := os.UserHomeDir() // this should be OS agnostic
	if err != nil {
		t.Errorf("%s", err)
	}
	config_file, err := os.OpenFile(home_dir+"/.fmtd/config.yaml", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
	if err != nil {
		// might have to create the .fmtd directory and try again
		err = os.Mkdir(home_dir+"/.fmtd", 0775)
		if err != nil {
			t.Errorf("%s", err)
		}
		config_file, err = os.OpenFile(home_dir+"/.fmtd/config.yaml", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
		if err != nil {
			t.Errorf("%s", err)
		}
	}
	// write to yaml file
	d_config := Config{
		DefaultLogDir: false,
		LogFileDir: "/home/vagrant/documents",
		ConsoleOutput: true,
		GrpcPort: 3567,
		MacaroonDBPath: default_macaroon_db_file,
		TLSCertPath: default_tls_cert_path,
		TLSKeyPath: default_tls_key_path,
		AdminMacPath: default_admin_macaroon_path,
	}
	_, err = config_file.WriteString(fmt.Sprintf("DefaultLogDir: %v\n", d_config.DefaultLogDir))
	if err != nil {
		t.Errorf("%s", err)
	}
	_, err = config_file.WriteString(fmt.Sprintf("LogFileDir: %v\n", d_config.LogFileDir))
	if err != nil {
		t.Errorf("%s", err)
	}
	_, err = config_file.WriteString(fmt.Sprintf("ConsoleOutput: %v\n", d_config.ConsoleOutput))
	if err != nil {
		t.Errorf("%s", err)
	}
	_, err = config_file.WriteString(fmt.Sprintf("GrpcPort: %v", d_config.GrpcPort))
	if err != nil {
		t.Errorf("%s", err)
	}
	config_file.Sync()
	config_file.Close()
	config, err := InitConfig()
	if err != nil {
		t.Errorf("%s", err)
	}
	if config != d_config {
		t.Errorf("InitConfig did not properly read the config file: %v", config)
	}
}

// TestDefaultLogDir tests that default_log_dir returns the expected default log directory
func TestDefaultLogDir(t *testing.T) {
	home_dir, err := os.UserHomeDir() // this should be OS agnostic
	if err != nil {
		t.Errorf("%s", err)
	}
	log_dir := home_dir+"/.fmtd"
	if log_dir != default_log_dir() {
		t.Errorf("default_log_dir not returning expected directory. Expected: %s\tReceived: %s", log_dir, default_log_dir())
	}
}

// TestDefaultConfig checks if default_config does return the expected default config struct
func TestDefaultConfig(t *testing.T) {
	d_config := Config{
		DefaultLogDir: true,
		LogFileDir: default_log_dir(),
		ConsoleOutput: false,
		GrpcPort: 7777,
		MacaroonDBPath: default_macaroon_db_file,
		TLSCertPath: default_tls_cert_path,
		TLSKeyPath: default_tls_key_path,
		AdminMacPath: default_admin_macaroon_path,
	}
	if d_config != default_config() {
		t.Errorf("default_config not returning expected config. Expected: %v\tReceived: %v", d_config, default_config())
	}
}