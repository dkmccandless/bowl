package main

import (
	"fmt"
	"testing"
)

func TestEval(t *testing.T) {
	tests := []struct{ exp, want string }{
		{"5", "5"},
		{"6", "6"},
		{"(quote a)", "a"},
		{"(abs -3)", "3"},
		{"(abs 3)", "3"},
		{"(append nil 3)", "(3)"},
		{"(append (list 3 4) 5)", "(3 4 5)"},
		{"(cons 3 4)", "(3 . 4)"},
		{"(list 3 4 5)", "(3 4 5)"},
	}

	for _, test := range tests {
		s, err := Parse(test.exp)
		if err != nil {
			t.Error(err)
			continue
		}
		if got := fmt.Sprint(Eval(s, GlobalEnv)); err != nil || got != test.want {
			t.Errorf("Eval(%q): got %q, want %q", test.exp, got, test.want)
		}
	}
}
