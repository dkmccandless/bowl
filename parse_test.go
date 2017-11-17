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
	want Value
}{
	{"()", snil},
	{"5", 5},
	{"(5)", Pair{5, snil}},
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

func TestPairString(t *testing.T) {
	for _, test := range []struct {
		p    Pair
		want string
	}{
		{Pair{"a", "b"}, "(a . b)"},
		{Pair{"a", Pair{"b", "c"}}, "(a b . c)"},
		{Pair{"a", Pair{"b", Pair{"c", Pair{"d", "x"}}}}, "(a b c d . x)"},
		{Pair{"a", snil}, "(a)"},
		{Pair{"a", Pair{"b", Pair{"c", Pair{"d", snil}}}}, "(a b c d)"},
	} {
		if got := test.p.String(); got != test.want {
			t.Errorf("%#v.String(): got %v, want %v", test.p, got, test.want)
		}
	}
}
