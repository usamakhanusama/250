package auth

import "time"

// RegisterRequest represents the payload for registering a new user (Admin only)
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=admin teacher student parent accountant"`
}

// LoginRequest represents user login payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
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
	NewPassword string `json:"new_password" validate:"required,min=6"`
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

// UserResponse represents safe user data returned in API
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
