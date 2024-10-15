package main

import (
	"reflect"
	"testing"
)

func Test_config_init(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		obj     config
		wantErr bool
	}{
		{"test 1", []string{"programname"}, config{
			programName:  "programname",
			isNeedPrefix: true,
			paths:        []string{},
		}, true},
		{
			"test 2",
			[]string{"programname", "-s"},
			config{
				programName:  "programname",
				isNeedPrefix: false,
				paths:        []string{},
			},
			true},
		{
			"test 3",
			[]string{"programname", "-s", "a1", "b2"},
			config{
				programName:  "programname",
				isNeedPrefix: false,
				paths:        []string{"a1", "b2"},
			},
			false},
	}
	for _, tt := range tests {
		var got config

		t.Run(tt.name, func(t *testing.T) {
			err := got.init(tt.args)
			got.rootPath = ""
			if !reflect.DeepEqual(got, tt.obj) {
				t.Errorf("config.init() = %v, want %v", got, tt.obj)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("config.init() error = %T, wantErr %T", err, tt.wantErr)
			}
		})
	}
}
