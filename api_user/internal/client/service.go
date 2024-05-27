package client

import (
	"context"
	"fmt"
	"time"

	"github.com/reversersed/go-web-services/tree/main/api_user/internal/email"
	"github.com/reversersed/go-web-services/tree/main/api_user/internal/rabbitmq"
	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/cache"
	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/logging"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	storage      Storage
	logger       *logging.Logger
	cache        cache.Cache
	rabbitSender *rabbitmq.Sender
}

func NewService(storage Storage, logger *logging.Logger, cache cache.Cache, rabbitSender *rabbitmq.Sender) *service {
	return &service{storage: storage, logger: logger, cache: cache, rabbitSender: rabbitSender}
}
func (s *service) SendEmailConfirmation(ctx context.Context, userId string) error {
	u, err := s.storage.FindById(ctx, userId)
	if err != nil {
		return err
	}
	if u.EmailConfirmed {
		return errormiddleware.BadRequestError([]string{"user's email already confirmed"}, "can't send message to already confirmed email")
	}
	if _, err := s.cache.Get([]byte(fmt.Sprintf("cd%s", userId))); err == nil {
		return errormiddleware.ForbiddenError([]string{"message resending cooldown still not expired"}, "can't send message now because of cooldown")
	}
	code := primitive.NewObjectID().Hex()
	ok := email.SendEmailConfirmationMessage(u.Email, u.Login, code)
	if !ok {
		return fmt.Errorf("can't send email message")
	}
	err = s.cache.Set([]byte(fmt.Sprintf("cd%s", userId)), []byte("cooldown"), int(time.Minute/time.Second))
	if err != nil {
		return err
	}

	err = s.cache.Set([]byte(fmt.Sprintf("code%s", userId)), []byte(code), int((10*time.Minute)/time.Second))
	if err != nil {
		return err
	}
	s.logger.Infof("email confirmation cached. now there are %d entries in cache", s.cache.EntryCount())
	return nil
}
func (s *service) ValidateEmailConfirmationCode(ctx context.Context, userId string, code string) error {
	cached_code, err := s.cache.Get([]byte(fmt.Sprintf("code%s", userId)))
	if err != nil {
		return errormiddleware.ValidationErrorByString([]string{"user has no stored code or code is expired"}, "can't find cached code by user's email")
	}
	if string(cached_code) != code {
		return errormiddleware.ValidationErrorByString([]string{"code is incorrect"}, "code has found in cache, but provided code is incorrect. maybe the wrong link")
	}
	s.cache.Delete([]byte(fmt.Sprintf("cd%s", userId)))
	s.cache.Delete([]byte(fmt.Sprintf("code%s", userId)))

	err = s.storage.ApproveUserEmail(ctx, userId)
	if err != nil {
		return err
	}
	return nil
}
func (s *service) SignInUser(ctx context.Context, model *AuthUserByLoginAndPassword) (*User, error) {
	u, err := s.storage.FindByLogin(ctx, model.Login)
	if err != nil {
		u, err = s.storage.FindByEmail(ctx, model.Login)
		if err != nil {
			return nil, err
		}
	}

	if err = bcrypt.CompareHashAndPassword(u.Password, []byte(model.Password)); err != nil {
		return nil, err
	}

	return u, nil
}
func (s *service) RegisterUser(ctx context.Context, model *RegisterUserQuery) (*User, error) {
	pass, err := bcrypt.GenerateFromPassword([]byte(model.Password), bcrypt.MinCost)
	if err != nil {
		return nil, err
	}
	user := User{
		Login:          model.Login,
		Password:       pass,
		Roles:          []string{"user"},
		Email:          model.Email,
		EmailConfirmed: false,
	}
	cntx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_, err = s.storage.FindByLogin(cntx, user.Login)
	if err == nil {
		s.logger.Warnf("user %s couldn't register because of login collision", user.Login)
		return nil, errormiddleware.NotUniqueError([]string{"user with provided login already exist"}, "error while registering user")
	}
	_, err = s.storage.FindByEmail(cntx, user.Email)
	if err == nil {
		s.logger.Warnf("user %s couldn't register because of email (%s) collision", user.Login, user.Email)
		return nil, errormiddleware.NotUniqueError([]string{"email already taken"}, "error while registering user")
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
func (s *service) GetUserById(ctx context.Context, userId string) (*User, error) {
	cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	u, err := s.storage.FindById(cntx, userId)
	if err != nil {
		return nil, errormiddleware.NotFoundError([]string{"user with provided id not found"}, err.Error())
	}
	return u, nil
}
func (s *service) GetUserByLogin(ctx context.Context, login string) (*User, error) {
	cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	u, err := s.storage.FindByLogin(cntx, login)
	if err != nil {
		return nil, errormiddleware.NotFoundError([]string{"user with provided login not found"}, err.Error())
	}
	return u, nil
}
