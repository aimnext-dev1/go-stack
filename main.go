package main

import (
	"fmt"
	"os"
	"sort"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "help" {
		printHelp()
		return
	}
	if os.Args[1] == "init" {
		if err := cmdInit(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "dtx: %v\n", err)
			os.Exit(1)
		}
		return
	}
	loadConfig()
	name := os.Args[1]
	a, ok := cmds[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "dtx: 알 수 없는 명령어 '%s'. 'dtx help'를 실행하세요.\n", name)
		os.Exit(1)
	}
	args := os.Args[2:]
	if len(args) < a.minArgs {
		fmt.Fprintf(os.Stderr, "dtx %s: 인자가 부족합니다\n  %s\n", name, a.usage)
		os.Exit(1)
	}
	if err := a.fn(args); err != nil {
		fmt.Fprintf(os.Stderr, "dtx: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	hdr := map[string]string{
		"init": "초기화 명령어", "stack": "스택 명령어", "svc":"서비스 명령어",
		"vol":"볼륨 명령어", "img":"이미지 명령어", "deploy":"배포/정리 명령어",
	}
	names := make([]string, 0, len(cmds))
	for k := range cmds { names = append(names, k) }
	sort.Strings(names)
	for _, g := range groupOrder {
		fmt.Printf("\n%s:\n", hdr[g])
		for _, k := range names {
			if cmds[k].group == g { fmt.Printf("  %s\n", cmds[k].usage) }
		}
	}
	fmt.Printf("\ndtx help\n\n")
}