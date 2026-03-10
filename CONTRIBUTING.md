**# Contributing to firmdiff

Thanks for your interest in improving firmdiff.

This project aims to help engineers understand firmware build differences.

---

# Development setup

Clone the repository:


git clone https://github.com/lostbinarylabs/firmdiff

cd firmdiff


Build the project:


go build ./cmd/firmdiff


Run tests:


go test ./...


---

# Code style

Please follow these conventions:

- keep functions small and focused
- avoid deep nesting
- add comments explaining non-obvious logic
- run `go fmt`

---

# Pull requests

Before submitting a PR:

- ensure tests pass
- run `go vet`
- ensure new code is documented

---

# Feature requests

If you'd like to propose a feature:

1. Open a GitHub issue
2. Describe the problem the feature solves
3. Provide example use cases

---

# Bug reports

Please include:

- firmdiff version
- operating system
- reproduction steps
- example firmware if possible
