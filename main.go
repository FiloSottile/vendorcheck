package main

import (
	"flag"
	"fmt"
	"go/build"
	"go/parser"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kisielk/gotool"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/refactor/importgraph"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: vendorcheck [-t] package [package ...]")
		flag.PrintDefaults()
	}
	tests := flag.Bool("t", false, "also check dependencies of test files")
	unused := flag.Bool("u", false, "print unused vendored packages instead")
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		args = append(args, "./...")
	}
	args = gotool.ImportPaths(args)
	if *unused {
		orphan(args)
	} else {
		missing(args, *tests)
	}
}

func missing(args []string, tests bool) {
	var conf loader.Config
	conf.ParserMode = parser.ImportsOnly
	conf.AllowErrors = true
	conf.TypeChecker.Error = func(error) {}
	for _, p := range args {
		if strings.Index(p, "/vendor/") != -1 {
			// ignore vendored packages (if they are not imported by real ones)
			continue
		}
		if tests {
			conf.ImportWithTests(p)
		} else {
			conf.Import(p)
		}
	}
	prog, err := conf.Load()
	if err != nil {
		log.Fatal(err)
	}

	exitCode := 0
	for _, pi := range prog.AllPackages {
		if len(pi.Files) == 0 {
			continue // virtual stdlib package
		}

		filename := prog.Fset.File(pi.Files[0].Pos()).Name()
		if strings.HasPrefix(filename, build.Default.GOROOT) && isStandardImportPath(pi.Pkg.Path()) {
			continue
		}

		internal := false
		for _, ini := range prog.InitialPackages() {
			if ini == pi || strings.HasPrefix(pi.Pkg.Path(), ini.Pkg.Path()+"/") {
				internal = true
				break
			}
		}
		if internal {
			continue
		}

		if strings.Index(pi.Pkg.Path(), "/vendor/") == -1 {
			fmt.Println("[!] dependency not vendored:", pi.Pkg.Path())
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}

func orphan(args []string) {
	_, clients, _ := importgraph.Build(&build.Default)
	exitCode := 0
	for _, p := range args {
		p = absImportPath(p)
		if strings.Index(p, "/vendor/") == -1 {
			continue
		}
		if len(clients[p]) == 0 {
			fmt.Println("[!] unused vendored package:", p)
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}

func isStandardImportPath(path string) bool {
	// from https://github.com/golang/go/blob/87bca88/src/cmd/go/pkg.go#L183-L194
	i := strings.Index(path, "/")
	if i < 0 {
		i = len(path)
	}
	elem := path[:i]
	return !strings.Contains(elem, ".")
}

func absImportPath(path string) string {
	if !build.IsLocalImport(path) {
		return path
	}
	pwd, _ := os.Getwd()
	fullPath := filepath.Clean(filepath.Join(pwd, path))
	for _, root := range build.Default.SrcDirs() {
		if strings.HasPrefix(fullPath, root+string(filepath.Separator)) {
			rel, _ := filepath.Rel(root, fullPath)
			return filepath.ToSlash(rel)
		}
	}
	panic("can't run on . or ./... outside $GOPATH")
}
