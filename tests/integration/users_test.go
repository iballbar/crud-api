package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	httprouter "crud-api/internal/adapters/http/router"
	httpuser "crud-api/internal/adapters/http/user"
	postgresadapter "crud-api/internal/adapters/postgres"
	redisadapter "crud-api/internal/adapters/redis"
	applicationuser "crud-api/internal/application/user"
	userdecorator "crud-api/internal/application/user/decorator"
	"crud-api/internal/db"
	"crud-api/internal/domain/user"
	"crud-api/internal/ports"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type UsersPostgresSuite struct {
	suite.Suite
	ctx    context.Context
	cancel context.CancelFunc
	srv    *httptest.Server
	rdb    *redis.Client
}

func (s *UsersPostgresSuite) SetupSuite() {
	// Ryuk (the reaper) can fail to start on some Docker Desktop/Windows setups.
	// Disabling it keeps the test reliable; the container is still terminated via t.Cleanup.
	s.T().Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 3*time.Minute)

	// Start Postgres
	pgReq := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "test_users",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("5432/tcp"),
			wait.ForLog("database system is ready to accept connections"),
		).WithDeadline(90 * time.Second),
	}

	pg, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: pgReq,
		Started:          true,
	})
	if err != nil {
		s.T().Fatalf("start postgres container: %v", err)
	}
	s.T().Cleanup(func() { _ = pg.Terminate(context.Background()) })

	// Start Redis
	redisReq := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}

	redisC, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: redisReq,
		Started:          true,
	})
	if err != nil {
		s.T().Fatalf("start redis container: %v", err)
	}
	s.T().Cleanup(func() { _ = redisC.Terminate(context.Background()) })

	pgHost, _ := pg.Host(s.ctx)
	pgPort, _ := pg.MappedPort(s.ctx, "5432/tcp")
	dsn := "host=" + pgHost + " user=postgres password=postgres dbname=test_users port=" + pgPort.Port() + " sslmode=disable TimeZone=UTC"

	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{PrepareStmt: true})
	if err != nil {
		s.T().Fatalf("open gorm: %v", err)
	}

	if err := db.AutoMigrate(gdb); err != nil {
		s.T().Fatalf("migrate: %v", err)
	}

	redisHost, _ := redisC.Host(s.ctx)
	redisPort, _ := redisC.MappedPort(s.ctx, "6379/tcp")
	s.rdb = redis.NewClient(&redis.Options{
		Addr: redisHost + ":" + redisPort.Port(),
	})

	repo := postgresadapter.NewUserRepository(gdb)
	cache := redisadapter.NewUserCache(s.rdb)

	var svc ports.UserService
	svc = applicationuser.NewService(repo)
	svc = userdecorator.NewCacheDecorator(svc, cache, 30*time.Second)

	handlers := httpuser.NewHandlers(svc)
	router := httprouter.New(handlers, gdb, s.rdb, "test")

	s.srv = httptest.NewServer(router)
	s.T().Cleanup(s.srv.Close)
}

func (s *UsersPostgresSuite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *UsersPostgresSuite) doJSON(method, path string, body any) (*http.Response, []byte) {
	s.T().Helper()
	var r *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	} else {
		r = bytes.NewReader(nil)
	}
	req, _ := http.NewRequestWithContext(s.ctx, method, s.srv.URL+path, r)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		s.T().Fatalf("request %s %s: %v", method, path, err)
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	return res, data
}

func (s *UsersPostgresSuite) TestUsersLifecycle() {
	var created user.User

	s.Run("Create", func() {
		res, data := s.doJSON(http.MethodPost, "/v1/users", map[string]any{
			"name":  "Ada Lovelace",
			"email": "ada@example.com",
		})
		s.Require().Equal(http.StatusCreated, res.StatusCode, "body=%s", string(data))
		s.Require().NoError(json.Unmarshal(data, &created))
		s.Require().NotEmpty(created.ID)
		s.Equal("Ada Lovelace", created.Name)
	})

	s.Run("Create (Conflict)", func() {
		res, _ := s.doJSON(http.MethodPost, "/v1/users", map[string]any{
			"name":  "Duplicate",
			"email": "ada@example.com",
		})
		s.Equal(http.StatusConflict, res.StatusCode)
	})

	s.Run("Create (Invalid Email)", func() {
		res, _ := s.doJSON(http.MethodPost, "/v1/users", map[string]any{
			"name":  "Invalid",
			"email": "not-an-email",
		})
		s.Equal(http.StatusBadRequest, res.StatusCode)
	})

	s.Run("Get", func() {
		res, data := s.doJSON(http.MethodGet, "/v1/users/"+created.ID, nil)
		s.Require().Equal(http.StatusOK, res.StatusCode, "body=%s", string(data))
		var u user.User
		s.Require().NoError(json.Unmarshal(data, &u))
		s.Equal(created.ID, u.ID)
		s.Equal(created.Email, u.Email)

		// Verify it was cached in Redis
		val, err := s.rdb.Get(s.ctx, "user:{"+created.ID+"}").Result()
		s.NoError(err, "should be in redis after GET")
		s.NotEmpty(val)
	})

	s.Run("Update", func() {
		res, data := s.doJSON(http.MethodPut, "/v1/users/"+created.ID, map[string]any{"name": "Ada L."})
		s.Require().Equal(http.StatusOK, res.StatusCode, "body=%s", string(data))
		var u user.User
		s.Require().NoError(json.Unmarshal(data, &u))
		s.Equal("Ada L.", u.Name)
		s.Equal(created.Email, u.Email)

		// Verify cache was invalidated
		_, err := s.rdb.Get(s.ctx, "user:{"+created.ID+"}").Result()
		s.ErrorIs(err, redis.Nil, "should be deleted from redis after UPDATE")
	})

	s.Run("List", func() {
		res, data := s.doJSON(http.MethodGet, "/v1/users?page=1&pageSize=10", nil)
		s.Require().Equal(http.StatusOK, res.StatusCode, "body=%s", string(data))
		var body struct {
			Items []user.User `json:"items"`
			Total int64       `json:"total"`
		}
		s.Require().NoError(json.Unmarshal(data, &body))
		s.GreaterOrEqual(body.Total, int64(1))

		found := false
		for _, item := range body.Items {
			if item.ID == created.ID {
				found = true
				break
			}
		}
		s.True(found, "created user not found in list")
	})

	s.Run("Delete", func() {
		res, data := s.doJSON(http.MethodDelete, "/v1/users/"+created.ID, nil)
		s.Require().Equal(http.StatusNoContent, res.StatusCode, "body=%s", string(data))

		// Verify cache was invalidated
		_, err := s.rdb.Get(s.ctx, "user:{"+created.ID+"}").Result()
		s.ErrorIs(err, redis.Nil, "should be deleted from redis after DELETE")
	})

	s.Run("Get (NotFound after delete)", func() {
		res, _ := s.doJSON(http.MethodGet, "/v1/users/"+created.ID, nil)
		s.Require().Equal(http.StatusNotFound, res.StatusCode)
	})
}

func TestUsersPostgresSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short")
	}
	suite.Run(t, new(UsersPostgresSuite))
}
