# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# dtx-docker-manager

`docker compose` 스택을 배포/관리/백업/복원하는 단일 Go 바이너리 CLI `go-stack`.
과거 bash 스크립트(`_script/*.sh` + `Makefile`) 모음을 Go로 재작성 중. 프로젝트 규모: ~9개 .go 파일, 700줄 안팎.

## 빌드/실행

```bash
go build -o go-stack .
go vet ./...
```

자동화된 테스트(`*_test.go`) 없음 — 아래 "검증 방식" 참고, 실제 docker로 수동 검증.

## 구조

각 파일이 명령어 그룹 하나씩 담당 (평면 구조, 패키지 분리 없음, 전부 `package main`):

- `main.go` — 엔트리포인트, 명령 디스패치, `help` 출력
- `types.go` — `cmds` map (명령어 이름 → 함수/usage/그룹/minArgs), `groupOrder`
- `config.go` — `stack.env` 로딩(`findRoot()`가 cwd에서 상위로 탐색, git처럼), 컨테이너 런타임 감지(docker/podman/docker-compose)
- `helpers.go` — 공용 유틸: `run`/`runOut`/`compose`/`composeOut`, `checkStackExists`, `resolveComposeFiles`, `validateTimestamp` 등
- `init.go` — `go-stack init` (stack.env 템플릿 생성 + .gitignore 등록)
- `stack.go` — up/down/update
- `service.go` — status/start/stop/restart/log/connect
- `volume.go` — pull/push/backup/restore (볼륨)
- `image.go` — isave/iload (이미지 백업/복원)
- `deploy.go` — deploy(S3에서 compose 파일 배포, 레거시 잔재 있음)/clear

## 프로젝트 구조 (사용자 stack.env 폴더)

`_project/` 서브폴더 **폐지함** — compose 파일들은 `stack.env`와 **같은 폴더에 직접** 위치:
```
my-service/
├── stack.env
├── docker-compose.yml
├── docker-compose.local.yml
├── docker-compose.dev.yml
└── docker-compose.prod.yml
```
`cfg.projectDir` 필드 없음, `resolveComposeFiles()`는 `cfg.root` 기준으로 파일 찾음.

## 이번 세션에서 고친 것 (시간순)

1. **버그 3개 수정**
   - `stack.env` 탐색을 바이너리 위치가 아닌 **cwd→상위 탐색**으로 (`findRoot()`) — 전역 설치 시 작동 안 하던 문제
   - `checkStackExists()`: `ps -a`(헤더 때문에 항상 비어있지 않음) → `ps -aq`로 수정. `go-stack up`이 영구 차단되던 버그
   - `go-stack connect`: shell 문자열 통짜를 `run()`에 단일 인자로 넘기던 것 → `"sh","-c",script` 로 argv 분리
2. **`go-stack init` 명령어 추가**: 현재 폴더에 `stack.env` 템플릿 생성(항목별 설명 주석 포함) + `.gitignore`에 자동 등록. `_project/` 폴더는 만들지 않음. 이미 `stack.env` 있으면 에러 후 아무것도 안 바꿈.
3. **인자 없이 `go-stack` 실행 시 help 출력**: `loadConfig()`(stack.env 필수) 호출 전에 처리하도록 순서 변경 — stack.env 없는 곳에서도 `go-stack`/`go-stack help`/`go-stack init` 동작
4. **`FOLDER_NAME` 환경변수 + `checkProjectDir()` 제거**: 실측으로 `docker compose stop/down`이 컨테이너 라벨만으로 동작함을 확인, `_project` 폴더 존재 여부 체크가 불필요한 차단이었음 (특히 `_project` 삭제된 뒤에도 `down`/`backup`으로 정리할 수 있어야 하는데 막고 있었음)
5. **전체 콘솔 출력 한글화**: `redLog`/`fatal`/`fmt.Errorf`/`fmt.Fprintf` 등 go-stack가 직접 찍는 메시지 전부 한글로. (단, `docker compose`가 자체적으로 찍는 출력은 번역 불가 — 도커 바이너리 소관)
6. **`_project` 서브폴더 개념 전체 제거**: `cfg.projectDir` 필드 삭제, compose 파일은 `cfg.root`(stack.env와 같은 폴더) 기준으로 직접 탐색. `deploy.go`의 S3 배포도 폴더 통째 스왑 대신 `COMPOSE_FILE_*`/`COMPOSE_BASE_FILE` 값으로 정해지는 개별 파일명만 다운로드/교체하도록 변경(`composeFileNames()`)

## 릴리즈 (yum 배포)

`git tag vX.Y.Z && git push origin vX.Y.Z` 하면 `.github/workflows/release.yml`이 실행됨:
빌드 → `nfpm.yaml` 기준 rpm 생성(gpg 서명 없음, x86_64만) → `createrepo_c`로 메타데이터 생성 → `gh-pages` 브랜치에 커밋(GitHub Pages로 호스팅, 무료).
사용자 설치는 README "yum" 섹션 참고. GitHub repo Settings → Pages source를 `gh-pages`로 최초 1회 수동 설정 필요.

## 검증 방식 (이 프로젝트에서 반드시 실전 테스트할 것)

`go build`만으로 안심 금지 — 매번 `/tmp`에 테스트 stack.env + docker-compose.yml 만들어서 `go-stack up/status/stop/down` 등 실제 docker로 검증함. 예:
```bash
mkdir /tmp/test && cd /tmp/test
printf 'STACK_NAME=test\nCOMPOSE_FILE_LOCAL=docker-compose.yml\n' > stack.env
printf 'services:\n  web:\n    image: alpine:latest\n    command: sleep 3600\n' > docker-compose.yml
/path/to/go-stack up local
```

## 아직 스코프 밖 (사용자가 결정 안 함, 건드리지 말 것)

- `getMount()`가 `docker` 커맨드 하드코딩 — podman 환경에서 깨짐
- `cfg.cmd[0]`을 compose 서브커맨드 아닌 곳(`volume ls`, `exec`, `cp`)에 재사용 — v1(`docker-compose`) 바이너리 환경에서 깨질 수 있음
- `deploy.go`의 `Makefile`/`_script/` S3 배포 부분 — 이미 저장소에서 삭제된 레거시 bash 산출물을 참조하는 죽은 코드, 별도 정리 필요
- **go-sdk(`github.com/docker/go-sdk`) 채택 여부** — 검증 완료, **비추천**. compose 오케스트레이션 미지원, pre-1.0 WIP("API may change"), 지금 1줄 shell-out으로 되는 걸 무거운 의존성으로 바꾸는 격. `--format json`으로 구조화 출력은 의존성 0으로 이미 해결됨
- (해결됨) 바이너리/CLI 개명 — `go-stack`으로 확정, 저장소명(`dtx-docker-manager`)은 미변경(사용자가 GitHub에서 별도 처리 예정)

## 컨벤션

- 커밋 메시지: `[feat]`/`[fix]`/`[docs]` 접두사, 한글 설명 (git log 참고)
- 코드 내 사용자 대상 문자열(에러/로그): 전부 한글. 기술 식별자(`STACK_NAME`, `stack.env`, `go-stack` 등)는 원문 유지
- 주석은 거의 안 씀 — 자명하지 않은 워크어라운드에만 최소한으로
