package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"daileycoding/backend/internal/domain"
)

type submissionRepo interface {
	ListByProblemID(ctx context.Context, problemID uuid.UUID) ([]domain.SubmissionSummary, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SubmissionDetail, error)
}

// SubmissionHandler は /api/submissions/* と /api/problems/:id/submissions を処理する
type SubmissionHandler struct {
	repo submissionRepo
}

func NewSubmissionHandler(repo submissionRepo) *SubmissionHandler {
	return &SubmissionHandler{repo: repo}
}

// --- レスポンス型 ---

type submissionSummaryResponse struct {
	ID            string    `json:"id"`
	Language      string    `json:"language"`
	OverallStatus string    `json:"overall_status"`
	Score         int16     `json:"score"`
	CreatedAt     time.Time `json:"created_at"`
}

type testCaseResultResponse struct {
	OrderIndex      int16   `json:"order_index"`
	IsSample        bool    `json:"is_sample"`
	Stdin           string  `json:"stdin,omitempty"`
	ExpectedStdout  string  `json:"expected_stdout,omitempty"`
	Status          string  `json:"status"`
	Stdout          string  `json:"stdout"`
	Stderr          string  `json:"stderr"`
	ExecutionTimeMs float32 `json:"execution_time_ms"`
}

type submissionDetailResponse struct {
	ID            string                   `json:"id"`
	ProblemID     string                   `json:"problem_id"`
	Code          string                   `json:"code"`
	Language      string                   `json:"language"`
	OverallStatus string                   `json:"overall_status"`
	Score         int16                    `json:"score"`
	CreatedAt     time.Time                `json:"created_at"`
	Results       []testCaseResultResponse `json:"results"`
}

// List GET /api/problems/:id/submissions
func (h *SubmissionHandler) List(c *echo.Context) error {
	problemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid problem id"})
	}

	subs, err := h.repo.ListByProblemID(c.Request().Context(), problemID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}

	resp := make([]submissionSummaryResponse, len(subs))
	for i, s := range subs {
		resp[i] = submissionSummaryResponse{
			ID:            s.ID.String(),
			Language:      s.Language,
			OverallStatus: s.OverallStatus,
			Score:         s.Score,
			CreatedAt:     s.CreatedAt,
		}
	}
	return c.JSON(http.StatusOK, resp)
}

// GetByID GET /api/submissions/:id
func (h *SubmissionHandler) GetByID(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid submission id"})
	}

	sub, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
	if sub == nil {
		return c.JSON(http.StatusNotFound, map[string]any{"error": "submission not found"})
	}

	results := make([]testCaseResultResponse, len(sub.Results))
	for i, r := range sub.Results {
		results[i] = testCaseResultResponse{
			OrderIndex:      r.OrderIndex,
			IsSample:        r.IsSample,
			Stdin:           r.Stdin,
			ExpectedStdout:  r.ExpectedStdout,
			Status:          r.Status,
			Stdout:          r.Stdout,
			Stderr:          r.Stderr,
			ExecutionTimeMs: r.ExecutionTimeMs,
		}
	}

	return c.JSON(http.StatusOK, submissionDetailResponse{
		ID:            sub.ID.String(),
		ProblemID:     sub.ProblemID.String(),
		Code:          sub.Code,
		Language:      sub.Language,
		OverallStatus: sub.OverallStatus,
		Score:         sub.Score,
		CreatedAt:     sub.CreatedAt,
		Results:       results,
	})
}
