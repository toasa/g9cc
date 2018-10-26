package regalloc

import (
    . "g9cc/common"
    . "g9cc/util"
    . "g9cc/regs"
    . "g9cc/ir"
)

// Register allocator
var Used []bool = make([]bool, Len_Regs)

// IRの命令数分の要素をもつ配列(alloc_regs()で初期化)
var Reg_map []int


func alloc(ir_reg int) int {
    // 演算子の場合こちらが走る
    if Reg_map[ir_reg] != -1 {
        var r int = Reg_map[ir_reg]
        Assert(Used[r], "allocation error")
        return r
    }

    // i はレジスタの配列regsのindex
    // 数値の場合こちらが走る
    for i := 0; i < Len_Regs; i++ {
        // index i のレジスタが使用済みの場合
        if Used[i] {
            continue
        }
        // index i のレジスタが未使用の場合
        Used[i] = true
        Reg_map[ir_reg] = i // registerへのmapping
        return i
    }

    Error("register exhausted")
    return 0 // ここには到達しないため(intを返さないと怒るコンパイラを鎮める他に)イミなし
}

func kill(r int) {
    Assert(Used[r], "kill error")
    Used[r] = false
}

// 中間表現の命令配列insの各要素に対し、必要ならレジスタを割り当てていく
func Alloc_regs(irv *Vector) {

    for i := 0; i < irv.Len; i++ {
        Reg_map = append(Reg_map, -1)
    }

    for i := 0; i < irv.Len; i++ {
        ir, ok := irv.Data[i].(*IR)
        if !ok {
            Error("Not *IR type is in irv.data[]")
        }

        // 各irに対し Get_irinfo()を呼び出し, irの種類(レジスタの使い方で類別)
        var info *IRInfo = Get_irinfo(ir)

        switch info.Ty {
        case IR_TY_REG, IR_TY_REG_IMM, IR_TY_REG_LABEL:
            ir.Lhs = alloc(ir.Lhs)
        case IR_TY_REG_REG:
            ir.Lhs = alloc(ir.Lhs)
            ir.Rhs = alloc(ir.Rhs)
        }

        if ir.Op == IR_KILL {
            kill(ir.Lhs)
            ir.Op = IR_NOP
        }
    }
}
