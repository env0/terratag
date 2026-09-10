# AGENTS.md

<!-- agent-docs:critical:start -->
- Read the nearest guide before editing.
- Every behaviour change ships with the test layer the root path table requires.
- A failed live-suite preflight blocks running the test, not writing it.
<!-- agent-docs:critical:end -->

Terratag is a Go CLI that applies tags/labels across a tree of OpenTofu/Terraform (and Terragrunt)
files, rewriting resource blocks in place.

## Components

| Path | What it does |
| --- | --- |
| `cmd/terratag/` | CLI entrypoint (`main.go`); builds and runs the binary; owns exit-code behaviour |
| `cli/` | Flag parsing and validation (`args.go`) |
| `terratag.go` (root) | Orchestrates a run: read HCL, compute tags, rewrite files |
| `internal/terraform/` | Runs `terraform`/`tofu` to get provider schema and init modules |
| `internal/tfschema/` | Maps a resource type to its schema-declared tag/label attribute |
| `internal/tagging/` | Per-cloud tagging rules (`aws.go`, `azure.go`, `gcp.go`) |
| `internal/convert/` | Rewrites HCL: inserts the tag expression, builds the `terratag_local_...` local |
| `internal/providers/` | Provider tag-support data, including the generated `azure_resource_tag_support.csv` |
| `internal/tag_keys/` | Tag-key defaults and merge/keep-existing behaviour |
| `internal/common/`, `internal/utils/` | Shared types and helpers — see Reuse below |
| `azapi/` | Standalone Python script (`generate.py`) that regenerates the azapi CSV; not part of the Go build |
| `test/tests/` | Fixture pairs (`input`/`expected`, or `in`/`expected` for Terragrunt) driving the integration suites in `terratag_test.go` |
| `test/fixture/<tf-version>/config.yaml` | Lists which `test/tests/*` suites run under each Terraform version |

Import boundary: `internal/*` packages are implementation detail; the root package (`terratag.go`) and
`cli/` are the only callers of `cmd/terratag`. Nothing outside `azapi/` should read or write
`internal/providers/azure_resource_tag_support.csv` by hand — see Reuse.

## Boundaries

- **Always do**: `go fmt ./...`, `go vet ./...`, `SKIP_INTEGRATION_TESTS=1 go test ./...` before
  calling a change done.
- **Ask first**: editing `go.mod`/`go.sum` (dependency changes), `.github/workflows/**`,
  `.goreleaser.yml` — these affect CI and the release pipeline for every contributor.
- **Never touch by hand**: `internal/providers/azure_resource_tag_support.csv`. It is generated
  from Microsoft's tag-support docs — regenerate it with the azapi tool below, never hand-edit it.

In review, an unmet contract row below is **blocking**, not a nit.

## Test-for-change contract

| Changed path | Required test |
| --- | --- |
| `` `^internal/tagging/.*\.go$` `` | `` `^internal/tagging/tagging_test\.go$` `` |
| `` `^terratag\.go$` `` | `` `^test/tests/[^/]+/(in\|input\|expected)/.*\.(tf\|hcl)$` `` |
| `` `^internal/(tagging\|terraform\|convert\|tfschema\|providers\|tag_keys)/.*\.go$` `` | `` `^test/tests/[^/]+/(in\|input\|expected)/.*\.(tf\|hcl)$` `` |
| `` `^cli/.*\.go$` `` | `` `^cmd/terratag/main_test\.go$` `` (`TestExitCodes`) |
| `` `^cmd/terratag/main\.go$` `` | `` `^cmd/terratag/main_test\.go$` `` (`TestExitCodes`) |

Machine-readable form: `.agents/test-contract` (kept in sync by tooling).

Rows apply independently; no row substitutes for another. Skipping the required layer is never a
budget decision — under time pressure, cut scope (a smaller change with its test beats a complete
change without one), or say so explicitly in the PR. A fixture pair still discharges its row when you
cannot run it locally: write it, mark it unrun, and name the missing preflight (below).

`internal/common/`, `internal/utils/`, `internal/file/` hold shared types and helpers, not
tagging behaviour, so they carry no fixture-test row — check Reuse before adding to them.

## Running each test layer

1. **Unit tests** (all packages, integration suites skipped):
   ```bash
   SKIP_INTEGRATION_TESTS=1 go test ./...
   ```
   Narrow to one package while iterating: `go test ./internal/tagging/...`.

2. **Terraform/OpenTofu fixture suites** (`terratag_test.go`, `TestTerraform12`…`TestTerraformlatest`,
   `TestOpenTofu`): each runs `terraform init` / `tofu` against every suite listed in
   `test/fixture/<version>/config.yaml`, runs terratag, then diffs the result against
   `test/tests/<suite>/expected`.
   ```bash
   go test -run '^TestTerraformlatest$' -v
   ```
   Preflight: a matching `terraform` (or `tofu`) binary on `PATH` for the version under test —
   `tfenv install && tfenv use` inside `test/tfenvconf/terraform_<version>/`, or `tofu` directly for
   the latest/OpenTofu suites. A missing or wrong-version binary blocks **running** this test, not
   writing the fixture pair.
   Adding a suite: create `test/tests/<name>/{input,expected}`, then add `<name>` to the `suites:`
   list in every `test/fixture/<version>/config.yaml` the suite should run under.

3. **Terragrunt suites** (`TestTerragruntWithCache`, `TestTerragruntRunAll`): same shape, fixtures
   under `test/tests/<name>/{in,expected}`.
   ```bash
   go test -run '^TestTerragrunt.*$' -v
   ```
   Preflight: `terragrunt` and `terraform` on `PATH`.

4. **CLI exit-code tests** (`cmd/terratag/main_test.go`): builds the binary and runs it as a
   subprocess.
   ```bash
   go test ./cmd/terratag/... -run TestExitCodes -v
   ```

5. **Lint**: `go vet ./...` and `go fmt ./...` are what CI enforces. `golangci-lint run` (config at
   `.golangci.yaml`) is available locally but is not run in CI — optional, not a contract row.

**Seeing a fix work**: this is a CLI with no server to start. Run the built binary against a fixture
directory and inspect the rewritten `.tf`/`.hcl` output directly, or run the one integration suite
that exercises the change (`go test -run '^TestTerraform<version>$/<suite-name>$' -v`) and read its
diff against `test/tests/<suite>/expected`.

## Reuse

- `internal/common/common.go` — shared types every tagging path imports: `IACType`, `Version`,
  `TaggingArgs`, `TerratagLocal`. Add a new cross-package type here, not in the package that happens
  to need it first.
- `internal/utils/utils.go` — `SortObjectKeys` is the only helper; used for deterministic tag-key
  ordering. Check here before writing another map-key sort.
- Promote a type or helper here only once a second package needs it; a single caller keeps it local.

Reimplementing something named above instead of importing it is **blocking**, not a nit.

## Comments and prose

State what the code cannot: a workaround, a non-obvious invariant, an ordering constraint, a git
issue reference for a fix that looks unnecessary otherwise. Never restate a signature or narrate what
a line does.

## Plans

A plan touching a `test-for-change` path names the fixture or unit test it adds, not just the code
change. A plan touching `internal/providers/azure_resource_tag_support.csv` says it will run
`azapi/generate.py`, not hand-edit the CSV.
