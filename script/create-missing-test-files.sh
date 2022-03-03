find . -name "*.go" | grep -v "_test.go"| sed 's/.go//g'|xargs -I{} sh -c "echo {}; if ! test -f '{}_test.go'; then sed '/^package/q' {}.go > {}_test.go; fi"
