#!/bin/bash

try() {
    expected="$1"
    input="$2"

    go run main.go "$input" > tmp.s
    # tmp.s tmp-plus.oをリンクし、実行ファイルtmpを作る
    gcc -o tmp tmp.s
    # gcc -o tmp tmp-plus.o tmp.o
    ./tmp
    actual="$?"

    if [ "$actual" == "$expected" ]; then
        echo "$input => $actual"
    else
        echo "$input: $expected expected, but got $actual"
        exit 1
    fi
}

# 入力ファイル名を"-"とすると、標準入力と解釈され、-xcオプションでc言語のソースコードと明示
# echo 'int plus(int x, int y) { return x + y; }' | gcc -dynamiclib -xc -o tmp-plus.dylib -

echo 'int plus(int x, int y) { return x + y; }' | gcc -o tmp-plus.o -c -xc -
ar rcs libstatic.a tmp-plus.o

try 10 'main() { return 2*3+4; }'
try 14 'main() { return 2+3*4; }'
try 26 'main() { return 2*3+4*5; } '
try 5 'main() { return 50/10; }'
try 153 'main() { return 1+2+3+4+5+6+7+8+9+10+11+12+13+14+15+16+17; }'

try 2 'main() { a=2; return a; }'
try 10 'main() { a=2; b=3+2; return a*b; }'
try 2 'main() { if (1) return 2; return 3; }'
try 3 'main() { if (0) return 2; return 3; }'
try 2 'main() { if (1) return 2; else return 3; }'
try 3 'main() { if (0) return 2; else return 3; }'
try 1 'one() { return 1; } main() { return one(); }'
try 3 'one() { return 1; } two() { return 2; } main() { return one() + two(); }'
# try 5 'main() { return plus(2, 3); }'

echo OK
