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
        echo "$expected expected, but got $actual"
        exit 1
    fi
}

try 0 0
try 46 46
try 14 ' 10 - 6 + 10 '
try 153 '1+2+3+4+5+6+7+8+9+10+11+12+13+14+15+16+17'
try 10 '2 * 3 + 4'
try 14 '2 + 3 * 4'
try 26 '2 * 3 + 4 * 5'
try 8 '64 / 8'
try 6 '3 * 4 / 2'

echo OK
