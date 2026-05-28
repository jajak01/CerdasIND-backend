package service

import (
	"context"
	"errors"
	"cerdasind-backend/internal/model"
	"cerdasind-backend/internal/repository"
	"cerdasind-backend/pkg/utils"
)

type AuthService interface {
	Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error)
	Register(ctx context.Context, req model.RegisterRequest) error
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("email tidak ditemukan")
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, errors.New("password salah")
	}

	token, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		Token:    token,
	}, nil
}

func (s *authService) Register(ctx context.Context, req model.RegisterRequest) error {
	existing, _ := s.userRepo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return errors.New("email sudah terdaftar")
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
		Role:         model.RolePeserta,
	}

	return s.userRepo.Create(ctx, user)
}
