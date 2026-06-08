package dto

type RegisterUser struct {
	Name     string `json:"name"     validate:"required,min=2,max=50"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginUser struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UpdateUser struct {
	Name  string  `json:"name"  validate:"omitempty,min=2,max=50"`
	Image *string `json:"image" validate:"omitempty,url"`
}
type GoogleCallbackRequest struct {
	Code  string `form:"code"  binding:"required"`
	State string `form:"state" binding:"required"`
}
type UpdatePassword struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8"`
	NewPassword     string `json:"new_password"     validate:"required,min=8"`
}

type ForgotPassword struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyOTP struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp"   validate:"required,len=6"`
}

type ResetPassword struct {
	Email       string `json:"email"        validate:"required,email"`
	OTP         string `json:"otp"          validate:"required,len=6"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type UpdateProfileImage struct {
	Image string `json:"image" validate:"required,url"`
}

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}
