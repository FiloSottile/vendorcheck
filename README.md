# vendorcheck
Check that all your Go dependencies are properly vendored

```
$ vendorcheck ./...
[!] dependency not vendored: golang.org/x/tools/go/buildutil
[!] dependency not vendored: github.com/kisielk/gotool
[!] dependency not vendored: golang.org/x/tools/go/loader
[!] dependency not vendored: golang.org/x/tools/go/ast/astutil
```

Run `vendorcheck -u` to list unused vendored packages instead.
