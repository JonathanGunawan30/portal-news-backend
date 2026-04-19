package request

type LoginRequest struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=6"`
	RecaptchaToken string `json:"recaptcha_token" validate:"required"`
}
