package metrics

import (
	"testing"
)

func Test_processTags(t *testing.T) {
	type args struct {
		tagKvs map[string]string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "",
			args: args{
				nil,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := processTags(tt.args.tagKvs); got != tt.want {
				t.Errorf("processTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getLocalHost(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "ip",
			want: "10.90.190.185"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLocalHost(); got != tt.want {
				t.Errorf("getLocalHost() = %v, want %v", got, tt.want)
			}
		})
	}
}
