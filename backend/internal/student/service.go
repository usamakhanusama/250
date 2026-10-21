// Package student - Authored by Usama (2023331088)
package student

import (
	"regexp"
	"strconv"
	"strings"
)

type Service struct {
	students map[string]StudentProfile
}

func NewService() *Service {
	s := &Service{
		students: make(map[string]StudentProfile),
	}
	s.initializeStudents()
	return s
}

// DetermineSection: In SUST CSE, even registration last digits belong to Section B,
// and odd registration last digits belong to Section A.
func DetermineSection(reg string) string {
	re := regexp.MustCompile(`\D`)
	digits := re.ReplaceAllString(reg, "")
	if len(digits) < 2 {
		return "B"
	}
	last2Str := digits[len(digits)-2:]
	last2, err := strconv.Atoi(last2Str)
	if err != nil {
		return "B"
	}
	if last2%2 == 0 {
		return "B" // Even -> Section B (e.g. Afif 20, Usama 88)
	}
	return "A" // Odd -> Section A (e.g. 01, 03, 19)
}

func (s *Service) initializeStudents() {
	s.students["2023331020"] = StudentProfile{
		ID:           "stu-001",
		Name:         "Afif Zaman Khan",
		Registration: "2023331020",
		Email:        "afif@student.sust.edu",
		Batch:        "2023–2024",
		Department:   "Computer Science & Engineering",
		Semester:     4,
		Section:      DetermineSection("2023331020"), // Section B
		CGPA:         3.72,
		Role:         "student",
	}

	s.students["2023331088"] = StudentProfile{
		ID:           "stu-002",
		Name:         "Usama",
		Registration: "2023331088",
		Email:        "usama@student.sust.edu",
		Batch:        "2023–2024",
		Department:   "Computer Science & Engineering",
		Semester:     4,
		Section:      DetermineSection("2023331088"), // Section B
		CGPA:         3.65,
		Role:         "student",
	}

	s.students["2023331019"] = StudentProfile{
		ID:           "stu-003",
		Name:         "Tanvir Ahmed (Odd Roll Demo)",
		Registration: "2023331019",
		Email:        "tanvir@student.sust.edu",
		Batch:        "2023–2024",
		Department:   "Computer Science & Engineering",
		Semester:     4,
		Section:      DetermineSection("2023331019"), // Section A
		CGPA:         3.68,
		Role:         "student",
	}
}

func (s *Service) GetStudentByRegistration(reg string) (*StudentProfile, bool) {
	student, exists := s.students[reg]
	if !exists {
		return nil, false
	}
	return &student, true
}

func (s *Service) GetAllStudents() []StudentProfile {
	var list []StudentProfile
	for _, st := range s.students {
		list = append(list, st)
	}
	return list
}

func (s *Service) GetBatchAnalytics() BatchAnalytics {
	secA, secB := 0, 0
	for _, st := range s.students {
		if st.Section == "A" {
			secA++
		} else {
			secB++
		}
	}
	return BatchAnalytics{
		Batch:         "2023–2024 (CSE 23)",
		Department:    "Computer Science & Engineering",
		TotalStudents: len(s.students),
		SectionACount: secA,
		SectionBCount: secB,
		EvenRollRule:  "Registration last 2 digits even -> Section B (e.g. 20, 88)",
		OddRollRule:   "Registration last 2 digits odd -> Section A (e.g. 19)",
	}
}

func (s *Service) SearchStudents(query string) []StudentProfile {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return s.GetAllStudents()
	}
	var results []StudentProfile
	for _, st := range s.students {
		if strings.Contains(strings.ToLower(st.Name), q) ||
			strings.Contains(st.Registration, q) ||
			strings.Contains(strings.ToLower(st.Section), q) {
			results = append(results, st)
		}
	}
	return results
}

// Day 2 Enhancement: Filter Students by Section (A or B)
func (s *Service) GetStudentsBySection(sec string) []StudentProfile {
	target := strings.ToUpper(strings.TrimSpace(sec))
	var res []StudentProfile
	for _, st := range s.students {
		if strings.ToUpper(st.Section) == target {
			res = append(res, st)
		}
	}
	return res
}

// Day 3 Implementation: Validate Credit Cap (Max 25.0 credits per semester in SUST)
const MaxSemesterCredits = 25.0

func (s *Service) ValidateAndEnrollCourse(reg string, courseCode string, courseCredits float64) CourseEnrollmentResponse {
	// Baseline 4th semester enrolled credits is 22.5
	currentCredits := 22.5
	projectedCredits := currentCredits + courseCredits

	if projectedCredits > MaxSemesterCredits {
		return CourseEnrollmentResponse{
			RegistrationNumber: reg,
			CourseCode:         courseCode,
			TotalCredits:       projectedCredits,
			Allowed:            false,
			Message:            "Credit cap exceeded! Maximum allowed semester load in SUST CSE is 25.0 credits.",
		}
	}

	return CourseEnrollmentResponse{
		RegistrationNumber: reg,
		CourseCode:         courseCode,
		TotalCredits:       projectedCredits,
		Allowed:            true,
		Message:            "Course enrollment validated and approved within semester credit bounds.",
	}
}
