// gen_ir.goが担っていた機能, 各識別子や変数をstoreやloadするために
// base pointerからの距離を, map varsに格納し、node.Offsetに代入する機能,
// これをsema.goに書いた. またグローバル変数の環境範囲（？）もここで行う

package sema

import (
    . "g9cc/common"
    . "g9cc/util"
    "fmt"
)

var int_ty Type = Type{Ty: INT, Ptr_of: nil, Ary_of: nil, Len: 0}

// Compound statement( {...; ...; ...;} )内に変数のスコープを制限するための構造体Env.
// 各comp stmtごとに変数を登録するためのmap, varsを持っている
type Env struct {
    // 各識別子において、rbpからのoffsetを登録するための辞書
    vars map[string]interface{}

    next *Env
}

var globals *Vector
var str_label int
var stacksize int

func new_env(next *Env) *Env{
    env := new(Env)
    env.vars = make(map[string]interface{})
    env.next = next
    return env
}

func new_global(ty *Type, name, data string, len int) *Var {
    var_ := new(Var)
    var_.Ty = ty
    var_.Is_local = false
    var_.Name = name
    var_.Data = data
    var_.Len = len
    return var_
}

func find(env *Env, name string) *Var {
    for ; env != nil; env = env.next {
        if env.vars[name] != nil {
            var_ := env.vars[name].(*Var)
            if var_ != nil {
                return var_
            }
        }
    }
    return nil
}

func swap(p **Node, q **Node) {
    var r *Node = *p
    *p = *q
    *q = r
}

func maybe_decay(base *Node, decay bool) *Node {
    if !decay || (base.Ty.Ty != ARY) {
        return base
    }

    node := new(Node)
    node.Op = ND_ADDR
    node.Ty = Ptr_of(base.Ty.Ary_of)

    copy := new(Node)

    // Cでは memcpy(copy, base, sizeof(Node))
    // と書かれた.
    *copy = *base
    node.Expr = copy

    return node
}

func check_lval(node *Node) {
    op := node.Op
    if op == ND_LVAR || op == ND_GVAR || op == ND_DEREF {
        return
    }
    Error(fmt.Sprintf("not an lvalue: %d (%s)", op, node.Name))
}

