package hasher

import (
	"strings"
	"testing"
)

func TestArgon2Hasher_Hash(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		params    *Params
		wantError bool
	}{
		{
			name:      "successful hash with default params",
			password:  "mysecretpassword",
			params:    nil,
			wantError: false,
		},
		{
			name:      "successful hash with custom params",
			password:  "anotherpassword",
			params:    &Params{Memory: 32 * 1024, Iterations: 2, Parallelism: 1, SaltLength: 16, KeyLength: 32},
			wantError: false,
		},
		{
			name:      "empty password",
			password:  "",
			params:    nil,
			wantError: false,
		},
		{
			name:      "long password",
			password:  strings.Repeat("a", 1000),
			params:    nil,
			wantError: false,
		},
		{
			name:      "special characters password",
			password:  "p@ssw0rd!§$%&/()=?*'#+~",
			params:    nil,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewArgon2Hasher(tt.params)
			hash, err := hasher.Hash(tt.password)

			if (err != nil) != tt.wantError {
				t.Errorf("Hash() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				// Проверяем формат хеша
				if !strings.HasPrefix(hash, "$argon2id$") {
					t.Errorf("Hash should start with $argon2id$ prefix, got: %s", hash)
				}

				parts := strings.Split(hash, "$")
				if len(parts) != 6 {
					t.Errorf("Hash should have 6 parts, got %d: %s", len(parts), hash)
				}

				// Проверяем что хеши разные для одного пароля (из-за соли)
				hash2, err := hasher.Hash(tt.password)
				if err != nil {
					t.Errorf("Second hash failed: %v", err)
				}
				if hash == hash2 {
					t.Errorf("Hashes should be different due to different salt: %s == %s", hash, hash2)
				}
			}
		})
	}
}

func TestArgon2Hasher_Verify(t *testing.T) {
	hasher := NewArgon2Hasher(nil)
	
	// Сначала создаем валидный хеш для тестов
	validPassword := "correctpassword"
	validHash, err := hasher.Hash(validPassword)
	if err != nil {
		t.Fatalf("Failed to create test hash: %v", err)
	}

	tests := []struct {
		name          string
		password      string
		encodedHash   string
		wantMatch     bool
		wantError     bool
		errorContains string
	}{
		{
			name:        "correct password",
			password:    validPassword,
			encodedHash: validHash,
			wantMatch:   true,
			wantError:   false,
		},
		{
			name:        "wrong password",
			password:    "wrongpassword",
			encodedHash: validHash,
			wantMatch:   false,
			wantError:   false,
		},
		{
			name:        "empty password with valid hash",
			password:    "",
			encodedHash: validHash,
			wantMatch:   false,
			wantError:   false,
		},
		{
			name:          "malformed hash - wrong format",
			password:      validPassword,
			encodedHash:   "invalidhashformat",
			wantMatch:     false,
			wantError:     true,
			errorContains: "invalid hash format",
		},
		{
			name:          "malformed hash - wrong algorithm",
			password:      validPassword,
			encodedHash:   "$md5$v=1$m=65536,t=3,p=2$salt$hash",
			wantMatch:     false,
			wantError:     true,
			errorContains: "unsupported algorithm",
		},
		{
			name:          "malformed hash - wrong version",
			password:      validPassword,
			encodedHash:   "$argon2id$v=999$m=65536,t=3,p=2$c2FsdA==$aGFzaA==",
			wantMatch:     false,
			wantError:     true,
			errorContains: "incompatible version",
		},
		{
			name:          "malformed hash - invalid base64 salt",
			password:      validPassword,
			encodedHash:   "$argon2id$v=19$m=65536,t=3,p=2$invalid-base64$hash",
			wantMatch:     false,
			wantError:     true,
			errorContains: "failed to decode salt",
		},
		{
			name:          "malformed hash - invalid base64 hash",
			password:      validPassword,
			encodedHash:   "$argon2id$v=19$m=65536,t=3,p=2$c2FsdA==$invalid-base64",
			wantMatch:     false,
			wantError:     true,
			errorContains: "failed to decode hash",
		},
		{
			name:        "empty hash",
			password:    validPassword,
			encodedHash: "",
			wantMatch:   false,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := hasher.Verify(tt.password, tt.encodedHash)

			if (err != nil) != tt.wantError {
				t.Errorf("Verify() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if tt.wantError && tt.errorContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Error should contain '%s', got: %v", tt.errorContains, err)
				}
			}

			if match != tt.wantMatch {
				t.Errorf("Verify() match = %v, wantMatch %v", match, tt.wantMatch)
			}
		})
	}
}

func TestArgon2Hasher_Verify_SamePasswordDifferentHashes(t *testing.T) {
	hasher := NewArgon2Hasher(nil)
	password := "testpassword"

	// Генерируем несколько хешей для одного пароля
	hashes := make([]string, 5)
	for i := 0; i < 5; i++ {
		hash, err := hasher.Hash(password)
		if err != nil {
			t.Fatalf("Failed to generate hash %d: %v", i, err)
		}
		hashes[i] = hash
	}

	// Проверяем что все хеши работают с правильным паролем
	for i, hash := range hashes {
		match, err := hasher.Verify(password, hash)
		if err != nil {
			t.Errorf("Verify failed for hash %d: %v", i, err)
		}
		if !match {
			t.Errorf("Hash %d should verify with correct password", i)
		}

		// Проверяем что неправильный пароль не работает
		match, err = hasher.Verify("wrongpassword", hash)
		if err != nil {
			t.Errorf("Verify with wrong password failed for hash %d: %v", i, err)
		}
		if match {
			t.Errorf("Hash %d should not verify with wrong password", i)
		}
	}
}

