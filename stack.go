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
	redLog("creating stack...")
	return compose(out...)
}

func cmdDown(args []string) error {
	if err := checkStackExists(); err != nil { return err }
	if !confirm("Delete the stack? All containers will be removed.") {
		fmt.Println("Cancelled.")
		return nil
	}
	redLog("deleting stack...")
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
	redLog("updating stack (--build)...")
	return compose(out...)
}