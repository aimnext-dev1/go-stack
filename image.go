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
		cName, _ := runOut(cfg.cmd[0], "inspect", "--format", "{{.Name}}", cid)
		cName = strings.TrimPrefix(cName, "/")
		if sourceOnly {
			img, _ := runOut(cfg.cmd[0], "inspect", "--format", "{{.Image}}", cid)
			run(cfg.cmd[0], "save", "-o", filepath.Join(dir, cName+".image.backup.tar"), img)
		} else {
			run(cfg.cmd[0], "commit", cid, cName+":backup")
			run(cfg.cmd[0], "save", "-o", filepath.Join(dir, cName+".image.backup.tar"), cName+":backup")
			run(cfg.cmd[0], "rmi", cName+":backup")
		}
	}
	redLog("압축 중...")
	run("tar", "-czf", dir+".tar.gz", "-C", dir, ".")
	os.RemoveAll(dir)
	redLog("이미지 백업: " + dir + ".tar.gz")
	return nil
}

func cmdImgRestore(args []string) error {
	if _, err := validateTimestamp(args[0]); err != nil { return err }
	tarFile := filepath.Join(cfg.backupDir, cfg.stackName+".image."+args[0]+".tar.gz")
	if _, err := os.Stat(tarFile); os.IsNotExist(err) { return fmt.Errorf("백업을 찾을 수 없습니다: %s", tarFile) }
	rd, _ := os.MkdirTemp(cfg.backupDir, "imgrest.")
	defer os.RemoveAll(rd)
	run("tar", "-xzf", tarFile, "-C", rd)
	ents, _ := os.ReadDir(rd)
	for _, e := range ents {
		if strings.HasSuffix(e.Name(), ".tar") {
			redLog("불러오는 중: " + e.Name())
			run(cfg.cmd[0], "load", "-i", filepath.Join(rd, e.Name()))
		}
	}
	redLog("이미지 복원 완료! compose 파일의 image 값을 수정한 뒤 실행하세요.")
	return nil
}