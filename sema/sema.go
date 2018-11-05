// gen_ir.goが担っていた機能, 各識別子や変数をstoreやloadするために
// base pointerからの距離を, map varsに格納し、node.Offsetに代入する機能,
// これをsema.goに書いた.

package sema

import (
    . "g9cc/common"
    . "g9cc/util"
    "fmt"
)

// 各識別子において、rbpからのoffsetを登録するための辞書
var vars map[string]interface{}
var stacksize int

func walk(node *Node) {
    switch node.Ty {
    case ND_NUM:
        return
    case ND_IDENT:
        _, ok := vars[node.Name]

        if !ok {
            Error(fmt.Sprintf("undefined variable: %s", node.Name))
        }
        node.Ty = ND_LVAR
        node.Offset = vars[node.Name].(int)
        return
    case ND_VARDEF:
        // varsに識別子の登録がされていない場合
        // 識別子をメモリ上へstoreしたり、メモリからloadする時のために、
        // base pointerからの距離を, map varsに格納しておく。
        stacksize += 8
        vars[node.Name] = stacksize
        node.Offset = stacksize
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
        return
    case ND_RETURN:
        walk(node.Expr)
        return
    case ND_CALL:
        for i := 0; i < node.Args.Len; i++ {
            walk(node.Args.Data[i].(*Node))
        }
        // walk(node.Body)
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
        Assert(node.Ty == ND_FUNC, "node type is not ND_FUNC")

        vars = make(map[string]interface{})
        stacksize = 0
        walk(node)
        node.Stacksize = stacksize
    }
}
