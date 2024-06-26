package main

import (
	"reflect"
	"testing"
	"time"
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
			operation:   operationTimeGapBack,
		}, false},
		{
			"test 2",
			[]string{"programname", "2024.01.20_12:33:44", "2024.01.20_13:33:44"},
			config{
				programName:      "programname",
				operation:        operationFilterByTyme,
				filterBeginTime:  time.Date(2024, 1, 20, 12, 33, 44, 0, time.Local),
				filterFinishTime: time.Date(2024, 1, 20, 13, 33, 44, 0, time.Local),
				filterEdge:       edgeStop,
			},
			false},
		{
			"test 2 1",
			[]string{"programname", "24012012.log:33:44.012345-44000000"},
			config{
				programName:      "programname",
				operation:        operationFilterByTyme,
				filterBeginTime:  time.Date(2024, 1, 20, 12, 33, 00, 12345000, time.Local),
				filterFinishTime: time.Date(2024, 1, 20, 12, 33, 44, 12345000, time.Local),
				filterEdge:       edgeStop,
			},
			false},
		{
			"test 2 11",
			[]string{"programname", "24022006.log:47:11.180041-171938"},
			config{
				programName:      "programname",
				operation:        operationFilterByTyme,
				filterBeginTime:  time.Date(2024, 2, 20, 6, 47, 11, 8103000, time.Local),
				filterFinishTime: time.Date(2024, 2, 20, 6, 47, 11, 180041000, time.Local),
				filterEdge:       edgeStop,
			},
			false},
		{
			"test 2 2",
			[]string{"programname", "24012012.log:33:44.012345-44000000", "start"},
			config{
				programName:      "programname",
				operation:        operationFilterByTyme,
				filterBeginTime:  time.Date(2024, 1, 20, 12, 33, 00, 12345000, time.Local),
				filterFinishTime: time.Date(2024, 1, 20, 12, 33, 44, 12345000, time.Local),
				filterEdge:       edgeStart,
			},
			false},
		{
			"test 3",
			[]string{"programname", "2024.01.20_12:33:44.012345", "2024.01.20_13:33:44.543210"},
			config{
				programName:      "programname",
				operation:        operationFilterByTyme,
				filterBeginTime:  time.Date(2024, 1, 20, 12, 33, 44, 12345000, time.Local),
				filterFinishTime: time.Date(2024, 1, 20, 13, 33, 44, 543210000, time.Local),
				filterEdge:       edgeStop,
			},
			false},
		{
			"test 3 1",
			[]string{"programname", "2024.01.20_12:33:44.012345", "2024.01.20_13:33:44.543210", "start"},
			config{
				programName:      "programname",
				operation:        operationFilterByTyme,
				filterBeginTime:  time.Date(2024, 1, 20, 12, 33, 44, 12345000, time.Local),
				filterFinishTime: time.Date(2024, 1, 20, 13, 33, 44, 543210000, time.Local),
				filterEdge:       edgeStart,
			},
			false},
		{
			"test 3 1",
			[]string{"programname", "2024.01.20_12:33:44.012345", "2024.01.20_13:33:44.543210", "start"},
			config{
				programName:      "programname",
				operation:        operationFilterByTyme,
				filterBeginTime:  time.Date(2024, 1, 20, 12, 33, 44, 12345000, time.Local),
				filterFinishTime: time.Date(2024, 1, 20, 13, 33, 44, 543210000, time.Local),
				filterEdge:       edgeStart,
			},
			false},
		{
			"test 4",
			[]string{"programname", "2024.01#20_12:33:44.012345", "2024.01.20_13:33:44.543210"},
			config{
				programName: "programname",
				operation:   operationFilterByTyme,
			},
			true},
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
