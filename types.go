package main

type action struct {
	fn      func([]string) error
	usage   string
	group   string
	minArgs int
}

var cmds = map[string]action{
    "init":    {cmdInit, "go-stack init                            create stack.env in the current folder","init", 0},
    "up":      {cmdUp, "go-stack up [local|dev|prod]                create the docker compose stack","stack", 0},
    "down":    {cmdDown, "go-stack down                              remove the docker compose stack","stack", 0},
    "update":  {cmdUpdate, "go-stack update [local|dev|prod]           rebuild changes and recreate the stack","stack", 0},
    "status":  {cmdStatus, "go-stack status                          show service status","svc", 0},
    "start":   {cmdSvcStart, "go-stack start [svc...]                  start services","svc", 0},
    "stop":    {cmdSvcStop, "go-stack stop [svc...]                   stop services","svc", 0},
    "restart": {cmdSvcRestart, "go-stack restart [svc...]               restart services","svc", 0},
    "log":     {cmdLog, "go-stack log [name]                        tail container logs","svc", 0},
    "connect": {cmdConnect, "go-stack connect [name]                  exec into a container","svc", 0},
    "pull":    {cmdVolPull, "go-stack pull                             pull volumes into _volume/","vol", 0},
    "push":    {cmdVolPush, "go-stack push                            push volumes into containers","vol", 0},
    "backup":  {cmdVolBackup, "go-stack backup [no-stop]                back up volumes to tar.gz","vol", 0},
    "restore": {cmdVolRestore, "go-stack restore <YYYYMMDD_HHMM> [no-stop] restore volumes","vol", 1},
    "isave":   {cmdImgBackup, "go-stack isave [source]                   back up images to tar.gz","img", 0},
    "iload":   {cmdImgRestore, "go-stack iload <YYYYMMDD_HHMM>            restore images","img", 1},
    "deploy":  {cmdDeploy, "go-stack deploy [dev|prod]             deploy from S3","deploy", 0},
    "clear":   {cmdClear, "go-stack clear                         prune unused images","deploy", 0},
}

var groupOrder = []string{"init","stack","svc","vol","img","deploy"}
