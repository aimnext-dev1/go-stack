# AGENTS.md

This file provides guidance to coding agents working with code in this repository.

# go-stack

`go-stack` is a single Go-binary CLI that deploys, manages, backs up, and restores `docker compose` stacks.
Rewrite of a bash script collection (`_script/*.sh` + `Makefile`, since deleted from the repo). 10 .go files, ~690 lines, `package main`, no internal packages, no external dependencies (stdlib only).

## Build/run

```bash
go build -o go-stack .
go vet ./...
```

No automated tests (`*_test.go`) — see "Verification" below; every change is tested against real docker.

To embed a version (as the release workflow does): `go build -ldflags "-X main.version=v0.2.0" -o go-stack .`. Plain builds report `dev`.

## Structure

Each file owns one command group (flat layout, everything is `package main`):

- `main.go` — entrypoint, command dispatch, `help`/`-v`/`--version`/`version` (handled before `loadConfig()`, so they work with no `stack.env` present)
- `types.go` — `cmds` map (command name → fn/usage/group/minArgs), `groupOrder` (controls help output order)
- `config.go` — `stack.env` loading (`findRoot()` walks up from cwd, like git), container runtime detection
- `helpers.go` — shared utils: `run`/`runOut`/`compose`/`composeOut`, `checkStackExists`, `resolveComposeFiles`, `validateTimestamp`, `getMount`, etc.
- `init.go` — `go-stack init` (generates the stack.env template + registers it in `.gitignore`)
- `stack.go` — up/down/update
- `service.go` — status/start/stop/restart/log/connect
- `volume.go` — pull/push/backup/restore (volumes)
- `image.go` — isave/iload (image backup/restore)
- `deploy.go` — deploy (compose file deployment from S3, has legacy remnants)/clear

## Config: stack.env

Each managed project has a `stack.env` file (loaded by `config.go:parseEnvFile`, a minimal `KEY=value` parser — no shell expansion, supports quoted values). `findRoot()` walks up from cwd looking for it, so `go-stack` works from any subdirectory of a project, like git.

Compose files live in the **same folder** as `stack.env` (no `_project/` subfolder).

Keys (see `init.go:stackEnvTemplate` for the canonical template `go-stack init` writes):
- `STACK_NAME` — docker compose project name (`-p`) and backup filename prefix
- `COMPOSE_FILE_LOCAL` / `COMPOSE_FILE_DEV` / `COMPOSE_FILE_PROD` — per-environment compose spec filename
- `COMPOSE_BASE_FILE` (optional) — shared base compose file; when set, runs `-f <base> -f <env file>` combined (`helpers.go:resolveComposeFiles`)
- `ENV_FILE_LOCAL` / `ENV_FILE_DEV` / `ENV_FILE_PROD` (optional) — passed as `--env-file` to compose
- `GO_STACK_CONTAINER` (optional) — force `docker` or `podman` instead of auto-detecting
- `DEPLOY_S3_BUCKET_DEV` / `DEPLOY_S3_BUCKET_PROD` — used only by `deploy`

## Container runtime detection (`config.go:detectContainer`)

Tries in order, first match wins:
1. `GO_STACK_CONTAINER` env override (`docker` or `podman`)
2. `docker` daemon reachable (`docker info`) **and** the v2 compose plugin actually present (`docker compose version` succeeds) → `cfg.cmd = ["docker","compose"]`
3. `podman` on PATH → `cfg.cmd = ["podman","compose"]`
4. standalone `docker-compose` binary on PATH (v1 CLI, or a v2-engine binary just published under that name) → `cfg.cmd = ["docker-compose"]`
5. `podman-compose` on PATH
6. none found → fatal error

`cfg.cmd` is only for compose subcommands (`compose()`/`composeOut()`/`composeLines()` in `helpers.go` prepend it + `-p <stackName>`). Direct, non-compose container commands (`docker logs`, `docker exec`, `docker inspect`, `docker cp`, `docker commit`, `docker save/load/rmi`, `docker volume ls`) must use `cfg.containerBin` instead (`"docker"` or `"podman"`, set alongside `cfg.cmd` in every branch of `detectContainer()`), **not** `cfg.cmd[0]` — under the `docker-compose`/`podman-compose` fallback branches, `cfg.cmd[0]` is the compose binary, not the container binary, and would break those commands.

## Commands

Grouped as they appear in `go-stack help` (`types.go`):

