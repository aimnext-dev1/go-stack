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
	if env != "dev" && env != "prod" { return fmt.Errorf("unsupported environment: %s (dev, prod)", env) }
	s3 := os.Getenv("DEPLOY_S3_BUCKET_" + strings.ToUpper(env))
	if s3 == "" { return fmt.Errorf("DEPLOY_S3_BUCKET_%s is not set in stack.env", strings.ToUpper(env)) }
	tmp, _ := os.MkdirTemp(".", ".deploy.")
	defer os.RemoveAll(tmp)
	redLog("downloading from S3...")
	run("aws", "s3", "cp", s3+"/Makefile", filepath.Join(tmp, "Makefile"))
	for _, f := range composeFileNames() {
		run("aws", "s3", "cp", s3+"/"+f, filepath.Join(tmp, f))
	}
	os.MkdirAll(filepath.Join(tmp, "_script"), 0755)
	run("aws", "s3", "cp", "--recursive", s3+"/_script", filepath.Join(tmp, "_script"))
	redLog("replacing...")
	os.Rename(filepath.Join(tmp, "Makefile"), filepath.Join(cfg.root, "Makefile"))
	for _, f := range composeFileNames() {
		os.Rename(filepath.Join(tmp, f), filepath.Join(cfg.root, f))
	}
	os.RemoveAll(filepath.Join(cfg.root, "_script"))
	os.Rename(filepath.Join(tmp, "_script"), filepath.Join(cfg.root, "_script"))
	redLog("deploy complete!")
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
	redLog("pruning unused images (-af)...")
	return run(cfg.cmd[0], "image", "prune", "-af")
}