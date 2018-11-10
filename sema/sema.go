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
func walk(node *Node, decay bool) *Node {
    switch node.Op {
    case ND_NUM:
        return node
    case ND_IDENT:
        var_, ok := vars[node.Name].(*Var)

        if !ok {
            Error(fmt.Sprintf("undefined variable: %s", node.Name))
        }
        //node.Ty = var_.ty
        node.Op = ND_LVAR
        node.Offset = var_.offset

        if decay && (var_.ty.Ty == ARY) {
            return addr_of(node, var_.ty.Ary_of)
        }

        node.Ty = var_.ty
        return node
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
            node.Init = walk(node.Init, true)
        }
        return node
    case ND_IF:
        node.Cond = walk(node.Cond, true)
        node.Then = walk(node.Then, true)
        if node.Els != nil {
            node.Els = walk(node.Els, true)
        }
        return node
    case ND_FOR:
        node.Init = walk(node.Init, true)
        node.Cond = walk(node.Cond, true)
        node.Inc = walk(node.Inc, true)
        node.Body = walk(node.Body, true)
        return node
    case '+', '-':
        node.Lhs = walk(node.Lhs, true)
        node.Rhs = walk(node.Rhs, true)

        if node.Rhs.Ty.Ty == PTR {
            swap(&node.Lhs, &node.Rhs)
        }
        if node.Rhs.Ty.Ty == PTR {
            Error(fmt.Sprintf("pointer %c pointer is not defined", node.Op))
        }

        node.Ty = node.Lhs.Ty
        return node
    case '=':
        node.Lhs = walk(node.Lhs, false)
        node.Rhs = walk(node.Rhs, true)
        node.Ty = node.Lhs.Ty
        return node
    case '*', '/', '<', ND_LOGAND, ND_LOGOR:
        node.Lhs = walk(node.Lhs, true)
        node.Rhs = walk(node.Rhs, true)
        node.Ty = node.Lhs.Ty
        return node
    case ND_ADDR:
        node.Expr = walk(node.Expr, true)
        node.Ty = Ptr_of(node.Expr.Ty)
        return node
    case ND_DEREF:
        node.Expr = walk(node.Expr, true)
        if node.Expr.Ty.Ty != PTR {
            Error("operand must be a pointer")
        }
        node.Ty = node.Expr.Ty.Ptr_of
        return node
    case ND_RETURN:
        node.Expr = walk(node.Expr, true)
        return node
    case ND_CALL:
        for i := 0; i < node.Args.Len; i++ {
            node.Args.Data[i] = walk(node.Args.Data[i].(*Node), true)
        }
        node.Ty = &int_ty
        return node
    case ND_FUNC:
        for i := 0; i < node.Args.Len; i++ {
            node.Args.Data[i] = walk(node.Args.Data[i].(*Node), true)
        }
        node.Body = walk(node.Body, true)
        return node
    case ND_COMP_STMT:
        for i := 0; i < node.Stmts.Len; i++ {
            node.Stmts.Data[i] = walk(node.Stmts.Data[i].(*Node), true)
        }
        return node
    case ND_EXPR_STMT:
        node.Expr = walk(node.Expr, true)
        return node
    default:
        Assert(false, "unknown node type")
    }

    err := new(Node)
    return err
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
