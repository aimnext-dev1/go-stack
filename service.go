package main

import "fmt"

func cmdStatus(args []string) error {
	if err := checkStackExists(); err != nil { return err }
	return compose("ps", "-a")
}

var svcActionKo = map[string]string{"start": "시작", "stop": "중지", "restart": "재시작"}

func svcAction(subcmd string, args []string) error {
	if err := checkStackExists(); err != nil { return err }
	redLog("전체 " + svcActionKo[subcmd])
	return compose(append([]string{subcmd}, args...)...)
}

func cmdSvcStart(args []string) error   { return svcAction("start", args) }
func cmdSvcStop(args []string) error    { return svcAction("stop", args) }
func cmdSvcRestart(args []string) error { return svcAction("restart", args) }

func cmdLog(args []string) error {
	name := resolveOneContainer(args)
	if name == "" { return fmt.Errorf("컨테이너 이름을 지정하세요") }
	redLog("로그 조회: " + name)
	return run(cfg.cmd[0], "logs", "-f", "-n", "10000", name)
}

func cmdConnect(args []string) error {
	name := resolveOneContainer(args)
	if name == "" { return fmt.Errorf("컨테이너 이름을 지정하세요") }
	redLog("접속 중: " + name)
	return run(cfg.cmd[0], "exec", "-it", name, "sh", "-c",
		`if command -v bash >/dev/null 2>&1; then exec bash; else exec sh; fi`)
}