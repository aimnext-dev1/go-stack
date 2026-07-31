# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# go-stack

`go-stack` is a single Go-binary CLI that deploys, manages, backs up, and restores `docker compose` stacks.
Rewrite in progress from a collection of bash scripts (`_script/*.sh` + `Makefile`) to Go. Project size: ~9 .go files, ~700 lines.

## Build/run

```bash
go build -o go-stack .
go vet ./...
```

No automated tests (`*_test.go`) — see "Verification" below; verify manually against real docker.

## Structure

Each file owns one command group (flat layout, no package split, everything is `package main`):

- `main.go` — entrypoint, command dispatch, `help` output
- `types.go` — `cmds` map (command name → fn/usage/group/minArgs), `groupOrder`
- `config.go` — `stack.env` loading (`findRoot()` walks up from cwd, like git), container runtime detection (docker/podman/docker-compose)
- `helpers.go` — shared utils: `run`/`runOut`/`compose`/`composeOut`, `checkStackExists`, `resolveComposeFiles`, `validateTimestamp`, etc.
- `init.go` — `go-stack init` (generates the stack.env template + registers it in .gitignore)
- `stack.go` — up/down/update
- `service.go` — status/start/stop/restart/log/connect
- `volume.go` — pull/push/backup/restore (volumes)
- `image.go` — isave/iload (image backup/restore)
- `deploy.go` — deploy (compose file deployment from S3, has legacy remnants)/clear

## Project layout (user's stack.env folder)

The `_project/` subfolder **is gone** — compose files live **directly in the same folder** as `stack.env`:
```
my-service/
├── stack.env
├── docker-compose.yml
├── docker-compose.local.yml
├── docker-compose.dev.yml
└── docker-compose.prod.yml
```
No `cfg.projectDir` field; `resolveComposeFiles()` finds files relative to `cfg.root`.

## Changes made this session (chronological)

1. **Fixed 3 bugs**
   - `stack.env` lookup changed from binary location to **cwd → parent search** (`findRoot()`) — fixed it not working after a global install
   - `checkStackExists()`: `ps -a` (never empty because of the header) → fixed to `ps -aq`. Fixed `go-stack up` being permanently blocked.
   - `go-stack connect`: was passing a single shell string blob to `run()` → split into argv as `"sh","-c",script`
2. **Added the `go-stack init` command**: generates a `stack.env` template (with per-field comments) in the current folder + auto-registers it in `.gitignore`. Does not create a `_project/` folder. If `stack.env` already exists, errors out and changes nothing.
3. **Print help when `go-stack` runs with no arguments**: reordered so this happens before `loadConfig()` (which requires stack.env) is called — `go-stack`/`go-stack help`/`go-stack init` now work even outside a stack.env folder
4. **Removed the `FOLDER_NAME` env var + `checkProjectDir()`**: confirmed empirically that `docker compose stop/down` works purely off container labels — the `_project` folder existence check was an unnecessary block (in particular it prevented cleaning up via `down`/`backup` after `_project` had already been deleted)
5. **Translated all console output**: `redLog`/`fatal`/`fmt.Errorf`/`fmt.Fprintf` messages that go-stack prints directly are all in English now. (Output that `docker compose` prints on its own can't be translated — that's up to the docker binary.)
6. **Removed the `_project` subfolder concept entirely**: deleted the `cfg.projectDir` field; compose files are now found directly relative to `cfg.root` (same folder as stack.env). `deploy.go`'s S3 deploy also changed from swapping the whole folder to downloading/replacing only the individual filenames determined by `COMPOSE_FILE_*`/`COMPOSE_BASE_FILE` (`composeFileNames()`)
7. **Renamed the binary/CLI from `dtx` to `go-stack`**: to better convey broader applicability. The GitHub repo was later renamed to match (`aimnext-dev1/go-stack`).
8. **Translated all user-facing strings to English**: the repo is now public; all error/log messages, README, and this file were translated from Korean.

## Release (yum distribution)

`git tag vX.Y.Z && git push origin vX.Y.Z` triggers `.github/workflows/release.yml`:
build → generate rpm per `nfpm.yaml` (no gpg signing, x86_64 only) → generate repo metadata with `createrepo_c` → commit to the `gh-pages` branch (hosted via GitHub Pages, free).
See the README "yum" section for user install instructions. GitHub repo Settings → Pages source needs to be set to `gh-pages` once, manually.

## Verification (always test for real in this project)

Don't trust `go build` alone — every change gets tested by creating a test stack.env + docker-compose.yml under `/tmp` and running `go-stack up/status/stop/down` etc. against real docker. Example:
```bash
mkdir /tmp/test && cd /tmp/test
printf 'STACK_NAME=test\nCOMPOSE_FILE_LOCAL=docker-compose.yml\n' > stack.env
printf 'services:\n  web:\n    image: alpine:latest\n    command: sleep 3600\n' > docker-compose.yml
/path/to/go-stack up local
```

## Out of scope for now (not decided by the user, don't touch)

- `getMount()` hardcodes the `docker` command — breaks under podman
- `cfg.cmd[0]` is reused in places that aren't compose subcommands (`volume ls`, `exec`, `cp`) — may break under a v1 (`docker-compose`) binary
- The `Makefile`/`_script/` S3 deploy part of `deploy.go` — dead code referencing legacy bash artifacts already deleted from the repo, needs separate cleanup
- **Whether to adopt go-sdk (`github.com/docker/go-sdk`)** — evaluated, **not recommended**. No compose orchestration support, pre-1.0 WIP ("API may change"), would trade a heavy dependency for what's currently a one-line shell-out. Structured output via `--format json` already solves this with zero dependencies.

## Conventions

- Commit messages: gitmoji prefix, English description (see git log)
- User-facing strings in code (errors/logs): all English. Technical identifiers (`STACK_NAME`, `stack.env`, `go-stack`, etc.) stay as-is
- Minimal comments — only for non-obvious workarounds
