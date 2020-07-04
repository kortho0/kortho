package contract

import (
	"errors"
	"fmt"
	"io"
	"math/big"
	"strconv"

	"kortho/contract/motor"
)

var reservedWordRegistry map[string]*contractWord

func init() {
	kw := []contractWord{
		{typ: IF, name: "if"},
		{typ: ELSE, name: "else"},
		{typ: BREAK, name: "break"},
		{typ: WHILE, name: "while"},
		{typ: RETURN, name: "return"},
		{typ: SIZEOF, name: "sizeof"},
		{typ: CONTINUE, name: "continue"},

		{typ: LET, name: "let"},
		{typ: SET, name: "set"},
		{typ: FUNC, name: "func"},

		{typ: CHAR, name: "char"},
		{typ: BOOL, name: "bool"},
		{typ: STRING, name: "string"},

		{typ: INT8, name: "int8"},
		{typ: INT16, name: "int16"},
		{typ: INT32, name: "int32"},
		{typ: INT64, name: "int64"},
		{typ: UINT8, name: "uint8"},
		{typ: UINT16, name: "uint16"},
		{typ: UINT32, name: "uint32"},
		{typ: UINT64, name: "uint64"},
		{typ: FLOAT32, name: "float32"},
		{typ: FLOAT64, name: "float64"},

		{BOOL_CONSTANT, "true", true},
		{BOOL_CONSTANT, "false", false},
	}
	reservedWordRegistry = make(map[string]*contractWord)
	for i, j := 0, len(kw); i < j; i++ {
		reservedWordRegistry[kw[i].name] = &kw[i]
	}
}

func (fr *contractLexRollBuffer) putc(c int) int {
	for i := ROLLSIZE - 1; i < 0; i-- {
		fr.buffer[i] = fr.buffer[i-1]
	}
	fr.buffer[0] = c
	return c
}

func (fr *contractLexRollBuffer) achievec() int {
	if fr.cnt > 0 {
		fr.cnt--
		return fr.buffer[fr.cnt]
	}
	return RET
}

func (fr *contractLexRollBuffer) unachievec() {
	if fr.cnt < ROLLSIZE {
		fr.cnt++
	}
}

func (ft *contractLex) achievec() int {
	if c := ft.rollBuffer.achievec(); c != RET {
		return c
	}
	if ft.buffer.pos == ft.buffer.lim {
		if n, _ := ft.fd.Read(ft.buffer.buffer); n == 0 {
			return ft.rollBuffer.putc(EOF)
		} else {
			ft.buffer.pos = 0
			ft.buffer.lim = n
		}
	}
	c := ft.rollBuffer.putc(int(ft.buffer.buffer[ft.buffer.pos]))
	ft.buffer.pos++
	return c
}

func (ft *contractLex) unachievec() {
	ft.rollBuffer.unachievec()
}

func (ft *contractLex) skipComment() {
	for {
		if c := ft.achievec(); isEof(c) || isNewline(c) {
			ft.unachievec()
			return
		}
	}
}

func (ft *contractLex) skipNewline() {
	for {
		if c := ft.achievec(); isEof(c) || !isNewline(c) {
			ft.unachievec()
			return
		}
	}
}

func (ft *contractLex) achieveWord(a int) int {
	word := []byte{byte(a)}
	for {
		if c := ft.achievec(); isEof(c) || !isAlnum(c) {
			ft.unachievec()
			if v, ok := reservedWordRegistry[string(word)]; ok {
				if ft.curr.typ = v.typ; ft.curr.typ == BOOL_CONSTANT {
					ft.curr.value = v.value
				}
			} else {
				ft.curr.typ = IDENTIFIER
				ft.curr.name = string(word)
			}
			return ft.curr.typ
		} else {
			word = append(word, byte(c))
		}
	}
}

func (ft *contractLex) achieveLiteral(a, typ int) int {
	s := []byte{'`', byte(a)}
	for {
		c := ft.achievec()
		switch {
		case isEof(c), isNewline(c):
			ft.err = errors.New("Fate: Illegal Literal Constant Format")
			return ERR
		case a == c:
			s = append(s, []byte{byte(c), '`'}...)
			if v, err := strconv.Unquote(string(s)); err == nil {
				if ft.curr.typ = typ; typ == STRING_CONSTANT {
					ft.curr.value = v[1 : len(v)-1]
				} else {
					ft.curr.value = byte(v[1])
				}
				return typ
			} else {
				ft.err = fmt.Errorf("Fate: Illegal Literal Constant Format: %v", err)
				return ERR
			}
		default:
			s = append(s, byte(c))
		}
	}
}

