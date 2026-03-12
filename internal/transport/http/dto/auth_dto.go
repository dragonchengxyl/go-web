package dto

// RegisterRequest represents registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
}

// VerifyEmailRequest represents an email verification request.
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// UserResponse represents user response
type UserResponse struct {
	ID              string  `json:"id"`
	Username        string  `json:"username"`
	Email           string  `json:"email"`
	Avatar          *string `json:"avatar,omitempty"`
	Bio             *string `json:"bio,omitempty"`
	Location        *string `json:"location,omitempty"`
	Role            string  `json:"role"`
	Status          string  `json:"status"`
	EmailVerifiedAt *string `json:"email_verified_at,omitempty"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int64         `json:"expires_in"`
}
