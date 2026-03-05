1. Linting is always done using `golangci-lint run -c init/.golangci.yaml`
2. Always run the linter after doing significant edits to a .go file, confirm the changes have not created new issues.
3. Confirm the unit tests are working after any major edits to a .go file.
4. When checking for specific lint issues, run the full linter output without grep to see all issues, as some linters may have exit code 1 even when they find issues.
5. Keep the "Summary" brief.
