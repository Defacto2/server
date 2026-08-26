module github.com/Defacto2/server/runner

go 1.26.6

// go list -m -u all
// go get -u

require github.com/evanw/esbuild v0.28.2

require golang.org/x/sys v0.47.0 // indirect

tool github.com/evanw/esbuild/cmd/esbuild
