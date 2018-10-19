package token

import (
    . "g9cc/common"
    . "g9cc/util"
    "fmt"
    "os"
)

// Tokenizer
func add_token(v *Vector, ty int, input string) *Token {
    t := new(Token)
    t.Ty = ty
    t.Input = input
    Vec_push(v, t)
    return t
}

func Tokenize(s string) *Vector {
    var v *Vector = New_vec()

    // index of input
    i_input := 0

    for s[i_input] != '\000' {

        // white space
        if Isspace(s[i_input]) {
            i_input++
            continue
        }

        // single-letter token
        if s[i_input] == '+' || s[i_input] == '-' || s[i_input] == '*' || s[i_input] == '/'{
            add_token(v, int(s[i_input]), string(s[i_input]))
            i_input++
            continue
        }

        // number
        if Isdigit(s[i_input]) {
            var num int = int(s[i_input] - '0')
            i_input++
            for ; Isdigit(s[i_input]); i_input++ {
                num = num * 10 + int(s[i_input] - '0')
            }

            var t *Token = add_token(v, TK_NUM, string(num))

            t.Val = num
            continue
        }

        fmt.Println("what's up guys")
        fmt.Printf("cannot tokenize: %s", s);
        os.Exit(1)
    }


    add_token(v, TK_EOF, s);
    return v
}

func Isdigit(c uint8) bool {
    if '0' <= c && c <= '9' {
        return true
    } else {
        return false
    }
}

func Isspace(c uint8) bool {
    if c == ' ' {
        return true
    } else {
        return false
    }
}
