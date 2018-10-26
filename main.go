// 大まかな流れ
//
// input from command line
// |
// |   Tokenize() in token.go
// v
// tokens *Vector (tokens.dataは*Token型の配列)
// |
// |   Parse() in parse.go
// v
// node *Node: 構文木のroot node (root nodeさえわかれば、他のnodeへは辿って行けるので*Vectorによるwrapは不要)
// |
// |   Gen_ir() in ir.go
// v
// irv *Vector (irv.dataは*IR型の配列)
// |
// |   Alloc_regs() in regalloc.go
// v
// irv : 中間表現の完成（一つ前のirvはレジスタ割当に無駄がある）
// |
// |   Gen_x86() in codegen.go
// v
// assembly code

package main

import (
    "fmt"
    "os"
    . "g9cc/common"
    . "g9cc/util"
    "g9cc/token"
    "g9cc/parse"
    "g9cc/ir"
    "g9cc/regalloc"
    "g9cc/codegen"
)

func main() {

    if os.Args[1] == "-test" {
        Util_test()
        return
    }

    var input string
    var dump_ir1 bool
    var dump_ir2 bool

    var argc int = len(os.Args)

    if argc == 3 && (os.Args[1] == "-dump-ir1") {
        dump_ir1 = true
        input = os.Args[2]
    } else if (argc == 3 && (os.Args[1] == "-dump-ir2")) {
        dump_ir2 = true
        input = os.Args[2]
    } else {
        if argc != 2 {
            Error("Usage: g9cc [-test] [-dump_ir] <code>")
        }
        input = os.Args[1]
    }


    // 標準入力からの文字列に終端文字を追加する. parseをかんたんにするため
    var tokens *Vector = token.Tokenize(input + "\000")
    // PrintVector(tokens)

    var node *Node = parse.Parse(tokens)
    // PrintAST(node)

    var irv *Vector = ir.Gen_ir(node)
    if dump_ir1 {
        ir.Dump_ir(irv)
    }

    // PrintVector(irv)
    regalloc.Alloc_regs(irv)
    if dump_ir2 {
        ir.Dump_ir(irv)
    }
    // PrintVector(irv)

    fmt.Println("    .intel_syntax noprefix")
    fmt.Println("    .globl _main") // ここを".global main",
    fmt.Println("_main:") // "main:"と書くと, ruiさんバージョンになる

    codegen.Gen_x86(irv)
}
