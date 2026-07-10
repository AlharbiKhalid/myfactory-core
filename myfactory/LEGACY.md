# LEGACY: Python CLI (migration reference only)

This directory contains the original Python implementation of the MyFactory
CLI. It has been **replaced by the standalone Go CLI** (`cmd/myfactory/`,
`internal/`), which is what the installers ship and what users run.

Status:

- Kept temporarily as the behavioral reference for the Go port.
- Not installed by any installer. `scripts/install-local.sh` and
  `scripts/install.sh` install the Go binary only.
- Do not add features here. Fix bugs in the Go CLI.
- Still runnable for comparison: `python -m myfactory --help` (Python 3.9+).

Related legacy files elsewhere in the repo:

- `pyproject.toml` — packaging for this Python package only.
- `scripts/create-product.py` — still supported (standalone helper for
  creating a brand-new product repo); not part of the CLI migration.

## Cleanup task (after Go parity has soaked)

Once the Go CLI has been used for a release cycle without parity regressions,
a cleanup task can delete:

- `myfactory/` (this whole package)
- `pyproject.toml`

and drop the "Python CLI migration status" section from README.md.
`scripts/create-product.py` should be kept until `myfactory new` replaces it.
