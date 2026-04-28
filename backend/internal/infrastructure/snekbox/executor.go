// Package snekbox は snekbox (https://github.com/python-discord/snekbox) を
// バックエンドとして使うコード実行エンジンを提供する。
//
// snekbox は nsjail でサンドボックス化された Python 実行環境であり、
// Docker コンテナ内で動く標準的なアプローチより強固なプロセス分離を提供する。
//
// セキュリティモデル:
//   - nsjail によるシステムコール制限 (seccomp-bpf)
//   - namespaces によるプロセス・ファイルシステム・ネットワーク分離
//   - メモリ・CPU・プロセス数の上限は snekbox 側の設定で制御
//
// API: POST http://<host>/eval
//
//	Request:  {"input": "<python code>"}
//	Response: {"stdout": "...", "returncode": 0, "files": []}
package snekbox

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"daileycoding/backend/internal/domain"
)

// Executor は snekbox HTTP API を呼び出してコードを実行する。
// usecase.Executor インターフェースを実装する。
type Executor struct {
	baseURL    string
	httpClient *http.Client
}

// NewExecutor は snekbox Executor を生成する。
// baseURL: snekbox サービスの URL (例: "http://snekbox:8060")
func NewExecutor(baseURL string) *Executor {
	return &Executor{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type evalRequest struct {
	Input string `json:"input"`
}

type evalResponse struct {
	Stdout     string `json:"stdout"`
	Returncode int    `json:"returncode"`
}

// Execute はコードを snekbox に送信して実行し結果を返す。
//
// stdin の扱い:
// snekbox API は stdin を直接サポートしないため、コードの先頭に
// sys.stdin を StringIO で差し替えるシムを注入する。
// stdin は base64 エンコードして埋め込むことでエスケープ問題を回避する。
func (e *Executor) Execute(ctx context.Context, req domain.ExecuteRequest) (domain.ExecuteResult, error) {
	code := req.Code
	if req.Stdin != "" {
		code = injectStdin(req.Code, req.Stdin)
	}

	body, err := json.Marshal(evalRequest{Input: code})
	if err != nil {
		return domain.ExecuteResult{}, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		e.baseURL+"/eval", bytes.NewReader(body))
	if err != nil {
		return domain.ExecuteResult{}, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := e.httpClient.Do(httpReq)
	if err != nil {
		return domain.ExecuteResult{}, fmt.Errorf("call snekbox: %w", err)
	}
	defer resp.Body.Close()
	elapsed := float64(time.Since(start).Milliseconds())

	if resp.StatusCode != http.StatusOK {
		return domain.ExecuteResult{}, fmt.Errorf("snekbox returned status %d", resp.StatusCode)
	}

	var evalResp evalResponse
	if err := json.NewDecoder(resp.Body).Decode(&evalResp); err != nil {
		return domain.ExecuteResult{}, fmt.Errorf("decode response: %w", err)
	}

	// snekbox は stdout に実行出力を返す。
	// returncode != 0 の場合は実行エラー扱いとし Stderr に出力を移す。
	result := domain.ExecuteResult{TimeMs: elapsed}
	if evalResp.Returncode == 0 {
		result.Stdout = evalResp.Stdout
	} else {
		result.Stderr = evalResp.Stdout
	}
	return result, nil
}

// injectStdin はユーザーコードの先頭に stdin モックシムを注入する。
// stdin の内容は base64 エンコードして埋め込むためエスケープが不要。
func injectStdin(code, stdin string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(stdin))
	shim := fmt.Sprintf(
		"import sys as __sys__,io as __io__,base64 as __b64__\n"+
			"__sys__.stdin=__io__.StringIO(__b64__.b64decode('%s').decode())\n"+
			"del __sys__,__io__,__b64__\n",
		encoded,
	)
	return shim + code
}
