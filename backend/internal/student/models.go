// Package student - Authored by Usama (2023331088)
// Module: Student Identity & Section Auto-Assignment Engine
package student

type StudentProfile struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Registration string  `json:"registration"`
	Email        string  `json:"email"`
	Batch        string  `json:"batch"`
	Department   string  `json:"department"`
	Semester     int     `json:"semester"`
	Section      string  `json:"section"`
	CGPA         float64 `json:"cgpa"`
	Role         string  `json:"role"`
}

type BatchAnalytics struct {
	Batch        string `json:"batch"`
	Department   string `json:"department"`
	TotalStudents int    `json:"totalStudents"`
	SectionACount int    `json:"sectionACount"`
	SectionBCount int    `json:"sectionBCount"`
	EvenRollRule  string `json:"evenRollRule"`
	OddRollRule   string `json:"oddRollRule"`
}

// Day 3 Models: Course Enrollment & Credit Cap Validation
type CourseEnrollmentRequest struct {
	RegistrationNumber string  `json:"registrationNumber"`
	CourseCode         string  `json:"courseCode"`
	Credits            float64 `json:"credits"`
}

type CourseEnrollmentResponse struct {
	RegistrationNumber string  `json:"registrationNumber"`
	CourseCode         string  `json:"courseCode"`
	TotalCredits       float64 `json:"totalCredits"`
	Allowed            bool    `json:"allowed"`
	Message            string  `json:"message"`
}
