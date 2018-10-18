// アセンブリ言語の構文には, AT&TのものとIntelのものがある。
// レジスタ名に%がつくのがAT&T, そうでないものがIntel

// 大まかな流れ
//
// input from command line
// |
// |   tokenize()
// v
// tokens *Vector (tokens.dataは*Token型の配列)
// |
// |   expr()
// v
// node *Node: 構文木のroot node (root nodeさえわかれば、他のnodeへは辿って行けるので*Vectorによるwrapは不要)
// |
// |   gen_ir()
// v
// irv *Vector (irv.dataは*IR型の配列)
// |
// |   alloc_regs()
// v
// irv : 中間表現の完成（一つ前のirvはレジスタ割当に無駄がある）
// |
// |   gen_x86()
// v
// assembly code

package main

import (
    "fmt"
    "os"
    // "strconv"
)

// Vector
// 異なるデータ型*Token, *IRなどをスライスとして扱うための構造体(wrapperみたいなものか?)
type Vector struct {
    data []interface{}
    capacity int
    len int
}

func new_vec() *Vector {
    var v *Vector = new(Vector)
    v.capacity = 16
    v.len = 0
    v.data = make([]interface{}, v.capacity)
    return v
}

func vec_push(v *Vector, elem interface{}) {
    if v.len == v.capacity {
        v.capacity *= 2
        // v.dataの容量を増やすための処理
        for i := 0; i < v.capacity; i++ {
            var a interface{}
            v.data = append(v.data, a)
        }
    }
    v.data[v.len] = elem
    v.len++
}

const (
    TK_NUM = iota + 256 // Number Literal
    TK_EOF
)

// Tokenizer
type Token struct {
    ty int // token type
    val int // number literal
    input string // token string
}

func add_token(v *Vector, ty int, input string) *Token {
    t := new(Token)
    t.ty = ty
    t.input = input
    vec_push(v, t)
    return t
}

func tokenize(s string) *Vector {
    var v *Vector = new_vec()

    // index of input
    i_input := 0

    for s[i_input] != '\000' {

        // white space
        if Isspace(s[i_input]) {
            i_input++
            continue
        }

        // + or -
        if s[i_input] == '+' || s[i_input] == '-' {
            add_token(v, int(s[i_input]), string(s[i_input]))
            i_input++
            continue
        }

        // number
        if Isdigit(s[i_input]) {
            var num int = int(s[i_input] - '0')
            i_input++
            for ; Isdigit(s[i_input]); i_input++ {
                num = num * 10 + int(s[i_input] - '0')
            }

            var t *Token = add_token(v, TK_NUM, string(num))

            t.val = num
            continue
        }

        fmt.Println("what's up guys")
        fmt.Printf("cannot tokenize: %s", s);
        os.Exit(1)
    }


    add_token(v, TK_EOF, s);
    return v
}

// Recursive-descent parser.

const (
    ND_NUM = 256
)

type Node struct {
    ty int // node type
    lhs *Node // left-hand side
    rhs *Node // right-hand side
    val int // number literal
}

var tokens *Vector
// "tokens" array's index
var pos int = 0

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
// func fail(i int) {
//     fmt.Printf("unexpected token: %s\n", tokens[i].input)
//     os.Exit(1)
// }

func error(msgs ...string) {
    for _, msg := range msgs {
        fmt.Println(msg)
    }
    os.Exit(1)
}

func assert(b bool, msg string) {
    if !b {
        error(msg)
    }
}

