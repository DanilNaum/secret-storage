package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTManager структура для управления JWT токенами
type JWTManager[T any] struct {
	secretKey     string
	tokenDuration time.Duration
}

// NewJWTManager создает новый менеджер JWT токенов
func NewJWTManager[T any](secretKey string, tokenDuration time.Duration) *JWTManager[T] {
	return &JWTManager[T]{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

// Claims кастомные claims с дженериком для хранения пользовательских данных
type Claims[T any] struct {
	UserData T `json:"user_data"`
	jwt.RegisteredClaims
}

// GenerateToken генерирует JWT токен с переданными данными пользователя
func (m *JWTManager[T]) GenerateToken(userData T) (string, error) {
	claims := &Claims[T]{
		UserData: userData,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

// ParseToken парсит и валидирует JWT токен, возвращает пользовательские данные
func (m *JWTManager[T]) ParseToken(tokenString string) (T, error) {
	var zero T

	token, err := jwt.ParseWithClaims(tokenString, &Claims[T]{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return zero, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims[T])
	if !ok || !token.Valid {
		return zero, errors.New("invalid token claims")
	}

	return claims.UserData, nil
}
