package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"daileycoding/backend/internal/domain"
)

type problemRepo interface {
	ListActive(ctx context.Context) ([]domain.Problem, error)
	ListAll(ctx context.Context, params domain.ListAllParams) ([]domain.Problem, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Problem, error)
	GetToday(ctx context.Context, today string) (*domain.Problem, error)
	Create(ctx context.Context, params domain.CreateProblemParams) (*domain.Problem, error)
	Update(ctx context.Context, id uuid.UUID, params domain.UpdateProblemParams) (*domain.Problem, error)
}

// ProblemHandler は /api/problems/* を処理する
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
	PublishedAt *string `json:"published_at"`
}

// List GET /api/problems
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

// ListAll GET /api/admin/problems — 管理用: filter/sort クエリパラメータに応じて絞込・並び替え
func (h *ProblemHandler) ListAll(c *echo.Context) error {
	params := domain.ListAllParams{
		Filter: c.QueryParam("filter"),
		Sort:   c.QueryParam("sort"),
	}
	problems, err := h.repo.ListAll(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
	resp := make([]problemResponse, len(problems))
	for i, p := range problems {
		resp[i] = toProblemResponse(p)
	}
	return c.JSON(http.StatusOK, resp)
}

// Today GET /api/problems/today — JST 当日の問題を1件返す
func (h *ProblemHandler) Today(c *echo.Context) error {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "load timezone: " + err.Error()})
	}
	today := time.Now().In(jst).Format("2006-01-02")

	p, err := h.repo.GetToday(c.Request().Context(), today)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
	if p == nil {
		return c.JSON(http.StatusNotFound, map[string]any{"error": "no problem for today"})
	}
	return c.JSON(http.StatusOK, toProblemResponse(*p))
}

// GetByID GET /api/problems/:id
func (h *ProblemHandler) GetByID(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid problem id"})
	}

	p, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
	if p == nil {
		return c.JSON(http.StatusNotFound, map[string]any{"error": "problem not found"})
	}
	return c.JSON(http.StatusOK, toProblemResponse(*p))
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
	PublishedAt *string                 `json:"published_at"` // "YYYY-MM-DD" or null
	TestCases   []createTestCaseRequest `json:"test_cases"`
}

// Create POST /api/problems
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
		if errors.Is(err, domain.ErrDateConflict) {
			return c.JSON(http.StatusConflict, map[string]any{"error": "この日付はすでに別の問題に使用されています"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, toProblemResponse(*p))
}

type updateProblemRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	IsActive    bool    `json:"is_active"`
	PublishedAt *string `json:"published_at"` // "YYYY-MM-DD" or null
}

// Update PATCH /api/problems/:id
func (h *ProblemHandler) Update(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid problem id"})
	}

	var req updateProblemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid request body"})
	}
	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "title is required"})
	}

	var publishedAt *time.Time
	if req.PublishedAt != nil && *req.PublishedAt != "" {
		t, err := time.Parse("2006-01-02", *req.PublishedAt)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid published_at format, expected YYYY-MM-DD"})
		}
		publishedAt = &t
	}

	p, err := h.repo.Update(c.Request().Context(), id, domain.UpdateProblemParams{
		Title:       req.Title,
		Description: req.Description,
		IsActive:    req.IsActive,
		PublishedAt: publishedAt,
	})
	if err != nil {
		if errors.Is(err, domain.ErrDateConflict) {
			return c.JSON(http.StatusConflict, map[string]any{"error": "この日付はすでに別の問題に使用されています"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
	if p == nil {
		return c.JSON(http.StatusNotFound, map[string]any{"error": "problem not found"})
	}
	return c.JSON(http.StatusOK, toProblemResponse(*p))
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
