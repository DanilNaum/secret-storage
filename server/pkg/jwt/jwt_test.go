package jwt

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTManager(t *testing.T) {
	secretKey := "test-secret-key"
	tokenDuration := time.Hour

	t.Run("NewJWTManager", func(t *testing.T) {
		manager := NewJWTManager[string](secretKey, tokenDuration)
		if manager == nil {
			t.Error("Expected non-nil manager")
		}
		if manager.secretKey != secretKey {
			t.Errorf("Expected secret key %s, got %s", secretKey, manager.secretKey)
		}
		if manager.tokenDuration != tokenDuration {
			t.Errorf("Expected token duration %v, got %v", tokenDuration, manager.tokenDuration)
		}
	})

	t.Run("GenerateToken and ParseToken with string", func(t *testing.T) {
		manager := NewJWTManager[string](secretKey, tokenDuration)
		testData := "test-user-data"

		token, err := manager.GenerateToken(testData)
		if err != nil {
			t.Fatalf("GenerateToken failed: %v", err)
		}

		if token == "" {
			t.Error("Expected non-empty token")
		}

		// Проверяем, что токен состоит из трех частей (header.payload.signature)
		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			t.Errorf("Expected JWT token to have 3 parts, got %d", len(parts))
		}

		parsedData, err := manager.ParseToken(token)
		if err != nil {
			t.Fatalf("ParseToken failed: %v", err)
		}

		if parsedData != testData {
			t.Errorf("Expected parsed data %s, got %s", testData, parsedData)
		}
	})

	t.Run("GenerateToken and ParseToken with UUID", func(t *testing.T) {
		manager := NewJWTManager[uuid.UUID](secretKey, tokenDuration)
		testUUID := uuid.New()

		token, err := manager.GenerateToken(testUUID)
		if err != nil {
			t.Fatalf("GenerateToken failed: %v", err)
		}

		parsedUUID, err := manager.ParseToken(token)
		if err != nil {
			t.Fatalf("ParseToken failed: %v", err)
		}

		if parsedUUID != testUUID {
			t.Errorf("Expected parsed UUID %s, got %s", testUUID.String(), parsedUUID.String())
		}
	})

	t.Run("GenerateToken and ParseToken with struct", func(t *testing.T) {
		type User struct {
			ID    uuid.UUID
			Name  string
			Email string
		}

		manager := NewJWTManager[User](secretKey, tokenDuration)
		testUser := User{
			ID:    uuid.New(),
			Name:  "John Doe",
			Email: "john@example.com",
		}

		token, err := manager.GenerateToken(testUser)
		if err != nil {
			t.Fatalf("GenerateToken failed: %v", err)
		}

		parsedUser, err := manager.ParseToken(token)
		if err != nil {
			t.Fatalf("ParseToken failed: %v", err)
		}

		if parsedUser.ID != testUser.ID {
			t.Errorf("Expected user ID %s, got %s", testUser.ID.String(), parsedUser.ID.String())
		}
		if parsedUser.Name != testUser.Name {
			t.Errorf("Expected user name %s, got %s", testUser.Name, parsedUser.Name)
		}
		if parsedUser.Email != testUser.Email {
			t.Errorf("Expected user email %s, got %s", testUser.Email, parsedUser.Email)
		}
	})

	t.Run("ParseToken with invalid token", func(t *testing.T) {
		manager := NewJWTManager[string](secretKey, tokenDuration)

		_, err := manager.ParseToken("invalid.token.here")
		if err == nil {
			t.Error("Expected error for invalid token")
		}
	})

	t.Run("ParseToken with wrong secret key", func(t *testing.T) {
		manager1 := NewJWTManager[string](secretKey, tokenDuration)
		manager2 := NewJWTManager[string]("different-secret-key", tokenDuration)

		testData := "test-data"
		token, err := manager1.GenerateToken(testData)
		if err != nil {
			t.Fatalf("GenerateToken failed: %v", err)
		}

		_, err = manager2.ParseToken(token)
		if err == nil {
			t.Error("Expected error for token signed with different secret")
		}
	})

	t.Run("ParseToken with expired token", func(t *testing.T) {
		shortDuration := time.Millisecond
		manager := NewJWTManager[string](secretKey, shortDuration)

		testData := "test-data"
		token, err := manager.GenerateToken(testData)
		if err != nil {
			t.Fatalf("GenerateToken failed: %v", err)
		}

		// Ждем пока токен истечет
		time.Sleep(2 * shortDuration)

		_, err = manager.ParseToken(token)
		if err == nil {
			t.Error("Expected error for expired token")
		}
		if !strings.Contains(err.Error(), "expired") {
			t.Errorf("Expected expired token error, got: %v", err)
		}
	})

	t.Run("ParseToken with tampered token", func(t *testing.T) {
		manager := NewJWTManager[string](secretKey, tokenDuration)

		testData := "test-data"
		token, err := manager.GenerateToken(testData)
		if err != nil {
			t.Fatalf("GenerateToken failed: %v", err)
		}

		// Подменяем часть токена
		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			t.Fatalf("Invalid token format")
		}
		tamperedToken := parts[0] + "." + parts[1] + ".tampered-signature"

		_, err = manager.ParseToken(tamperedToken)
		if err == nil {
			t.Error("Expected error for tampered token")
		}
	})

	t.Run("ParseToken with empty token", func(t *testing.T) {
		manager := NewJWTManager[string](secretKey, tokenDuration)

		_, err := manager.ParseToken("")
		if err == nil {
			t.Error("Expected error for empty token")
		}
	})

	t.Run("ParseToken with malformed token", func(t *testing.T) {
		manager := NewJWTManager[string](secretKey, tokenDuration)

		_, err := manager.ParseToken("not.a.jwt.token")
		if err == nil {
			t.Error("Expected error for malformed token")
		}
	})

	t.Run("GenerateToken with empty secret key", func(t *testing.T) {
		manager := NewJWTManager[string]("", tokenDuration)
		testData := "test-data"

		token, err := manager.GenerateToken(testData)
		if err != nil {
			t.Fatalf("GenerateToken should work with empty secret key: %v", err)
		}

		// Должен уметь парсить тот же менеджер
		parsedData, err := manager.ParseToken(token)
		if err != nil {
			t.Fatalf("ParseToken failed: %v", err)
		}

		if parsedData != testData {
			t.Errorf("Expected parsed data %s, got %s", testData, parsedData)
		}
	})

	t.Run("Zero value return on error", func(t *testing.T) {
		manager := NewJWTManager[string](secretKey, tokenDuration)

		result, err := manager.ParseToken("invalid.token")
		if err == nil {
			t.Error("Expected error")
		}
		if result != "" {
			t.Errorf("Expected zero value, got: %s", result)
		}
	})

	t.Run("Token with different generic types", func(t *testing.T) {
		// Тестируем разные типы данных
		testCases := []struct {
			name    string
			data    interface{}
			checkFn func(interface{}, interface{}) bool
		}{
			{
				name:    "int",
				data:    42,
				checkFn: func(a, b interface{}) bool { return a.(int) == b.(int) },
			},
			{
				name:    "bool",
				data:    true,
				checkFn: func(a, b interface{}) bool { return a.(bool) == b.(bool) },
			},
			{
				name:    "float",
				data:    3.14,
				checkFn: func(a, b interface{}) bool { return a.(float64) == b.(float64) },
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				switch data := tc.data.(type) {
				case int:
					manager := NewJWTManager[int](secretKey, tokenDuration)
					token, err := manager.GenerateToken(data)
					if err != nil {
						t.Fatalf("GenerateToken failed: %v", err)
					}

					parsed, err := manager.ParseToken(token)
					if err != nil {
						t.Fatalf("ParseToken failed: %v", err)
					}

					if parsed != data {
						t.Errorf("Expected %v, got %v", data, parsed)
					}

				case bool:
					manager := NewJWTManager[bool](secretKey, tokenDuration)
					token, err := manager.GenerateToken(data)
					if err != nil {
						t.Fatalf("GenerateToken failed: %v", err)
					}

					parsed, err := manager.ParseToken(token)
					if err != nil {
						t.Fatalf("ParseToken failed: %v", err)
					}

					if parsed != data {
						t.Errorf("Expected %v, got %v", data, parsed)
					}

				case float64:
					manager := NewJWTManager[float64](secretKey, tokenDuration)
					token, err := manager.GenerateToken(data)
					if err != nil {
						t.Fatalf("GenerateToken failed: %v", err)
					}

					parsed, err := manager.ParseToken(token)
					if err != nil {
						t.Fatalf("ParseToken failed: %v", err)
					}

					if parsed != data {
						t.Errorf("Expected %v, got %v", data, parsed)
					}
				}
			})
		}
	})
}
