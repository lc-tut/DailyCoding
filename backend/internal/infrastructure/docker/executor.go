package docker

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	moby "github.com/moby/moby/client"

	"daileycoding/backend/internal/domain"
)

const (
	defaultMemoryBytes = 64 * 1024 * 1024 // 64MB
	defaultCPUQuota    = 50_000            // 50% of one CPU core
	defaultCPUPeriod   = 100_000
	defaultPidsLimit   = int64(50)
	execTimeout        = 10 * time.Second
)

// Executor は Docker コンテナを使ってコードを実行する
// domain.Executor インターフェースを実装する
type Executor struct {
	cli   *moby.Client
	image string
}

// NewExecutor は Docker クライアントを初期化して Executor を返す
// image: 使用する Docker イメージ名 (例: "python:3.14-slim")
func NewExecutor(image string) (*Executor, error) {
	cli, err := moby.New(moby.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}
	return &Executor{cli: cli, image: image}, nil
}

// EnsureImage はイメージがローカルになければ pull する
// サーバー起動時に一度だけ呼び出す想定
func (e *Executor) EnsureImage(ctx context.Context) error {
	log.Printf("pulling image %s ...", e.image)
	resp, err := e.cli.ImagePull(ctx, e.image, moby.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("pull image %s: %w", e.image, err)
	}
	defer resp.Close()
	// Wait で pull 完了を待つ（進捗は破棄）
	if err := resp.Wait(ctx); err != nil {
		return fmt.Errorf("wait image pull: %w", err)
	}
	log.Printf("image %s ready", e.image)
	return nil
}

// Execute はコンテナを起動してコードを実行し、結果を返す
func (e *Executor) Execute(ctx context.Context, req domain.ExecuteRequest) (domain.ExecuteResult, error) {
	// Base64エンコードでシェルエスケープを回避する
	encodedCode := base64.StdEncoding.EncodeToString([]byte(req.Code))
	stdinWithNewline := req.Stdin
	if len(stdinWithNewline) > 0 && stdinWithNewline[len(stdinWithNewline)-1] != '\n' {
		stdinWithNewline += "\n"
	}
	encodedStdin := base64.StdEncoding.EncodeToString([]byte(stdinWithNewline))

	// コードをファイルに書き出し、stdinをパイプして実行する
	cmd := []string{
		"sh", "-c",
		fmt.Sprintf(
			"echo %s | base64 -d > /tmp/s.py && echo %s | base64 -d | python3 /tmp/s.py",
			encodedCode, encodedStdin,
		),
	}

	pidsLimit := defaultPidsLimit
	createResult, err := e.cli.ContainerCreate(ctx, moby.ContainerCreateOptions{
		Image: e.image,
		Config: &container.Config{
			Cmd:          cmd,
			AttachStdout: true,
			AttachStderr: true,
		},
		HostConfig: &container.HostConfig{
			NetworkMode: "none", // ネットワーク無効化（セキュリティ）
			Resources: container.Resources{
				Memory:     defaultMemoryBytes,
				MemorySwap: defaultMemoryBytes, // スワップなし
				CPUPeriod:  defaultCPUPeriod,
				CPUQuota:   defaultCPUQuota,
				PidsLimit:  &pidsLimit, // fork bomb 対策
			},
		},
	})
	if err != nil {
		return domain.ExecuteResult{}, fmt.Errorf("create container: %w", err)
	}

	containerID := createResult.ID
	// 関数終了時に必ずコンテナを削除する
	defer func() {
		rmCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = e.cli.ContainerRemove(rmCtx, containerID, moby.ContainerRemoveOptions{Force: true})
	}()

	if _, err := e.cli.ContainerStart(ctx, containerID, moby.ContainerStartOptions{}); err != nil {
		return domain.ExecuteResult{}, fmt.Errorf("start container: %w", err)
	}

	// 実行時間の計測開始
	startTime := time.Now()

	// タイムアウト付きで待機
	execCtx, cancel := context.WithTimeout(ctx, execTimeout)
	defer cancel()

	waitResult := e.cli.ContainerWait(execCtx, containerID, moby.ContainerWaitOptions{
		Condition: container.WaitConditionNotRunning,
	})

	select {
	case err := <-waitResult.Error:
		return domain.ExecuteResult{}, fmt.Errorf("wait container: %w", err)
	case <-waitResult.Result:
		// コンテナ終了
	}

	elapsed := time.Since(startTime).Seconds() * 1000

	// ログ（stdout/stderr）を取得する
	logStream, err := e.cli.ContainerLogs(ctx, containerID, moby.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return domain.ExecuteResult{}, fmt.Errorf("get logs: %w", err)
	}
	defer logStream.Close()

	// Docker のマルチプレクスされたログストリームを分離する
	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, &stderr, logStream); err != nil && err != io.EOF {
		return domain.ExecuteResult{}, fmt.Errorf("read logs: %w", err)
	}

	return domain.ExecuteResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		TimeMs: elapsed,
	}, nil
}
