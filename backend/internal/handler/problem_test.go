package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"daileycoding/backend/internal/domain"
	"daileycoding/backend/internal/handler"
)

// ---- モックリポジトリ ----

type mockProblemRepo struct {
	activeProblems []domain.Problem
	allProblems    []domain.Problem
	createdProblem *domain.Problem
	updatedProblem *domain.Problem
}

func (m *mockProblemRepo) ListActive(_ context.Context) ([]domain.Problem, error) {
	return m.activeProblems, nil
}
func (m *mockProblemRepo) ListAll(_ context.Context, _ domain.ListAllParams) ([]domain.Problem, error) {
	return m.allProblems, nil
}
func (m *mockProblemRepo) GetByID(_ context.Context, _ uuid.UUID) (*domain.Problem, error) {
	if len(m.allProblems) > 0 {
		return &m.allProblems[0], nil
	}
	return nil, nil
}
func (m *mockProblemRepo) GetToday(_ context.Context, _ string) (*domain.Problem, error) {
	if len(m.activeProblems) > 0 {
		return &m.activeProblems[0], nil
	}
	return nil, nil
}
func (m *mockProblemRepo) Create(_ context.Context, _ domain.CreateProblemParams) (*domain.Problem, error) {
	return m.createdProblem, nil
}
func (m *mockProblemRepo) Update(_ context.Context, _ uuid.UUID, _ domain.UpdateProblemParams) (*domain.Problem, error) {
	return m.updatedProblem, nil
}

// ---- テストヘルパー ----

func newEchoContext(method, path, body string) (*echo.Echo, *echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return e, c, rec
}

func parseBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var result map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response body: %v\nbody: %s", err, rec.Body.String())
	}
	return result
}

func parseArrayBody(t *testing.T, rec *httptest.ResponseRecorder) []map[string]any {
	t.Helper()
	var result []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response body: %v\nbody: %s", err, rec.Body.String())
	}
	return result
}

// ---- テストデータ ----

var (
	activeID   = uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	inactiveID = uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002")
	now        = time.Now()
)

func activeProblem() domain.Problem {
	return domain.Problem{
		ID: activeID, Title: "公開中の問題", Description: "本文A",
		IsActive: true, CreatedAt: now, UpdatedAt: now,
	}
}

func inactiveProblem() domain.Problem {
	return domain.Problem{
		ID: inactiveID, Title: "未公開の問題", Description: "本文B",
		IsActive: false, CreatedAt: now, UpdatedAt: now,
	}
}

// ---- List (GET /api/problems) ----

// List は is_active=true の問題だけを返すこと
func TestList_ReturnsOnlyActiveProblems(t *testing.T) {
	repo := &mockProblemRepo{
		activeProblems: []domain.Problem{activeProblem()},
		allProblems:    []domain.Problem{activeProblem(), inactiveProblem()},
	}
	h := handler.NewProblemHandler(repo)

	_, c, rec := newEchoContext(http.MethodGet, "/api/problems", "")
	if err := h.List(c); err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	items := parseArrayBody(t, rec)
	if len(items) != 1 {
		t.Fatalf("expected 1 problem, got %d", len(items))
	}
	if items[0]["id"] != activeID.String() {
		t.Errorf("expected active problem id, got %v", items[0]["id"])
	}
}

// List が空のとき空配列を返すこと
func TestList_EmptyWhenNoActiveProblems(t *testing.T) {
	repo := &mockProblemRepo{
		activeProblems: []domain.Problem{},
		allProblems:    []domain.Problem{inactiveProblem()},
	}
	h := handler.NewProblemHandler(repo)

	_, c, rec := newEchoContext(http.MethodGet, "/api/problems", "")
	if err := h.List(c); err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	items := parseArrayBody(t, rec)
	if len(items) != 0 {
		t.Fatalf("expected 0 problems, got %d", len(items))
	}
}

// ---- ListAll (GET /api/admin/problems) ----

