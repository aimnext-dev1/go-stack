package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func cmdImgBackup(args []string) error {
	if err := checkStackExists(); err != nil { return err }
	sourceOnly := len(args) > 0 && args[0] == "source"
	ts := time.Now().Format("20060102_1504")
	dir := filepath.Join(cfg.backupDir, cfg.stackName+".image."+ts)
	os.MkdirAll(dir, 0755)
	for _, cid := range containerIDs() {
		cName, _ := runOut(cfg.containerBin, "inspect", "--format", "{{.Name}}", cid)
		cName = strings.TrimPrefix(cName, "/")
		if sourceOnly {
			img, _ := runOut(cfg.containerBin, "inspect", "--format", "{{.Image}}", cid)
			run(cfg.containerBin, "save", "-o", filepath.Join(dir, cName+".image.backup.tar"), img)
		} else {
			run(cfg.containerBin, "commit", cid, cName+":backup")
			run(cfg.containerBin, "save", "-o", filepath.Join(dir, cName+".image.backup.tar"), cName+":backup")
			run(cfg.containerBin, "rmi", cName+":backup")
		}
	}
	redLog("compressing...")
	run("tar", "-czf", dir+".tar.gz", "-C", dir, ".")
	os.RemoveAll(dir)
	redLog("image backup: " + dir + ".tar.gz")
	return nil
}

func cmdImgRestore(args []string) error {
	if _, err := validateTimestamp(args[0]); err != nil { return err }
	tarFile := filepath.Join(cfg.backupDir, cfg.stackName+".image."+args[0]+".tar.gz")
	if _, err := os.Stat(tarFile); os.IsNotExist(err) { return fmt.Errorf("backup not found: %s", tarFile) }
	rd, _ := os.MkdirTemp(cfg.backupDir, "imgrest.")
	defer os.RemoveAll(rd)
	run("tar", "-xzf", tarFile, "-C", rd)
	ents, _ := os.ReadDir(rd)
	for _, e := range ents {
		if strings.HasSuffix(e.Name(), ".tar") {
			redLog("loading: " + e.Name())
			run(cfg.containerBin, "load", "-i", filepath.Join(rd, e.Name()))
		}
	}
	redLog("image restore complete! Update the image value in your compose file, then run.")
	return nil
}