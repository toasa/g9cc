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

// tokenizer
type Token struct {
    ty int // token type
    val int // number literal
    input string // token string
}

// tokenized済のtokenをこの配列に格納する
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

// Recursive-descent parser.

// "tokens" array's index
var pos int = 0

const (
    ND_NUM = 256
)

type Node struct {
    ty int // node type
    lhs *Node // left-hand side
    rhs *Node // right-hand side
    val int // number literal
}

func new_node(op int, lhs *Node, rhs *Node) *Node {
    n := new(Node) // new()関数でNode型のポインタを返す
    n.ty = op
    n.lhs = lhs
    n.rhs = rhs

    return n
}

func new_node_num(val int) *Node {
    n := new(Node)
    n.ty = TK_NUM
    n.val = val
    return n
}

// An error reporting function.
func fail(i int) {
    fmt.Printf("unexpected token: %s\n", tokens[i].input)
    os.Exit(1)
}

func error(msgs ...string) {
    for _, msg := range msgs {
        fmt.Println(msg)
    }
    os.Exit(1)
}

func number() *Node {
    if tokens[pos].ty == TK_NUM {
        n := new_node_num(tokens[pos].val)
        pos++
        return n;
    }

    error(fmt.Sprintf("number expected, but got %s", tokens[pos].input))
    return nil
}

// expression
func expr() *Node {
    var lhs *Node = number()

    // この文はすごい
    // 2 + 3 - 4から以下の構文木を作成した
    //
    //       ---(-)---
    //       |       |
    //    --(+)--    4
    //    |     |
    //    2     3

    for true {
        op := tokens[pos].ty
        if !(op == '+' || op == '-') {
            break
        }
        pos++
        lhs = new_node(op, lhs, number())
    }

    if tokens[pos].ty != TK_EOF {
        error(fmt.Sprintf("stray token: %s", tokens[pos].input))
    }
    
    return lhs
}

// regs(registers array)'s index
var cur int;

func gen(node *Node) string {

    // go-langで配列orスライスをグローバル変数として定義(var regs []string = {...}のように)できないのか？
    regs := []string{"rdi", "rsi", "r10", "r11", "r12", "r13", "r14", "r15", "NULL"}

    if node.ty == ND_NUM {
        reg := regs[cur]
        cur++
        if reg == "NULL" {
            error("register exhausted")
        }
        // 汎用レジスタregにnode.valを代入
        fmt.Printf("	mov %s, %d\n", reg, node.val)
        return reg
    }

    // lhs, rhsの値がそれぞれレジスタ(string型)dst, srcに格納されている
    // destination, source
    dst := gen(node.lhs)
    src := gen(node.rhs)

    switch node.ty {
    case '+':
        fmt.Printf("	add %s, %s\n", dst, src)
        return dst
    case '-':
        fmt.Printf("	sub %s, %s\n", dst, src)
        return dst
    default:
        error("unknown operator")
    }

    return "NULL"
}

func main() {
    if len(os.Args) != 2 {
        fmt.Println("Usage: g9cc <code>")

        return
    }

    // 標準入力からの文字列に終端文字を追加する. parseをかんたんにするため
    tokenize(os.Args[1] + "\000")

    // print tokens
    // for _, t := range tokens {
    //     if t.ty == TK_EOF {
    //         fmt.Printf("EOF\n\n")
    //         break
    //     }
    //     fmt.Printf("%+v\n", t)
    // }

    var node *Node = expr()

    // fmt.Println("	.section	__TEXT,__text,regular,pure_instructions")
    // fmt.Println("	.macosx_version_min 10, 12")
    fmt.Println("	.intel_syntax noprefix")
    fmt.Println("	.globl	_main")
    fmt.Println("_main:")

    fmt.Printf("	mov rax, %s\n", gen(node))
    fmt.Printf("	ret\n")
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
