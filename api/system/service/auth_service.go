package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/models/dto"
)

type options struct {
	issuer        string
	signingMethod jwt.SigningMethod
	signingKey    interface{}
	keyfunc       jwt.Keyfunc
	expired       int
	tokenType     string
}

type AuthService struct {
	opts  *options
	cache lib.Cache
}

func NewAuthService(cache lib.Cache, config lib.Config) AuthService {
	issuer := config.Name
	signingKey := fmt.Sprintf("Jwt:%s", issuer)

	opts := &options{
		issuer:        issuer,
		tokenType:     "Bearer",
		expired:       config.Auth.TokenExpired,
		signingMethod: jwt.SigningMethodHS512,
		signingKey:    []byte(signingKey),
		keyfunc: func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, apperrors.AuthTokenInvalid
			}
			return []byte(signingKey), nil
		},
	}

	return AuthService{cache: cache, opts: opts}
}

func wrapperAuthKey(key string) string {
	return fmt.Sprintf("auth:%s", key)
}

func (a AuthService) GenerateToken(user *system.User) (*dto.LoginResponse, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(a.opts.expired) * time.Second)
	claims := &dto.JwtClaims{
		ID:       user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(a.opts.signingMethod, claims)
	expired := expiresAt.Sub(time.Now())

	err := a.cache.Set(wrapperAuthKey(claims.Username), 1, expired)
	if err != nil {
		return nil, err
	}

	accessToken, err := token.SignedString(a.opts.signingKey)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: "", // 暂未实现刷新令牌
		TokenType:    a.opts.tokenType,
		ExpiresIn:    a.opts.expired,
	}, nil
}

func (a AuthService) ParseToken(tokenString string) (*dto.JwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &dto.JwtClaims{}, a.opts.keyfunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, apperrors.AuthTokenMalformed
		} else if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperrors.AuthTokenExpired
		} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, apperrors.AuthTokenNotValidYet
		} else {
			return nil, apperrors.AuthTokenInvalid
		}
	}

	if token != nil {
		if claims, ok := token.Claims.(*dto.JwtClaims); ok && token.Valid {
			return claims, nil
		}
	}

	return nil, apperrors.AuthTokenInvalid
}

func (a AuthService) DestroyToken(username string) error {
	_, err := a.cache.Delete(wrapperAuthKey(username))
	return err
}
