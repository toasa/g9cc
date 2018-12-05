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
    {"_Alignof", TK_ALIGNOF}, {"break", TK_BREAK},
    {"char", TK_CHAR},
    {"do", TK_DO},{"else", TK_ELSE},
    {"extern", TK_EXTERN}, {"for", TK_FOR},
    {"if", TK_IF}, {"int", TK_INT},
    {"return", TK_RETURN}, {"sizeof", TK_SIZEOF},
    {"struct", TK_STRUCT}, {"typedef", TK_TYPEDEF},
    {"void", TK_VOID}, {"while", TK_WHILE},
    {"!=", TK_NE}, {"&&", TK_LOGAND},
    {"++", TK_INC}, {"--", TK_DEC},
    {"->", TK_ARROW}, {"<<", TK_SHL},
    {"<=", TK_LE}, {"==", TK_EQ},
    {">=", TK_GE},{">>", TK_SHR},
    {"||", TK_LOGOR}, {"NULL", 0},
}

var escaped [256]int32;
func init_escaped() {
    escaped['a'] = '\a'
    escaped['b'] = '\b'
    escaped['f'] = '\f'
    escaped['n'] = '\n'
    escaped['r'] = '\r'
    escaped['t'] = '\t'
    escaped['v'] = '\v'
    escaped['e'] = '\033'
    escaped['E'] = '\033'
}

func read_char(result *int, s string) int {
    s_i := 0
    if s[s_i] == '\000' {
        Error("premature end of input")
    }

    if s[s_i] != '\\' {
        *result = int(s[s_i])
        s_i++
    } else {
        s_i++
        if s[s_i] == '\000' {
            Error("premature end of input")
        }
        esc := escaped[s[s_i]]
        if esc != 0 {
            *result = int(esc)
        } else {
            *result = int(s[s_i])
        }
        s_i++
    }

    if s[s_i] != '\'' {
        Error("unclosed character literal")
    }
    s_i++
    return s_i
}

func read_string(s string) (string, int) {

    s_i := 0
    var str string

    for s[s_i] != '"' {
        if s[s_i] == '\000' {
            Error("premature end of input")
        }

        if s[s_i] != '\\' {
            str += string(s[s_i])
            s_i++
            continue
        }

        s_i++
        if s[s_i] == '\000' {
            Error("premature end of input")
        }
        esc := escaped[s[s_i]]
        if esc != 0 {
            str += string(esc)
        } else {
            str += string(s[s_i])
        }
        s_i++
    }

    return str, s_i + 1
}

func Tokenize(s string) *Vector {
    var v *Vector = New_vec()

    init_escaped()

    // index of input
    i_input := 0

    loop:
        for s[i_input] != '\000' {

            // skip white space, new line and tub.
            if isspace(s[i_input]) || s[i_input] == '\n' || s[i_input] == '\t'{
                i_input++
                continue
            }

            // Line comment
            if s[i_input:i_input+2] == "//" {
                for s[i_input] != '\000' && s[i_input] != '\n' {
                    i_input++
                }
                continue
            }

            // Block comment
            if s[i_input:i_input+2] == "/*" {
                for i_input += 2; ; i_input++ {
                    if s[i_input:i_input+2] != "*/" {
                        continue
                    }
                    i_input += 2
                    goto loop
                }
                Error("unclosed comment")
            }

            // Character literal
            if s[i_input] == '\'' {
                t := add_token(v, TK_NUM, string(s[i_input]))
                i_input++
                i_input += read_char(&t.Val, s[i_input:])
                continue
            }

            // String literal
            if s[i_input] == '"' {
                t := add_token(v, TK_STR, string(s[i_input]))
                i_input++

                t.Str, t.Len = read_string(s[i_input:])
                i_input += t.Len
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

            // Single-letter token
            if strings.Contains("+-*/;=(),{}<>[]&.!?:|^%", string(s[i_input])) {
                add_token(v, int(s[i_input]), string(s[i_input]))
                i_input++
                continue
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
