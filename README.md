# firmdiff

**Build A. Build B. Compare firmware outputs.**  
A small tool for embedded teams to detect what *actually changed* between two builds (flags, toolchain, env).

## Why
Zombie firmware projects bite when:
- the original build environment is lost
- a compiler upgrade changes size or behavior
- an IDE hides flags and link steps

`firmdiff` makes build differences visible.

## Install

### Homebrew (later)
Coming soon.

### Go install
```bash
go install github.com/lostbinarylabs/firmdiff/cmd/firmdiff@latest