package main

import (
	"flag"
	"fmt"
	"go/build"
	"go/parser"
	"log"
	"os"
	"strings"

	"github.com/kisielk/gotool"

	"golang.org/x/tools/go/loader"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: vendorcheck [-t] package [package ...]")
		flag.PrintDefaults()
	}
	tests := flag.Bool("t", false, "also check dependencies of test files")
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		args = append(args, "./...")
	}

	var conf loader.Config
	conf.ParserMode = parser.ImportsOnly
	conf.AllowErrors = true
	conf.TypeChecker.Error = func(error) {}
	for _, p := range gotool.ImportPaths(args) {
		if strings.Index(p, "/vendor/") != -1 {
			// ignore vendored packages (if they are not imported by real ones)
			continue
		}
		if *tests {
			conf.ImportWithTests(p)
		} else {
			conf.Import(p)
		}
	}
	prog, err := conf.Load()
	if err != nil {
		log.Fatal(err)
	}

	initial := make(map[*loader.PackageInfo]bool)
	for _, pi := range prog.InitialPackages() {
		initial[pi] = true
	}

	var packages []*loader.PackageInfo
	for _, pi := range prog.AllPackages {
		if initial[pi] {
			continue
		}
		if len(pi.Files) == 0 {
			continue // virtual stdlib package
		}
		filename := prog.Fset.File(pi.Files[0].Pos()).Name()
		if !strings.HasPrefix(filename, build.Default.GOROOT) || !isStandardImportPath(pi.Pkg.Path()) {
			packages = append(packages, pi)
		}
	}

	exitCode := 0
	for _, pi := range packages {
		if strings.Index(pi.Pkg.Path(), "/vendor/") == -1 {
			fmt.Println("[!] dependency not vendored:", pi.Pkg.Path())
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
