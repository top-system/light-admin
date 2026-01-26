package dto

type Login struct {
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required"`
	CaptchaID   string `json:"captchaId" validate:"required"`
	CaptchaCode string `json:"captchaCode" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int    `json:"expiresIn"`
}
