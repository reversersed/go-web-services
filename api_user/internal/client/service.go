package client

import (
	"context"
	"time"

	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/logging"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	storage Storage
	logger  *logging.Logger
}

func NewService(storage Storage, logger *logging.Logger) *service {
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
func (s *service) RegisterUser(ctx context.Context, login, password string) (*User, error) {
	pass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return nil, err
	}
	user := User{
		Login:    login,
		Password: pass,
		Roles:    []string{"user"},
	}
	cntx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_, err = s.storage.FindByLogin(cntx, user.Login)
	if err == nil {
		s.logger.Warnf("user %s couldn't register because of login collision", user.Login)
		return nil, errormiddleware.NotUniqueError([]string{"user with provided login already exist"}, "error while registering user")
	}
	result, err := s.storage.AddUser(cntx, &user)
	if err != nil {
		return nil, err
	}
	user.Id, err = primitive.ObjectIDFromHex(result)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
