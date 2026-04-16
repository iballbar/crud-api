package user_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	applicationuser "crud-api/internal/application/user"
	"crud-api/internal/domain/user"
	"crud-api/internal/ports"
	"crud-api/internal/ports/mocks"
)

type ServiceSuite struct {
	suite.Suite
	repo *mocks.MockUserRepository
	svc  ports.UserService
	ctx  context.Context
}

func (s *ServiceSuite) SetupTest() {
	s.repo = mocks.NewMockUserRepository(s.T())
	s.svc = applicationuser.NewService(s.repo)
	s.ctx = context.Background()
}

func (s *ServiceSuite) TestCreate() {
	tests := []struct {
		name      string
		input     ports.CreateUserInput
		setup     func()
		wantErr   error
		checkUser func(*user.User)
	}{
		{
			name:  "Email already taken",
			input: ports.CreateUserInput{Name: "Ada", Email: "ADA@EXAMPLE.COM"},
			setup: func() {
				s.repo.EXPECT().ExistsByEmail(s.ctx, "ada@example.com").Return(true, nil).Once()
			},
			wantErr: user.ErrEmailTaken,
		},
		{
			name:  "Success and normalizes data",
			input: ports.CreateUserInput{Name: " Ada ", Email: " ADA@EXAMPLE.COM "},
			setup: func() {
				s.repo.EXPECT().ExistsByEmail(s.ctx, "ada@example.com").Return(false, nil).Once()
				s.repo.EXPECT().Create(s.ctx, mock.AnythingOfType("*user.User")).Return(nil).Once()
			},
			wantErr: nil,
			checkUser: func(u *user.User) {
				s.NotNil(u)
				s.Equal("Ada", u.Name)
				s.Equal("ada@example.com", u.Email)
				s.NotEmpty(u.ID)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setup()
			u, err := s.svc.Create(s.ctx, tt.input)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
				s.Nil(u)
			} else {
				s.NoError(err)
				if tt.checkUser != nil {
					tt.checkUser(u)
				}
			}
		})
	}
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}
