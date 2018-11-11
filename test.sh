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

try 33 'int main() { return 33; }'
try 10 'int main() { return 2*3+4; }'
try 14 'int main() { return 2+3*4; }'
try 26 'int main() { return 2*3+4*5; } '
try 5 'int main() { return 50/10; }'
try 153 'int main() { return 1+2+3+4+5+6+7+8+9+10+11+12+13+14+15+16+17; }'

try 2 'int main() { int a=2; return a; }'
try 10 'int main() { int a=2; int b=3+2; return a*b; }'
try 2 'int main() { if (1) return 2; return 3; }'
try 3 'int main() { if (0) return 2; return 3; }'
try 2 'int main() { if (1) return 2; else return 3; }'
try 3 'int main() { if (0) return 2; else return 3; }'
try 1 'int one() { return 1; } int main() { return one(); }'
try 3 'int one() { return 1; } int two() { return 2; } int main() { return one() + two(); }'
try 5 'int plus(int x, int y) { return x + y; } int main() { return plus(2, 3); }'
try 6 'int mul(int a, int b) { return a * b; } int main() { return mul(2, 3); }'
try 21 'int add(int a, int b, int c, int d, int e, int f) { return a+b+c+d+e+f; } int main() { return add(1,2,3,4,5,6); }'

try 0 'int main() { return 0||0; }'
try 1 'int main() { return 1||0; }'
try 1 'int main() { return 0||1; }'
try 1 'int main() { return 1||1; }'

try 0 'int main() { return 0&&0; }'
try 0 'int main() { return 1&&0; }'
try 0 'int main() { return 0&&1; }'
try 1 'int main() { return 1&&1; }'

try 0 'int main() { return 0<0; }'
try 0 'int main() { return 1<0; }'
try 1 'int main() { return 0<1; }'
try 0 'int main() { return 1<1; }'
try 0 'int main() { return 0>0; }'
try 1 'int main() { return 1>0; }'
try 0 'int main() { return 0>1; }'
try 0 'int main() { return 1>1; }'

try 60 'int main() {int sum=0; int i; for (i=10; i<15; i=i+1) sum = sum + i; return sum; }'
try 89 'int main() {int i=1; int j=1; for (int k=0; k<10; k=k+1) {int m=i+j; i=j; j=m; } return i; }'

try 9 'int main() { int *p = alloc1(4, 5); return *p + *(p + 1); }'
try 15 'int main() { int *p = alloc2(10, 5); return *(p - 1) + *p; }'
try 46 'int main() { int **p = alloc_ptr_ptr(46); return **p; }'

try 3 'int main() { int ary[2]; *ary=1; *(ary+1)=2; return *ary + *(ary+1); }'
try 5 'int main() { int x; int *p = &x; x = 5; return *p; }'

try 3 'int main() { int ary[2]; ary[0]=1; ary[1]=2; return ary[0] + ary[1]; }'
try 5 'int main() { int x; int *p = &x; x = 5; return p[0]; }'

try 1 'int main() { char x; return sizeof x; }'
try 4 'int main() { int x; return sizeof(x); }'
try 8 'int main() { int *x; return sizeof(x); }'
try 16 'int main() { int x[4]; return sizeof(x); }'

try 5 'int main() { char x = 5; return x; }'
try 42 'int main() { int x = 0; char *p = &x; p[0] = 42; return x; }'

echo OK
