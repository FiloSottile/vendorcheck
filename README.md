# vendorcheck

**Deprecated**: use [`go mod vendor`](https://golang.org/ref/mod#vendoring),
which [starting in Go 1.14](https://golang.org/doc/go1.14#go-command) will
automatically check that the vendor folder is complete.

Check that all your Go dependencies are properly vendored

```
$ vendorcheck ./...
[!] dependency not vendored: golang.org/x/tools/go/buildutil
[!] dependency not vendored: github.com/kisielk/gotool
[!] dependency not vendored: golang.org/x/tools/go/loader
[!] dependency not vendored: golang.org/x/tools/go/ast/astutil
```

Run `vendorcheck -u` to list unused vendored packages instead.
