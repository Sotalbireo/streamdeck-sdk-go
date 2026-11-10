# AGENTS.md

## File editing

- Do not edit files with Python (`python` / `python3`) scripts. On Windows, `python` is the Microsoft Store stub:
  it can fail silently without modifying the file, so a change may look applied when it is not.
- Use the editor tools (Edit / Write) to change files. Use `gofmt` for formatting.
- After editing, verify the change (`git diff`, `go vet ./...`, `go test ./...`) instead of assuming it applied.
