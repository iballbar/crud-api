package user

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	domain "crud-api/internal/domain/user"
	"crud-api/internal/ports"
)

type Service struct {
	repo ports.UserRepository
}

func NewService(repo ports.UserRepository) ports.UserService {
	return &Service{repo: repo}
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (s *Service) Create(ctx context.Context, req ports.CreateUserInput) (*domain.User, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	exists, err := s.repo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrEmailTaken
	}

	now := time.Now().UTC()
	u := &domain.User{
		ID:        uuid.NewString(),
		Name:      strings.TrimSpace(req.Name),
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) Get(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, page, pageSize int) ([]domain.User, int64, error) {
	page = clamp(page, 1, 10_000)
	pageSize = clamp(pageSize, 1, 200)
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}

func (s *Service) Update(ctx context.Context, id string, req ports.UpdateUserInput) (*domain.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Email != nil {
		email := strings.TrimSpace(strings.ToLower(*req.Email))
		if email != u.Email {
			exists, err := s.repo.ExistsByEmail(ctx, email)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, domain.ErrEmailTaken
			}
			u.Email = email
		}
	}
	if req.Name != nil {
		u.Name = strings.TrimSpace(*req.Name)
	}

	u.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
