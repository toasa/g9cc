// gen_ir.goが担っていた機能, 各識別子や変数をstoreやloadするために
// base pointerからの距離を, map varsに格納し、node.Offsetに代入する機能,
// これをsema.goに書いた.

package sema

import (
    . "g9cc/common"
    . "g9cc/util"
    "fmt"
)

var int_ty Type = Type{Ty: INT, Ptr_of: nil, Ary_of: nil, Len: 0}

type Var struct {
    ty *Type
    offset int
}

// 各識別子において、rbpからのoffsetを登録するための辞書
var vars map[string]interface{}
var stacksize int

func swap(p **Node, q **Node) {
    var r *Node = *p
    *p = *q
    *q = r
}

func addr_of(base *Node, ty *Type) *Node {
    node := new(Node)
    node.Op = ND_ADDR
    node.Ty = Ptr_of(ty)

    copy := new(Node)

    // Cでは memcpy(copy, base, sizeof(Node))
    // と書かれた.
    *copy = *base
    node.Expr = copy

    return node
}

// ASTを渡り歩く
func walk(node *Node, decay bool) {
    switch node.Op {
    case ND_NUM:
        return
    case ND_IDENT:
        var_, ok := vars[node.Name].(*Var)

        if !ok {
            Error(fmt.Sprintf("undefined variable: %s", node.Name))
        }
        //node.Ty = var_.ty
        node.Op = ND_LVAR
        node.Offset = var_.offset

        if decay && (var_.ty.Ty == ARY) {
            *node = *addr_of(node, var_.ty.Ary_of)
        } else {
            node.Ty = var_.ty
        }
        return
    case ND_VARDEF:
        // varsに識別子の登録がされていない場合
        // 識別子をメモリ上へstoreしたり、メモリからloadする時のために、
        // base pointerからの距離を, map varsに格納しておく。
        stacksize += Size_of(node.Ty)
        node.Offset = stacksize

        var_ := new(Var)
        var_.ty = node.Ty
        var_.offset = stacksize
        vars[node.Name] = var_

        if node.Init != nil {
            walk(node.Init, true)
        }
        return
    case ND_IF:
        walk(node.Cond, true)
        walk(node.Then, true)
        if node.Els != nil {
            walk(node.Els, true)
        }
        return
    case ND_FOR:
        walk(node.Init, true)
        walk(node.Cond, true)
        walk(node.Inc, true)
        walk(node.Body, true)
        return
    case '+', '-':
        walk(node.Lhs, true)
        walk(node.Rhs, true)

        if node.Rhs.Ty.Ty == PTR {
            swap(&node.Lhs, &node.Rhs)
        }
        if node.Rhs.Ty.Ty == PTR {
            Error(fmt.Sprintf("pointer %c pointer is not defined", node.Op))
        }

        node.Ty = node.Lhs.Ty
        return
    case '=':
        walk(node.Lhs, false)
        walk(node.Rhs, true)
        node.Ty = node.Lhs.Ty
        return
    case '*', '/', '<', ND_LOGAND, ND_LOGOR:
        walk(node.Lhs, true)
        walk(node.Rhs, true)
        node.Ty = node.Lhs.Ty
        return
    case ND_DEREF:
        walk(node.Expr, true)
        if node.Expr.Ty.Ty != PTR {
            Error("operand must be a pointer")
        }
        node.Ty = node.Expr.Ty.Ptr_of
        return
    case ND_RETURN:
        walk(node.Expr, true)
        return
    case ND_CALL:
        for i := 0; i < node.Args.Len; i++ {
            walk(node.Args.Data[i].(*Node), true)
        }
        node.Ty = &int_ty
        return
    case ND_FUNC:
        for i := 0; i < node.Args.Len; i++ {
            walk(node.Args.Data[i].(*Node), true)
        }
        walk(node.Body, true)
        return
    case ND_COMP_STMT:
        for i := 0; i < node.Stmts.Len; i++ {
            walk(node.Stmts.Data[i].(*Node), true)
        }
        return
    case ND_EXPR_STMT:
        walk(node.Expr, true)
        return
    default:
        Assert(false, "unknown node type")
    }
}

func Sema(nodes *Vector) {
    for i := 0; i < nodes.Len; i++ {
        node := nodes.Data[i].(*Node)
        Assert(node.Op == ND_FUNC, "node Op is not ND_FUNC")

        // 各関数のローカル変数郡のためのmap
        vars = make(map[string]interface{})
        stacksize = 0
        walk(node, true)
        node.Stacksize = stacksize
    }
}
