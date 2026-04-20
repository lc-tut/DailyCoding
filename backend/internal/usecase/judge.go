package usecase

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"

	"daileycoding/backend/internal/domain"
)

// Executor はコード実行エンジンの抽象インターフェース
type Executor interface {
	Execute(ctx context.Context, req domain.ExecuteRequest) (domain.ExecuteResult, error)
}

// ProblemRepository は問題・テストケース取得に必要な操作を定義する
type ProblemRepository interface {
	GetTestCasesByProblemID(ctx context.Context, problemID uuid.UUID) ([]domain.TestCase, error)
}

// SubmissionRepository は提出結果の永続化に必要な操作を定義する
type SubmissionRepository interface {
	Save(ctx context.Context, params domain.CreateSubmissionParams, results []domain.CreateSubmissionResultParams) (*domain.Submission, error)
}

// JudgeUsecase はコード採点のビジネスロジックを担う
type JudgeUsecase struct {
	executor       Executor
	problemRepo    ProblemRepository
	submissionRepo SubmissionRepository
}

func NewJudgeUsecase(
	executor Executor,
	problemRepo ProblemRepository,
	submissionRepo SubmissionRepository,
) *JudgeUsecase {
	return &JudgeUsecase{
		executor:       executor,
		problemRepo:    problemRepo,
		submissionRepo: submissionRepo,
	}
}

// Judge はテストケースを DB から取得し、実行・採点・保存を行う
func (u *JudgeUsecase) Judge(ctx context.Context, problemID uuid.UUID, code string, userID uuid.UUID) (*domain.JudgeResponse, error) {
	// 1. テストケースを DB から取得
	testCases, err := u.problemRepo.GetTestCasesByProblemID(ctx, problemID)
	if err != nil {
		return nil, fmt.Errorf("get test cases: %w", err)
	}
	if len(testCases) == 0 {
		return nil, fmt.Errorf("no test cases found for problem %s", problemID)
	}

	// 2. 各テストケースを実行・採点
	judgeResults := make([]domain.TestResult, 0, len(testCases))
	resultParams := make([]domain.CreateSubmissionResultParams, 0, len(testCases))
	overall := "Pass"
	passedWeight, totalWeight := 0, 0

	for _, tc := range testCases {
		totalWeight += int(tc.ScoreWeight)

		execResult, err := u.executor.Execute(ctx, domain.ExecuteRequest{
			Code:     code,
			Language: "python",
			Stdin:    tc.Stdin,
		})
		if err != nil {
			return nil, fmt.Errorf("execute case %d: %w", tc.OrderIndex, err)
		}

		status := "Pass"
		if strings.TrimSpace(execResult.Stdout) != strings.TrimSpace(tc.ExpectedStdout) || execResult.Stderr != "" {
			status = "Fail"
			overall = "Fail"
		} else {
			passedWeight += int(tc.ScoreWeight)
		}

		judgeResults = append(judgeResults, domain.TestResult{
			CaseIndex: int(tc.OrderIndex),
			Status:    status,
			Stdout:    execResult.Stdout,
			Expected:  tc.ExpectedStdout,
			Stderr:    execResult.Stderr,
			TimeMs:    execResult.TimeMs,
		})
		resultParams = append(resultParams, domain.CreateSubmissionResultParams{
			TestCaseID:      tc.ID,
			Status:          status,
			Stdout:          execResult.Stdout,
			Stderr:          execResult.Stderr,
			ExecutionTimeMs: float32(execResult.TimeMs),
		})
	}

	// 3. スコア計算: floor(通過重み / 合計重み × 100)
	score := int16(0)
	if totalWeight > 0 {
		score = int16(math.Floor(float64(passedWeight) / float64(totalWeight) * 100))
	}

	// 4. 提出をトランザクションで保存
	if _, err = u.submissionRepo.Save(ctx, domain.CreateSubmissionParams{
		UserID:        userID, // 認証必須のため常に有効な UUID
		ProblemID:     problemID,
		Code:          code,
		Language:      "python",
		OverallStatus: overall,
		Score:         score,
	}, resultParams); err != nil {
		return nil, fmt.Errorf("save submission: %w", err)
	}

	return &domain.JudgeResponse{
		Overall: overall,
		Results: judgeResults,
	}, nil
}
