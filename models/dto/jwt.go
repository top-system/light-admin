package dto

import (
	"github.com/golang-jwt/jwt/v5"
)

type JwtClaims struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
