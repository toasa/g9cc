// gen_ir.goが担っていた機能, 各識別子や変数をstoreやloadするために
// base pointerからの距離を, map varsに格納し、node.Offsetに代入する機能,
// これをsema.goに書いた.

package sema

import (
    . "g9cc/common"
    . "g9cc/util"
    "fmt"
)

var int_ty Type = Type{Ty: INT, Ptr_of: nil}

type Var struct {
    ty *Type
    offset int
}

// 各識別子において、rbpからのoffsetを登録するための辞書
var vars map[string]interface{}
var stacksize int

func walk(node *Node) {
    switch node.Op {
    case ND_NUM:
        return
    case ND_IDENT:
        var_, ok := vars[node.Name].(*Var)

        if !ok {
            Error(fmt.Sprintf("undefined variable: %s", node.Name))
        }
        node.Ty = var_.ty
        node.Op = ND_LVAR
        node.Offset = var_.offset
        return
    case ND_VARDEF:
        // varsに識別子の登録がされていない場合
        // 識別子をメモリ上へstoreしたり、メモリからloadする時のために、
        // base pointerからの距離を, map varsに格納しておく。
        stacksize += 8
        node.Offset = stacksize

        var_ := new(Var)
        var_.ty = node.Ty
        var_.offset = stacksize
        vars[node.Name] = var_

        if node.Init != nil {
            walk(node.Init)
        }
        return
    case ND_IF:
        walk(node.Cond)
        walk(node.Then)
        if node.Els != nil {
            walk(node.Els)
        }
        return
    case ND_FOR:
        walk(node.Init)
        walk(node.Cond)
        walk(node.Inc)
        walk(node.Body)
        return
    case '+', '-', '*', '/', '=', '<', ND_LOGAND, ND_LOGOR:
        walk(node.Lhs)
        walk(node.Rhs)
        node.Ty = node.Lhs.Ty
        return
    case ND_DEREF, ND_RETURN:
        walk(node.Expr)
        return
    case ND_CALL:
        for i := 0; i < node.Args.Len; i++ {
            walk(node.Args.Data[i].(*Node))
        }
        node.Ty = &int_ty
        return
    case ND_FUNC:
        for i := 0; i < node.Args.Len; i++ {
            walk(node.Args.Data[i].(*Node))
        }
        walk(node.Body)
        return
    case ND_COMP_STMT:
        for i := 0; i < node.Stmts.Len; i++ {
            walk(node.Stmts.Data[i].(*Node))
        }
        return
    case ND_EXPR_STMT:
        walk(node.Expr)
        return
    default:
        Assert(false, "unknown node type")
    }
}

func Sema(nodes *Vector) {
    for i := 0; i < nodes.Len; i++ {
        node := nodes.Data[i].(*Node)
        Assert(node.Op == ND_FUNC, "node Op is not ND_FUNC")

        vars = make(map[string]interface{})
        stacksize = 0
        walk(node)
        node.Stacksize = stacksize
    }
}
