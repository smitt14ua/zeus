# ZEUS Documentation

## Structure

```
docs/
├── README.md                         # this file
├── glossary.md                       # terms, abbreviations, aliases
├── formats/
│   ├── server_cfg.md                 # server.cfg parameter reference
│   ├── server_cfg_missions.md        # mission rotation class format
│   └── basic_cfg.md                  # basic.cfg network tuning reference
├── protocols/
│   └── startup_params.md             # Arma 3 startup parameters (server-focused)
└── ai/
    ├── conventions.md                # naming, struct tags, YAML keys, file paths
    ├── architecture_rules.md         # package boundaries, invariants, data flow
    └── common_gotchas.md             # edge cases, dangerous assumptions, traps
```

## Navigation

### I want to understand what a server.cfg field does
→ [formats/server_cfg.md](formats/server_cfg.md)

### I want to understand mission rotation configuration
→ [formats/server_cfg_missions.md](formats/server_cfg_missions.md)

### I want to understand basic.cfg / network tuning
→ [formats/basic_cfg.md](formats/basic_cfg.md)

### I want to understand what startup parameters ZEUS uses
→ [protocols/startup_params.md](protocols/startup_params.md)

### I'm an AI agent working on the codebase
→ Start with [../CLAUDE.md](../CLAUDE.md) (Claude Code) or [../AGENTS.md](../AGENTS.md) (other agents),
  then [ai/architecture_rules.md](ai/architecture_rules.md),
  [ai/conventions.md](ai/conventions.md),
  [ai/common_gotchas.md](ai/common_gotchas.md)

### I want to contribute
→ [../CONTRIBUTING.md](../CONTRIBUTING.md)

### I need to look up a term or abbreviation
→ [glossary.md](glossary.md)

## Source Material

The `formats/` and `protocols/` documentation is derived from the official Arma 3
community wiki. The original wiki files are preserved at:

- `internal/arma/server_config.wiki`
- `internal/arma/basic_server_config.wiki`
- `internal/arma/startup_params.wiki`

## Conventions

- All filenames use `snake_case`.
- Cross-links use relative paths.
- Version notes use the form *(since X.YY)*.
- Code blocks use `cpp` for Arma 3 config syntax and `go` for Go snippets.
- `NOTE:` blocks mark preserved ambiguities from source documentation.
