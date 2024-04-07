package main

import "testing"

func Test_isTrueFile(t *testing.T) {
	{
		tests := []struct {
			name     string
			fileName string
			want     bool
		}{
			{"test 1 1", "24040712.log", true},
		}

		var obj fileChecker
		obj.init([]string{})
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := obj.isTrueFile(tt.fileName); got != tt.want {
					t.Errorf("isTrueFile() = %v, want %v", got, tt.want)
				}
			})
		}
	}

	{
		tests := []struct {
			name     string
			fileName string
			want     bool
		}{
			{"test 2 1", "24040712.log", true},
			{"test 2 2", "24050712.log", false},
		}

		var obj fileChecker
		obj.init([]string{"24040712"})
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := obj.isTrueFile(tt.fileName); got != tt.want {
					t.Errorf("isTrueFile() = %v, want %v", got, tt.want)
				}
			})
		}
	}
}
