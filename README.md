# go-stack — Docker Stack Management CLI 🐳

[![release](https://github.com/aimnext-dev1/go-stack/actions/workflows/release.yml/badge.svg)](https://github.com/aimnext-dev1/go-stack/actions/workflows/release.yml)

`go-stack` is a single-binary CLI that makes it easy to **deploy, manage, back up, and restore** `docker compose` stacks.
In the past every project copied a whole `_script/` + `Makefile` set; now install `go-stack` globally once, and each
project folder only needs `stack.env` + `docker-compose*.yml`.

## Contents

- [Install](#-install)
- [Upgrade](#-upgrade)
- [Project layout](#-project-layout-per-service-folder-managed-by-go-stack)
- [Before you start](#-before-you-start)
- [Basic usage](#basic-usage)
- [Backup filename convention](#-backup-filename-convention)
- [Example](#-example)

---

## 📦 Install

### yum (RHEL/CentOS/Fedora, x86_64)

```bash
sudo curl -o /etc/yum.repos.d/go-stack.repo https://aimnext-dev1.github.io/go-stack/repo/go-stack.repo
sudo yum install go-stack
```

### Build from source

```bash
git clone https://github.com/aimnext-dev1/go-stack.git
cd go-stack
go build -o go-stack .
sudo mv go-stack /usr/local/bin/     # or anywhere on PATH
```

Requires Go 1.26+ (only for building; once installed, only the binary is needed).

---

## ⬆️ Upgrade

### yum

New versions are only picked up after a fresh `git tag vX.Y.Z` is pushed (re-pushing the same tag does not trigger a rebuild). yum also caches repo metadata, so refresh it before upgrading:

```bash
sudo yum clean all
sudo yum upgrade go-stack
```

### Source build

```bash
git pull
go build -o go-stack .
sudo mv go-stack /usr/local/bin/
```

---

## 📁 Project layout (per service folder managed by go-stack)

```text
my-service/
├── stack.env                  # this stack's config (hand-written, not committed to git)
├── docker-compose.yml         # base: shared service definitions (optional)
├── docker-compose.local.yml   # local only
├── docker-compose.dev.yml
└── docker-compose.prod.yml
```

`go-stack` walks up from the current directory looking for `stack.env` (the same way git looks for `.git`) —
you can run `go-stack` from anywhere inside a service folder.

`_backup/` (backup output) and `_volume/` (pull/push working dir) are created automatically by `go-stack` as needed.

---

## 🙌 Before you start

### Requirements

`go-stack` auto-detects the container runtime, in this order:
1. `docker` with the `compose` v2 CLI plugin
2. `docker` daemon only, with a standalone `docker-compose` binary on PATH (works whether it's the old v1 CLI or a v2-engine binary published under the `docker-compose` name — go-stack just shells out to whichever is found)
3. `podman` with the `compose` plugin
4. `podman-compose`

If auto-detection picks the wrong one, force it with `GO_STACK_CONTAINER=docker` or `GO_STACK_CONTAINER=podman` in `stack.env`.

(Unlike the old bash version, there's no `jq` dependency — the Go binary handles JSON directly.)

### Initial setup (stack.env)

```bash
mkdir my-service && cd my-service
go-stack init                          # creates stack.env + registers it in .gitignore
vi stack.env                           # fill in values (see table below)
vi docker-compose.yml                  # write service definitions (same folder as stack.env)
```

> `go-stack init` only creates `stack.env`. Write the compose file yourself.
> If `stack.env` already exists in the folder, it prints an error and changes nothing.

Values to fill in `stack.env`:
```text
STACK_NAME            # docker compose project name (-p) / backup filename prefix

COMPOSE_FILE_LOCAL     # local environment compose spec filename (relative to stack.env's folder)
COMPOSE_FILE_DEV       # dev environment compose spec filename
COMPOSE_FILE_PROD      # prod environment compose spec filename

COMPOSE_BASE_FILE      # (optional) shared base compose filename — see "Per-environment overrides" below

ENV_FILE_LOCAL         # (optional) local environment env file path
ENV_FILE_DEV           # (optional) dev environment env file path
ENV_FILE_PROD          # (optional) prod environment env file path

GO_STACK_CONTAINER     # (optional) docker | podman — force the runtime instead of auto-detecting
```

> It's recommended to add `stack.env` to `.gitignore` (per-project values, no need to commit).

### Per-environment overrides

Maintaining completely separate compose files per environment duplicates service definitions.
The recommended approach is the standard compose overlay pattern: put shared service definitions
in a base file, and only the differing values in each environment file. Setting
`COMPOSE_BASE_FILE=docker-compose.yml` in `stack.env` runs the combination
`-f docker-compose.yml -f docker-compose.<env>.yml` automatically.
If unset, the single per-environment file is used as a complete standalone spec.

---

## Basic usage

### 🔹 Start / remove the stack

```bash
go-stack up [env]       # e.g. go-stack up local (defaults to local if unspecified)
go-stack down
go-stack update [env]   # rebuild changes and recreate the stack (compose up -d --build)
```

### 🔹 Service control
```bash
go-stack start [svc...]        # starts all services if left empty
go-stack stop [svc...]         # stops all services if left empty
go-stack restart [svc...]      # restarts all services if left empty
go-stack status
go-stack log [container]       # auto-selects if exactly one container, else lists them
go-stack connect [container]   # exec into the container (bash if available, else sh)
```

### 🔹 Backup / restore
```bash
go-stack backup [no-stop]
go-stack restore <timestamp> [no-stop]   # e.g. go-stack restore 20250331_1325

go-stack isave [source]
go-stack iload <timestamp>               # e.g. go-stack iload 20250331_1325
```

> `backup`/`restore` stop the stack before running and start it again afterward, for data consistency.
> To skip stopping, pass the `no-stop` argument (e.g. `go-stack backup no-stop`).
>
> `isave` (image backup) by default commits the container's current state (including runtime changes) before backing up.
> To back up the original image as-is (faster, smaller), use `go-stack isave source`.
> `iload` (image restore) only loads images locally — afterward you must manually update the compose file's
> `image:` value to the restored image name and run `go-stack up`.

### 🔹 Apply volume changes
```bash
go-stack pull
# edit the data under ./_volume, then
go-stack push
```

### 🔹 Deploy from S3
```bash
go-stack deploy [dev|prod]   # defaults to dev
```
Requires the `aws` CLI installed and credentials configured, plus `DEPLOY_S3_BUCKET_DEV` / `DEPLOY_S3_BUCKET_PROD` set in `stack.env`.

### 🔹 Misc
```bash
go-stack clear          # prune unused images (docker image prune -af)
go-stack help           # print command help
```

## 🧩 Backup filename convention

Backups are saved automatically under `_backup/`, named as follows:

### Image backup
> `<stack name>.image.<backup timestamp>.tar.gz`
> e.g. `iot-db.image.20250331_1325.tar.gz`

### Volume backup
> `<stack name>.volume.<backup timestamp>.tar.gz`
> e.g. `iot-db.volume.20250331_1325.tar.gz`

## 🧪 Example
```bash
mkdir my-service && cd my-service
go-stack init
# fill in stack.env, write docker-compose.yml, then
go-stack up local
go-stack status
go-stack backup
go-stack restore 20250331_1325
```
