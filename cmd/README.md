# Commandline interface (CLI) README

The CLI options for the `server` should be kept to a minimum.

## Distributed application for download

- Embed an SQLite v3 database?

- Allow the importing and exporting of data.

Instead of using args, parse the stdin and stdout requests to determine the request.

`server > anyname.sql` `server > anyname.json` etc.
`server < anyname.sql`

- Always run in DEV mode with no panic protection.

- Have a BOLD note to never run in production.

---

[Full CLI API ref.](https://cli.urfave.org/v2/examples/full-api-example/)