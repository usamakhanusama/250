// Package student_test - Authored by Usama (2023331088)
package student_test

import (
	"testing"

	"sust-cse-portal/internal/student"
)

func TestEvenRollNumbersAssignedToSectionB(t *testing.T) {
	// Afif Zaman Khan (2023331020) ends in 20 (Even) -> Section B
	if sec := student.DetermineSection("2023331020"); sec != "B" {
		t.Errorf("expected 2023331020 to be Section B, got %s", sec)
	}

	// Usama (2023331088) ends in 88 (Even) -> Section B
	if sec := student.DetermineSection("2023331088"); sec != "B" {
		t.Errorf("expected 2023331088 to be Section B, got %s", sec)
	}
}

func TestOddRollNumbersAssignedToSectionA(t *testing.T) {
	// Tanvir (2023331019) ends in 19 (Odd) -> Section A
	if sec := student.DetermineSection("2023331019"); sec != "A" {
		t.Errorf("expected 2023331019 to be Section A, got %s", sec)
	}

	if sec := student.DetermineSection("2023331001"); sec != "A" {
		t.Errorf("expected 2023331001 to be Section A, got %s", sec)
	}
}

func TestStudentProfileLookup(t *testing.T) {
	s := student.NewService()
	st, found := s.GetStudentByRegistration("2023331020")
	if !found {
		t.Fatal("expected Afif's profile to be found")
	}
	if st.Name != "Afif Zaman Khan" || st.Section != "B" {
		t.Errorf("unexpected profile data: %+v", st)
	}
}

func TestFilterBySection(t *testing.T) {
	svc := NewService()
	secBList := svc.GetStudentsBySection("B")
	if len(secBList) == 0 {
		t.Errorf("Expected students in Section B, got 0")
	}
	for _, st := range secBList {
		if st.Section != "B" {
			t.Errorf("Expected Section B student, got %s", st.Section)
		}
	}
}
