package main

import "fmt"

func cmdStatus(args []string) error {
	if err := checkStackExists(); err != nil { return err }
	return compose("ps", "-a")
}

var svcActionLabel = map[string]string{"start": "starting", "stop": "stopping", "restart": "restarting"}

func svcAction(subcmd string, args []string) error {
	if err := checkStackExists(); err != nil { return err }
	redLog(svcActionLabel[subcmd] + " all services")
	return compose(append([]string{subcmd}, args...)...)
}

func cmdSvcStart(args []string) error   { return svcAction("start", args) }
func cmdSvcStop(args []string) error    { return svcAction("stop", args) }
func cmdSvcRestart(args []string) error { return svcAction("restart", args) }

func cmdLog(args []string) error {
	name := resolveOneContainer(args)
	if name == "" { return fmt.Errorf("specify a container name") }
	redLog("tailing logs: " + name)
	return run(cfg.cmd[0], "logs", "-f", "-n", "10000", name)
}

func cmdConnect(args []string) error {
	name := resolveOneContainer(args)
	if name == "" { return fmt.Errorf("specify a container name") }
	redLog("connecting: " + name)
	return run(cfg.cmd[0], "exec", "-it", name, "sh", "-c",
		`if command -v bash >/dev/null 2>&1; then exec bash; else exec sh; fi`)
}