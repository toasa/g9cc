package ir

import (
    . "g9cc/common"
    . "g9cc/util"
    "strings"
    "fmt"
)

var code *Vector

func add(op int, lhs int, rhs int) *IR {
    var ir *IR = new(IR)
    ir.Op = op
    ir.Lhs = lhs
    ir.Rhs = rhs
    Vec_push(code, ir)
    return ir
}

var regno int

func gen_expr(node *Node) int {

    if node.Ty == ND_NUM {
        var r int = regno
        regno++
        add(IR_IMM, r, node.Val)
        return r;
    }

    Assert((strings.Contains("+-*/", string(node.Ty))), "operator expected")

    // lhs, rhsどちらも数値が格納されている
    var lhs int = gen_expr(node.Lhs)
    var rhs int = gen_expr(node.Rhs)

    add(node.Ty, lhs, rhs)
    add(IR_KILL, rhs, 0)

    return lhs
}

func gen_stmt(node *Node) {
    if node.Ty == ND_RETURN {
        r := gen_expr(node.Expr)
        add(IR_RETURN, r, 0)
        add(IR_KILL, r, 0)
        return
    }

    if node.Ty == ND_EXPR_STMT {
        r := gen_expr(node.Expr)
        add(IR_KILL, r, 0)
        return
    }

    if node.Ty == ND_COMP_STMT {
        for i := 0; i < node.Stmts.Len; i++ {
            n, _ := node.Stmts.Data[i].(*Node)
            gen_stmt(n)
        }
        return
    }

    Error(fmt.Sprintf("unknown node: %d", node.Ty))
}

// ASTを引数にとり、中間表現を返す
// irは {op: , lhs: , rhs : }からなる
// op = numのとき => {IR_IMM, register_index, num_value}
// op = '+'のとき => {'+', lhsの値が格納されたregisterのindex, rhsの値が格納されたregisterのindex}
// opが'+'or'-'の直後 => {IR_KILL, rhsの値が格納されたregisterのindex, 0}
// ここで決めたregisterのindexは確定ではなく, alloc_regs()で配列insを一つひとつ読みながら
// 最終的なregister を決定する
func Gen_ir(node *Node) *Vector{
    Assert(node.Ty == ND_COMP_STMT, "Type of root node type is not ND_COMP_STMT")
    code = New_vec()
    gen_stmt(node)
    return code
}
