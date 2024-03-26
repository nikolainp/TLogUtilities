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
			programName: "programname",
		}, true},
		{"test 2 1",
			[]string{"programname", "./m1app", "24022006.log", "../case_2024.02.20/tj/m1app"},
			config{
				programName:       "programname",
				sourceFolder:      "./m1app",
				destinationFolder: "../case_2024.02.20/tj/m1app",
				fileNames:         []string{"24022006.log"},
				transferType:      dataCopy,
			}, false},
		{"test 2 2",
			[]string{"programname", "./", "//merope.dept07/csr/error/tj"},
			config{
				programName:       "programname",
				sourceFolder:      "./",
				destinationFolder: "//merope.dept07/csr/error/tj",
				transferType:      dataCopy,
			}, false},
		{"test 3 1",
			[]string{"programname", "-m", "./m1app", "24022006.log", "24022007.log", "../case_2024.02.20/tj/m1app"},
			config{
				programName:       "programname",
				sourceFolder:      "./m1app",
				destinationFolder: "../case_2024.02.20/tj/m1app",
				fileNames:         []string{"24022006.log", "24022007.log"},
				transferType:      dataMove,
			}, false},
		{"test 3 2",
			[]string{"programname", "-m", "./", "//merope.dept07/csr/error/tj"},
			config{
				programName:       "programname",
				sourceFolder:      "./",
				destinationFolder: "//merope.dept07/csr/error/tj",
				transferType:      dataMove,
			}, false},
	}
	for _, tt := range tests {
		var got config

		t.Run(tt.name, func(t *testing.T) {
			err := got.init(tt.args)
			if !reflect.DeepEqual(got, tt.obj) {
				t.Errorf("config.init() = %v, want %v", got, tt.obj)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("config.init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
