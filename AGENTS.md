# e-timologisi-utils

Shared Go package for `e-timologisi-api` and `mydata-api` (module `github.com/e-timologisi/e-timologisi-utils`). Each service pins its own version in `go.mod` (currently v1.2.43 in both), so a change reaches a service only when it bumps the version.

## What lives here

- The shared user profile: `invoiceusers/invoice_user.go`, `type User struct` (and `BratnetProvider`, `EParoxosProvider`, Mellon, DIAS, Ilyda flags ...). **New user profile fields are added here** and nowhere else: do not define or duplicate profile fields in e-timologisi-api, mydata-api or the GUI.
- Shared helpers: `logging`, `logger`, `resterrors`, `http_utils`, `crypto_utils`, `erganiusers`.

## What must not be here

No AADE (myDATA) or invoice-provider send, build, read or error-handling logic (E-Paroxos, Bratnet, Simply, Ilyda). That is `mydata-api`. This package holds data shapes and neutral helpers only.

## Adding a profile field

- Additive only. Never rename, remove or retag an existing field or JSON tag, and never change a type: the two services (and the stored Elasticsearch user documents) read the same JSON and upgrade at different times. Use `json:"snake_case"` tags like the neighbours; a missing field must mean "off" / zero value.
- Never write flags or data on prod user profiles, and do not put secrets or real user data in tests or fixtures.
- Run `go build ./... && go test ./...` before the PR.

## Release (as seen in the repo; there is no CI or release workflow here)

1. Merge the PR to `main` (squash or merge commit, e.g. #2 became `6276963`).
2. Tag that commit on `main` with the next patch version (`v1.2.44` after `v1.2.43`) and push the tag. Existing tags v1.2.x sit directly on the merged commits. GitHub Releases were created only up to v1.2.37; later versions are tags only. Do not use `v2.x`: a `v2.0.0` tag exists on an old commit and the module path has no `/v2`.
3. Bump the consumers in this order: `mydata-api` first (`go get github.com/e-timologisi/e-timologisi-utils@vX.Y.Z`, `go mod tidy`), then `e-timologisi-api`, then the GUI if it needs the field.
4. Both services' PR checks must pass after the bump, including their e2e jobs (`E2E Bratnet route`, `E2E E-Paroxos route` in mydata-api; `E2E invoice chain` in e-timologisi-api). A bump that weakens or skips an e2e check is not acceptable.

Every merge to a consuming service's `main` deploys to production. Agents do not push tags unless the owner asks.

## Before you write code (extend first)
1. Read `../e-timologisi-docs/docs/as-is/catalog-existing-solutions.md`, then `../e-timologisi-docs/docs/as-is/e-timologisi-utils/overview.md`.
2. Find the owning repo (`../e-timologisi-docs/docs/architecture/service-map.md`) and the nearest existing example. Extend it; never build a parallel solution.
3. Smallest refactor in a separate commit if it does not fit; new code only in the same layout, naming, error and test style.
4. Rules: `../e-timologisi-docs/docs/rules/extend-first.md`.
5. Every task and PR body needs the Reuse note: `../e-timologisi-docs/docs/rules/reuse-note.md`. Fiscal changes need their e2e case.
6. If the code and these docs disagree, the code wins; fix the doc in e-timologisi-docs (committed straight to main there; the service repos still use PRs).
