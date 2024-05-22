package client

import (
	"context"

	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/logging"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	SignInUser(ctx context.Context, login, password string) (*User, error)
}
type service struct {
	storage Storage
	logger  *logging.Logger
}

func NewService(storage Storage, logger *logging.Logger) Service {
	return &service{storage: storage, logger: logger}
}
func (s *service) SignInUser(ctx context.Context, login, password string) (*User, error) {
	u, err := s.storage.FindByLogin(ctx, login)
	if err != nil {
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword(u.Password, []byte(password)); err != nil {
		return nil, err
	}

	return u, nil
}
