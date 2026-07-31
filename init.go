package main

import (
	"fmt"
	"os"
	"strings"
)

const stackEnvTemplate = `# docker compose 프로젝트명(-p) / 백업 파일명 접두사
STACK_NAME=

# 환경별 compose 명세파일명 (stack.env와 같은 폴더 기준)
COMPOSE_FILE_LOCAL=docker-compose.yml
COMPOSE_FILE_DEV=docker-compose.yml
COMPOSE_FILE_PROD=docker-compose.yml

# (선택) 공통 베이스 compose 파일명. 지정 시 -f <base> -f <env파일> 로 조합 실행
# COMPOSE_BASE_FILE=docker-compose.yml

# (선택) 환경별 환경변수 파일 경로
# ENV_FILE_LOCAL=
# ENV_FILE_DEV=
# ENV_FILE_PROD=

# (선택) 컨테이너 런타임 강제 지정: docker | podman (미지정시 자동 감지)
# GO_STACK_CONTAINER=docker
`

func cmdInit(args []string) error {
	if _, err := os.Stat("stack.env"); err == nil {
		return fmt.Errorf("이 폴더에 이미 stack.env가 있습니다. 다시 만들려면 먼저 삭제하세요.")
	}
	if err := os.WriteFile("stack.env", []byte(stackEnvTemplate), 0644); err != nil {
		return fmt.Errorf("stack.env 생성 실패: %w", err)
	}
	if err := ensureGitignored("stack.env"); err != nil {
		return err
	}
	redLog("stack.env 생성 완료. 값을 채운 뒤 'go-stack up'을 실행하세요.")
	return nil
}

func ensureGitignored(entry string) error {
	data, err := os.ReadFile(".gitignore")
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf(".gitignore 읽기 실패: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == entry {
			return nil
		}
	}
	content := string(data)
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += entry + "\n"
	return os.WriteFile(".gitignore", []byte(content), 0644)
}
