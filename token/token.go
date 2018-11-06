package token

import (
    . "g9cc/common"
    . "g9cc/util"
    "fmt"
    "strings"
)

// Tokenizer
func add_token(v *Vector, ty int, input string) *Token {
    t := new(Token)
    t.Ty = ty
    t.Input = input
    Vec_push(v, t)
    return t
}

// var keywords *Map
// var keywords map[string]interface{}

var symbols = []struct {
    name string
    ty int
}{
    {"else", TK_ELSE}, {"for", TK_FOR}, {"if", TK_IF},
    {"int", TK_INT}, {"return", TK_RETURN},
    {"&&", TK_LOGAND}, {"||", TK_LOGOR}, {"NULL", 0},
}

func Tokenize(s string) *Vector {
    var v *Vector = New_vec()

    // index of input
    i_input := 0

    loop:
        for s[i_input] != '\000' {

            // skip white space
            if isspace(s[i_input]) {
                i_input++
                continue
            }

            // Single-letter token
            if strings.Contains("+-*/;=(),{}<>", string(s[i_input])) {
                add_token(v, int(s[i_input]), string(s[i_input]))
                i_input++
                continue
            }

            // Multi-letter token
            for i := 0; symbols[i].name != "NULL"; i++ {
                name := symbols[i].name
                l := len(name)

                // 下のs[i_input:i_input+l]でスライスの右側が文字列の長さを超えることがあり、
                // errorが出ていたため、ここを記述.
                if len(s) <= i_input + l {
                    continue
                }
                if s[i_input:i_input+l] != name {
                    continue
                }

                add_token(v, symbols[i].ty, name)
                //i++
                i_input += l
                goto loop
            }

            // Identifier
            if isalpha(s[i_input]) || s[i_input] == '_' {

                len := 1
                // identifierを切りだすための添字の取得
                for i := len + i_input; isalpha(s[i]) || isdigit(s[i]) || s[i] == '_'; {
                    len++
                    i = len + i_input
                }

                t := add_token(v, TK_IDENT, s[i_input:len + i_input])

                t.Name = s[i_input:len + i_input]
                i_input += len
                continue
            }

            // number
            if isdigit(s[i_input]) {
                var num int = int(s[i_input] - '0')
                i_input++
                for ; isdigit(s[i_input]); i_input++ {
                    num = num * 10 + int(s[i_input] - '0')
                }

                var t *Token = add_token(v, TK_NUM, string(num))

                t.Val = num
                continue
            }

            fmt.Println("what's up guys")
            Error(fmt.Sprintf("cannot tokenize: %s", s));
        }

    add_token(v, TK_EOF, s);
    return v
}

func isdigit(c uint8) bool {
    if '0' <= c && c <= '9' {
        return true
    } else {
        return false
    }
}

func isspace(c uint8) bool {
    if c == ' ' {
        return true
    } else {
        return false
    }
}

func isalpha(c uint8) bool {
    if ('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z') {
        return true
    } else {
        return false
    }
}
