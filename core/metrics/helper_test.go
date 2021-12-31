package metrics

import (
	"reflect"
	"testing"
)

func Test_parseNameAndTags(t *testing.T) {
	type args struct {
		src string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 map[string]string
		want2 bool
	}{
		{
			name: "test1",
			args: args{
				src: "request.count+name:qps|type:succ|other:xxx+yyy",
			},
			want: "request.count",
			want1: map[string]string{
				"name":  "qps",
				"type":  "succ",
				"other": "xxx+yyy",
			},
			want2: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := parseNameAndTags(tt.args.src)
			if got != tt.want {
				t.Errorf("parseNameAndTags() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("parseNameAndTags() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("parseNameAndTags() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
