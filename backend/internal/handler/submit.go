package handler

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"daileycoding/backend/internal/domain"
	"daileycoding/backend/internal/usecase"
)

type judger interface {
	Judge(ctx context.Context, problemID uuid.UUID, code string) (*domain.JudgeResponse, error)
}

// SubmitHandler は POST /api/submit を処理する
type SubmitHandler struct {
	usecase judger
}

func NewSubmitHandler(uc *usecase.JudgeUsecase) *SubmitHandler {
	return &SubmitHandler{usecase: uc}
}

func (h *SubmitHandler) Handle(c *echo.Context) error {
	var req domain.SubmitRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid request body"})
	}
	if req.Code == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "code is required"})
	}
	if req.ProblemID == uuid.Nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "problem_id is required"})
	}

	result, err := h.usecase.Judge(c.Request().Context(), req.ProblemID, req.Code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}
