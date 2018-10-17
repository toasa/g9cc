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

// Tokenizer
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

// Intermediate representation
const (
    IR_IMM = iota // immediate value
    IR_MOV
    IR_RETURN
    IR_KILL
    IR_NOP
)

type IR struct {
    op int
    lhs int
    rhs int
}

func new_ir(op int, lhs int, rhs int) *IR {
    var ir *IR = new(IR)
    ir.op = op
    ir.lhs = lhs
    ir.rhs = rhs
    return ir
}

// 中間表現を格納する配列
var ins [1000]*IR
// index of ins
var inp int
// register number
var regno int

// regs(registers array)'s index
var cur int;

func assert(b bool) {
    if !b {
        error("assert error")
    }
}

func gen_ir_sub(node *Node) int {

    if node.ty == ND_NUM {
        var r int = regno
        regno++
        ins[inp] = new_ir(IR_IMM, r, node.val)
        inp++
        return r;
    }

    if !(node.ty == '+' || node.ty == '-') {
        error("operator expected")
    }

    var lhs int = gen_ir_sub(node.lhs)
    var rhs int = gen_ir_sub(node.rhs)

    ins[inp] = new_ir(node.ty, lhs, rhs)
    inp++
    ins[inp] = new_ir(IR_KILL, rhs, 0)
    inp++

    return lhs
}

// ASTを引数にとり、中間表現を返す
// irは {op: , lhs: , rhs : }からなる
// op = numのとき => {IR_IMM, register_index, num_value}
// op = '+'のとき => {'+', lhsの値が格納されたregisterのindex, rhsの値が格納されたregisterのindex}
// opが'+'or'-'の直後 => {IR_KILL, rhsの値が格納されたregisterのindex, 0}
// ここで決めたregisterのindexは確定ではなく, alloc_regs()で配列insを一つひとつ読みながら
// 最終的なregister を決定する
func gen_ir(node *Node) {
    var r int = gen_ir_sub(node)
    ins[inp] = new_ir(IR_RETURN, r, 0)
    inp++
}

// Register allocator
var regs [8]string

var used [8]bool

var reg_map[1000]int

func alloc(ir_reg int) int {
    if reg_map[ir_reg] != -1 {
        var r int = reg_map[ir_reg]
        assert(used[r])
        return r
    }

    for i := 0; i < len(regs); i++ {
        if used[i] {
            continue
        }
        used[i] = true
        reg_map[ir_reg] = i
        return i
    }

    error("register exhausted")
    return 0 // ここには到達しないため(intを返さないと怒るコンパイラを鎮める他に)イミなし
}

func kill(r int) {
    assert(used[r])
    used[r] = false
}

// 中間表現の命令配列insの各要素に対し、必要ならレジスタを割り当てていく
func alloc_regs() {
    for i := 0; i < inp; i++ {
        var ir *IR = ins[i]

        switch ir.op {
        case IR_IMM:
            ir.lhs = alloc(ir.lhs)
        case IR_MOV, '+', '-':
            ir.lhs = alloc(ir.lhs)
            ir.rhs = alloc(ir.rhs)
        case IR_RETURN:
            kill(reg_map[ir.lhs])
        case IR_KILL:
            kill(reg_map[ir.lhs])
            ir.op = IR_NOP
        default:
            error("unknown operator")
        }
    }
}

// Code generator
// ins内の各irに対し、ir.opから機械的にアセンブリを生成していく
func gen_x86() {

    regs = [8]string{"rdi", "rsi", "r10", "r11", "r12", "r13", "r14", "r15"}

    for i := 0; i < inp; i++ {
        var ir *IR = ins[i]

        switch ir.op {
        case IR_IMM:
            fmt.Printf("    mov %s, %d\n", regs[ir.lhs], ir.rhs)
        case IR_MOV:
            fmt.Printf("    mov %s, %s\n", regs[ir.lhs], regs[ir.rhs])
        case IR_RETURN:
            fmt.Printf("    mov rax, %s\n", regs[ir.lhs])
            fmt.Printf("    ret\n")
        case '+':
            fmt.Printf("    add %s, %s\n", regs[ir.lhs], regs[ir.rhs])
        case '-':
            fmt.Printf("    sub %s, %s\n", regs[ir.lhs], regs[ir.rhs])
        case IR_NOP:

        default:
            error("unknown operator")
        }
    }
}

func main() {
    if len(os.Args) != 2 {
        fmt.Println("Usage: g9cc <code>")

        return
    }

    for i := 0; i < len(reg_map); i++ {
        reg_map[i] = -1
    }

    // 標準入力からの文字列に終端文字を追加する. parseをかんたんにするため
    tokenize(os.Args[1] + "\000")

    var node *Node = expr()

    gen_ir(node)

    // printIns()
    //
    // fmt.Println("==================================")

    alloc_regs()

    // printIns()

    // fmt.Println("	.section	__TEXT,__text,regular,pure_instructions")
    // fmt.Println("	.macosx_version_min 10, 12")
    fmt.Println("    .intel_syntax noprefix")
    fmt.Println("    .globl _main") // ここを".global main",
    fmt.Println("_main:") // "main:"と書くと, ruiさんバージョンになる

    gen_x86()
}

func printTokens() {
    for _, t := range tokens {
        if t.ty == TK_EOF {
            fmt.Printf("EOF\n\n")
            break
        }
        fmt.Printf("%+v\n", t)
    }
}

func printIns() {
    for i := 0; i < inp; i++ {
        fmt.Printf("%+v\n", ins[i])
    }
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
