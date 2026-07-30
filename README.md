# dtx-docker-manager

# Docker Stack Management CLI 🐳

`docker compose` 스택을 손쉽게 **배포/관리/백업/복원**할 수 있는 단일 바이너리 CLI `dtx`입니다.
과거에는 프로젝트마다 `_script/`+`Makefile` 전체를 복사해서 썼지만, 이제 `dtx`를 한 번 전역 설치해두면
각 프로젝트 폴더에는 `stack.env` + `docker-compose*.yml`만 있으면 됩니다.

---

## 📦 설치

### yum (RHEL/CentOS/Fedora, x86_64)

```bash
sudo curl -o /etc/yum.repos.d/dtx.repo https://aimnext-dev1.github.io/dtx-docker-manager/repo/dtx.repo
sudo yum install dtx
```

### 소스 빌드

```bash
git clone <이 저장소>
cd dtx-docker-manager
go build -o dtx .
sudo mv dtx /usr/local/bin/     # 또는 PATH가 잡힌 아무 위치
```

Go 1.26 이상이 필요합니다(빌드 시에만; 설치 후에는 바이너리만 있으면 됩니다).

---

## 📁 프로젝트 구조 (dtx가 관리하는 각 서비스 폴더)

```text
my-service/
├── stack.env                  # 이 스택의 설정값 (직접 작성, git에는 커밋하지 않음)
├── docker-compose.yml         # 베이스: 서비스 공통 정의 (선택)
├── docker-compose.local.yml   # local 전용
├── docker-compose.dev.yml
└── docker-compose.prod.yml
```

`dtx`는 현재 디렉토리에서 상위로 올라가며 `stack.env`를 찾습니다(git이 `.git`을 찾는 방식과 동일) —
서비스 폴더 안 어디서든 `dtx` 명령을 실행할 수 있습니다.

`_backup/`(백업 결과물), `_volume/`(pull/push 작업 폴더)는 필요할 때 `dtx`가 자동 생성합니다.

---

## 🙌 사용에 앞서

### 사용 요건
* Docker CLI / `docker compose`(v2), 또는 Podman / `docker-compose`(v1)가 PATH에 있어야 함

(과거 bash 버전과 달리 `jq` 의존성은 없습니다 — Go 바이너리 내부에서 JSON을 직접 처리합니다.)

### 초기 설정 (stack.env)

```bash
mkdir my-service && cd my-service
dtx init                               # stack.env 생성 + .gitignore에 등록
vi stack.env                          # 값 입력 (아래 표 참고)
vi docker-compose.yml                 # 서비스 정의 작성 (stack.env와 같은 폴더)
```

> `dtx init`은 `stack.env`만 생성합니다. compose 파일은 직접 작성하세요.
> 이미 `stack.env`가 있는 폴더에서 실행하면 에러 메시지를 띄우고 아무것도 바꾸지 않습니다.

`stack.env`에 채워야 할 값:
```text
STACK_NAME           # docker compose 프로젝트명(-p) / 백업 파일명 접두사

COMPOSE_FILE_LOCAL    # 로컬환경 도커 컴포즈 명세파일명 (stack.env와 같은 폴더 기준)
COMPOSE_FILE_DEV      # 개발환경 도커 컴포즈 명세파일명
COMPOSE_FILE_PROD     # 운영환경 도커 컴포즈 명세파일명

COMPOSE_BASE_FILE     # (선택) 공통 베이스 compose 파일명 — 아래 "환경별 오버라이드" 참고

ENV_FILE_LOCAL        # (선택) 로컬환경 환경변수 파일 경로
ENV_FILE_DEV          # (선택) 개발환경 환경변수 파일 경로
ENV_FILE_PROD         # (선택) 운영환경 환경변수 파일 경로

DTX_CONTAINER         # (선택) docker | podman — 자동 감지 우선순위를 강제로 지정
```

> `stack.env`는 `.gitignore`에 포함하는 것을 권장합니다(프로젝트별 값이므로 커밋 불필요).

### 환경별 오버라이드

환경마다 완전히 다른 compose 파일을 따로 관리하면 서비스 정의가 중복됩니다.
공통 서비스 정의는 베이스 파일에 두고, 환경별 파일에는 차이나는 값만 작성하는
compose 표준 오버레이 방식을 권장합니다. `stack.env`에서 `COMPOSE_BASE_FILE=docker-compose.yml`을
설정하면 `-f docker-compose.yml -f docker-compose.<환경>.yml`로 자동 조합되어 실행됩니다.
설정하지 않으면 환경별 파일 하나를 완전한 단독 스펙으로 사용합니다.

---

## 🛠️ 기본 사용법

### 🔹 스택 실행 / 제거

```bash
dtx up [환경]       # ex) dtx up local (환경 미지정시 local)
dtx down
dtx update [환경]   # 변경분 빌드 후 재생성 (compose up -d --build)
```

### 🔹 서비스 제어
```bash
dtx start [서비스명...]       # 비워놓을 경우 전체 시작
dtx stop [서비스명...]        # 비워놓을 경우 전체 중지
dtx restart [서비스명...]     # 비워놓을 경우 전체 재시작
dtx status
dtx log [컨테이너명]           # 미지정시 컨테이너 1개면 자동 선택, 여러 개면 목록 출력
dtx connect [컨테이너명]        # 컨테이너 안으로 접속 (bash 있으면 bash, 없으면 sh)
```

### 🔹 백업 / 복원
```bash
dtx backup [no-stop]
dtx restore <백업시간> [no-stop]   # 예: dtx restore 20250331_1325

dtx isave [source]
dtx iload <백업시간>                # 예: dtx iload 20250331_1325
```

> `backup`/`restore`는 데이터 정합성을 위해 진행 전 스택을 중지하고, 완료 후 다시 시작합니다.
> 중지 없이 진행하려면 `no-stop` 인자를 붙이세요(예: `dtx backup no-stop`).
>
> `isave`(이미지 백업)는 기본적으로 컨테이너의 현재 상태(런타임 변경분 포함)를 커밋해 백업합니다.
> 원본 이미지 그대로(더 빠르고 용량이 작음) 백업하려면 `dtx isave source`를 사용하세요.
> `iload`(이미지 복원)는 이미지를 로컬로 불러오기만 합니다 — 이후 compose 파일의 `image:` 값을
> 복원한 이미지명으로 수동으로 바꾸고 `dtx up`을 실행해야 합니다.

### 🔹 볼륨 변경사항 적용
```bash
dtx pull
# ./_volume 에 받은 데이터를 수정한 뒤
dtx push
```

### 🔹 기타
```bash
dtx clear          # 미사용 이미지 정리 (docker image prune -af)
dtx help           # 명령어 도움말 출력
```

## 🧩 백업 파일명 규칙

백업은 `_backup/` 폴더에 자동 저장되며, 다음 규칙으로 이름이 생성됩니다:

### 이미지 백업
> `<스택이름>.image.<백업날짜_시간>.tar.gz`
> 예: `iot-db.image.20250331_1325.tar.gz`

### 볼륨 백업
> `<스택이름>.volume.<백업날짜_시간>.tar.gz`
> 예: `iot-db.volume.20250331_1325.tar.gz`

## 🧪 예시
```bash
mkdir my-service && cd my-service
dtx init
# stack.env 값 입력, docker-compose.yml 작성 후
dtx up local
dtx status
dtx backup
dtx restore 20250331_1325
```
