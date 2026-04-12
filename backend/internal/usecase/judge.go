package usecase

import (
	"context"
	"strings"

	"daileycoding/backend/internal/domain"
)


// Executor はコード実行エンジンの抽象インターフェース
// Piston以外のエンジン（Judge0等）に差し替え可能にする
type Executor interface {
	Execute(ctx context.Context, req domain.ExecuteRequest) (domain.ExecuteResult, error)
}

// JudgeUsecase はコード採点のビジネスロジックを担う
type JudgeUsecase struct {
	executor  Executor
	testCases []domain.TestCase
}

func NewJudgeUsecase(executor Executor) *JudgeUsecase {
	return &JudgeUsecase{
		executor: executor,
		// テストケースはここにハードコード
		// 将来はDBやファイルから読み込む想定
		testCases: []domain.TestCase{
			{Input: "2 3", Expected: "5"},
			{Input: "10 20", Expected: "30"},
		},
	}
}

// Judge は与えられたコードを全テストケースで実行し採点結果を返す
func (u *JudgeUsecase) Judge(ctx context.Context, code string) (*domain.JudgeResponse, error) {
	results := make([]domain.TestResult, 0, len(u.testCases))
	overall := "Pass"

	for i, tc := range u.testCases {
		execResult, err := u.executor.Execute(ctx, domain.ExecuteRequest{
			Code:     code,
			Language: "python",
			Stdin:    tc.Input,
		})
		if err != nil {
			return nil, err
		}

		actual := strings.TrimSpace(execResult.Stdout)
		expected := strings.TrimSpace(tc.Expected)

		status := "Pass"
		if actual != expected || execResult.Stderr != "" {
			status = "Fail"
			overall = "Fail"
		}

		results = append(results, domain.TestResult{
			CaseIndex: i + 1,
			Status:    status,
			Stdout:    execResult.Stdout,
			Expected:  tc.Expected,
			Stderr:    execResult.Stderr,
			TimeMs:    execResult.TimeMs,
		})
	}

	return &domain.JudgeResponse{
		Overall: overall,
		Results: results,
	}, nil
}
