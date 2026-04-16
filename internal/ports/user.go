package ports

import (
	"context"
	"crud-api/internal/domain/user"
	"time"
)

type CreateUserInput struct {
	Name  string `json:"name" binding:"required,min=1,max=200"`
	Email string `json:"email" binding:"required,email,max=320"`
}

type UpdateUserInput struct {
	Name  *string `json:"name" binding:"omitempty,min=1,max=200"`
	Email *string `json:"email" binding:"omitempty,email,max=320"`
}

//go:generate mockery --name=UserRepository --output=mocks --outpkg=mocks --filename=user_repository_mock.go
type UserRepository interface {
	Create(ctx context.Context, u *user.User) error
	GetByID(ctx context.Context, id string) (*user.User, error)
	List(ctx context.Context, offset, limit int) ([]user.User, int64, error)
	Update(ctx context.Context, u *user.User) error
	Delete(ctx context.Context, id string) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

//go:generate mockery --name=UserCache --output=mocks --outpkg=mocks --filename=user_cache_mock.go
type UserCache interface {
	Set(ctx context.Context, u *user.User, ttl time.Duration) error
	Get(ctx context.Context, id string) (*user.User, error)
	Delete(ctx context.Context, id string) error
}

//go:generate mockery --name=UserService --output=mocks --outpkg=mocks --filename=user_service_mock.go
type UserService interface {
	Create(ctx context.Context, req CreateUserInput) (*user.User, error)
	Get(ctx context.Context, id string) (*user.User, error)
	List(ctx context.Context, page, pageSize int) ([]user.User, int64, error)
	Update(ctx context.Context, id string, req UpdateUserInput) (*user.User, error)
	Delete(ctx context.Context, id string) error
}
