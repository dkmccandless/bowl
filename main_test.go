package main

import (
	"reflect"
	"testing"
)

var Invalid = []string{
	"",
	"(",
	")",
	"5 5",
}

var Valid = []struct {
	in   string
	want Expression
}{
	{"()", nil},
	{"5", 5},
	{"(5)", Pair{a: 5}},
	{"5+5", "5+5"},
	{"55", 55},
}

func TestParseInvalid(t *testing.T) {
	for _, test := range Invalid {
		if _, err := Parse(test); err == nil {
			t.Errorf("Parse(%q) err is nil, want error", test)
		}
	}
}

func TestParseValid(t *testing.T) {
	for _, test := range Valid {
		if got, err := Parse(test.in); err != nil || !reflect.DeepEqual(got, test.want) {
			t.Errorf("Parse(%q): got %v, %v; want %v, nil", test.in, got, err, test.want)
		}
	}
}
