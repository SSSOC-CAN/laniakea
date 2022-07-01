/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package fmtd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"time"

	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/utils"
	bg "github.com/SSSOCPaulCote/blunderguard"
	flags "github.com/jessevdk/go-flags"
	e "github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

const (
	ErrInvalidPluginName  = bg.Error("invalid plugin name")
	ErrInvalidPluginExec  = bg.Error("invalid plugin executable")
	ErrPluginExecNotFound = bg.Error("plugin executable not found")
)

type PluginConfig struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	ExecName    string `yaml:"exec"`
	Timeout     int64  `yaml:"timeout"`
	MaxTimeouts int64  `yaml:"maxTimeouts"`
}

// Config is the object which will hold all of the config parameters
type Config struct {
	DefaultLogDir  bool            `yaml:"DefaultLogDir"`
	LogFileDir     string          `yaml:"LogFileDir" long:"logfiledir" description:"Choose the directory where the log file is stored"`
	MaxLogFiles    int64           `yaml:"MaxLogFiles" long:"maxlogfiles" description:"Maximum number of logfiles in the log rotation (0 for no rotation)"`
	MaxLogFileSize int64           `yaml:"MaxLogFileSize" long:"maxlogfilesize" description:"Maximum size of a logfile in MB"`
	ConsoleOutput  bool            `yaml:"ConsoleOutput" long:"consoleoutput" description:"Whether log information is printed to the console"`
	GrpcPort       int64           `yaml:"GrpcPort" long:"grpc_port" description:"The port where the fmtd listens for gRPC API requests"`
	RestPort       int64           `yaml:"RestPort" long:"rest_port" description:"The port where the fmtd listens for REST API requests"`
	TCPPort        int64           `yaml:"TCPPort" long:"tcp_port" description:"The port where the fmtd listens for TCP requests"`
	TCPAddr        string          `yaml:"TCPAddr" long:"tcp_addr" description:"The address where the fmtd listens for TCP requests"`
	DataOutputDir  string          `yaml:"DataOutputDir" long:"dataoutputdir" description:"Choose the directory where the recorded data is stored"`
	ExtraIPAddr    []string        `yaml:"ExtraIPAddr" long:"tlsextraip" description:"Adds an extra ip to the generated certificate"` // optional parameter
	InfluxURL      string          `yaml:"InfluxURL" long:"influxurl" description:"The InfluxDB URL for writing"`
	InfluxAPIToken string          `yaml:"InfluxAPIToken" long:"influxapitoken" description:"The InfluxDB API Token used to read and write"`
	PluginDir      string          `yaml:"PluginDir" long:"plugindir" description:"The directory where plugin executables will live and be run from. Must be absolute path"`
	Plugins        []*PluginConfig `yaml:"Plugins"`
	MacaroonDBPath string
	TLSCertPath    string
	TLSKeyPath     string
	AdminMacPath   string
	TestMacPath    string
	WSPingInterval time.Duration
	WSPongWait     time.Duration
}

// default_config returns the default configuration
// default_log_dir returns the default log directory
// default_grpc_port is the the default grpc port
var (
	config_file_name  string = "config.yaml"
	default_grpc_port int64  = 7777
	default_rest_port int64  = 8080
	default_tcp_port  int64  = 10024
	default_tcp_addr  string = "0.0.0.0"
	default_log_dir          = func() string {
		return utils.AppDataDir("fmtd", false)
	}
	default_macaroon_db_file    string = default_log_dir() + "/macaroon.db"
	default_tls_cert_path       string = default_log_dir() + "/tls.cert"
	default_tls_key_path        string = default_log_dir() + "/tls.key"
	default_admin_macaroon_path string = default_log_dir() + "/admin.macaroon"
	test_macaroon_path          string = default_log_dir() + "/test.macaroon"
	default_data_output_dir     string = default_log_dir()
	default_ws_ping_interval           = time.Second * 30
	default_ws_pong_wait               = time.Second * 5
	default_influx_url          string = "https://174.113.21.199:8088"
	default_log_file_size       int64  = 10
	default_max_log_files       int64  = 0
	default_plugin_dir          string = default_log_dir()
	default_plugin_timeout      int64  = 30
	default_plugin_max_timeouts int64  = 3
	default_plugin_version      string = "1.0"
	default_config                     = func() Config {
		return Config{
			DefaultLogDir:  true,
			LogFileDir:     default_log_dir(),
			MaxLogFiles:    default_max_log_files,
			MaxLogFileSize: default_log_file_size,
			ConsoleOutput:  true,
			GrpcPort:       default_grpc_port,
			RestPort:       default_rest_port,
			TCPPort:        default_tcp_port,
			TCPAddr:        default_tcp_addr,
			DataOutputDir:  default_data_output_dir,
			MacaroonDBPath: default_macaroon_db_file,
			TLSCertPath:    default_tls_cert_path,
			TLSKeyPath:     default_tls_key_path,
			AdminMacPath:   default_admin_macaroon_path,
			TestMacPath:    test_macaroon_path,
			WSPingInterval: default_ws_ping_interval,
			WSPongWait:     default_ws_pong_wait,
			InfluxURL:      default_influx_url,
			PluginDir:      default_plugin_dir,
		}
	}
)

