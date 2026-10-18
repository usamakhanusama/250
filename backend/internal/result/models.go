// Package result - Authored by Usama (2023331088)
// Module: Academic Result & CGPA Engine
package result

type CourseResult struct {
	Code    string  `json:"code"`
	Name    string  `json:"name"`
	Credits float64 `json:"credits"`
	Grade   string  `json:"grade"`
	Points  float64 `json:"points"`
}

type SemesterRecord struct {
	Semester     int            `json:"semester"`
	Term         string         `json:"term"`
	Session      string         `json:"session"`
	GPA          float64        `json:"gpa"`
	Credits      float64        `json:"credits"`
	CoursesCount int            `json:"coursesCount"`
	Courses      []CourseResult `json:"courses"`
}

type TranscriptResponse struct {
	Registration   string           `json:"registration"`
	StudentName    string           `json:"studentName"`
	Department     string           `json:"department"`
	CGPA           float64          `json:"cgpa"`
	TotalCredits   float64          `json:"totalCredits"`
	SemesterReport []SemesterRecord `json:"semesterReport"`
}

type GPASimulationRequest struct {
	Registration string  `json:"registration"`
	TargetCGPA   float64 `json:"targetCgpa"`
	NewCredits   float64 `json:"newCredits"`
}

type GPASimulationResponse struct {
	Registration   string  `json:"registration"`
	CurrentCGPA    float64 `json:"currentCgpa"`
	TargetCGPA     float64 `json:"targetCgpa"`
	CurrentCredits float64 `json:"currentCredits"`
	NewCredits     float64 `json:"newCredits"`
	RequiredGPA    float64 `json:"requiredGpa"`
	Achievable     bool    `json:"achievable"`
	Message        string  `json:"message"`
}
