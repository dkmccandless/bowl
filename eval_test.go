package main

import (
	"fmt"
	"testing"
)

func TestEval(t *testing.T) {
	tests := []struct{ exp, want string }{
		{"5", "5"},
		{"6", "6"},
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
