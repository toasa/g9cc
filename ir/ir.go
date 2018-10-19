package ir

import (
    . "g9cc/common"
)

func new_ir(op int, lhs int, rhs int) *IR {
    var ir *IR = new(IR)
    ir.Op = op
    ir.Lhs = lhs
    ir.Rhs = rhs
    return ir
}


var regno int

func gen(v *Vector, node *Node) int {
    // var regno int

    if node.Ty == ND_NUM {
        var r int = regno
        regno++
        Vec_push(v, new_ir(IR_IMM, r, node.Val))
        return r;
    }

    Assert((node.Ty == '+' || node.Ty == '-' || node.Ty == '*'), "operator expected")

    var lhs int = gen(v, node.Lhs)
    var rhs int = gen(v, node.Rhs)

    Vec_push(v, new_ir(node.Ty, lhs, rhs))
    Vec_push(v, new_ir(IR_KILL, rhs, 0))

    return lhs
}

// ASTを引数にとり、中間表現を返す
// irは {op: , lhs: , rhs : }からなる
// op = numのとき => {IR_IMM, register_index, num_value}
// op = '+'のとき => {'+', lhsの値が格納されたregisterのindex, rhsの値が格納されたregisterのindex}
// opが'+'or'-'の直後 => {IR_KILL, rhsの値が格納されたregisterのindex, 0}
// ここで決めたregisterのindexは確定ではなく, alloc_regs()で配列insを一つひとつ読みながら
// 最終的なregister を決定する
func Gen_ir(node *Node) *Vector{
    var v *Vector = New_vec()
    var r int = gen(v, node)
    Vec_push(v, new_ir(IR_RETURN, r, 0))
    return v
}
