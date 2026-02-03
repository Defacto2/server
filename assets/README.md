# Asset Build Process

The assets/ directory contains original JavaScript and CSS source files.
These files are minified and copied to the public/ directory during the build process.

To add a new JavaScript or CSS file to the project, follow these steps:

## Step-by-Step Instructions

1. **Create the source file**
   - Add your new file to assets/js/ or assets/css/ directory

2. **Add build configuration**
   - Add a new function in runner/runner.go that returns api.BuildOptions for your file
   - Call this function in the main() function's bundles slice

3. **Define constants**
   - Add two new constants in handler/app/app.go:
     - A constant for the minified file path (e.g., `NewJS`)
     - A constant for the public path (e.g., `NewPub`)

4. **Add SRI verification**
   - Add a new field to the SRI struct in handler/app/app.go to store the Subresource Integrity hash
   - Add a corresponding line in the Verify function to compute the integrity hash

5. **Create the route**
   - Add a new route in handler/router.go using e.FileFS() to serve the file

6. **Add template functions**
   - Create a new TemplateFuncMap entry in handler/app/funcmap.go for the file path and SRI hash

7. **Update templates**
   - Add a new <script> or <link> element in view/app/layout.tmpl linking to the minified file with integrity attribute

## Example: Adding a new.js File

### 1. Create source file
```
assets/js/new.js
```

### 2. Add build configuration in runner/runner.go
```go
func New() api.BuildOptions {
	min := "new.min.js"
	entryjs := filepath.Join("assets", "js", "new.js")
	output := filepath.Join("public", "js", min)
	return api.BuildOptions{
		EntryPoints:       []string{entryjs},
		Outfile:           output,
		Target:            ECMAScript,
		Write:             true,
		Bundle:            true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Banner: map[string]string{
			"js": fmt.Sprintf("/* %s %s %s */", min, C, time.Now().Format("2006")),
		},
	}
}
```

Add to main() function's bundles slice:
```go
New(),
```

### 3. Define constants in handler/app/app.go
```go
const (
	NewJS  = "/js/new.min.js"
	NewPub = public + NewJS
)
```

### 4. Add SRI struct field
```go
type SRI struct {
	// ... existing fields ...
	NewSRI string
}
```

Add to Verify function:
```go
s.NewSRI, err = helper.Integrity(NewPub, fs)
if err != nil {
	// ... handle error ...
}
```

### 5. Create route in handler/router.go
```go
e.FileFS(app.NewJS, app.NewPub, public)
```

### 6. Add template function in handler/app/funcmap.go
```go
"sriNew": func() string { return web.Subresource.NewSRI },
"newjs":  func() string { return app.NewJS },
```

### 7. Update template in view/app/layout.tmpl
```html
<script async src="{{ newjs }}" integrity="{{ sriNew }}" crossorigin="anonymous"></script>
```