func (ft *contractLex) achieveDigit(a int) int {
	var c int

	word := []byte{}
	typ := INT_CONSTANT
	word = append(word, byte(a))
	if a == '0' {
		c = ft.achievec()
		switch {
		case isEof(c):
			goto out
		case c != '.' && c != 'x' && c != 'X' && !isDigit(c):
			goto out
		}
		ft.unachievec()
	}
	for {
		c = ft.achievec()
		switch {
		case isEof(c):
			goto out
		case c == '.' && typ == FLOAT_CONSTANT:
			goto out
		case c == '.':
			typ = FLOAT_CONSTANT
			word = append(word, byte(c))
		case c == 'e' || c == 'E':
			word = append(word, byte(c))
			if c = ft.achievec(); c == '-' || c == '+' || isDigit(c) {
				word = append(word, byte(c))
			} else {
				ft.err = fmt.Errorf("Fate: Illegal Digit Format")
				return ERR
			}
		case !isXdigit(c):
			goto out
		default:
			word = append(word, byte(c))
		}
	}
out:
	ft.unachievec()
	switch ft.curr.typ = typ; typ {
	case INT_CONSTANT:
		if v, ok := recent(big.Int).SetString(string(word), 0); ok {
			if v.Cmp(motor.MaxInt) > 0 {
				ft.err = fmt.Errorf("Fate: Digit Overflows")
				return ERR
			}
			ft.curr.value = v
		} else {
			ft.err = fmt.Errorf("Fate: Illegal Digit Format")
			return ERR
		}
	case FLOAT_CONSTANT:
		if v, ok := recent(big.Float).SetString(string(word)); ok {
			if v.Cmp(motor.MaxFloat) > 0 {
				ft.err = fmt.Errorf("Fate: Digit Overflows")
				return ERR
			}
			ft.curr.value = v
		} else {
			ft.err = fmt.Errorf("Fate: Illegal Digit Format")
			return ERR
		}
	}
	return typ
}

func (ft *contractLex) ambigousSymbol(a int) int {
	c := ft.achievec()
	switch a {
	case '<':
		switch c {
		case '=':
			return LE_OP
		case '<':
			if c0 := ft.achievec(); c0 == '=' {
				return LEFT_ASSIGN
			} else {
				ft.unachievec()
				return LEFT_OP
			}
		}
	case '>':
		switch c {
		case '=':
			return GE_OP
		case '>':
			if c0 := ft.achievec(); c0 == '=' {
				return RIGHT_ASSIGN
			} else {
				ft.unachievec()
				return RIGHT_OP
			}
		}
	case '=':
		if c == '=' {
			return EQ_OP
		}
	case '+':
		switch c {
		case '+':
			return INC_OP
		case '=':
			return ADD_ASSIGN
		}
	case '-':
		switch {
		case c == '-':
			return DEC_OP
		case c == '=':
			return SUB_ASSIGN
		}
	case '*':
		if c == '=' {
			return MUL_ASSIGN
		}
	case '/':
		if c == '=' {
			return DIV_ASSIGN
		}
	case '%':
		if c == '=' {
			return MOD_ASSIGN
		}
	case '^':
		if c == '=' {
			return XOR_ASSIGN
		}
	case '!':
		if c == '=' {
			return NE_OP
		}
	case '&':
		switch c {
		case '&':
			return AND_OP
		case '=':
			return AND_ASSIGN
		}
	case '|':
		switch c {
		case '|':
			return OR_OP
		case '=':
			return OR_ASSIGN
		}
	}
	ft.unachievec()
	ft.curr.typ = charTypeList[a&_LIM]
	return ft.curr.typ
}

func (ft *contractLex) symbol() int {
	var c int

	for {
		c = ft.achievec()
		switch {
		case isEof(c):
			return c
		case isSpace(c):
		case isComment(c):
			ft.skipComment()
		case isNewline(c):
			ft.skipNewline()
		case isAlpha(c):
			return ft.achieveWord(c)
		case isDigit(c):
			return ft.achieveDigit(c)
		case isAmbiguous(c):
			ft.curr.typ = ft.ambigousSymbol(c)
			return ft.curr.typ
		case isSingleQuotation(c):
			return ft.achieveLiteral(c, CHAR_CONSTANT)
		case isDoubleQuotation(c):
			return ft.achieveLiteral(c, STRING_CONSTANT)
		default:
			ft.curr.typ = charTypeList[c&_LIM]
			return ft.curr.typ
		}
	}
}

func NewLex(fd io.Reader) *contractLex {
	return &contractLex{
		fd:   fd,
		curr: &contractWord{},
		buffer: &contractLexBuffer{
			buffer: make([]byte, BUFFSIZE),
		},
		rollBuffer: &contractLexRollBuffer{
			buffer: make([]int, ROLLSIZE),
		},
	}
}
