// gen_ir.goが担っていた機能, 各識別子や変数をstoreやloadするために
// base pointerからの距離を, map varsに格納し、node.Offsetに代入する機能,
// これをsema.goに書いた. またグローバル変数の環境範囲（？）もここで行う

package sema

import (
    . "g9cc/common"
    . "g9cc/util"
    "fmt"
    // "github.com/k0kubun/pp"
)

var int_ty Type = Type{Ty: INT, Size: 4, Align: 4}

// Compound statement( {...; ...; ...;} )内に変数のスコープを制限するための構造体Env.
// 各comp stmtごとに変数を登録するためのmap, varsを持っている
type Env struct {
    // 各識別子において、rbpからのoffsetを登録するための辞書
    vars map[string]interface{}
    next *Env
}

var globals *Vector
var env *Env
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

func find_var(name string) *Var {
    for e := env; e != nil; e = e.next {
        if e.vars[name] != nil {
            var_ := e.vars[name].(*Var)
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

    // decayがtrueでかつ、base.Ty.TyがARYだった場合
    node := new(Node)
    node.Op = ND_ADDR
    node.Ty = Ptr_to(base.Ty.Ary_of)

    copy := new(Node)

    // Cでは memcpy(copy, base, sizeof(Node))
    // と書かれた.
    *copy = *base
    node.Expr = copy

    return node
}

func check_lval(node *Node) {
    op := node.Op
    if op != ND_LVAR && op != ND_GVAR && op != ND_DEREF && op != ND_DOT {
        Error(fmt.Sprintf("not an lvalue: %d (%s)", op, node.Name))
    }
}

func new_int(val int) *Node {
    node := new(Node)
    node.Op = ND_NUM
    //node.Ty = INT
    node.Val = val
    return node
}

// ASTを渡り歩く
func walk(node *Node, decay bool) *Node {
    switch node.Op {
    case ND_NUM, ND_NULL, ND_BREAK:
        return node
    case ND_STR:
        // 文字列はグローバル変数として扱う
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
        var_ := find_var(node.Name)

        // pp.Print(node)
        // fmt.Println("------------------")

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
        stacksize = Roundup(stacksize, node.Ty.Align)
        stacksize += node.Ty.Size
        node.Offset = stacksize

        var_ := new(Var)
        var_.Ty = node.Ty
        var_.Is_local = true
        var_.Offset = stacksize
        env.vars[node.Name] = var_

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
        env = new_env(env)
        node.Init = walk(node.Init, true)
        if node.Cond != nil {
            node.Cond = walk(node.Cond, true)
        }
        if node.Inc != nil {
            node.Inc = walk(node.Inc, true)
        }
        node.Body = walk(node.Body, true)
        env = env.next
        return node
    case ND_DO_WHILE:
        node.Cond = walk(node.Cond, true)
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
    case '=', ND_MUL_EQ, ND_DIV_EQ, ND_MOD_EQ, ND_ADD_EQ, ND_SUB_EQ, ND_SHL_EQ, ND_SHR_EQ, ND_BITAND_EQ, ND_XOR_EQ, ND_BITOR_EQ:
        node.Lhs = walk(node.Lhs, false)
        check_lval(node.Lhs)
        node.Rhs = walk(node.Rhs, true)
        node.Ty = node.Lhs.Ty
        return node
    case ND_DOT:
        node.Expr = walk(node.Expr, true)
        if node.Expr.Ty.Ty != STRUCT {
            Error("struct expect before '.'")
        }

        ty := node.Expr.Ty
        if ty.Members == nil {
            Error("incomplete type")
        }

        for i := 0; i < ty.Members.Len; i++ {
            m := ty.Members.Data[i].(*Node)
            if m.Name != node.Name {
                continue
            }
            node.Ty = m.Ty
            node.Offset = m.Ty.Offset
            return maybe_decay(node, decay)
        }
        Error(fmt.Sprintf("member missing: %s", node.Name))
    case '?':
        node.Cond = walk(node.Cond, true)
        node.Then = walk(node.Then, true)
        node.Els = walk(node.Els, true)
        node.Ty = node.Then.Ty
        return node
    case '*', '/', '%', '<', '|', '^', '&', ND_EQ, ND_NE, ND_LE, ND_SHL, ND_SHR, ND_LOGAND, ND_LOGOR:
        node.Lhs = walk(node.Lhs, true)
        node.Rhs = walk(node.Rhs, true)
        node.Ty = node.Lhs.Ty
        return node
    case ',':
        node.Lhs = walk(node.Lhs, true)
        node.Rhs = walk(node.Rhs, true)
        node.Ty = node.Rhs.Ty
        return node
    case ND_PRE_INC, ND_PRE_DEC, ND_POST_INC, ND_POST_DEC, ND_NEG, '!':
        node.Expr = walk(node.Expr, true)
        node.Ty = node.Expr.Ty
        return node
    case ND_ADDR:
        node.Expr = walk(node.Expr, true)
        check_lval(node.Expr)
        node.Ty = Ptr_to(node.Expr.Ty)
        return node
    case ND_DEREF:
        node.Expr = walk(node.Expr, true)
        if node.Expr.Ty.Ty != PTR {
            Error("operand must be a pointer")
        }
        if node.Expr.Ty.Ptr_to.Ty == VOID {
            Error("cannot dereference void pointer")
        }
        node.Ty = node.Expr.Ty.Ptr_to
        return node
    case ND_RETURN, ND_EXPR_STMT:
        node.Expr = walk(node.Expr, true)
        return node
    case ND_SIZEOF:
        expr := walk(node.Expr, false)
        return new_int(expr.Ty.Size)
    case ND_ALIGNOF:
        expr := walk(node.Expr, false)
        return new_int(expr.Ty.Align)
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
        env = new_env(env)
        for i := 0; i < node.Stmts.Len; i++ {
            node.Stmts.Data[i] = walk(node.Stmts.Data[i].(*Node), true)
        }
        env = env.next
        return node
    case ND_STMT_EXPR:
        node.Body = walk(node.Body, true)
        node.Ty = &int_ty
        return node
    default:
        Assert(false, "unknown node type")
    }

    err := new(Node)
    return err
}

func Sema(nodes *Vector) *Vector {
    env = new_env(nil)
    globals = New_vec()

    for i := 0; i < nodes.Len; i++ {
        node := nodes.Data[i].(*Node)

        if node.Op == ND_VARDEF {
            var_ := new_global(node.Ty, node.Name, node.Data, node.Len)
            var_.Is_extern = node.Is_extern
            Vec_push(globals, var_)
            env.vars[node.Name] = var_
            continue
        }

        Assert(node.Op == ND_FUNC, "node Op is not ND_FUNC")

        stacksize = 0

        walk(node, true)
        node.Stacksize = stacksize
    }

    return globals
}
