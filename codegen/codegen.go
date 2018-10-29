package codegen

import (
    . "g9cc/common"
    . "g9cc/util"
    . "g9cc/regs"
    "fmt"
)

// Code generator
// irv.data[]内の各ir(中間表現)に対し、ir.opから機械的にアセンブリを生成していく

var label int

func gen(fn *Function) {

    var ret string = fmt.Sprintf(".Lend%d", label)
    label++

    fmt.Printf(".globl _%s\n", fn.Name)
    fmt.Printf("_%s:\n", fn.Name)
    fmt.Printf("    push r12\n")
    fmt.Printf("    push r13\n")
    fmt.Printf("    push r14\n")
    fmt.Printf("    push r15\n")
    fmt.Printf("    push rbp\n")
    fmt.Printf("    mov rbp, rsp\n")

    for i := 0; i < fn.Ir.Len; i++ {
        ir := fn.Ir.Data[i].(*IR)

        switch ir.Op {
        case IR_IMM:
            fmt.Printf("    mov %s, %d\n", Regs[ir.Lhs], ir.Rhs)
        case IR_ADD_IMM:
            fmt.Printf("    add %s, %d\n", Regs[ir.Lhs], ir.Rhs)
        case IR_MOV:
            fmt.Printf("    mov %s, %s\n", Regs[ir.Lhs], Regs[ir.Rhs])
        case IR_RETURN: // lhsに格納された値をアキュムレータに渡し、戻り値とする
            fmt.Printf("    mov rax, %s\n", Regs[ir.Lhs])
            fmt.Printf("    jmp %s\n", ret)
        case IR_CALL:

            var arg []string = []string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}

            for i := 0; i < ir.Nargs; i++ {
                fmt.Printf("    mov %s, %s\n", arg[i], Regs[ir.Args[i]])
            }

            fmt.Printf("    push r10\n")
            fmt.Printf("    push r11\n")
            fmt.Printf("    mov rax, 0\n")
            // 関数名の先頭に"_"が必要
            fmt.Printf("    call _%s\n", ir.Name)
            fmt.Printf("    pop r11\n")
            fmt.Printf("    pop r10\n")

            fmt.Printf("    mov %s, rax\n", Regs[ir.Lhs])
        case IR_LABEL:
            fmt.Printf(".L%d:\n", ir.Lhs)
        case IR_JMP:
            fmt.Printf("    jmp .L%d\n", ir.Lhs)
        case IR_UNLESS:
            // 今の所, lhsの(レジスタの)値が0ならラベルに飛ぶ
            fmt.Printf("    cmp %s, 0\n", Regs[ir.Lhs])
            fmt.Printf("    je .L%d\n", ir.Rhs)
        case IR_ALLOCA:
            if Int2bool(ir.Rhs) {
                fmt.Printf("    sub rsp, %d\n", ir.Rhs)
            }
            fmt.Printf("    mov %s, rsp\n", Regs[ir.Lhs])
        case IR_LOAD:
            fmt.Printf("    mov %s, [%s]\n", Regs[ir.Lhs], Regs[ir.Rhs])
        case IR_STORE:
            fmt.Printf("    mov [%s], %s\n", Regs[ir.Lhs], Regs[ir.Rhs])
        case '+':
            fmt.Printf("    add %s, %s\n", Regs[ir.Lhs], Regs[ir.Rhs])
        case '-':
            fmt.Printf("    sub %s, %s\n", Regs[ir.Lhs], Regs[ir.Rhs])
        case '*':
            fmt.Printf("    mov rax, %s\n", Regs[ir.Rhs])
            // mul reg: 予めrax(アキュムレータ)に格納された値と
            //          regに格納された値の掛け算を行い,結果をraxに格納する
            fmt.Printf("    mul %s\n", Regs[ir.Lhs])
            fmt.Printf("    mov %s, rax\n", Regs[ir.Lhs])
        case '/':
            // raxに左辺値を代入
            fmt.Printf("    mov rax, %s\n", Regs[ir.Lhs])
            // convert quadword to octaword: 符号拡張(アキュムレータを拡張する)
            // wordは4byteのこと? => ocは32byte?
            fmt.Printf("    cqo\n")
            fmt.Printf("    div %s\n", Regs[ir.Rhs])
            fmt.Printf("    mov %s, rax\n", Regs[ir.Lhs])
        case IR_NOP:

        default:
            Error("unknown operator")
        }
    }

    fmt.Printf("%s:\n", ret)
    fmt.Printf("    mov rsp, rbp\n")
    fmt.Printf("    pop rbp\n")
    fmt.Printf("    pop r15\n")
    fmt.Printf("    pop r14\n")
    fmt.Printf("    pop r13\n")
    fmt.Printf("    pop r12\n")
    fmt.Printf("    ret\n")
}

func Gen_x86(fns *Vector) {
    fmt.Printf(".intel_syntax noprefix\n")

    for i := 0; i < fns.Len; i++ {
        fn := fns.Data[i].(*Function)
        gen(fn)
    }
}
