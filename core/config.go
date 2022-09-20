/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/09/20
*/

package core

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"time"

	"github.com/SSSOC-CAN/fmtd/lanirpc"
	"github.com/SSSOC-CAN/fmtd/plugins"
	"github.com/SSSOC-CAN/fmtd/utils"
	flags "github.com/jessevdk/go-flags"
	e "github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// Config is the object which will hold all of the config parameters
type Config struct {
	DefaultLogDir       bool                    `yaml:"DefaultLogDir"`
	LogFileDir          string                  `yaml:"LogFileDir" long:"logfiledir" description:"Choose the directory where the log file is stored"`
	MaxLogFiles         int64                   `yaml:"MaxLogFiles" long:"maxlogfiles" description:"Maximum number of logfiles in the log rotation (0 for no rotation)"`
	MaxLogFileSize      int64                   `yaml:"MaxLogFileSize" long:"maxlogfilesize" description:"Maximum size of a logfile in MB"`
	ConsoleOutput       bool                    `yaml:"ConsoleOutput" long:"consoleoutput" description:"Whether log information is printed to the console"`
	GrpcPort            int64                   `yaml:"GrpcPort" long:"grpc_port" description:"The port where Laniakea listens for gRPC API requests"`
	RestPort            int64                   `yaml:"RestPort" long:"rest_port" description:"The port where Laniakea listens for REST API requests"`
	DataOutputDir       string                  `yaml:"DataOutputDir" long:"dataoutputdir" description:"Choose the directory where the recorded data is stored"`
	ExtraIPAddr         []string                `yaml:"ExtraIPAddr" long:"tlsextraip" description:"Adds an extra ip to the generated certificate"` // optional parameter
	PluginDir           string                  `yaml:"PluginDir" long:"plugindir" description:"The directory where plugin executables will live and be run from. Must be absolute path"`
	Plugins             []*lanirpc.PluginConfig `yaml:"Plugins"`
	RegenerateMacaroons bool                    `long:"regenmacaroons" description:"Boolean to determine whether macaroons are regenerated"`
	MacaroonDBPath      string
	TLSCertPath         string
	TLSKeyPath          string
	AdminMacPath        string
	TestMacPath         string
	WSPingInterval      time.Duration
	WSPongWait          time.Duration
}

// default_config returns the default configuration
// default_log_dir returns the default log directory
// default_grpc_port is the the default grpc port
var (
	config_file_name  string = "config.yaml"
	default_grpc_port int64  = 7777
	default_rest_port int64  = 8080
	default_log_dir          = func() string {
		return utils.AppDataDir("laniakea", false)
	}
	default_macaroon_db_file    string = default_log_dir() + "/macaroon.db"
	default_tls_cert_path       string = default_log_dir() + "/tls.cert"
	default_tls_key_path        string = default_log_dir() + "/tls.key"
	default_admin_macaroon_path string = default_log_dir() + "/admin.macaroon"
	test_macaroon_path          string = default_log_dir() + "/test.macaroon"
	default_data_output_dir     string = default_log_dir()
	default_ws_ping_interval           = time.Second * 30
	default_ws_pong_wait               = time.Second * 5
	default_log_file_size       int64  = 10
	default_max_log_files       int64  = 0
	default_plugin_dir          string = default_log_dir()
	default_config                     = func() Config {
		return Config{
			DefaultLogDir:  true,
			LogFileDir:     default_log_dir(),
			MaxLogFiles:    default_max_log_files,
			MaxLogFileSize: default_log_file_size,
			ConsoleOutput:  true,
			GrpcPort:       default_grpc_port,
			RestPort:       default_rest_port,
			DataOutputDir:  default_data_output_dir,
			MacaroonDBPath: default_macaroon_db_file,
			TLSCertPath:    default_tls_cert_path,
			TLSKeyPath:     default_tls_key_path,
			AdminMacPath:   default_admin_macaroon_path,
			TestMacPath:    test_macaroon_path,
			WSPingInterval: default_ws_ping_interval,
			WSPongWait:     default_ws_pong_wait,
			PluginDir:      default_plugin_dir,
		}
	}
)

// InitConfig returns the `Config` struct with either default values or values specified in `config.yaml`
func InitConfig(isTesting bool) (Config, error) {
	// Check if laniakea directory exists, if no then create it
	if !utils.FileExists(utils.AppDataDir("laniakea", false)) {
		err := os.Mkdir(utils.AppDataDir("laniakea", false), 0700)
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
					err = plugins.ValidatePluginConfig(cfg, config.PluginDir, isTesting)
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
