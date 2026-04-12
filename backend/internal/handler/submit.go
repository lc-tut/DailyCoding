package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"daileycoding/backend/internal/domain"
	"daileycoding/backend/internal/usecase"
)

type judger interface {
	Judge(ctx context.Context, code string) (*domain.JudgeResponse, error)
}

// SubmitHandler は POST /api/submit を処理する
type SubmitHandler struct {
	usecase judger
}

func NewSubmitHandler(uc *usecase.JudgeUsecase) *SubmitHandler {
	return &SubmitHandler{usecase: uc}
}

func (h *SubmitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Code == "" {
		http.Error(w, "code is required", http.StatusBadRequest)
		return
	}

	result, err := h.usecase.Judge(r.Context(), req.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
