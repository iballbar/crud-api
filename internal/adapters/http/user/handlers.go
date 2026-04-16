package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"crud-api/internal/ports"
)

type Handlers struct {
	userService ports.UserService
}

func NewHandlers(userService ports.UserService) *Handlers {
	return &Handlers{userService: userService}
}

func (h *Handlers) Create(c *gin.Context) {
	var req ports.CreateUserInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	u, err := h.userService.Create(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, u)
}

func (h *Handlers) Get(c *gin.Context) {
	id := c.Param("id")

	u, err := h.userService.Get(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, u)
}

func (h *Handlers) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))

	items, total, err := h.userService.List(c.Request.Context(), page, pageSize)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": total,
	})
}

func (h *Handlers) Update(c *gin.Context) {
	id := c.Param("id")
	var req ports.UpdateUserInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	u, err := h.userService.Update(c.Request.Context(), id, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, u)
}

func (h *Handlers) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.userService.Delete(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
