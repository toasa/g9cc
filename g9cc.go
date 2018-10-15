// アセンブリ言語の構文には, AT&TのものとIntelのものがある。
// レジスタ名に%がつくのがAT&T, そうでないものがIntel

package main

import (
    "fmt"
    "os"
    // "strconv"
)

const (
    TK_NUM = iota + 256
    TK_EOF
)

type Token struct {
    ty int // token type
    val int // number literal
    input string // token string
}

var tokens [100]Token

func tokenize(s string) {
    // index of tokens, input
    i_tokens := 0
    i_input := 0

    for s[i_input] != '\000' {

        // white space
        if Isspace(s[i_input]) {
            i_input++
            continue
        }

        // + or -
        if s[i_input] == '+' || s[i_input] == '-' {
            tokens[i_tokens].ty = int(s[i_input])
            tokens[i_tokens].input = string(s[i_input])
            i_input++
            i_tokens++
            continue
        }

        // number
        if Isdigit(s[i_input]) {
            var num int = int(s[i_input] - '0')
            i_input++
            for ; Isdigit(s[i_input]); i_input++ {
                num = num * 10 + int(s[i_input] - '0')
            }

            tokens[i_tokens].ty = TK_NUM
            tokens[i_tokens].input = string(num)
            tokens[i_tokens].val = num
            i_tokens++
            continue
        }

        fmt.Println("what's up guys")
        fmt.Printf("cannot tokenize: %s", s);
        os.Exit(1)
    }

    tokens[i_tokens].ty = TK_EOF
}

// An error reporting function.

func fail(i int) {
    fmt.Printf("unexpected token: %s\n", tokens[i].input)
    os.Exit(1)
}

func main() {
    if len(os.Args) != 2 {
        fmt.Println("Usage: g9cc <code>")

        return
    }

    // argv := []rune(os.Args[1])
    argv := os.Args[1] + "\000"

    tokenize(argv)

    // fmt.Println("	.section	__TEXT,__text,regular,pure_instructions")
    // fmt.Println("	.macosx_version_min 10, 12")
    fmt.Println("	.globl	_main")
    // fmt.Println("	.p2align	4, 0x90")
    fmt.Println("_main:                                  ## @main")
    // fmt.Println(".cfi_startproc")
    // fmt.Println("## BB#0:")
    // fmt.Println("	pushq	%rbp")
    // fmt.Println("Ltmp0:")
    // fmt.Println("	.cfi_def_cfa_offset 16")
    // fmt.Println("Ltmp1:")
    // fmt.Println("	.cfi_offset %rbp, -16")
    // fmt.Println("	movq	%rsp, %rbp")
    // fmt.Println("Ltmp2:")
    // fmt.Println("	.cfi_def_cfa_register %rbp")

    if (tokens[0].ty != TK_NUM) {
        fail(0)
    }
    fmt.Printf("	movl	$%d, %%eax    ## imm = 0x4B0\n", tokens[0].val)

    i := 1
    for tokens[i].ty != TK_EOF {

        if tokens[i].ty == '+' {
            i++
            if tokens[i].ty != TK_NUM {
                fmt.Println("hom")
            }
            fmt.Printf("	add	   $%d, %%eax\n", tokens[i].val);
            i++
            continue
        }

        if tokens[i].ty == '-' {
            i++
            if tokens[i].ty != TK_NUM {
                fmt.Println("pom")
            }
            fmt.Printf("	sub	   $%d, %%eax\n", tokens[i].val);
            i++
            continue
        }

        fmt.Println("dom")
        fail(i)
    }

    // fmt.Println("	movl	$0, -4(%rbp)")
    // fmt.Println("	popq	%rbp")
    fmt.Println("	retq")
    // fmt.Println("	.cfi_endproc")
    fmt.Printf("\n")
    // fmt.Println(".subsections_via_symbols")

}

func Isdigit(c uint8) bool {
    if '0' <= c && c <= '9' {
        return true
    } else {
        return false
    }
}

func Isspace(c uint8) bool {
    if c == ' ' {
        return true
    } else {
        return false
    }
}
