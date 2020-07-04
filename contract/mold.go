package contract

	_ODG = 0x02 
	_XDG = 0x04 
	_PHA = 0x08
	_SPC = 0x10
	_AMP = 0x20  
	_NLN = 0x40  
	_CMT = 0x80  
	_SQU = 0x100 
	_DQU = 0x200 
	_LIM = 0xFF
)

const (
	_HT = _SPC
	_VT = _SPC
	_FF = _SPC
	_LF = _NLN
	_CR = _NLN
)

var charList = []int{

	0, 0, 0, 0, 0, 0, 0, 0,

	0, _HT, _LF, _VT, _FF, _CR, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	_SPC, _AMP, _DQU, _CMT, 0, _AMP, _AMP, _SQU,

	0, 0, _AMP, _AMP, 0, _AMP, 0, _AMP,

	_ODG | _DDG | _XDG, _ODG | _DDG | _XDG, _ODG | _DDG | _XDG,
	_ODG | _DDG | _XDG, _ODG | _DDG | _XDG, _ODG | _DDG | _XDG,
	_ODG | _DDG | _XDG, _ODG | _DDG | _XDG,

	_DDG | _XDG, _DDG | _XDG, 0, 0, _AMP, _AMP, _AMP, 0,

	0, _PHA | _XDG, _PHA | _XDG, _PHA | _XDG, _PHA | _XDG, _PHA | _XDG, _PHA | _XDG, _PHA,

	_PHA, _PHA, _PHA, _PHA, _PHA, _PHA, _PHA, _PHA,

	_PHA, _PHA, _PHA, _PHA, _PHA, _PHA, _PHA, _PHA,

	_PHA, _PHA, _PHA, 0, 0, 0, _AMP, _PHA,

	0, _PHA | _XDG, _PHA | _XDG, _PHA | _XDG, _PHA | _XDG, _PHA | _XDG, _PHA | _XDG, _PHA,

	_PHA, _PHA, _PHA, _PHA, _PHA, _PHA, _PHA, _PHA,

	_PHA, _PHA, _PHA, _PHA, _PHA, _PHA, _PHA, _PHA,

	_PHA, _PHA, _PHA, 0, _AMP, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,
}

func isEof(c int) bool {
	return c == EOF
}

func isSpace(c int) bool {
	return (charList[c&_LIM] & _SPC) != 0
}
func isDigit(c int) bool {
	return (charList[c&_LIM] & _DDG) != 0
}

func isOdigit(c int) bool {
	return (charList[c&_LIM] & _ODG) != 0
}
func isXdigit(c int) bool {
	return (charList[c&_LIM] & _XDG) != 0
}

func isAlpha(c int) bool {
	return (charList[c&_LIM] & _PHA) != 0
}

func isAlnum(c int) bool {
	return (charList[c&_LIM] & (_PHA | _DDG)) != 0
}

func isNewline(c int) bool {
	return (charList[c&_LIM] & _NLN) != 0
}

func isComment(c int) bool {
	return (charList[c&_LIM] & _CMT) != 0
}

func isSingleQuotation(c int) bool {
	return (charList[c&_LIM] & _SQU) != 0
}

func isDoubleQuotation(c int) bool {
	return (charList[c&_LIM] & _DQU) != 0
}

func isAmbiguous(c int) bool {
	return (charList[c&_LIM] & _AMP) != 0
}
