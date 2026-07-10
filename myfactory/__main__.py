"""Allow running the CLI as `python -m myfactory`."""

from myfactory.cli import main

if __name__ == "__main__":
    raise SystemExit(main())
