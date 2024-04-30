package main

import (
	"bytes"
	"testing"
)

func Test_timeGap_lineProcessor(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		wantWriter string
	}{
		{"test 1", []byte(""), ""},
		{"test 2", []byte(`.\rphost_2345\24011215.log:32:47.733007-0,EXCP,OSTread=1,
.\rphost_2345\24011215.log:22:47.733007-0,EXCP,OSTread=2,
.\rphost_2346\24011215.log:34:47.733007-0,EXCP,OSTread=1,
.\rphost_2345\24011215.log:35:00.000001-0,EXCP,OSTread=1,
.\rphost_2345\24011215.log:35:00.003003-0,EXCP,OSTread=2,
.\rphost_2346\24011215.log:35:00.003003-1000,EXCP,OSTread=1,`),
			`.\rphost_2345\24011215.log:32:47.733007-0,EXCP,OSTread=1,
.\rphost_2345\24011215.log:22:47.733007-0,EXCP,OSTread=2,
.\rphost_2346\24011215.log:34:47.733007-0,EXCP,OSTread=1,
.\rphost_2345\24011215.log:35:00.000001-132266,TIMEGAP,OSTread=1
.\rphost_2345\24011215.log:35:00.000001-0,EXCP,OSTread=1,
.\rphost_2345\24011215.log:35:00.003003-732269,TIMEGAP,OSTread=2
.\rphost_2345\24011215.log:35:00.003003-0,EXCP,OSTread=2,
.\rphost_2346\24011215.log:35:00.002003-12268,TIMEGAP,OSTread=1
.\rphost_2346\24011215.log:35:00.003003-1000,EXCP,OSTread=1,
`},
	}

	var obj timeGap
	obj.init()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &bytes.Buffer{}
			splitData := bytes.Split(tt.data, []byte("\n"))
			for i := range splitData {
				obj.lineProcessor(splitData[i], writer)
			}
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t.Errorf("timeGap.lineProcessor() = %v, want %v", gotWriter, tt.wantWriter)
			}
		})
	}
}
