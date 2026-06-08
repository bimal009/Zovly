package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/bimal009/Zovly/internal/config"
	"github.com/bimal009/Zovly/internal/dto"
	"github.com/bimal009/Zovly/internal/models"
	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/bimal009/Zovly/pkg/utils"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthService interface {
	ForgotPassword(ctx context.Context, email string) error
	VerifyOTP(ctx context.Context, email, otp string) error
	ResetPassword(ctx context.Context, email, otp, newPassword string) error
}

type authService struct {
	db          *sqlx.DB
	userRepo    repository.UserRepo
	accountRepo repository.AccountRepo
	rdb         *redis.Client
	logger      *slog.Logger
	jwtUtil     *utils.JWTUtil
}

func NewAuthService(
	db *sqlx.DB,
	userRepo repository.UserRepo,
	accountRepo repository.AccountRepo,
	rdb *redis.Client,
	logger *slog.Logger,
	jwtUtil *utils.JWTUtil,
) AuthService {
	return &authService{
		db:          db,
		userRepo:    userRepo,
		accountRepo: accountRepo,
		rdb:         rdb,
		logger:      logger,
		jwtUtil:     jwtUtil,
	}
}


func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil || u == nil {
		return nil
	}

	otp, err := utils.GenerateOTP()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	if err := s.rdb.Set(ctx, fmt.Sprintf("otp:reset:%s", email), otp, 10*time.Minute).Err(); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	s.enqueuePasswordResetEmail(u.Name, email, otp)
	return nil
}


func (s *authService) VerifyOTP(ctx context.Context, email, otp string) error {
	stored, err := s.rdb.Get(ctx, fmt.Sprintf("otp:reset:%s", email)).Result()
	if err != nil {
		return errors.New("OTP expired or not found")
	}
	if stored != otp {
		return errors.New("invalid OTP")
	}
	return nil
}


func (s *authService) ResetPassword(ctx context.Context, email, otp, newPassword string) error {
	if err := s.VerifyOTP(ctx, email, otp); err != nil {
		return err
	}

	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil || u == nil {
		return errors.New("user not found")
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.accountRepo.UpdatePassword(ctx, u.ID, string(models.AccountTypeCredential), hash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.rdb.Del(ctx, fmt.Sprintf("otp:reset:%s", email))
	return nil
}
