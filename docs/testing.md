# Testing Policy

This document owns the repository testing policy. Other documentation should link here instead of copying command lists.

## Local Evidence

Use the cheapest check that can fail for the behavior being changed. Do not run a full suite by default.

| Change                      | Required local evidence                                                                |
| --------------------------- | -------------------------------------------------------------------------------------- |
| Go package behavior         | `cd src && go test ./internal/<affected-package> [./internal/<affected-package>/test]` |
| Vue component behavior      | `cd vue && npm test -- <target-spec>`                                                  |
| Vue source quality          | `cd vue && npm run lint -- <changed-files>`                                            |
| Frontend build or packaging | `cd vue && npm run build`                                                              |
| Docker Compose structure    | `bash scripts/docker-verify.sh`                                                        |
| Governance files            | `bash scripts/check-governance.sh`                                                     |

Run a build only when the change affects TypeScript project configuration, bundling, dependencies, public assets, or production packaging. Record the exact command and result; a source inspection is not runtime or deployment proof.

When an affected Go owner has a sibling `test` package, run both paths in the same command. In particular, handler, bazi, fortune, and ziwei changes include `./internal/handler/test`, `./internal/service/bazi/test`, `./internal/service/fortune/test`, or `./internal/service/ziwei/test` respectively. A non-recursive package pattern does not discover these siblings.

## Risky Checks

`scripts/test-e2e.sh` creates users and persisted application data. Run it only against an explicitly approved disposable environment. Do not point it at an arbitrary `BASE_URL`, and do not publish its token-bearing output.

Scripts that read credentials, call external services, publish artifacts, deploy, or delete data require explicit approval for the named target and effect. Missing approval blocks only that external effect, not local reading, editing, or focused checks.

## CI Coverage

The pattern-profile workflow validates only its named release-anchor contract. The governance workflow validates repository metadata, command drift, and the static Docker port contract. Focused workflows replay selected Go package pairs and the recent-chart/ZiWei frontend tests, lint, and TypeScript contract only when their owned paths change. None of these workflows proves the full backend, frontend, browser flow, deployment, or provider behavior.

An exhaustive backend/frontend CI matrix is not yet present. Until it is added and required by branch protection, do not describe this repository as having full regression coverage.

## Evidence Rules

- Unit and component tests prove only the exercised package or rendered contract.
- Builds prove compilation and packaging, not browser behavior or deployment health.
- Static Docker checks prove declared structure, not that containers start successfully.
- Real endpoint or browser evidence must name the environment and any data it changed.
- CI must replay expected outputs read-only; it must never rewrite fixtures or repository state.
