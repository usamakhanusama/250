// Package result - Authored by Usama (2023331088)
package result

import (
	"math"
)

type Service struct {
	transcripts map[string]TranscriptResponse
}

func NewService() *Service {
	s := &Service{
		transcripts: make(map[string]TranscriptResponse),
	}
	s.initializeTranscripts()
	return s
}

func (s *Service) initializeTranscripts() {
	// Standard semester 1, 2, 3 records for SUST CSE 2023-2024 session
	semesters := []SemesterRecord{
		{
			Semester: 1, Term: "1st Year, 1st Semester (1/1)", Session: "2023", GPA: 3.75, Credits: 16.5, CoursesCount: 6,
			Courses: []CourseResult{
				{Code: "CSE-1101", Name: "Introduction to Computing", Credits: 3.0, Grade: "A", Points: 4.0},
				{Code: "CSE-1102", Name: "Programming Fundamentals", Credits: 3.0, Grade: "A-", Points: 3.7},
				{Code: "CSE-1103", Name: "Programming Fundamentals Lab", Credits: 1.5, Grade: "A", Points: 4.0},
				{Code: "MATH-1101", Name: "Calculus & Differential Equations", Credits: 3.0, Grade: "B+", Points: 3.3},
				{Code: "PHY-1101", Name: "Physics", Credits: 3.0, Grade: "A-", Points: 3.7},
				{Code: "EEE-1101", Name: "Basic Electrical Engineering", Credits: 3.0, Grade: "B+", Points: 3.3},
			},
		},
		{
			Semester: 2, Term: "1st Year, 2nd Semester (1/2)", Session: "2023–2024", GPA: 3.68, Credits: 18.0, CoursesCount: 7,
			Courses: []CourseResult{
				{Code: "CSE-1201", Name: "Object Oriented Programming (Java)", Credits: 3.0, Grade: "A", Points: 4.0},
				{Code: "CSE-1202", Name: "OOP Lab", Credits: 1.5, Grade: "A", Points: 4.0},
				{Code: "CSE-1203", Name: "Discrete Mathematics", Credits: 3.0, Grade: "A-", Points: 3.7},
				{Code: "MATH-1201", Name: "Linear Algebra & Matrices", Credits: 3.0, Grade: "B+", Points: 3.3},
				{Code: "STAT-1201", Name: "Probability & Statistics", Credits: 3.0, Grade: "A-", Points: 3.7},
				{Code: "HUM-1201", Name: "English Communication", Credits: 2.0, Grade: "A", Points: 4.0},
				{Code: "CSE-1204", Name: "Digital Logic Design", Credits: 2.5, Grade: "B+", Points: 3.3},
			},
		},
		{
			Semester: 3, Term: "2nd Year, 1st Semester (2/1)", Session: "2024", GPA: 3.72, Credits: 19.5, CoursesCount: 8,
			Courses: []CourseResult{
				{Code: "CSE-2101", Name: "Computer Networks & Protocols", Credits: 3.0, Grade: "A", Points: 4.0},
				{Code: "CSE-2102", Name: "Networks Lab", Credits: 1.5, Grade: "A-", Points: 3.7},
				{Code: "CSE-2103", Name: "Database Systems", Credits: 3.0, Grade: "A", Points: 4.0},
				{Code: "CSE-2104", Name: "Database Systems Lab", Credits: 1.5, Grade: "A", Points: 4.0},
				{Code: "CSE-2105", Name: "Operating Systems Principles", Credits: 3.0, Grade: "A-", Points: 3.7},
				{Code: "MATH-2101", Name: "Complex Variables", Credits: 3.0, Grade: "B+", Points: 3.3},
				{Code: "CSE-2106", Name: "Software Engineering & Architecture", Credits: 3.0, Grade: "A-", Points: 3.7},
				{Code: "CSE-2107", Name: "Software Engineering Project Lab", Credits: 1.5, Grade: "A", Points: 4.0},
			},
		},
	}

	totalCredits := 16.5 + 18.0 + 19.5
	totalPoints := (3.75 * 16.5) + (3.68 * 18.0) + (3.72 * 19.5)
	calcCGPA := math.Round((totalPoints/totalCredits)*100) / 100

	s.transcripts["2023331020"] = TranscriptResponse{
		Registration:   "2023331020",
		StudentName:    "Afif Zaman Khan",
		Department:     "Computer Science & Engineering",
		CGPA:           calcCGPA,
		TotalCredits:   totalCredits,
		SemesterReport: semesters,
	}

	s.transcripts["2023331088"] = TranscriptResponse{
		Registration:   "2023331088",
		StudentName:    "Usama",
		Department:     "Computer Science & Engineering",
		CGPA:           3.65,
		TotalCredits:   totalCredits,
		SemesterReport: semesters,
	}
}

func (s *Service) GetTranscriptByRegistration(reg string) (*TranscriptResponse, bool) {
	t, found := s.transcripts[reg]
	if !found {
		// Return Afif's transcript as benchmark fallback
		t = s.transcripts["2023331020"]
		t.Registration = reg
		return &t, true
	}
	return &t, true
}

func CalculateCGPA(semesters []SemesterRecord) float64 {
	var totalWeightedPoints float64
	var totalCredits float64

	for _, sem := range semesters {
		totalWeightedPoints += sem.GPA * sem.Credits
		totalCredits += sem.Credits
	}

	if totalCredits == 0 {
		return 0.0
	}
	return math.Round((totalWeightedPoints/totalCredits)*100) / 100
}

func (s *Service) SimulateTargetGPA(reg string, targetCGPA float64, newCredits float64) GPASimulationResponse {
	transcript, _ := s.GetTranscriptByRegistration(reg)
	currentCGPA := transcript.CGPA
	currentCredits := transcript.TotalCredits

	if newCredits <= 0 {
		newCredits = 22.5 // Default to 4th semester credit load
	}

	totalCredits := currentCredits + newCredits
	neededPoints := (targetCGPA * totalCredits) - (currentCGPA * currentCredits)
	requiredGPA := math.Round((neededPoints/newCredits)*100) / 100

	achievable := true
	message := "Target CGPA is realistically achievable with disciplined performance."

	if requiredGPA > 4.00 {
		achievable = false
		message = "Target CGPA exceeds 4.00 maximum SUST grade ceiling for this semester load."
	} else if requiredGPA < 2.00 {
		message = "Target is comfortably achieved even with minimum passing marks."
	}

	return GPASimulationResponse{
		Registration:   reg,
		CurrentCGPA:    currentCGPA,
		TargetCGPA:     targetCGPA,
		CurrentCredits: currentCredits,
		NewCredits:     newCredits,
		RequiredGPA:    requiredGPA,
		Achievable:     achievable,
		Message:        message,
	}
}
