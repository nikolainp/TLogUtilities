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
		{"test 2", []byte(`.\rphost_2345\24011215.log:32:47.733007-0,EXCP,OSThread=1,
.\rphost_2345\24011215.log:22:47.733007-0,EXCP,OSThread=2,
.\rphost_2346\24011215.log:34:47.733007-0,EXCP,OSThread=1,
.\rphost_2345\24011215.log:35:00.000001-0,EXCP,OSThread=1,
.\rphost_2345\24011215.log:35:00.003003-0,EXCP,OSThread=2,
.\rphost_2346\24011215.log:35:00.003003-1000,EXCP,OSThread=1,`),
			`.\rphost_2345\24011215.log:32:47.733007-0,EXCP,OSThread=1,
.\rphost_2345\24011215.log:22:47.733007-600000000,TIMEBACK
.\rphost_2345\24011215.log:22:47.733007-0,EXCP,OSThread=2,
.\rphost_2346\24011215.log:34:47.733007-0,EXCP,OSThread=1,
.\rphost_2345\24011215.log:35:00.000001-132266994,TIMEGAP,0,OSThread=1,
.\rphost_2345\24011215.log:35:00.000001-0,EXCP,OSThread=1,
.\rphost_2345\24011215.log:35:00.003003-732269996,TIMEGAP,0,OSThread=2,
.\rphost_2345\24011215.log:35:00.003003-0,EXCP,OSThread=2,
.\rphost_2346\24011215.log:35:00.002003-12268996,TIMEGAP,0,OSThread=1,
.\rphost_2346\24011215.log:35:00.003003-1000,EXCP,OSThread=1,
`},
		{"test 3", []byte(`tj_server/rphost_25279/24051420.log:14:57.386003-0,CLSTR,3,process=rphos,OSThread=29360,
tj_server/rphost_25279/24051420.log:40:01.690000-1000,SCALL,3,process=rphost,OSThread=29360,`),
			`tj_server/rphost_25279/24051420.log:14:57.386003-0,CLSTR,3,process=rphos,OSThread=29360,
tj_server/rphost_25279/24051420.log:40:01.689000-1504302997,TIMEGAP,3,OSThread=29360,
tj_server/rphost_25279/24051420.log:40:01.690000-1000,SCALL,3,process=rphost,OSThread=29360,
`},
		{"test 4",
			[]byte(`./tj/33/rphost_1744/24100115.log:29:50.267156-0,CLSTR,3,OSThread=18900,
./tj/33/rphost_1744/24100115.log:29:52.041078-2866083051,CALL,1,OSThread=18900
./tj/33/rphost_1744/24100115.log:29:58.267156-0,CLSTR,3,OSThread=18900
./tj/33/rphost_1744/24100115.log:30:00.041078-7000000,CALL,1,OSThread=18900`),
			`./tj/33/rphost_1744/24100115.log:29:50.267156-0,CLSTR,3,OSThread=18900,
./tj/33/rphost_1744/24100115.log:29:52.041078-2866083051,CALL,1,OSThread=18900
./tj/33/rphost_1744/24100115.log:29:58.267156-0,CLSTR,3,OSThread=18900
./tj/33/rphost_1744/24100115.log:29:53.041078-1000000,TIMEGAP,1,OSThread=18900,
./tj/33/rphost_1744/24100115.log:30:00.041078-7000000,CALL,1,OSThread=18900
`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var obj timeGap
			obj.init()

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
