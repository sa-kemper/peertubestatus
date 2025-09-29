package Response

import (
	"testing"
)

func Test_extractValFromEnvBytes(t *testing.T) {
	configBytes := []byte(`
	# This is an example config file.
	foo=bar
	advanced-foo  = "advanced -bar"
	green-green = " yum havea problem #in yum brain" # if you know, you know
	broken =
	bro=ken # valid
	k=#
`)
	type args struct {
		bytes []byte
		name  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "BaseCase", args: args{
			bytes: configBytes,
			name:  "foo",
		}, want: "bar"},
		{name: "Advanced case", args: args{
			bytes: configBytes,
			name:  "advanced-foo",
		}, want: "advanced -bar"},
		{name: "Green-Green case", args: args{
			bytes: configBytes,
			name:  "green-green",
		}, want: " yum havea problem #in yum brain"},
		{name: "Broken config value", args: args{bytes: configBytes, name: "broken"}, want: ""},
		{name: "bro can do this", args: args{bytes: configBytes, name: "bro"}, want: "ken"},
		{name: "no value but looks like one", args: args{bytes: configBytes, name: "k"}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractValFromEnvBytes(tt.args.bytes, tt.args.name); got != tt.want {
				t.Errorf("extractValFromEnvBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
