package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	domain "crud-api/internal/domain/user"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		var status int
		var message string

		switch {
		case errors.Is(err, domain.ErrNotFound):
			status = http.StatusNotFound
			message = "resource not found"
		case errors.Is(err, domain.ErrEmailTaken):
			status = http.StatusConflict
			message = "email already taken"
		default:
			status = http.StatusInternalServerError
			message = "internal server error"
		}

		c.AbortWithStatusJSON(status, gin.H{"error": message})
	}
}
