package hasher

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// ErrMismatchedHashAndPassword возвращается когда пароль не соответствует хешу
var ErrMismatchedHashAndPassword = errors.New("password does not match hash")

// Params параметры для алгоритма Argon2
type Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultParams стандартные параметры для Argon2
var DefaultParams = &Params{
	Memory:      64 * 1024, // 64 MB
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

// Argon2Hasher реализация Hasher для Argon2
type Argon2Hasher struct {
	params *Params
}

// NewArgon2Hasher создает новый Argon2Hasher
func NewArgon2Hasher(params *Params) *Argon2Hasher {
	if params == nil {
		params = DefaultParams
	}
	return &Argon2Hasher{params: params}
}

// Hash создает хеш пароля
func (h *Argon2Hasher) Hash(password string) (string, error) {
	// Генерируем случайную соль
	salt := make([]byte, h.params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Создаем хеш с помощью Argon2
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.params.Iterations,
		h.params.Memory,
		h.params.Parallelism,
		h.params.KeyLength,
	)

	// Кодируем хеш в строку
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Форматируем строку хеша
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.params.Memory,
		h.params.Iterations,
		h.params.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

// Verify проверяет соответствие пароля хешу
func (h *Argon2Hasher) Verify(password, encodedHash string) (bool, error) {
	// Парсим закодированный хеш
	params, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// Создаем хеш для проверки
	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	// Сравниваем хеши безопасным способом (constant-time comparison)
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}

	return false, nil
}

// decodeHash парсит закодированный хеш
func decodeHash(encodedHash string) (*Params, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, errors.New("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return nil, nil, nil, errors.New("unsupported algorithm")
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse version: %w", err)
	}

	if version != argon2.Version {
		return nil, nil, nil, errors.New("incompatible version")
	}

	params := &Params{}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", 
		&params.Memory, &params.Iterations, &params.Parallelism)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode salt: %w", err)
	}
	params.SaltLength = uint32(len(salt))

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode hash: %w", err)
	}
	params.KeyLength = uint32(len(hash))

	return params, salt, hash, nil
}