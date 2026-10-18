package student

import "testing"

func TestSection(t *testing.T) {
	if Section("2023331088") != "B" {
		t.Fatal("88 should be B")
	}
	if Section("2023331019") != "A" {
		t.Fatal("19 should be A")
	}
}