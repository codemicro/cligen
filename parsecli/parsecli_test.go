package parsecli

import (
	"reflect"
	"strings"
	"testing"
)

func TestSlice(t *testing.T) {
	type args struct {
		input []string
	}
	tests := []struct {
		name      string
		args      args
		wantFlags map[string]string
		wantArgs  []string
		wantErr   bool
	}{
		{args: args{[]string{"--hello=world"}}, wantFlags: map[string]string{"hello": "world"}},
		{args: args{[]string{"--hello=\"world", "this", "is", "cool\""}}, wantFlags: map[string]string{"hello": "world this is cool"}, wantErr: false},
		{args: args{[]string{"-hello=\"world", "this", "is", "cool\""}}, wantFlags: map[string]string{"e": "true", "h": "true", "l": "true", "o": "world this is cool"}, wantErr: false},

		{args: args{strings.Split(`banana this is a thing "hello what oh wow ok"`, " ")}, wantFlags: map[string]string{}, wantArgs: []string{"banana", "this", "is", "a", "thing", "hello what oh wow ok"}},
		{args: args{strings.Split(`banana this is a thing 'hello what oh wow ok'`, " ")}, wantFlags: map[string]string{}, wantArgs: []string{"banana", "this", "is", "a", "thing", "hello what oh wow ok"}},
		{args: args{strings.Split(`--hello banana this is a thing "hello what oh wow ok"`, " ")}, wantFlags: map[string]string{"hello": "true"}, wantArgs: []string{"banana", "this", "is", "a", "thing", "hello what oh wow ok"}},
		{args: args{strings.Split(`-he banana this is a thing "hello what oh wow ok"`, " ")}, wantFlags: map[string]string{"h": "true", "e": "true"}, wantArgs: []string{"banana", "this", "is", "a", "thing", "hello what oh wow ok"}},
		{args: args{strings.Split(`-h=banana`, " ")}, wantFlags: map[string]string{"h": "banana"}},

		{args: args{strings.Split("---hello", " ")}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFlags, gotArgs, err := Slice(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Slice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFlags, tt.wantFlags) {
				t.Errorf("Slice() gotFlags = %v, want %v", gotFlags, tt.wantFlags)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("Slice() gotArgs = %#v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func Test_countPrefixLength(t *testing.T) {
	type args struct {
		s      string
		prefix rune
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{args: args{"--hello", '-'}, want: 2},
		{args: args{"--hell-o--", '-'}, want: 2},
	}
		for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := countPrefixLength(tt.args.s, tt.args.prefix); got != tt.want {
				t.Errorf("countPrefixLength() = %v, want %v", got, tt.want)
			}
		})
	}
}