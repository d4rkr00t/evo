# @evobuild/cli

## 0.7.0

### Minor Changes

- 5769dfd: output evo_cache_key to use as an input for github actions cache key or similar

### Patch Changes

- f7bf598: Fix build hanging sometimes

## 0.6.0

### Minor Changes

- 28c9bb5: Add progress reporter with a spinner
- 2b7c397: feat: improve reporter output with the number of failed/succeeded tasks
- 1955d33: Simplify cache hit message when outputs match

### Patch Changes

- 8fd1e84: Overrides are applied based on suffixes match not prefixes
- 186c8b7: reporter enable spinner acutally enables spinner
- 3774dad: Fix installation with improved postinstall script
- 3f35741: revert changes in reporter that produce 0 output in CombindOutput mode

## 0.5.1

### Patch Changes

- e0b7fef: fix missing dependency

## 0.5.0

### Minor Changes

- ac60948: evo run outputs a list of available top level targets
- 334c415: `evo list` command show a list of available targets
- f37ca96: go 1.17 -> 1.18
- 9aa1074: Move evo config to .evo.json file
- ef3f0ad: chore: internal refactor
- 3206cd7: Log errors returned from Evo internal CLI command
- 53642de: Always log saved time after command execution

### Patch Changes

- 46da5a0: Fix excludes path comparisons
- bae7ef2: Fix race conditions highlighted by -race flag
- 6ddb74c: Use concurent-map in stats to avoid race conditions

## 0.4.0

### Minor Changes

- f2c51b0: Add --since flag

### Patch Changes

- 133bb0d: Make --since work as only

## 0.2.3

### Patch Changes

- faa3de5: Test CI release

## 0.2.2

### Patch Changes

- test release