// ASTを渡り歩く
func walk(env *Env, node *Node, decay bool) *Node {
    switch node.Op {
    case ND_NUM:
        return node
    case ND_STR:
        var_ := new_global(node.Ty, fmt.Sprintf("L_.str%d", str_label),
            node.Data, node.Len)
        str_label++
        Vec_push(globals, var_)

        ret := new(Node)
        ret.Op = ND_GVAR
        ret.Ty = node.Ty
        ret.Name = var_.Name
        return maybe_decay(ret, decay)
    case ND_IDENT:
        var_ := find(env, node.Name)

        if var_ == nil {
            Error(fmt.Sprintf("undefined variable: %s", node.Name))
        }

        if var_.Is_local {
            ret := new(Node)
            ret.Op = ND_LVAR
            ret.Ty = var_.Ty
            ret.Offset = var_.Offset
            return maybe_decay(ret, decay)
        }

        // global var
        ret := new(Node)
        ret.Op = ND_GVAR
        ret.Ty = var_.Ty
        ret.Name = var_.Name
        return maybe_decay(ret, decay)
    case ND_VARDEF:
        // varsに識別子の登録がされていない場合
        // 識別子をメモリ上へstoreしたり、メモリからloadする時のために、
        // base pointerからの距離を, map varsに格納しておく。
        stacksize += Size_of(node.Ty)
        node.Offset = stacksize

        var_ := new(Var)
        var_.Ty = node.Ty
        var_.Is_local = true
        var_.Offset = stacksize
        env.vars[node.Name] = var_

        if node.Init != nil {
            node.Init = walk(env, node.Init, true)
        }
        return node
    case ND_IF:
        node.Cond = walk(env, node.Cond, true)
        node.Then = walk(env, node.Then, true)
        if node.Els != nil {
            node.Els = walk(env, node.Els, true)
        }
        return node
    case ND_FOR:
        node.Init = walk(env, node.Init, true)
        node.Cond = walk(env, node.Cond, true)
        node.Inc = walk(env, node.Inc, true)
        node.Body = walk(env, node.Body, true)
        return node
    case ND_DO_WHILE:
        node.Cond = walk(env, node.Cond, true)
        node.Body = walk(env, node.Body, true)
        return node
    case '+', '-':
        node.Lhs = walk(env, node.Lhs, true)
        node.Rhs = walk(env, node.Rhs, true)

        if node.Rhs.Ty.Ty == PTR {
            swap(&node.Lhs, &node.Rhs)
        }
        if node.Rhs.Ty.Ty == PTR {
            Error(fmt.Sprintf("pointer %c pointer is not defined", node.Op))
        }

        node.Ty = node.Lhs.Ty
        return node
    case '=':
        node.Lhs = walk(env, node.Lhs, false)
        check_lval(node.Lhs)
        node.Rhs = walk(env, node.Rhs, true)
        node.Ty = node.Lhs.Ty
        return node
    case '*', '/', '<', ND_EQ, ND_NE, ND_LOGAND, ND_LOGOR:
        node.Lhs = walk(env, node.Lhs, true)
        node.Rhs = walk(env, node.Rhs, true)
        node.Ty = node.Lhs.Ty
        return node
    case ND_ADDR:
        node.Expr = walk(env, node.Expr, true)
        check_lval(node.Expr)
        node.Ty = Ptr_of(node.Expr.Ty)
        return node
    case ND_DEREF:
        node.Expr = walk(env, node.Expr, true)
        if node.Expr.Ty.Ty != PTR {
            Error("operand must be a pointer")
        }
        node.Ty = node.Expr.Ty.Ptr_of
        return node
    case ND_RETURN:
        node.Expr = walk(env, node.Expr, true)
        return node
    case ND_SIZEOF:
        expr := walk(env, node.Expr, false)

        ret := new(Node)
        ret.Op = ND_NUM
        ret.Ty = &Type{Ty: INT, Ptr_of: nil, Ary_of: nil, Len: 0}
        ret.Val = Size_of(expr.Ty)
        return ret
    case ND_CALL:
        for i := 0; i < node.Args.Len; i++ {
            node.Args.Data[i] = walk(env, node.Args.Data[i].(*Node), true)
        }
        node.Ty = &int_ty
        return node
    case ND_FUNC:
        for i := 0; i < node.Args.Len; i++ {
            node.Args.Data[i] = walk(env, node.Args.Data[i].(*Node), true)
        }
        node.Body = walk(env, node.Body, true)
        return node
    case ND_COMP_STMT:
        newenv := new_env(env)
        for i := 0; i < node.Stmts.Len; i++ {
            node.Stmts.Data[i] = walk(newenv, node.Stmts.Data[i].(*Node), true)
        }
        return node
    case ND_EXPR_STMT:
        node.Expr = walk(env, node.Expr, true)
        return node
    default:
        Assert(false, "unknown node type")
    }

    err := new(Node)
    return err
}

func Sema(nodes *Vector) *Vector {
    globals = New_vec()
    topenv := new_env(nil)

    for i := 0; i < nodes.Len; i++ {
        node := nodes.Data[i].(*Node)

        if node.Op == ND_VARDEF {
            var_ := new_global(node.Ty, node.Name, node.Data, node.Len)
            var_.Is_extern = node.Is_extern
            Vec_push(globals, var_)
            topenv.vars[node.Name] = var_
            continue
        }

        Assert(node.Op == ND_FUNC, "node Op is not ND_FUNC")

        stacksize = 0

        walk(topenv, node, true)
        node.Stacksize = stacksize
    }

    return globals
}
