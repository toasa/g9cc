#!/bin/bash

try() {
    expect="$1"
    filename="$2"

    go run main.go -export-file "$filename" > tmp.s
    gcc -o tmp tmp.s

    ./tmp
    actual="$?"

    if [ "$actual" == "$expect" ]; then
        echo "$filename => $actual"
    else
        echo "$input: $expect expected, but got $actual"
        exit 1
    fi
}

try 46 './examples/test.c'
