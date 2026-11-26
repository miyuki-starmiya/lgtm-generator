# AGENTS.md

lgtm-generator is a Go CLI that overlays an LGTM image on pictures in `input/`
and writes them to `output/`. See README.md for usage.

## Before creating or updating a pull request

Run every check below from the repository root. All of them must pass before
you push the branch and create or update the PR. If one fails, fix the cause;
do not skip the check or loosen the lint config to make it pass.

| Check | Command | Passes when |
|---|---|---|
| Format | `gofmt -l .` | it prints nothing |
| Build (compile) | `go build -o /dev/null ./...` | exit code 0 |
| Static analysis | `go vet ./...` | exit code 0 |
| Static analysis | `golangci-lint run ./...` | it prints `0 issues.` |
| Test | `go test -count=1 ./...` | every test passes |

All at once:

```sh
test -z "$(gofmt -l .)" && go build -o /dev/null ./... && go vet ./... && golangci-lint run ./... && go test -count=1 ./...
```

- `go build` compiles the program. Keep `-o /dev/null`: without it,
  `go build` overwrites the tracked `lgtm-generator` binary. `go vet` and
  `go test` also compile the `_test.go` files, so a broken test fails there.
- `golangci-lint` uses `.golangci.yml` (v2 config). It includes staticcheck,
  so the standalone `staticcheck` is not needed.
- Write the commands you ran and their results in the PR description.

## Conventions

- Go files use LF line endings (`.gitattributes`). CRLF makes `gofmt -l` fail.
- A file opened for writing must check the error of `Close`, because a failed
  `Close` can mean the output is incomplete. `.golangci.yml` excludes
  `(*os.File).Close` from errcheck only so that closing files opened for
  reading stays short.
- Tests create their own images in `t.TempDir()`. Do not depend on files in
  `input/` or `output/`; both are git-ignored and differ per machine.
- When CLI behavior changes, update `usageText` in `main.go` (the `--help`
  output), README.md, and the tests together.
