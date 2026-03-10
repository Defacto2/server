<!-- Please read: You should always tell an AI agent to read this file before doing any work,
      however, because the neural network these agents use is massive and random,
      there is a good chance the agent will ignore this as it copies and pastes
      other people's code into your project, so beware! -->

- NEVER EVER EVER: Attempt to run the server in the background.
- If you cannot find the web application running, ask me, the operator to start it.
- All custom debugging or other temporary .go files should begin with: // +build ignore

- Linting is always done using `golangci-lint run -c init/.golangci.yaml`
- Always run the linter after doing significant edits to a .go file, confirm the changes have not created new issues.
- When checking for specific lint issues, run the full linter output without grep to see all issues, as some linters may have exit code 1 even when they find issues.

- The application code is running in the background and is reachable at http://localhost:1323/
- Code changes are monitored by Air which will reload the web application with the new code.

- Confirm the unit tests are working after any major edits to a .go file.
- Never create new tests in the root directory, the use of main will always break the application.

- Whenever I ask for a review of something, never automatically apply or make any changes.
- Never undo any changes using git, without first confirming that only the active task will be undone.

- We use Bootstrap 5.3, so it should take priority over bespoke styles.
- When making changes to CSS or JS, the command, "task assets" must always be run to rebuild with the changes.
