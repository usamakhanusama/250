// Package result_test - Authored by Usama (2023331088)
package result_test

import (
	"testing"

	"sust-cse-portal/internal/result"
)

func TestCalculateCGPA(t *testing.T) {
	sems := []result.SemesterRecord{
		{Semester: 1, GPA: 4.0, Credits: 15.0},
		{Semester: 2, GPA: 3.0, Credits: 15.0},
	}
	cgpa := result.CalculateCGPA(sems)
	if cgpa != 3.50 {
		t.Errorf("expected CGPA 3.50, got %f", cgpa)
	}
}

func TestTranscriptRetrieval(t *testing.T) {
	s := result.NewService()
	tr, found := s.GetTranscriptByRegistration("2023331020")
	if !found {
		t.Fatal("expected transcript to be found")
	}
	if tr.TotalCredits != 54.0 {
		t.Errorf("expected 54.0 total credits, got %f", tr.TotalCredits)
	}
}