// InitConfig returns the `Config` struct with either default values or values specified in `config.yaml`
func InitConfig(isTesting bool) (Config, error) {
	// Check if fmtd directory exists, if no then create it
	if !utils.FileExists(utils.AppDataDir("fmtd", false)) {
		err := os.Mkdir(utils.AppDataDir("fmtd", false), 0700)
		if err != nil {
			log.Println(err)
		}
	}
	config := default_config()
	if utils.FileExists(path.Join(default_log_dir(), config_file_name)) {
		filename, _ := filepath.Abs(path.Join(default_log_dir(), config_file_name))
		config_file, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Println(err)
			return default_config(), nil
		}
		err = yaml.Unmarshal(config_file, &config)
		if err != nil {
			log.Println(err)
			config = default_config()
		} else {
			// Need to check if any config parameters aren't defined in `config.yaml` and assign them a default value
			config = check_yaml_config(config)
			if len(config.Plugins) > 0 {
				for _, cfg := range config.Plugins {
					err = validatePluginConfig(cfg, config.PluginDir)
					if err != nil {
						return Config{}, e.Wrapf(err, "Error validating %s plugin", cfg.Name)
					}
				}
			}
		}
		config.WSPingInterval = default_ws_ping_interval
		config.WSPongWait = default_ws_pong_wait
	}
	// now to parse the flags
	if !isTesting {
		if _, err := flags.Parse(&config); err != nil {
			return Config{}, err
		}
	}
	return config, nil
}

// validatePluginConfig takes a PluginConfig and validates the parameters
func validatePluginConfig(cfg *PluginConfig, pluginDir string) error {
	// checks that the plugin has a valid name and executable
	if !utils.ValidatePluginName(cfg.Name) {
		return ErrInvalidPluginName
	} else if !utils.ValidatePluginExec(cfg.ExecName) {
		return ErrInvalidPluginExec
	}
	// check that the plugin type is either a datasource or controller
	if cfg.Type != "datasource" && cfg.Type != "controller" {
		return errors.ErrInvalidPluginType
	}
	if !utils.FileExists(filepath.Join(pluginDir, cfg.ExecName)) {
		return ErrPluginExecNotFound
	}
	// if timeout or maxtimeout or version are 0 set to default values
	if cfg.Timeout == int64(0) {
		cfg.Timeout = default_plugin_timeout
	}
	if cfg.MaxTimeouts == int64(0) {
		cfg.MaxTimeouts = default_plugin_max_timeouts
	}
	return nil
}

// change_field changes the value of a specified field from the config struct
func change_field(field reflect.Value, new_value interface{}) {
	if field.IsValid() {
		if field.CanSet() {
			f := field.Kind()
			switch f {
			case reflect.String:
				if v, ok := new_value.(string); ok {
					field.SetString(v)
				} else {
					log.Fatal(fmt.Sprintf("Type of new_value: %v does not match the type of the field: string", new_value))
				}
			case reflect.Bool:
				if v, ok := new_value.(bool); ok {
					field.SetBool(v)
				} else {
					log.Fatal(fmt.Sprintf("Type of new_value: %v does not match the type of the field: bool", new_value))
				}
			case reflect.Int64:
				if v, ok := new_value.(int64); ok {
					field.SetInt(v)
				} else {
					log.Fatal(fmt.Sprintf("Type of new_value: %v does not match the type of the field: int64", new_value))
				}
			}
		}
	}
}

// check_yaml_config iterates over the Config struct fields and changes blank fields to default values
func check_yaml_config(config Config) Config {
	pv := reflect.ValueOf(&config)
	v := pv.Elem()
	field_names := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		field_name := field_names.Field(i).Name
		switch field_name {
		case "LogFileDir":
			if f.String() == "" {
				change_field(f, default_log_dir())
				dld := v.FieldByName("DefaultLogDir")
				change_field(dld, true)
			}
		case "MaxLogFileSize":
			if f.Int() == 0 {
				change_field(f, default_log_file_size)
			}
		case "GrpcPort":
			if f.Int() == 0 { // This may end up being a range of values
				change_field(f, default_grpc_port)
			}
		case "RestPort":
			if f.Int() == 0 { // This may end up being a range of values
				change_field(f, default_rest_port)
			}
		case "TCPPort":
			if f.Int() == 0 {
				change_field(f, default_tcp_port)
			}
		case "TCPAddr":
			if f.String() == "" {
				change_field(f, default_tcp_addr)
			}
		case "MacaroonDBPath":
			if f.String() == "" {
				change_field(f, default_macaroon_db_file)
			}
		case "TLSCertPath":
			if f.String() == "" {
				change_field(f, default_tls_cert_path)
				tls_key := v.FieldByName("TLSKeyPath")
				change_field(tls_key, default_tls_key_path)
			}
		case "TLSKeyPath":
			if f.String() == "" {
				change_field(f, default_tls_key_path)
				tls_cert := v.FieldByName("TLSCertPath")
				change_field(tls_cert, default_tls_cert_path)
			}
		case "AdminMacPath":
			if f.String() == "" {
				change_field(f, default_admin_macaroon_path)
			}
		case "TestMacPath":
			if f.String() == "" {
				change_field(f, test_macaroon_path)
			}
		case "DataOutputDir":
			if f.String() == "" {
				change_field(f, default_data_output_dir)
			}
		case "InfluxURL":
			if f.String() == "" {
				change_field(f, default_influx_url)
			}
		case "PluginDir":
			if f.String() == "" {
				change_field(f, default_plugin_dir)
			}
		default:
			continue
		}
	}
	return config
}
