#!/bin/bash

try() {
    expected="$1"
    input="$2"

    go run main.go "$input" > tmp.s
    gcc -o tmp tmp.s
    ./tmp
    actual="$?"

    if [ "$actual" == "$expected" ]; then
        echo "$input => $actual"
    else
        echo "$input: $expected expected, but got $actual"
        exit 1
    fi
}

try 0 'return 0;'
try 46 'return 46;'
try 36 'return 12 + 34 - 10;'
try 153 'return 1+2+3+4+5+6+7+8+9+10+11+12+13+14+15+16+17;'
try 10 'return 2 * 3 + 4;'
try 14 'return 2 + 3 * 4;'
try 26 'return 2 * 3 + 4 * 5;'
try 8 'return 64 / 8;'
try 21 '1+2; return 5+20-4;'
try 6 'return 3 * 4 / 2;'
try 77 '(3+4) * (5 + 6);'

try 2 'a=2; return a;'
try 10 'a=2; b=3+2; return a*b;'
try 2 'if (1) return 2; return 3;'
try 3 'if (0) return 2; return 3;'

echo OK