// ListAll は is_active に関わらず全問題を返すこと — 管理画面のバグ回帰テスト
func TestListAll_ReturnsAllProblems(t *testing.T) {
	repo := &mockProblemRepo{
		activeProblems: []domain.Problem{activeProblem()},
		allProblems:    []domain.Problem{activeProblem(), inactiveProblem()},
	}
	h := handler.NewProblemHandler(repo)

	_, c, rec := newEchoContext(http.MethodGet, "/api/admin/problems", "")
	if err := h.ListAll(c); err != nil {
		t.Fatalf("ListAll returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	items := parseArrayBody(t, rec)
	if len(items) != 2 {
		t.Fatalf("expected 2 problems (active+inactive), got %d\nbody: %s", len(items), rec.Body.String())
	}

	ids := map[string]bool{}
	for _, item := range items {
		ids[item["id"].(string)] = true
	}
	if !ids[activeID.String()] {
		t.Errorf("active problem %s not found in response", activeID)
	}
	if !ids[inactiveID.String()] {
		t.Errorf("inactive problem %s not found in response", inactiveID)
	}
}

// ListAll と List が同じエンドポイントに対応していないことを確認
func TestListAll_IncludesInactiveProblems_ListDoesNot(t *testing.T) {
	repo := &mockProblemRepo{
		activeProblems: []domain.Problem{},
		allProblems:    []domain.Problem{inactiveProblem()},
	}
	h := handler.NewProblemHandler(repo)

	// List (公開用) → 0件
	_, cList, recList := newEchoContext(http.MethodGet, "/api/problems", "")
	if err := h.List(cList); err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	listItems := parseArrayBody(t, recList)
	if len(listItems) != 0 {
		t.Errorf("List should return 0 items for inactive-only problems, got %d", len(listItems))
	}

	// ListAll (管理用) → 1件
	_, cAll, recAll := newEchoContext(http.MethodGet, "/api/admin/problems", "")
	if err := h.ListAll(cAll); err != nil {
		t.Fatalf("ListAll returned error: %v", err)
	}
	allItems := parseArrayBody(t, recAll)
	if len(allItems) != 1 {
		t.Errorf("ListAll should return 1 item for inactive-only problems, got %d", len(allItems))
	}
}

// ---- Create (POST /api/problems) ----

// Create が成功したとき 201 と問題を返すこと
func TestCreate_Success(t *testing.T) {
	created := activeProblem()
	repo := &mockProblemRepo{createdProblem: &created}
	h := handler.NewProblemHandler(repo)

	body := `{"title":"テスト問題","description":"本文","is_active":true,"test_cases":[{"stdin":"1","expected_stdout":"1","order_index":1,"is_sample":true,"score_weight":1}]}`
	_, c, rec := newEchoContext(http.MethodPost, "/api/problems", body)
	if err := h.Create(c); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d\nbody: %s", rec.Code, rec.Body.String())
	}
	result := parseBody(t, rec)
	if result["id"] != activeID.String() {
		t.Errorf("unexpected id: %v", result["id"])
	}
}

// Create でタイトルが空のとき 400 を返すこと
func TestCreate_MissingTitle(t *testing.T) {
	repo := &mockProblemRepo{}
	h := handler.NewProblemHandler(repo)

	body := `{"title":"","description":"本文","test_cases":[{"stdin":"1","expected_stdout":"1","order_index":1}]}`
	_, c, rec := newEchoContext(http.MethodPost, "/api/problems", body)
	if err := h.Create(c); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// Create でテストケースが0のとき 400 を返すこと
func TestCreate_NoTestCases(t *testing.T) {
	repo := &mockProblemRepo{}
	h := handler.NewProblemHandler(repo)

	body := `{"title":"タイトル","description":"本文","test_cases":[]}`
	_, c, rec := newEchoContext(http.MethodPost, "/api/problems", body)
	if err := h.Create(c); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
