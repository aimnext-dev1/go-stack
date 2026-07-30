package main

import "fmt"

func cmdUp(args []string) error {
	if err := checkStackNotExist(); err != nil { return err }
	env := "local"
	if len(args) > 0 { env = args[0] }
	files, envFile := resolveComposeFiles(env)
	out := make([]string, 0, len(files)+8)
	out = append(out, files...)
	if envFile != "" { out = append(out, "--env-file", envFile) }
	out = append(out, "up", "-d")
	redLog("스택 생성 중...")
	return compose(out...)
}

func cmdDown(args []string) error {
	if err := checkStackExists(); err != nil { return err }
	if !confirm("스택을 삭제하시겠습니까? 모든 컨테이너가 제거됩니다.") {
		fmt.Println("취소됨.")
		return nil
	}
	redLog("스택 삭제 중...")
	return compose("down")
}

func cmdUpdate(args []string) error {
	if err := checkStackExists(); err != nil { return err }
	env := "local"
	if len(args) > 0 { env = args[0] }
	files, envFile := resolveComposeFiles(env)
	out := make([]string, 0, len(files)+8)
	out = append(out, files...)
	if envFile != "" { out = append(out, "--env-file", envFile) }
	out = append(out, "up", "-d", "--build")
	redLog("스택 갱신 중 (--build)...")
	return compose(out...)
}