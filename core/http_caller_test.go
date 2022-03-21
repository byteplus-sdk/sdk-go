package core

import (
	"testing"

	"github.com/byteplus-sdk/sdk-go/core/option"
)

func TestHttpCaller_withOptionQueries(t *testing.T) {
	type fields struct {
		context *Context
	}
	type args struct {
		options *option.Options
		url     string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "empty_with_stage_and_query",
			fields: fields{},
			args: args{
				options: &option.Options{
					Stage: "pre",
					Queries: map[string]string{
						"query1": "value1",
					},
				},
				url: "https://www.bytedance.com",
			},
			want: "https://www.bytedance.com?stage=pre&query1=value1",
		},
		{
			name:   "exist_with_stage_and_query",
			fields: fields{},
			args: args{
				options: &option.Options{
					Stage: "pre",
					Queries: map[string]string{
						"query2": "value2",
					},
				},
				url: "https://www.bytedance.com?query1=value1",
			},
			want: "https://www.bytedance.com?query1=value1&stage=pre&query2=value2",
		},
		{
			name:   "exist_with_empty",
			fields: fields{},
			args: args{
				options: &option.Options{},
				url:     "https://www.bytedance.com?query1=value1",
			},
			want: "https://www.bytedance.com?query1=value1",
		},
		{
			name:   "empty",
			fields: fields{},
			args: args{
				options: &option.Options{},
				url:     "https://www.bytedance.com",
			},
			want: "https://www.bytedance.com",
		},
		{
			name:   "exist_with_query",
			fields: fields{},
			args: args{
				options: &option.Options{
					Queries: map[string]string{
						"query2": "value2",
					},
				},
				url: "https://www.bytedance.com?query1=value1",
			},
			want: "https://www.bytedance.com?query1=value1&query2=value2",
		},
		{
			name:   "exist_with_stage",
			fields: fields{},
			args: args{
				options: &option.Options{
					Stage: "pre",
				},
				url: "https://www.bytedance.com?query1=value1",
			},
			want: "https://www.bytedance.com?query1=value1&stage=pre",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HTTPCaller{
				context: tt.fields.context,
			}
			if got := c.withOptionQueries(tt.args.options, tt.args.url); got != tt.want {
				t.Errorf("withOptionQueries() = %v, want %v", got, tt.want)
			}
		})
	}
}