func number() *Node {
    t, ok := tokens.data[pos].(*Token)
    if !ok {
        error("Not *Token type is in tokens.data[]")
    }

    if t.ty != TK_NUM {
        error(fmt.Sprintf("number expected, but got %s", t.input))
    }
    pos++

    return new_node_num(t.val)
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
        t, ok := tokens.data[pos].(*Token)
        if !ok {
            error("Not *Token type is in tokens.data[]")
        }

        op := t.ty
        if !(op == '+' || op == '-') {
            break
        }
        pos++
        lhs = new_node(op, lhs, number())
    }

    t, ok := tokens.data[pos].(*Token)
    if !ok {
        error("Not *Token type is in tokens.data[]")
    }

    if t.ty != TK_EOF {
        error(fmt.Sprintf("stray token: %s", t.input))
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


var regno int

func gen_ir_sub(v *Vector, node *Node) int {
    // var regno int

    if node.ty == ND_NUM {
        var r int = regno
        regno++
        vec_push(v, new_ir(IR_IMM, r, node.val))
        return r;
    }

    assert((node.ty == '+' || node.ty == '-'), "operator expected")

    var lhs int = gen_ir_sub(v, node.lhs)
    var rhs int = gen_ir_sub(v, node.rhs)

    vec_push(v, new_ir(node.ty, lhs, rhs))
    vec_push(v, new_ir(IR_KILL, rhs, 0))

    return lhs
}

// ASTを引数にとり、中間表現を返す
// irは {op: , lhs: , rhs : }からなる
// op = numのとき => {IR_IMM, register_index, num_value}
// op = '+'のとき => {'+', lhsの値が格納されたregisterのindex, rhsの値が格納されたregisterのindex}
// opが'+'or'-'の直後 => {IR_KILL, rhsの値が格納されたregisterのindex, 0}
// ここで決めたregisterのindexは確定ではなく, alloc_regs()で配列insを一つひとつ読みながら
// 最終的なregister を決定する
func gen_ir(node *Node) *Vector{
    var v *Vector = new_vec()
    var r int = gen_ir_sub(v, node)
    vec_push(v, new_ir(IR_RETURN, r, 0))
    return v
}

// Register allocator
var regs [8]string
var used [len(regs)]bool

// IRの命令数分の要素をもつ配列(alloc_regs()で初期化)
var reg_map []int

func alloc(ir_reg int) int {
    // 演算子の場合こちらが走る
    if reg_map[ir_reg] != -1 {
        var r int = reg_map[ir_reg]
        assert(used[r], "allocation error")
        return r
    }

    // i はレジスタの配列regsのindex
    // 数値の場合こちらが走る
    for i := 0; i < len(regs); i++ {
        // index i のレジスタが使用済みの場合
        if used[i] {
            continue
        }
        // index i のレジスタが未使用の場合
        used[i] = true
        reg_map[ir_reg] = i // registerへのmapping
        return i
    }

    error("register exhausted")
    return 0 // ここには到達しないため(intを返さないと怒るコンパイラを鎮める他に)イミなし
}

func kill(r int) {
    assert(used[r], "kill error")
    used[r] = false
}

// 中間表現の命令配列insの各要素に対し、必要ならレジスタを割り当てていく
func alloc_regs(irv *Vector) {

    for i := 0; i < irv.len; i++ {
        reg_map = append(reg_map, -1)
    }

    for i := 0; i < irv.len; i++ {
        ir, ok := irv.data[i].(*IR)
        if !ok {
            error("Not *IR type is in irv.data[]")
        }

        switch ir.op {
        case IR_IMM:
            // 数値のとき格納先のレジスタのindexを調整する(->数値の値自体(rhs)はいじらない)
            ir.lhs = alloc(ir.lhs)
        case IR_MOV, '+', '-':
            ir.lhs = alloc(ir.lhs)
            ir.rhs = alloc(ir.rhs)
        case IR_RETURN:
            kill(reg_map[ir.lhs])
        case IR_KILL:
            // レジスタに格納された即値で、不要になったときに、そのレジスタを開放する操作
            kill(reg_map[ir.lhs])
            ir.op = IR_NOP
        default:
            error("unknown operator")
        }
    }
}

// Code generator
// irv.data[]内の各ir(中間表現)に対し、ir.opから機械的にアセンブリを生成していく
func gen_x86(irv *Vector) {

    regs = [8]string{"rdi", "rsi", "r10", "r11", "r12", "r13", "r14", "r15"}

    for i := 0; i < irv.len; i++ {
        ir, ok := irv.data[i].(*IR)
        if !ok {
            error("Not *IR type is in irv.data[]")
        }

        switch ir.op {
        case IR_IMM:
            fmt.Printf("    mov %s, %d\n", regs[ir.lhs], ir.rhs)
        case IR_MOV:
            fmt.Printf("    mov %s, %s\n", regs[ir.lhs], regs[ir.rhs])
        case IR_RETURN: // lhsに格納された値をアキュムレータに渡し、戻り値とする
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

    // 標準入力からの文字列に終端文字を追加する. parseをかんたんにするため
    tokens = tokenize(os.Args[1] + "\000")

    var node *Node = expr()

    var irv *Vector = gen_ir(node)

    // printVector(irv)
    alloc_regs(irv)
    // printVector(irv)

    // fmt.Println("	.section	__TEXT,__text,regular,pure_instructions")
    // fmt.Println("	.macosx_version_min 10, 12")
    fmt.Println("    .intel_syntax noprefix")
    fmt.Println("    .globl _main") // ここを".global main",
    fmt.Println("_main:") // "main:"と書くと, ruiさんバージョンになる

    gen_x86(irv)
}

func printVector(v *Vector) {
    switch v.data[0].(type) {
    case *Token:
        for i := 0; i < v.len; i++ {
            fmt.Printf("%+v\n", v.data[i])
        }
        fmt.Printf("=== END OF PRINT TOKEN ===\n\n")

    case *IR:
        for i := 0; i < v.len; i++ {
            fmt.Printf("%+v\n", v.data[i])
        }
        fmt.Printf("=== END OF PRINT IR ===\n\n")
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
