package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func cmdDeploy(args []string) error {
	env := "dev"
	if len(args) > 0 { env = args[0] }
	if env != "dev" && env != "prod" { return fmt.Errorf("지원하지 않는 환경입니다: %s (dev, prod)", env) }
	s3 := os.Getenv("DEPLOY_S3_BUCKET_" + strings.ToUpper(env))
	if s3 == "" { return fmt.Errorf("DEPLOY_S3_BUCKET_%s가 stack.env에 설정되지 않았습니다", strings.ToUpper(env)) }
	tmp, _ := os.MkdirTemp(".", ".deploy.")
	defer os.RemoveAll(tmp)
	redLog("S3에서 다운로드 중...")
	run("aws", "s3", "cp", s3+"/Makefile", filepath.Join(tmp, "Makefile"))
	for _, f := range composeFileNames() {
		run("aws", "s3", "cp", s3+"/"+f, filepath.Join(tmp, f))
	}
	os.MkdirAll(filepath.Join(tmp, "_script"), 0755)
	run("aws", "s3", "cp", "--recursive", s3+"/_script", filepath.Join(tmp, "_script"))
	redLog("교체 중...")
	os.Rename(filepath.Join(tmp, "Makefile"), filepath.Join(cfg.root, "Makefile"))
	for _, f := range composeFileNames() {
		os.Rename(filepath.Join(tmp, f), filepath.Join(cfg.root, f))
	}
	os.RemoveAll(filepath.Join(cfg.root, "_script"))
	os.Rename(filepath.Join(tmp, "_script"), filepath.Join(cfg.root, "_script"))
	redLog("배포 완료!")
	return nil
}

func composeFileNames() []string {
	seen := map[string]bool{}
	var names []string
	for _, key := range []string{"COMPOSE_BASE_FILE", "COMPOSE_FILE_LOCAL", "COMPOSE_FILE_DEV", "COMPOSE_FILE_PROD"} {
		f := os.Getenv(key)
		if f == "" || seen[f] { continue }
		seen[f] = true
		names = append(names, f)
	}
	return names
}

func cmdClear(args []string) error {
	redLog("미사용 이미지 정리 중 (-af)...")
	return run(cfg.cmd[0], "image", "prune", "-af")
}