package main

import (
    "bufio"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"
)

func fatal(format string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, format+"\n", args...)
    os.Exit(1)
}

func redLog(msg string) { fmt.Printf("\n\033[1;31m*** %s ***\033[0m\n", msg) }

func confirm(msg string) bool {
    fmt.Printf("%s (y/n): ", msg)
    rd := bufio.NewReader(os.Stdin)
    line, _ := rd.ReadString('\n')
    return strings.TrimSpace(strings.ToLower(line)) == "y"
}

func run(args ...string) error {
    c := exec.Command(args[0], args[1:]...)
    c.Stdin  = os.Stdin
    c.Stdout = os.Stdout
    c.Stderr = os.Stderr
    return c.Run()
}

func runOut(args ...string) (string, error) {
    c := exec.Command(args[0], args[1:]...)
    c.Stderr = os.Stderr
    b, err := c.Output()
    return strings.TrimSpace(string(b)), err
}

func runLines(args ...string) []string {
    out, _ := runOut(args...)
    if out == "" { return nil }
    var rv []string
    for _, l := range strings.Split(out, "\n") {
        if l = strings.TrimSpace(l); l != "" { rv = append(rv, l) }
    }
    return rv
}

func compose(args ...string) error {
    a := make([]string, 0, len(args)+2+len(cfg.cmd))
    a = append(a, cfg.cmd...)
    a = append(a, "-p", cfg.stackName)
    a = append(a, args...)
    return run(a...)
}

func composeOut(args ...string) (string, error) {
    a := make([]string, 0, len(args)+2+len(cfg.cmd))
    a = append(a, cfg.cmd...)
    a = append(a, "-p", cfg.stackName)
    a = append(a, args...)
    return runOut(a...)
}

func composeLines(args ...string) []string {
    a := make([]string, 0, len(args)+2+len(cfg.cmd))
    a = append(a, cfg.cmd...)
    a = append(a, "-p", cfg.stackName)
    a = append(a, args...)
    return runLines(a...)
}

func checkStackExists() error {
	out, _ := composeOut("ps", "-aq")
	if out == "" {
		return fmt.Errorf("'%s' 스택을 찾을 수 없습니다. 먼저 'go-stack up'을 실행하세요.", cfg.stackName)
	}
	return nil
}

func checkStackNotExist() error {
    err := checkStackExists()
    if err == nil {
        return fmt.Errorf("'%s' 스택이 이미 존재합니다. 먼저 'go-stack down'을 실행하세요.", cfg.stackName)
    }
    return nil
}

func containerNames() []string {
    return composeLines("ps", "-a", "--format", "{{.Names}}")
}

func containerIDs() []string {
    return composeLines("ps", "-q")
}

func volumeList() []string {
    return runLines(cfg.cmd[0], "volume", "ls",
        "--filter", "label=com.docker.compose.project="+cfg.stackName,
        "--format", "{{.Name}}")
}

func getMount(container, volume string) (src, dst string) {
    const tmpl = `{{range .Mounts}}{{if eq .Name "%s"}}{{.Source}}{{"\n"}}{{.Destination}}{{end}}{{end}}`
    s, _ := runOut("docker", "inspect", container, "--format", fmt.Sprintf(tmpl, volume))
    lines := strings.Split(strings.TrimSpace(s), "\n")
    if len(lines) >= 2 { src, dst = lines[0], lines[1] }
    return
}

func resolveComposeFiles(env string) ([]string, string) {
    var cfgFile string
    switch env {
    case "local","": cfgFile = os.Getenv("COMPOSE_FILE_LOCAL")
    case "dev":     cfgFile = os.Getenv("COMPOSE_FILE_DEV")
    case "prod":    cfgFile = os.Getenv("COMPOSE_FILE_PROD")
    default:        fatal("지원하지 않는 환경입니다: %s (local, dev, prod)", env)
    }
    envUpper := strings.ToUpper(env)
    if cfgFile == "" { fatal("COMPOSE_FILE_%s가 stack.env에 설정되지 않았습니다", envUpper) }
    var files []string
    if base := os.Getenv("COMPOSE_BASE_FILE"); base != "" {
        bf := filepath.Join(cfg.root, base)
        if _, e := os.Stat(bf); os.IsNotExist(e) { fatal("베이스 compose 파일을 찾을 수 없습니다: %s", bf) }
        files = append(files, "-f", bf)
    }
    cf := filepath.Join(cfg.root, cfgFile)
    if _, e := os.Stat(cf); os.IsNotExist(e) { fatal("compose 파일을 찾을 수 없습니다: %s", cf) }
    files = append(files, "-f", cf)
    var envFile string
    switch env {
    case "local","": envFile = os.Getenv("ENV_FILE_LOCAL")
    case "dev":      envFile = os.Getenv("ENV_FILE_DEV")
    case "prod":     envFile = os.Getenv("ENV_FILE_PROD")
    }
    return files, envFile
}

func resolveOneContainer(args []string) string {
    if len(args) > 0 { return args[0] }
    list := containerNames()
    if len(list) == 1 {
        redLog("자동 선택: " + list[0])
        return list[0]
    }
    fmt.Println("컨테이너 목록:")
    for _, n := range list { fmt.Printf("  %s\n", n) }
    return ""
}

func validateTimestamp(ts string) (string, error) {
    if len(ts) != 13 || ts[8] != '_' { return "", fmt.Errorf("형식: YYYYMMDD_HHMM 예) 20240131_1341") }
    t, err := time.Parse("20060102", ts[:8])
    if err != nil { return "", fmt.Errorf("잘못된 날짜입니다: %s", ts[:8]) }
    if _, err := time.Parse("1504", ts[9:]); err != nil { return "", fmt.Errorf("잘못된 시간입니다: %s", ts[9:]) }
    return t.Format("20060102") + "_" + ts[9:], nil
}