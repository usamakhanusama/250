package auth

import "time"

// SignupRequest is the public student self-registration payload.
// Role is NOT accepted from the client: every self-signup becomes a "student".
// (If we accepted "role", anyone could sign up as "admin".)
type SignupRequest struct {
	RegNo           string `json:"reg_no" validate:"required"`
	FullName        string `json:"full_name" validate:"required"`
	Email           string `json:"email" validate:"required,email,max=255"`
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
}

// LoginRequest: "login" can be either the registration number or the email.
type LoginRequest struct {
	Login    string `json:"login" validate:"required,max=255"`
	Password string `json:"password" validate:"required,max=72"`
}

// RefreshRequest represents token refresh payload
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequest represents logout payload
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ChangePasswordRequest represents password update payload
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

// StudentProfileBrief provides essential student details
type StudentProfileBrief struct {
	StudentID       string `json:"student_id,omitempty"`
	RegNo           string `json:"reg_no,omitempty"`
	FullName        string `json:"full_name,omitempty"`
	CurrentSemester string `json:"current_semester,omitempty"`
}

// TeacherProfileBrief provides essential teacher details
type TeacherProfileBrief struct {
	TeacherID   string `json:"teacher_id,omitempty"`
	TeacherCode string `json:"teacher_code,omitempty"`
	FullName    string `json:"full_name,omitempty"`
}

// UserResponse is the SAFE user data we send to the browser.
// Notice there is no password / password_hash field here at all, so it is
// impossible to leak the password through the JSON response.
type UserResponse struct {
	ID        string               `json:"id"`
	Email     string               `json:"email"`
	Role      string               `json:"role"`
	Student   *StudentProfileBrief `json:"student,omitempty"`
	Teacher   *TeacherProfileBrief `json:"teacher,omitempty"`
	CreatedAt time.Time            `json:"created_at"`
}

// LoginResponse represents the successful login payload
type LoginResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
}
