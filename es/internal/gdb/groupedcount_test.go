package gdb

import "testing"

func TestIdentifierRegexp(t *testing.T) {
	valid := []string{"routing_number", "us_bank_account_form_state", "state", "_x", "Col1"}
	for _, s := range valid {
		if !identifierRegexp.MatchString(s) {
			t.Errorf("expected %q to be a valid identifier", s)
		}
	}

	invalid := []string{"", "1col", "a b", "a;b", "count(*)", "a-b", "drop table x", "a,b", "a)"}
	for _, s := range invalid {
		if identifierRegexp.MatchString(s) {
			t.Errorf("expected %q to be rejected as an identifier", s)
		}
	}
}
