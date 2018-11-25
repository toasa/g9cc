package main

import (
    "os"
    . "g9cc/common"
    . "g9cc/util"
    "g9cc/token"
    "g9cc/parse"
    "g9cc/sema"
    "g9cc/irdump"
    "g9cc/gen_ir"
    "g9cc/regalloc"
    "g9cc/gen_x86"
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
            Error("Usage: g9cc [-test] [-dump-ir1] [-dump-ir2] <code>")
        }
        input = os.Args[1]
    }


    // 標準入力からの文字列に終端文字を追加する. parseをかんたんにするため
    var tokens *Vector = token.Tokenize(input + "\000")
    // PrintVector(tokens)

    var nodes *Vector = parse.Parse(tokens)
    // PrintVector(node)

    var globals *Vector = sema.Sema(nodes)
    var fns *Vector = gen_ir.Gen_ir(nodes)

    if dump_ir1 {
        irdump.Dump_ir(fns)
    }
    //PrintVector(fns)

    regalloc.Alloc_regs(fns)

    if dump_ir2 {
        irdump.Dump_ir(fns)
    }
    // PrintVector(irv)

    codegen.Gen_x86(globals, fns)
}