- **init**: `go-stack init` — writes `stack.env` template to cwd + adds it to `.gitignore`. Errors without changes if `stack.env` already exists.
- **stack**: `up [env]` (default `local`) / `down` (confirms interactively) / `update [env]` (`up -d --build`)
- **svc**: `status` (`compose ps -a`) / `start` / `stop` / `restart` (`[svc...]`, all services if empty) / `log [name]` (`docker logs -f -n 10000`; auto-selects the sole container if name omitted and exactly one exists, else lists them) / `connect [name]` (`docker exec -it`, prefers bash, falls back to sh)
- **vol**: `pull` (copies each volume's data from containers into `./_volume/`, writes `volume-map.json`) / `push` (copies `./_volume/` data back into containers using that map, then `chown`s to match original owner) / `backup [no-stop]` (stops the stack unless `no-stop`, copies volumes to `_backup/<stack>.volume.<ts>.tar.gz`) / `restore <ts> [no-stop]`
- **img**: `isave [source]` (default: commits each container's current state before saving, so runtime changes are included; `source`: saves the original image as-is, faster/smaller) / `iload <ts>` (loads images locally only — compose file's `image:` must be updated manually afterward, then run `up`)
- **deploy**: `deploy [dev|prod]` (default `dev`; downloads compose files + `Makefile`/`_script/` from `DEPLOY_S3_BUCKET_*` via `aws s3 cp`, has legacy remnants, see below) / `clear` (`docker image prune -af`)

Timestamps everywhere are `YYYYMMDD_HHMM` (`helpers.go:validateTimestamp`).

## Backup filename convention

- Image: `_backup/<STACK_NAME>.image.<YYYYMMDD_HHMM>.tar.gz`
- Volume: `_backup/<STACK_NAME>.volume.<YYYYMMDD_HHMM>.tar.gz`

Both are directories built up under `_backup/`, tarred, then the directory is removed — the `.tar.gz` is the only artifact left. Each also contains a `volume-map.json` (or is paired with one from the same backup) mapping container → `{volume, destination}`, produced by `getMount()` inspecting each container's mounts.

## Release (yum distribution)

`git tag vX.Y.Z && git push origin vX.Y.Z` triggers `.github/workflows/release.yml`:
build (with `-ldflags -X main.version=$GITHUB_REF_NAME`) → generate rpm per `nfpm.yaml` (no gpg signing, x86_64 only) → generate repo metadata with `createrepo_c` → commit to the `gh-pages` branch (hosted via GitHub Pages, free). rpm output accumulates in `repo/x86_64/` across releases (old versions aren't pruned), so `yum upgrade` can move between any published version.

See the README "Install"/"Upgrade" sections for user-facing instructions. GitHub repo Settings → Pages source needs to be `gh-pages` (one-time manual setup). Re-pushing the same tag does **not** re-trigger the workflow — a new tag is required per release, and yum clients must refresh their cache (`yum clean all`) since metadata is cached.

## Verification (always test for real in this project)

Don't trust `go build` alone — every change gets tested by creating a test stack.env + docker-compose.yml under `/tmp` and running `go-stack up/status/stop/down` etc. against real docker. Example:
```bash
mkdir /tmp/test && cd /tmp/test
printf 'STACK_NAME=test\nCOMPOSE_FILE_LOCAL=docker-compose.yml\n' > stack.env
printf 'services:\n  web:\n    image: alpine:latest\n    command: sleep 3600\n' > docker-compose.yml
/path/to/go-stack up local
```

## Known issues / out of scope (not decided by the user, don't touch)

- The `Makefile`/`_script/` S3 deploy part of `deploy.go` (`cmdDeploy`) — dead code referencing legacy bash artifacts already deleted from the repo, needs separate cleanup
- **Whether to adopt go-sdk (`github.com/docker/go-sdk`)** — evaluated, **not recommended**. No compose orchestration support, pre-1.0 WIP ("API may change"), would trade a heavy dependency for what's currently a one-line shell-out. Structured output via `--format json` already solves this with zero dependencies.

## Conventions

- Commit messages: gitmoji prefix, English description, explains *why* (see `git log` for style/history — the full change log lives there, not in this file)
- User-facing strings in code (errors/logs): all English. Technical identifiers (`STACK_NAME`, `stack.env`, `go-stack`, etc.) stay as-is
- Minimal comments — only for non-obvious workarounds
