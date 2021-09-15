package fmtd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	yaml "gopkg.in/yaml.v2"
)

// Config is the object which will hold all of the config parameters
type Config struct {
	DefaultLogDir	bool 	`yaml:"DefaultLogDir"`
	LogFileDir 		string	`yaml:"LogFileDir"`
	ConsoleOutput	bool	`yaml:"ConsoleOutput"`
	GrpcPort		int64	`yaml:"GrpcPort"`
}

// default_config returns the default configuration
// default_log_dir returns the default log directory
// default_grpc_port is the the default grpc port
var (
	default_grpc_port int64 = 7777
	default_log_dir = func() string {
		home_dir, err := os.UserHomeDir() // this should be OS agnostic
		if err != nil {
			log.Fatal(err)
		}
		return home_dir+"/.fmtd"
	}
	default_config = func() Config {
		return Config{
			DefaultLogDir: true,
			LogFileDir: default_log_dir(),
			ConsoleOutput: false,
			GrpcPort: default_grpc_port,
		}
	}
)

// InitConfig returns the `Config` struct with either default values or values specified in `config.yaml`
func InitConfig() Config {
	filename, _ := filepath.Abs(default_log_dir()+"/config.yaml")
	config_file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println(err)
		return default_config()
	}
	var config Config
	err = yaml.Unmarshal(config_file, &config)
	if err != nil {
		log.Println(err)
		config = default_config()
	} else {
		// Need to check if any config parameters aren't defined in `config.yaml` and assign them a default value
		config = check_yaml_config(config)
	}
	return config
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
		case "GrpcPort":
			if f.Int() == 0 { // This may end up being a range of values
				change_field(f, default_grpc_port)
			}
		}
	}
	return config
}
