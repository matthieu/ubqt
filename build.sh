#!/usr/bin/env sh

mkdir -p out
# 6g -o out/parsec.6 etc/go-parse/parsec.go
# 6g -o out/parser.6 -I out src/parser.go
# 6l -L out -o out/witty out/parser.6

# 6g -o out/ubqt.6 src/defs.go src/chunk.go
# 6g -o out/vm.6 -I out/ src/vm.go
# 6l -L out -o out/vm out/vm.6

6g -o out/ubqt.6 src/token.go src/parse.go
6g -I out -o out/repl.6 src/repl.go
6l -L out -o out/ubqt out/repl.6
