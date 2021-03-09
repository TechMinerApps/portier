package log

import "testing"

func TestConvertToLoggerType(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want LoggerType
	}{
		{
			name: "HUMAN",
			args: args{
				input: "human",
			},
			want: HUMAN,
		},
		{
			name: "MACHINE",
			args: args{
				input: "machine",
			},
			want: MACHINE,
		},
		{
			name: "SOME_RANDOME",
			args: args{
				input: "aldkfhlakjsdhflkasdjfhl",
			},
			want: HUMAN,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertToLoggerType(tt.args.input); got != tt.want {
				t.Errorf("ConvertToLoggerType() = %v, want %v", got, tt.want)
			}
		})
	}
}
