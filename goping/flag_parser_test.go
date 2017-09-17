package goping

import (
	"io/ioutil"
	"reflect"
	"testing"
	"time"
)

type AnyError struct{}

func (err AnyError) Error() string { return "" }

var flagtests = []struct {
	args            []string
	expected_result Params
	expected_error  error
}{
	{[]string{}, Params{}, WrongNumberOfArguments{expected: 1, actual: 0}},
	{[]string{"http://some_url.com", "http://another_url"}, Params{}, WrongNumberOfArguments{expected: 1, actual: 2}},
	{[]string{"-help"}, Params{}, Cancelled{"Help is called."}},
	{[]string{"-someNonExistingFlag"}, Params{}, AnyError{}},
	{[]string{"http://some_url.com"},
		Params{DefaultTimeout, DefaultCount, DefaultDeadline, "http://some_url.com"}, nil},
	{[]string{"-timeout", "5s", "-count", "3", "-deadline", "10h", "http://url.com"},
		Params{time.Second * 5, 3, time.Hour * 10, "http://url.com"}, nil},
}

func TestFlagParser(t *testing.T) {
	for _, tt := range flagtests {
		res, err := ParseCommandLine(tt.args, ioutil.Discard)

		switch {
		case tt.expected_error == (AnyError{}):
			if err == nil {
				t.Errorf("On input %v expected an error.", tt.args)
			}
		case reflect.TypeOf(err) != reflect.TypeOf(tt.expected_error):
			t.Errorf("On input %v expected an error of type: %T, actual error type: %T.",
				tt.args, tt.expected_error, err)
		case res != tt.expected_result:
			t.Errorf("On input %v expected result: %v, actual result: %v", tt.args, tt.expected_result, res)
		}
	}
}
