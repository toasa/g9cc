package codegen

import (
    . "g9cc/common"
    . "g9cc/regalloc"
    "fmt"
)

// Code generator
// irv.data[]内の各ir(中間表現)に対し、ir.opから機械的にアセンブリを生成していく
func Gen_x86(irv *Vector) {

    Regs = [8]string{"rdi", "rsi", "r10", "r11", "r12", "r13", "r14", "r15"}

    for i := 0; i < irv.Len; i++ {
        ir, ok := irv.Data[i].(*IR)
        if !ok {
            Error("Not *IR type is in irv.data[]")
        }

        switch ir.Op {
        case IR_IMM:
            fmt.Printf("    mov %s, %d\n", Regs[ir.Lhs], ir.Rhs)
        case IR_MOV:
            fmt.Printf("    mov %s, %s\n", Regs[ir.Lhs], Regs[ir.Rhs])
        case IR_RETURN: // lhsに格納された値をアキュムレータに渡し、戻り値とする
            fmt.Printf("    mov rax, %s\n", Regs[ir.Lhs])
            fmt.Printf("    ret\n")
        case '+':
            fmt.Printf("    add %s, %s\n", Regs[ir.Lhs], Regs[ir.Rhs])
        case '-':
            fmt.Printf("    sub %s, %s\n", Regs[ir.Lhs], Regs[ir.Rhs])
        case IR_NOP:

        default:
            Error("unknown operator")
        }
    }
}
