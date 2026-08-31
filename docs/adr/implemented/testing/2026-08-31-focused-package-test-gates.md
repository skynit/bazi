# ADR: Focused package test gates
Status: implemented

## Problem

Go tests in sibling `test` packages are not discovered by commands such as `go test ./internal/handler`. The repository also has no stable full-suite baseline, so a broad CI matrix would mix unrelated evidence and generated-artifact failures into every change.

## Decision

Local verification runs the changed owner package together with its sibling `test` package when one exists. CI uses path-triggered workflows for selected stable Go package pairs and the recent-chart/ZiWei frontend contract. Docker-related paths run the static port-contract verifier. Full backend, frontend, browser, database, and deployment checks remain separate evidence levels.

## Alternatives considered

`go test ./...` and a full frontend test/build matrix were rejected because they violate the repository's scoped-test policy and currently combine unrelated failures. Testing only owner packages was rejected because it silently omits external-package integration contracts.

## Consequences

Nested integration tests now run when their owners change, and CI remains bounded enough to diagnose failures by domain. The focused workflows are intentionally incomplete: package-specific research artifacts, browser behavior, MySQL, containers, deployment, and real devices still require separately named verification.
