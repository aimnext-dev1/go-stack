package main

type action struct {
	fn      func([]string) error
	usage   string
	group   string
	minArgs int
}

var cmds = map[string]action{
    "init":    {cmdInit, "go-stack init                            현재 폴더에 stack.env 생성","init", 0},
    "up":      {cmdUp, "go-stack up [local|dev|prod]                docker compose 스택 생성","stack", 0},
    "down":    {cmdDown, "go-stack down                              docker compose 스택 제거","stack", 0},
    "update":  {cmdUpdate, "go-stack update [local|dev|prod]           변경분 빌드 후 스택 재생성","stack", 0},
    "status":  {cmdStatus, "go-stack status                          서비스 상태 조회","svc", 0},
    "start":   {cmdSvcStart, "go-stack start [svc...]                  서비스 시작","svc", 0},
    "stop":    {cmdSvcStop, "go-stack stop [svc...]                   서비스 중지","svc", 0},
    "restart": {cmdSvcRestart, "go-stack restart [svc...]               서비스 재시작","svc", 0},
    "log":     {cmdLog, "go-stack log [name]                        컨테이너 로그 조회","svc", 0},
    "connect": {cmdConnect, "go-stack connect [name]                  컨테이너 접속","svc", 0},
    "pull":    {cmdVolPull, "go-stack pull                             볼륨을 _volume/으로 가져오기","vol", 0},
    "push":    {cmdVolPush, "go-stack push                            볼륨을 컨테이너로 적용","vol", 0},
    "backup":  {cmdVolBackup, "go-stack backup [no-stop]                볼륨을 tar.gz로 백업","vol", 0},
    "restore": {cmdVolRestore, "go-stack restore <YYYYMMDD_HHMM> [no-stop] 볼륨 복원","vol", 1},
    "isave":   {cmdImgBackup, "go-stack isave [source]                   이미지를 tar.gz로 백업","img", 0},
    "iload":   {cmdImgRestore, "go-stack iload <YYYYMMDD_HHMM>            이미지 복원","img", 1},
    "deploy":  {cmdDeploy, "go-stack deploy [dev|prod]             S3에서 배포","deploy", 0},
    "clear":   {cmdClear, "go-stack clear                         미사용 이미지 정리","deploy", 0},
}

var groupOrder = []string{"init","stack","svc","vol","img","deploy"}
