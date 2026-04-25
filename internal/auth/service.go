package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"ams-ai/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now()
}

type Service struct {
	repo   Repository
	secret string
	ttl    time.Duration
	clock  Clock
}

func NewService(repo Repository, secret string, ttl time.Duration, clock Clock) *Service {
	if clock == nil {
		clock = realClock{}
	}
	return &Service{repo: repo, secret: secret, ttl: ttl, clock: clock}
}

func (s *Service) Login(ctx context.Context, email, password string) (Token, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return Token{}, domain.ErrUnauthorized
		}
		return Token{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return Token{}, domain.ErrUnauthorized
	}
	expiresAt := s.clock.Now().Add(s.ttl)
	token, err := s.signToken(user.ID, user.Role, expiresAt)
	if err != nil {
		return Token{}, err
	}
	return Token{Token: token, ExpiresAt: expiresAt, User: user}, nil
}

func (s *Service) UserFromToken(ctx context.Context, token string) (domain.User, error) {
	userID, role, expiresAt, err := s.verifyToken(token)
	if err != nil {
		return domain.User{}, domain.ErrUnauthorized
	}
	if s.clock.Now().After(expiresAt) {
		return domain.User{}, domain.ErrUnauthorized
	}
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}
	if user.Role != role {
		return domain.User{}, domain.ErrUnauthorized
	}
	return user, nil
}

func (s *Service) ListUsers(ctx context.Context, user domain.User) ([]domain.User, error) {
	if user.Role != domain.RoleAdmin {
		return []domain.User{user}, nil
	}
	return s.repo.ListUsers(ctx)
}

func (s *Service) UpdateProfile(ctx context.Context, user domain.User, fullName, password string) (domain.User, error) {
	fullName = strings.TrimSpace(fullName)
	if fullName == "" {
		return domain.User{}, fmt.Errorf("%w: full name is required", domain.ErrInvalid)
	}
	var passwordHash *string
	if password != "" {
		if len(password) < 6 {
			return domain.User{}, fmt.Errorf("%w: password must be at least 6 characters", domain.ErrInvalid)
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return domain.User{}, err
		}
		hashed := string(hash)
		passwordHash = &hashed
	}
	return s.repo.UpdateUserProfile(ctx, user.ID, fullName, passwordHash)
}
