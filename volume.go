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
	if len(vols) == 0 { return fmt.Errorf("no volumes found") }
	ctrs := containerNames()
	redLog("pulling volumes...")
	os.RemoveAll(cfg.volumeDir)
	os.MkdirAll(cfg.volumeDir, 0755)
	m := map[string]volMap{}
	for _, vol := range vols {
		for _, ctr := range ctrs {
			_, dst := getMount(ctr, vol)
			if dst == "" { continue }
			m[ctr] = volMap{Volume: vol, Destination: dst}
			run(cfg.containerBin, "cp", ctr+":"+dst, filepath.Join(cfg.volumeDir, vol))
		}
	}
	b, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(cfg.volumeDir, "volume-map.json"), b, 0644)
	redLog("volume pull complete!")
	return nil
}

func cmdVolPush(args []string) error {
	if err := checkStackExists(); err != nil { return err }
	data, err := os.ReadFile(filepath.Join(cfg.volumeDir, "volume-map.json"))
	if err != nil { return fmt.Errorf("volume-map.json not found. Run 'go-stack pull' first.") }
	var vmap map[string]volMap
	if err := json.Unmarshal(data, &vmap); err != nil { return fmt.Errorf("volume-map.json is corrupted: %w", err) }
	redLog("pushing volumes...")
	for ctr, e := range vmap {
		src := filepath.Join(cfg.volumeDir, e.Volume)
		if _, e2 := os.Stat(src); os.IsNotExist(e2) {
			fmt.Fprintf(os.Stderr, "  skipped (no data): %s\n", src)
			continue
		}
		run(cfg.containerBin, "cp", src+"/.", ctr+":"+e.Destination)
		usr, _ := runOut(cfg.containerBin, "exec", ctr, "stat", "-c", "%U", e.Destination)
		if usr == "" { usr = "root" }
		grp, _ := runOut(cfg.containerBin, "exec", ctr, "stat", "-c", "%G", e.Destination)
		if grp == "" { grp = "root" }
		run(cfg.containerBin, "exec", "-u", "root", ctr, "chown", "-R", usr+":"+grp, e.Destination)
	}
	redLog("volume push complete!")
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
		redLog("stopping stack for a consistent backup...")
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
			run(cfg.containerBin, "cp", ctr+":"+dst, dest)
		}
	}
	b, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(dir, "volume-map.json"), b, 0644)
	redLog("compressing...")
	run("tar", "-czf", dir+".tar.gz", "-C", dir, ".")
	os.RemoveAll(dir)
	redLog("volume backup: " + dir + ".tar.gz")
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
		return fmt.Errorf("backup not found: %s", tarFile)
	}
	rdir, _ := os.MkdirTemp(cfg.backupDir, "volrest.")
	defer os.RemoveAll(rdir)
	run("tar", "-xzf", tarFile, "-C", rdir)
	d, _ := os.ReadFile(filepath.Join(rdir, "volume-map.json"))
	var vmap map[string]volMap
	json.Unmarshal(d, &vmap)
	if !noStop {
		redLog("stopping stack for restore...")
		compose("stop")
		defer compose("start")
	}
	for ctr, e := range vmap {
		src := filepath.Join(rdir, e.Volume)
		if _, e2 := os.Stat(src); os.IsNotExist(e2) { continue }
		run(cfg.containerBin, "cp", src+"/.", ctr+":"+e.Destination)
		usr, _ := runOut(cfg.containerBin, "exec", ctr, "stat", "-c", "%U", e.Destination)
		if usr == "" { usr = "root" }
		grp, _ := runOut(cfg.containerBin, "exec", ctr, "stat", "-c", "%G", e.Destination)
		if grp == "" { grp = "root" }
		run(cfg.containerBin, "exec", "-u", "root", ctr, "chown", "-R", usr+":"+grp, e.Destination)
	}
	redLog("volume restore complete!")
	return nil
}