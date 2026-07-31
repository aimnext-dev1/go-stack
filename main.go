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
			fmt.Fprintf(os.Stderr, "go-stack: %v\n", err)
			os.Exit(1)
		}
		return
	}
	loadConfig()
	name := os.Args[1]
	a, ok := cmds[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "go-stack: unknown command '%s'. Run 'go-stack help'.\n", name)
		os.Exit(1)
	}
	args := os.Args[2:]
	if len(args) < a.minArgs {
		fmt.Fprintf(os.Stderr, "go-stack %s: missing arguments\n  %s\n", name, a.usage)
		os.Exit(1)
	}
	if err := a.fn(args); err != nil {
		fmt.Fprintf(os.Stderr, "go-stack: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	hdr := map[string]string{
		"init": "Init", "stack": "Stack", "svc":"Service",
		"vol":"Volume", "img":"Image", "deploy":"Deploy/Cleanup",
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
	fmt.Printf("\ngo-stack help\n\n")
}