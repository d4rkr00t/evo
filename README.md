- [x] Replace `build` with `run`
- [x] Add configuration
  - [x] Namespace config
- [x] Move more caching logic to a task
  - [x] Cache tasks with a task name prefix
- [x] Add logging and progress
- [x] Build stats
- [x] REFACTOR: Get rid of runner and project in favor of context
  - [x] Remove runner struct
  - [x] Remove project struct
  - [x] Split Run function into stages
  - [x] Create task runner
- [x] Proper dependency tracking for deciding when WS is updated
  - [x] Cache WS state
  - [x] Use WS_files_hash + dependencies_hash as a WS_hash
  - [x] Use WS states to build dependencies cache
- [x] Rebuild affected by a rule change
  - [x] Use a hash of all rules that apply to a WS
  - [x] Move preprocessed rules to WS
- [x] Add reusable command definition to a config – should be able to run a command defined in a config in multiple rules like `@typescript <params>`
- [x] Store STDOUT + STDERR of a command and replay output
- [x] Rename to "evo" from "evoke"
- [x] Throw an error when pnpm errors
- [x] Throw if target doesn't exist
- [x] Stricter overrides
  - [x] Replace glob with a relative path to a group or a certain package
- [x] Simple Error handling
  - [x] Task execution
  - [x] Task dependencies
- [ ] Workspaces struct
  - [ ] Store all WS
  - [ ] Store all updated WS
  - [ ] Store all affected WS
- [ ] Add dependency install and linking
  - [x] Cache project state
  - [x] Install packages
  - [x] Check if node_modules folder exists
  - [x] Link packages
  - [x] Link dev dependencies
  - [ ] Link binaries
  - [ ] Link peer dependencies
- [ ] Validations
  - [x] Validate external dependencies
  - [x] Validate dep cycles
  - [x] Duplicate WS
  - [ ] Rules
    - [ ] check that dependencies exist
    - [ ] check cycles
    - [ ] check that command exist
- [ ] Different info
  - [x] Show what's included in hash for a workspace – `evo show-hash pkg-a`
  - [x] Show all rules for a WS with overrides – `evo show-rules pkg-a`
  - [ ] Show rule with all overrides – `evo show-rule build pkg-a` (?)
- [ ] REFACTOR: Generic create a task from a rule
- [ ] Scoped runs – `evo run build pkg-a`
- [ ] REFACTOR: Refactor logger to interfaces
- [ ] REFACTOR: Refactor cache to interfaces
- [ ] Throw an error when not in an EVO project
- [ ] Pretty print duration
- [ ] Watch file changes during task run
  - [ ] Update FileSystem cache only when a file changes
- [ ] FileSystem module
  - [ ] In memory cache of file checksums, update only when update time of a file changed
    - [ ] Preserve cache on disk
  - [ ] Add / remove files from cache
  - [ ] Error handling
- [ ] More commands
  - [ ] `evo add dep@ver` to add a dependency
  - [ ] `evo remove dep` to remove a dependency
  - [ ] `evo clear cache`
  - [ ] `evo clear output` – clears all outputs from packages
- [ ] Generators
  - [ ] Generate a project from pnpm/yarn workspace and npm scripts
- [ ] TESTS!!!
- [ ] Watch mode
- [ ] Rebuild examples with a real world use cases
- [ ] Remote cache
- [ ] Per rule inputs config (?)
- [ ] `--force` to force run a command, ignoring cache
- [ ] Implicit dependencies
- [ ] Skip target for a WS, e.g. skip tests for a flakey package. Target should have `skip: true`
- [ ] Log Group, do not print extra line if no log before end
