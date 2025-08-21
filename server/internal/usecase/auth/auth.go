package auth

import (
	"context"
	"fmt"

	"github.com/DanilNaum/secret-storage-server/internal/entity"

	"github.com/thanhpk/randstr"
)

const saltLength = 10

type repository interface {
	CreateUser(ctx context.Context, user *entity.UserDTO) (int, error)
	GetUserByLogin(ctx context.Context, login string) (*entity.UserDTO, error)
}

type jwtManager interface {
	GenerateToken(uuid int) (string, error)
	ParseToken(token string) (int, error)
}

type hasher interface {
	Verify(password, encodedHash string) (bool, error)
	Hash(password string) (string, error)
}

type authUsecase struct {
	repo       repository
	jwtManager jwtManager
	hasher     hasher
}

func NewAuthUsecase(repo repository, jwtManager jwtManager, hasher hasher) *authUsecase {
	return &authUsecase{
		repo:       repo,
		jwtManager: jwtManager,
		hasher:     hasher,
	}
}

func (u *authUsecase) Authenticate(ctx context.Context, login, password string) (*entity.AuthData, error) {
	user, err := u.repo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("invalid login or password")
	}
	if ok, err := u.hasher.Verify(password, user.PasswordHash); !ok || err != nil {

		return nil, fmt.Errorf("invalid login or password")
	}
	token, err := u.jwtManager.GenerateToken(user.ID)

	return &entity.AuthData{
		JWTToken: token,
		Salt:     user.Salt,
	}, nil

}
func (u *authUsecase) Register(ctx context.Context, login, password string) (*entity.AuthData, error) {
	hash, err := u.hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("unexpected server error, failed to hash password: %w", err)
	}
	user := &entity.UserDTO{
		Login:        login,
		PasswordHash: hash,
		Salt:         randstr.String(saltLength),
	}
	id, err := u.repo.CreateUser(ctx, user)
	if err != nil {
		// TODO: обработать ошибки, в случае дублирования логина
		return nil, fmt.Errorf("unexpected server error, failed to create user")
	}

	token, err := u.jwtManager.GenerateToken(id)

	return &entity.AuthData{
		JWTToken: token,
		Salt:     user.Salt,
	}, nil
}
