package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

var cfg struct {
    cmd   []string
    podman bool
    stackName   string
    root        string
    backupDir   string
    volumeDir   string
}

func loadConfig() {
    cfg.root = findRoot()

    parseEnvFile(filepath.Join(cfg.root, "stack.env"))

    cfg.stackName = os.Getenv("STACK_NAME")
    if cfg.stackName == "" { fatal("STACK_NAME이 stack.env에 설정되지 않았습니다") }

    cfg.backupDir  = filepath.Join(cfg.root, "_backup")
    cfg.volumeDir  = filepath.Join(cfg.root, "_volume")
    os.MkdirAll(cfg.backupDir, 0755)

    cfg.cmd = detectContainer()
}

func findRoot() string {
    dir, _ := os.Getwd()
    for {
        if _, err := os.Stat(filepath.Join(dir, "stack.env")); err == nil {
            return dir
        }
        parent := filepath.Dir(dir)
        if parent == dir { fatal("stack.env를 찾을 수 없습니다 (현재 폴더부터 상위로 탐색함)") }
        dir = parent
    }
}

func parseEnvFile(path string) {
    data, err := os.ReadFile(path)
    if err != nil { fmt.Fprintf(os.Stderr,"stack.env 읽기 실패: %v\n",err); os.Exit(1) }
    for _, line := range strings.Split(strings.ReplaceAll(string(data),"\r\n","\n"),"\n") {
        line = strings.TrimSpace(line)
        if line == "" || line[0] == '#' { continue }
        eq := strings.IndexByte(line, '=')
        if eq < 0 { continue }
        k := strings.TrimSpace(line[:eq])
        v := strings.TrimSpace(line[eq+1:])
        if len(v) >= 2 && (v[0]=='"'&&v[len(v)-1]=='"' || v[0]=='\''&&v[len(v)-1]=='\'') { v = v[1:len(v)-1] }
        os.Setenv(k, v)
    }
}

func detectContainer() []string {
    if v := os.Getenv("DTX_CONTAINER"); v != "" {
        switch strings.ToLower(v) {
        case "podman": cfg.podman = true; return []string{"podman","compose"}
        case "docker": return []string{"docker","compose"}
        }
    }
    if _, e := exec.LookPath("docker"); e == nil {
        if exec.Command("docker","info").Run() == nil { return []string{"docker","compose"} }
    }
    if _, e := exec.LookPath("podman"); e == nil { cfg.podman = true; return []string{"podman","compose"} }
    if _, e := exec.LookPath("docker-compose"); e == nil { return []string{"docker-compose"} }
    if _, e := exec.LookPath("podman-compose"); e == nil { cfg.podman = true; return []string{"podman-compose"} }
    fmt.Fprintln(os.Stderr,"컨테이너 런타임을 찾을 수 없습니다.")
    os.Exit(1)
    return nil
}