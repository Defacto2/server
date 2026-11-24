package flags

import "fmt"

const (
	ansiReset = "\033[0m"
	ansiBold  = "\033[1m"
	ansiDim   = "\033[2m"
	ansiCyan  = "\033[36m"
)

func ansi(color bool, code, s string) string {
	if !color || s == "" {
		return s
	}
	return code + s + ansiReset
}

func appHelpTemplate(color bool) string {
	h := func(s string) string { return ansi(color, ansiBold, s) }
	d := func(s string) string { return ansi(color, ansiDim, s) }
	c := func(s string) string { return ansi(color, ansiCyan, s) }

	return fmt.Sprintf(`%s — {{.Usage}}
%s{{if .Description}}

{{.Description}}{{end}}

%s
  {{.HelpName}} [command] [options]
{{if .VisibleCategories}}
%s{{range .VisibleCategories}}{{if .Name}}
  %s{{end}}{{range .VisibleCommands}}
    %s`+"\t"+`{{.Usage}}{{end}}
{{end}}{{end}}{{if .VisibleFlags}}
%s{{range .VisibleFlags}}
  {{.}}{{end}}
{{end}}
%s
  Website`+"\t"+`%s
  GitHub`+"\t"+`%s

%s
`,
		h("{{.HelpName}}"),
		d("{{.Version}}"),
		h("USAGE"),
		h("COMMANDS"),
		d("{{.Name}}"),
		c(`{{join .Names ", "}}`),
		h("OPTIONS"),
		h("LINKS"),
		c("https://"+Domain),
		c(Repository),
		d("{{.Copyright}} <"+Email+">"),
	)
}

func commandHelpTemplate(color bool) string {
	h := func(s string) string { return ansi(color, ansiBold, s) }
	c := func(s string) string { return ansi(color, ansiCyan, s) }

	return fmt.Sprintf(`%s — {{.Usage}}{{if .Description}}

{{.Description}}{{end}}

%s
  {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [options]{{end}}{{end}}{{if .Aliases}}

%s
  %s{{end}}{{if .VisibleFlags}}

%s{{range .VisibleFlags}}
  {{.}}{{end}}{{end}}
`,
		h("{{.HelpName}}"),
		h("USAGE"),
		h("ALIASES"),
		c(`{{join .Aliases ", "}}`),
		h("OPTIONS"),
	)
}
