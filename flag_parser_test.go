package goping

import (
	"reflect"
	"testing"
	"io/ioutil"
)

type AnyError struct{}

func (err AnyError) Error() string { return "" }

func TestFlagParser(t *testing.T) {
	var flagTests = []struct {
		testCaseName   string
		args           []string
		expectedResult Params
		expectedError  error
	}{
		{"Minimal_working_example",
			[]string{"some_url.com"},
			Params{DefaultTimeout, DefaultInterval, DefaultCount, DefaultDeadline, "some_url.com"}, nil},
		{"No_url_provided",
			[]string{},
			Params{}, WrongNumberOfArguments{}},
		{"Two_urls_provided",
			[]string{"http://some_url.com", "another_url"},
			Params{}, WrongNumberOfArguments{}},
		{"Nonexistent_flag",
			[]string{"-someNonExistentFlag", "some_url.com"},
			Params{}, AnyError{}},
		{"Complete_example",
			[]string{"-timeout", "5", "-interval", "4", "-count", "3", "-deadline", "10", "url.com"},
			Params{5, 4, 3, 10, "url.com"}, nil},
	}

	for _, tc := range flagTests {
		t.Run(tc.testCaseName, func(t *testing.T) {
			res, err := ParseCommandLine(tc.args, ioutil.Discard)

			switch {
			case tc.expectedError == (AnyError{}):
				if err == nil {
					t.Errorf("On input %v expected an error.", tc.args)
				}
			case reflect.TypeOf(err) != reflect.TypeOf(tc.expectedError):
				t.Errorf("On input %v expected an error of type: %T, actual error type: %T.",
					tc.args, tc.expectedError, err)
			case res != tc.expectedResult:
				t.Errorf("On input %v expected result: %v, actual result: %v", tc.args, tc.expectedResult, res)
			}
		})
	}
}
