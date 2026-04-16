package user_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	httprouter "crud-api/internal/adapters/http/router"
	httpuser "crud-api/internal/adapters/http/user"
	"crud-api/internal/domain/user"
	"crud-api/internal/ports"
	"crud-api/internal/ports/mocks"
)

type HandlersSuite struct {
	suite.Suite
	router *gin.Engine
	svc    *mocks.MockUserService
}

func (s *HandlersSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.svc = mocks.NewMockUserService(s.T())
	handlers := httpuser.NewHandlers(s.svc)
	s.router = httprouter.New(handlers, nil, nil, "test")
}

func (s *HandlersSuite) do(method, path string, body any) *httptest.ResponseRecorder {
	s.T().Helper()
	var r *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	} else {
		r = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, r)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

func (s *HandlersSuite) TestCreate() {
	tests := []struct {
		name       string
		body       any
		setup      func()
		wantStatus int
	}{
		{
			name: "Success",
			body: map[string]any{"name": "Ada Lovelace", "email": "ada@example.com"},
			setup: func() {
				s.svc.EXPECT().Create(mock.Anything, ports.CreateUserInput{Name: "Ada Lovelace", Email: "ada@example.com"}).
					Return(&user.User{ID: "u1", Name: "Ada Lovelace", Email: "ada@example.com"}, nil).
					Once()
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "Email Conflict",
			body: map[string]any{"name": "A", "email": "dup@example.com"},
			setup: func() {
				s.svc.EXPECT().Create(mock.Anything, ports.CreateUserInput{Name: "A", Email: "dup@example.com"}).
					Return((*user.User)(nil), user.ErrEmailTaken).
					Once()
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "Invalid JSON",
			body:       "{invalid",
			setup:      func() {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setup()
			var w *httptest.ResponseRecorder
			if strBody, ok := tt.body.(string); ok && strBody == "{invalid" {
				// Special case for manual bad JSON
				req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewReader([]byte(strBody)))
				req.Header.Set("Content-Type", "application/json")
				w = httptest.NewRecorder()
				s.router.ServeHTTP(w, req)
			} else {
				w = s.do(http.MethodPost, "/v1/users", tt.body)
			}
			s.Equal(tt.wantStatus, w.Code)
		})
	}
}

func (s *HandlersSuite) TestGet() {
	tests := []struct {
		name       string
		userID     string
		setup      func()
		wantStatus int
	}{
		{
			name:   "Success",
			userID: "u1",
			setup: func() {
				s.svc.EXPECT().Get(mock.Anything, "u1").
					Return(&user.User{ID: "u1", Name: "Ada"}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "Not Found",
			userID: "u2",
			setup: func() {
				s.svc.EXPECT().Get(mock.Anything, "u2").
					Return((*user.User)(nil), user.ErrNotFound).
					Once()
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setup()
			w := s.do(http.MethodGet, "/v1/users/"+tt.userID, nil)
			s.Equal(tt.wantStatus, w.Code)
		})
	}
}

func (s *HandlersSuite) TestUpdate() {
	s.Run("Success", func() {
		updated := &user.User{ID: "u1", Name: "Ada L.", Email: "ada@example.com"}
		s.svc.EXPECT().Update(mock.Anything, "u1", ports.UpdateUserInput{Name: ptr("Ada L."), Email: nil}).
			Return(updated, nil).
			Once()
		w := s.do(http.MethodPut, "/v1/users/u1", map[string]any{"name": "Ada L."})
		s.Equal(http.StatusOK, w.Code)
	})
}

func (s *HandlersSuite) TestList() {
	s.Run("Success", func() {
		s.svc.EXPECT().List(mock.Anything, 1, 10).Return([]user.User{}, int64(0), nil).Once()
		w := s.do(http.MethodGet, "/v1/users?page=1&pageSize=10", nil)
		s.Equal(http.StatusOK, w.Code)
	})
}

func (s *HandlersSuite) TestDelete() {
	s.Run("Success", func() {
		s.svc.EXPECT().Delete(mock.Anything, "u1").Return(nil).Once()
		w := s.do(http.MethodDelete, "/v1/users/u1", nil)
		s.Equal(http.StatusNoContent, w.Code)
	})
}

func TestHandlersSuite(t *testing.T) {
	suite.Run(t, new(HandlersSuite))
}

func ptr(v string) *string { return &v }
