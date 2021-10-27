package testplan

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"
	"github.com/SSSOC-CAN/fmtd/utils"
	yaml "gopkg.in/yaml.v2"
)

var (
	stateFileName = "state.json"
	possibleActions = map[string]actionFunc{
		"WaitForTime": WaitForTime,
	}
	WaitForTime actionFunc = func(waitTime int64) error {
		time.Sleep(time.Duration(waitTime)*time.Second)
		return nil
	}
	defaultStateFilePath = func() string {
		return filepath.Join(utils.AppDataDir("fmtd", false), stateFileName)
	}
)

// actionFunc there will be a handful of standard actionFuncs like one to prompt a user and return when user confirms or wait a specific amount of time
type actionFunc func(int64) error

type TestPlan struct {
	Name				string			`yaml:"plan_name"`
	TestDuration		int64			`yaml:"test_duration"`
	DataProviders		[]*struct{
		Name			string		`yaml:"provider_name"`
		Driver			string		`yaml:"driver"`
		//dependencies	string		`yaml:"dependencies"`
		NumDataPoints	int64		`yaml:"num_data_points"`
	} `yaml:"data_providers"`
	Alerts				[]*struct{
		Name			string		`yaml:"alert_name"`
		ActionName		string		`yaml:"action"`
		ActionArg		int64		`yaml:"action_arg"`
		ActionStartTime	int64		`yaml:"action_start_time"`
		ActionFunc		actionFunc
	} `yaml:"alerts"`
	TestStateFilePath	string
}

// loadTestPlan will load the inputted test plan file into a TestPlan struct and return it
func loadTestPlan(path_to_test_plan string) (*TestPlan, error) {
	filename, err := filepath.Abs(path_to_test_plan)
	if err != nil {
		return nil, err
	}
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var testPlan TestPlan
	err = yaml.Unmarshal(file, &testPlan)
	if err != nil {
		return nil, err
	}
	err = assignActions(&testPlan)
	if err != nil {
		return nil, err
	}
	testPlan.TestStateFilePath = defaultStateFilePath()
	return &testPlan, nil
}

// assignActions takes the action names and maps them to a function
func assignActions(tp *TestPlan) error {
	for _, alert := range tp.Alerts {
		if aFunc, ok := possibleActions[alert.ActionName]; ok {
			alert.ActionFunc = aFunc
		} else {
			return fmt.Errorf("Action provided is not recognized: %v", alert.ActionName)
		}
	}
	return nil
}

// func main() {
// 	tp, err := loadTestPlan("C:\\Users\\Michael Graham\\Downloads\\testplan.yaml")
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	fmt.Println(tp)
// 	fmt.Println(tp.DataProviders[0], tp.DataProviders[1], tp.Alerts[0])
// }