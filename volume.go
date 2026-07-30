package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type volMap struct {
	Volume      string `json:"volume"`
	Destination string `json:"destination"`
}

func cmdVolPull(args []string) error {
	if err := checkStackExists(); err != nil { return err }
	vols := volumeList()
	if len(vols) == 0 { return fmt.Errorf("볼륨을 찾을 수 없습니다") }
	ctrs := containerNames()
	redLog("볼륨 가져오는 중...")
	os.RemoveAll(cfg.volumeDir)
	os.MkdirAll(cfg.volumeDir, 0755)
	m := map[string]volMap{}
	for _, vol := range vols {
		for _, ctr := range ctrs {
			_, dst := getMount(ctr, vol)
			if dst == "" { continue }
			m[ctr] = volMap{Volume: vol, Destination: dst}
			run(cfg.cmd[0], "cp", ctr+":"+dst, filepath.Join(cfg.volumeDir, vol))
		}
	}
	b, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(cfg.volumeDir, "volume-map.json"), b, 0644)
	redLog("볼륨 가져오기 완료!")
	return nil
}

func cmdVolPush(args []string) error {
	if err := checkStackExists(); err != nil { return err }
	data, err := os.ReadFile(filepath.Join(cfg.volumeDir, "volume-map.json"))
	if err != nil { return fmt.Errorf("volume-map.json을 찾을 수 없습니다. 먼저 'dtx pull'을 실행하세요.") }
	var vmap map[string]volMap
	if err := json.Unmarshal(data, &vmap); err != nil { return fmt.Errorf("volume-map.json이 손상되었습니다: %w", err) }
	redLog("볼륨 적용 중...")
	for ctr, e := range vmap {
		src := filepath.Join(cfg.volumeDir, e.Volume)
		if _, e2 := os.Stat(src); os.IsNotExist(e2) {
			fmt.Fprintf(os.Stderr, "  건너뜀 (데이터 없음): %s\n", src)
			continue
		}
		run(cfg.cmd[0], "cp", src+"/.", ctr+":"+e.Destination)
		usr, _ := runOut(cfg.cmd[0], "exec", ctr, "stat", "-c", "%U", e.Destination)
		if usr == "" { usr = "root" }
		grp, _ := runOut(cfg.cmd[0], "exec", ctr, "stat", "-c", "%G", e.Destination)
		if grp == "" { grp = "root" }
		run(cfg.cmd[0], "exec", "-u", "root", ctr, "chown", "-R", usr+":"+grp, e.Destination)
	}
	redLog("볼륨 적용 완료!")
	return nil
}

func cmdVolBackup(args []string) error {
	if err := checkStackExists(); err != nil { return err }
	noStop := len(args) > 0 && args[0] == "no-stop"
	ts := time.Now().Format("20060102_1504")
	dir := filepath.Join(cfg.backupDir, cfg.stackName+".volume."+ts)
	os.MkdirAll(dir, 0755)
	m := map[string]volMap{}
	if !noStop {
		redLog("정합성 있는 백업을 위해 스택 중지 중...")
		compose("stop")
		defer compose("start")
	}
	for _, vol := range volumeList() {
		dest := filepath.Join(dir, vol)
		os.MkdirAll(dest, 0755)
		for _, ctr := range containerNames() {
			_, dst := getMount(ctr, vol)
			if dst == "" { continue }
			m[ctr] = volMap{Volume: vol, Destination: dst}
			run(cfg.cmd[0], "cp", ctr+":"+dst, dest)
		}
	}
	b, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(dir, "volume-map.json"), b, 0644)
	redLog("압축 중...")
	run("tar", "-czf", dir+".tar.gz", "-C", dir, ".")
	os.RemoveAll(dir)
	redLog("볼륨 백업: " + dir + ".tar.gz")
	return nil
}

func cmdVolRestore(args []string) error {
	if err := checkStackExists(); err != nil { return err }
	if _, err := validateTimestamp(args[0]); err != nil { return err }
	noStop := false
	for _, a := range args[1:] {
		if a == "no-stop" { noStop = true; break }
	}
	tarFile := filepath.Join(cfg.backupDir, cfg.stackName+".volume."+args[0]+".tar.gz")
	if _, err := os.Stat(tarFile); os.IsNotExist(err) {
		return fmt.Errorf("백업을 찾을 수 없습니다: %s", tarFile)
	}
	rdir, _ := os.MkdirTemp(cfg.backupDir, "volrest.")
	defer os.RemoveAll(rdir)
	run("tar", "-xzf", tarFile, "-C", rdir)
	d, _ := os.ReadFile(filepath.Join(rdir, "volume-map.json"))
	var vmap map[string]volMap
	json.Unmarshal(d, &vmap)
	if !noStop {
		redLog("복원을 위해 스택 중지 중...")
		compose("stop")
		defer compose("start")
	}
	// chown 후보 높은 빈도
	for ctr, e := range vmap {
		src := filepath.Join(rdir, e.Volume)
		if _, e2 := os.Stat(src); os.IsNotExist(e2) { continue }
		run(cfg.cmd[0], "cp", src+"/.", ctr+":"+e.Destination)
		usr, _ := runOut(cfg.cmd[0], "exec", ctr, "stat", "-c", "%U", e.Destination)
		if usr == "" { usr = "root" }
		grp, _ := runOut(cfg.cmd[0], "exec", ctr, "stat", "-c", "%G", e.Destination)
		if grp == "" { grp = "root" }
		run(cfg.cmd[0], "exec", "-u", "root", ctr, "chown", "-R", usr+":"+grp, e.Destination)
	}
	redLog("볼륨 복원 완료!")
	return nil
}