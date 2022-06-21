/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

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
	defaultStateFilePath = func(testPlanName string) string {
		return filepath.Join(utils.AppDataDir("fmtd", false), testPlanName+"_"+stateFileName)
	}
)

// actionFunc there will be a handful of standard actionFuncs like one to prompt a user and return when user confirms or wait a specific amount of time
type actionFunc func(int64) error

type AlertState string

var (
	ALERTSTATE_PENDING AlertState = "PENDING"
	ALERTSTATE_COMPLETED AlertState = "COMPLETED"
	ALERTSTATE_EXPIRED AlertState = "EXPIRED"
)

type Alert struct {
	Name			string		`yaml:"alert_name"`
	ActionName		string		`yaml:"action"`
	ActionArg		int64		`yaml:"action_arg"`
	ActionStartTime	int			`yaml:"action_start_time"`
	ActionFunc		actionFunc
	ExecutionState	AlertState
}

type TestPlan struct {
	Name				string			`yaml:"plan_name"`
	TestDuration		int64			`yaml:"test_duration"`
	DataProviders		[]*struct{
		Name			string		`yaml:"provider_name"`
		Driver			string		`yaml:"driver"`
		//dependencies	string		`yaml:"dependencies"`
		NumDataPoints	int64		`yaml:"num_data_points"`
	} `yaml:"data_providers"`
	Alerts				[]*Alert 	`yaml:"alerts"`
	TestStateFilePath	string
	TestReportFilePath 	string		`yaml:"report_file_path"`
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
	testPlan.TestStateFilePath = defaultStateFilePath(testPlan.Name)
	return &testPlan, nil
}

// assignActions takes the action names and maps them to a function and also gives the default PENDING State
func assignActions(tp *TestPlan) error {
	for _, alert := range tp.Alerts {
		if aFunc, ok := possibleActions[alert.ActionName]; ok {
			alert.ActionFunc = aFunc
			alert.ExecutionState = ALERTSTATE_PENDING
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