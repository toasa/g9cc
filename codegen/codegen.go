package codegen

import (
    . "g9cc/common"
    . "g9cc/util"
    . "g9cc/regalloc"
    "fmt"
)

// Code generator
// irv.data[]内の各ir(中間表現)に対し、ir.opから機械的にアセンブリを生成していく

var label_num int
func gen_label() string {
    var buf string
    buf = fmt.Sprintf(".L%d", label_num)
    //buf = fmt.Sprintf("Ltmp%d", label_num)
    label_num++
    return buf
}

func Gen_x86(irv *Vector) {

    Regs = [8]string{"rdi", "rsi", "r10", "r11", "r12", "r13", "r14", "r15"}

    var ret string = gen_label()
    fmt.Printf("    push rbp\n")
    fmt.Printf("    mov rbp, rsp\n")

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
            fmt.Printf("    jmp %s\n", ret)
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
    fmt.Printf("    mov rsp, rbp\n")
    fmt.Printf("    pop rbp\n")
    fmt.Printf("    ret\n")
}
