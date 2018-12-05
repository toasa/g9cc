package irdump

import (
    . "g9cc/common"
    . "g9cc/util"
    "fmt"
    "os"
)

// 中の要素の順序はcommon.go内のirのconstと一致させる
var Irinfo_arr []IRInfo = []IRInfo{
    // name, ty
    {"ADD", IR_TY_REG_REG},
    {"ADD", IR_TY_REG_IMM},
    {"SUB", IR_TY_REG_REG},
    {"SUB", IR_TY_REG_IMM},
    {"MUL", IR_TY_REG_REG},
    {"MUL", IR_TY_REG_IMM},
    {"DIV", IR_TY_REG_REG},
    {"IMM", IR_TY_REG_IMM},
    {"BPREL", IR_TY_REG_IMM},
    {"MOV", IR_TY_REG_REG},
    {"RET", IR_TY_REG},
    {"CALL", IR_TY_CALL},
    {"", IR_TY_LABEL},
    {"LABEL_ADDR", IR_TY_LABEL_ADDR},
    {"EQ", IR_TY_REG_REG},
    {"NE", IR_TY_REG_REG},
    {"LE", IR_TY_REG_REG},
    {"LT", IR_TY_REG_REG},
    {"AND", IR_TY_REG_REG},
    {"OR", IR_TY_REG_REG},
    {"XOR", IR_TY_REG_REG},
    {"SHL", IR_TY_REG_REG},
    {"SHR", IR_TY_REG_REG},
    {"MOD", IR_TY_REG_REG},
    {"NEG", IR_TY_REG},
    {"JMP", IR_TY_JMP},
    {"IF", IR_TY_REG_LABEL},
    {"UNLESS", IR_TY_REG_LABEL},
    {"LOAD", IR_TY_MEM},
    {"STORE8", IR_TY_REG_REG},
    {"STORE32", IR_TY_REG_REG},
    {"STORE64", IR_TY_REG_REG},
    {"STORE8_ARG", IR_TY_IMM_IMM},
    {"STORE32_ARG", IR_TY_IMM_IMM},
    {"STORE64_ARG", IR_TY_IMM_IMM},
    {"KILL", IR_TY_REG},
    {"SAVE_ARGS", IR_TY_IMM},
    {"NOP", IR_TY_NOARG},
}

func tostr(ir *IR) string {
    info := Irinfo_arr[ir.Op]
    switch info.Ty {
    case IR_TY_LABEL:
        return fmt.Sprintf(".L%d:", ir.Lhs)
    case IR_TY_LABEL_ADDR:
        return fmt.Sprintf("  %s r%d, %s", info.Name, ir.Lhs, ir.Name)
    case IR_TY_IMM:
        return fmt.Sprintf("  %s %d", info.Name, ir.Lhs)
    case IR_TY_REG:
        return fmt.Sprintf("  %s r%d", info.Name, ir.Lhs)
    case IR_TY_JMP:
        return fmt.Sprintf("  %s .L%d", info.Name, ir.Lhs)
    case IR_TY_REG_REG:
        return fmt.Sprintf("  %s r%d, r%d", info.Name, ir.Lhs, ir.Rhs)
    case IR_TY_MEM:
        return fmt.Sprintf("  %s%d r%d, r%d", info.Name, ir.Size, ir.Lhs, ir.Rhs)
    case IR_TY_REG_IMM:
        return fmt.Sprintf("  %s r%d, %d", info.Name, ir.Lhs, ir.Rhs)
    case IR_TY_IMM_IMM:
        return fmt.Sprintf("  %s %d, %d", info.Name, ir.Lhs, ir.Rhs)
    case IR_TY_REG_LABEL:
        return fmt.Sprintf("  %s r%d, .L%d", info.Name, ir.Lhs, ir.Rhs)
    case IR_TY_CALL:
        var sb string
        sb = fmt.Sprintf("  r%d = %s(", ir.Lhs, ir.Name)
        for i := 0; i < ir.Nargs; i++ {
            if i != 0 {
                sb += ", "
            }
            sb += fmt.Sprintf("r%d", ir.Args[i])
        }
        sb += ")"
        return sb + "\000"
    default:
        Assert(info.Ty == IR_TY_NOARG, "not IR_TY_NOARG")
        return fmt.Sprintf("  %s", info.Name)
    }
}

func Dump_ir(irv *Vector) {
    for i := 0; i < irv.Len; i++ {
        fn, _ := irv.Data[i].(*Function)

        fmt.Fprintf(os.Stderr, "%s():\n", fn.Name);
        for j := 0; j < fn.Ir.Len; j++ {
            ir := fn.Ir.Data[j].(*IR)
            fmt.Fprintf(os.Stderr, "%s\n", tostr(ir))
        }
    }
}
