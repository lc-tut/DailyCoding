package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"

	"daileycoding/backend/internal/domain"
)

type problemRepo interface {
	ListActive(ctx context.Context) ([]domain.Problem, error)
	Create(ctx context.Context, params domain.CreateProblemParams) (*domain.Problem, error)
}

// ProblemHandler は GET /api/problems と POST /api/problems を処理する
type ProblemHandler struct {
	repo problemRepo
}

func NewProblemHandler(repo problemRepo) *ProblemHandler {
	return &ProblemHandler{repo: repo}
}

// problemResponse はクライアントへ返す問題の表現（テストケース解答は含まない）
type problemResponse struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	IsActive    bool    `json:"is_active"`
	PublishedAt *string `json:"published_at,omitempty"`
}

func (h *ProblemHandler) List(c *echo.Context) error {
	problems, err := h.repo.ListActive(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}

	resp := make([]problemResponse, len(problems))
	for i, p := range problems {
		resp[i] = toProblemResponse(p)
	}

	return c.JSON(http.StatusOK, resp)
}

type createTestCaseRequest struct {
	Stdin          string  `json:"stdin"`
	ExpectedStdout string  `json:"expected_stdout"`
	OrderIndex     int16   `json:"order_index"`
	IsSample       bool    `json:"is_sample"`
	Hint           *string `json:"hint"`
	ScoreWeight    int16   `json:"score_weight"`
}

type createProblemRequest struct {
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	IsActive    bool                    `json:"is_active"`
	PublishedAt *string                 `json:"published_at"` // "YYYY-MM-DD"
	TestCases   []createTestCaseRequest `json:"test_cases"`
}

func (h *ProblemHandler) Create(c *echo.Context) error {
	var req createProblemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid request body"})
	}
	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "title is required"})
	}
	if len(req.TestCases) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "at least one test case is required"})
	}

	var publishedAt *time.Time
	if req.PublishedAt != nil && *req.PublishedAt != "" {
		t, err := time.Parse("2006-01-02", *req.PublishedAt)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid published_at format, expected YYYY-MM-DD"})
		}
		publishedAt = &t
	}

	tcParams := make([]domain.CreateTestCaseParams, len(req.TestCases))
	for i, tc := range req.TestCases {
		tcParams[i] = domain.CreateTestCaseParams{
			Stdin:          tc.Stdin,
			ExpectedStdout: tc.ExpectedStdout,
			OrderIndex:     tc.OrderIndex,
			IsSample:       tc.IsSample,
			Hint:           tc.Hint,
			ScoreWeight:    tc.ScoreWeight,
		}
	}

	p, err := h.repo.Create(c.Request().Context(), domain.CreateProblemParams{
		Title:       req.Title,
		Description: req.Description,
		IsActive:    req.IsActive,
		PublishedAt: publishedAt,
		TestCases:   tcParams,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, toProblemResponse(*p))
}

func toProblemResponse(p domain.Problem) problemResponse {
	var publishedAt *string
	if p.PublishedAt != nil {
		s := p.PublishedAt.Format("2006-01-02")
		publishedAt = &s
	}
	return problemResponse{
		ID:          p.ID.String(),
		Title:       p.Title,
		Description: p.Description,
		IsActive:    p.IsActive,
		PublishedAt: publishedAt,
	}
}
