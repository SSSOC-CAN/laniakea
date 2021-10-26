package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"
	"github.com/SSSOC-CAN/fmtd/utils"
	yaml "gopkg.in/yaml.v3"
)

var (
	possibleActions = map[string]actionFunc{
		"WaitForTime": WaitForTime,
	}
	WaitForTime = func(waitTime int64) error {
		time.Sleep(time.Duration(waitTime)*time.Second)
		return nil
	}
	defaultStateFilePath = func() string {
		return utils.AppDataDir("fmtd", false)
	}
)

// actionFunc there will be a handful of standard actionFuncs like one to prompt a user and return when user confirms or wait a specific amount of time
type actionFunc func(int64) error

type TestPlan struct {
	Name				string			`yaml:"plan_name"`
	TestDuration		int64			`yaml:"test_duration"`
	DataProviders		map[string][]struct{
		name			string		`yaml:"name"`
		driver			string		`yaml:"driver"`
		// dependencies	string		`yaml:"dependencies"`
		numdatapoints	int64		`yaml:"num_data_points"`
	} `yaml:"data_providers"`
	Alerts				map[string][]struct{
		name			string		`yaml:"name"`
		actionName		string		`yaml:"action"`
		actionArg		int64		`yaml:"action_arg"`
		action			actionFunc
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
	testPlan, err = assignActions(testPlan)
	testPlan.TestStateFilePath = defaultStateFilePath()
	return &testPlan, nil
}

// assignActions takes the action names and maps them to a function
func assignActions(tp TestPlan) (TestPlan, error) {
	for _, alert := range tp.Alerts {
		if aFunc, ok := possibleActions[alert[0].actionName]; ok {
			alert[0].action = aFunc
		} else {
			return TestPlan{}, fmt.Errorf("Action provided is not recognized: %v", alert[0].actionName)
		}
	}
	return tp, nil
}

func main() {
	tp, err := loadTestPlan("C:\\Users\\Michael Graham\\Downloads\\testplan.yaml")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(*tp)
}