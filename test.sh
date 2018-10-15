#!/bin/bash

try() {
    expected="$1"
    input="$2"

    go run g9cc.go "$input" > tmp.s
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
try 24 ' 20 - 6 + 10'

echo OK
