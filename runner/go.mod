module github.com/Defacto2/server/runner

go 1.26.1

// go list -m -u all
// go get -u

require github.com/evanw/esbuild v0.27.3

require golang.org/x/sys v0.41.0 // indirect

tool github.com/evanw/esbuild/cmd/esbuild
