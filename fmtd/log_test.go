package fmtd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// TestInitLoggerOutput makes sure both console and logfile output work
func TestInitLoggerOutput(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	// start with True
	config := Config{
		DefaultLogDir: true,
		LogFileDir: default_log_dir(),
		ConsoleOutput: true,
	}
	log, err := InitLogger(&config)
	if err != nil {
		t.Errorf("%s", err)
	}
	log.Info().Msg("Testing both outputs...")
	// False
	config = Config{
		DefaultLogDir: true,
		LogFileDir: default_log_dir(),
		ConsoleOutput: false,
	}
	log, err = InitLogger(&config)
	if err != nil {
		t.Errorf("%s", err)
	}
	log.Info().Msg("This shouldn't appear in the console...")

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdout = old
	out := <-outC //reading last line of console output
	
	if strings.Contains(out, "This shouldn't appear in the console...") {
		t.Errorf("InitLogger produced a logger that prints to console when it shouldn't")
	}
}

// TestNewSubLogger tests to ensure a NewSubLogger can be created and behaves as expected
func TestNewSubLogger(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	// true first
	config := Config{
		DefaultLogDir: true,
		LogFileDir: default_log_dir(),
		ConsoleOutput: true,
	}
	log, err := InitLogger(&config)
	if err != nil {
		t.Errorf("%s", err)
	}
	test_sub_log := NewSubLogger(&log, "TEST")
	test_sub_log.SubLogger.Info().Msg("Testing both outputs...")
	// false
	config = Config{
		DefaultLogDir: true,
		LogFileDir: default_log_dir(),
		ConsoleOutput: false,
	}
	log, err = InitLogger(&config)
	if err != nil {
		t.Errorf("%s", err)
	}
	test_sub_log = NewSubLogger(&log, "TEST")
	test_sub_log.SubLogger.Info().Msg("This shouldn't appear in the console...")
	
	outC := make(chan string)

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdout = old
	out := <-outC //reading last line of console output
	
	if strings.Contains(out, "This shouldn't appear in the console...") {
		t.Errorf("NewSubLogger produced a logger that prints to console when it shouldn't")
	}
}

// TestLogWithErrors ensures that if a bad LogLevel is provided to Log, it will log an error
func TestLogWithErrors(t *testing.T) {
	config, err := InitConfig()
	if err != nil {
		t.Errorf("%s", err)
	}
	log, err := InitLogger(&config)
	if err != nil {
		t.Errorf("%s", err)
	}
	test_sub_log := NewSubLogger(&log, "TEST")
	tables := []struct{
		level	string
		msg		string
	}{
		{"INFO", "T1"},
		{"DEBUG", "T2"},
		{"TRACE", "T3"},
		{"ERROR", "T4"},
		{"FATAL", "T5"},
		{"PANIC", "T6"},
		{"TEST", "T7"},
	}
	var level, msg string
	for i, table := range tables {
		level, msg = table.level, table.msg
		err := test_sub_log.LogWithErrors(level, msg)
		if i == len(tables)-1 {
			if err == nil {
				t.Errorf("Log created a log with an invalid log level: %v", level)
			}
		}
	}
}