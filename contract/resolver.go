package contract

import (
	"errors"
	"fmt"
	"io"

	"kortho/contract/motor"
)

func init() {
	resolverRegistry = map[int]contractParserFunc{
		VariableDefinition:       resolver0,
		FunctionDefinition:       resolver1,
		Declarator:               resolver2,
		TypeSpecifier:            resolver3,
		CompoundStatement:        resolver4,
		BlockItem:                resolver5,
		ParameterList:            resolver6,
		Statement:                resolver7,
		JumpStatement:            resolver8,
		IterationStatement:       resolver9,
		SelectionStatement:       resolver10,
		ExpressionStatement:      resolver11,
		Expression:               resolver12,
		AssignmentExpression:     resolver13,
		UnaryExpression:          resolver14,
		AssignmentOperator:       resolver15,
		UnaryOperator:            resolver16,
		PostfixExpression:        resolver17,
		PrimaryExpression:        resolver18,
		ArgumentExpressionList:   resolver19,
		LogicalOrExpression:      resolver20,
		LogicalAndExpression:     resolver21,
		InclusiveOrExpression:    resolver22,
		ExclusiveOrExpression:    resolver23,
		AndExpression:            resolver24,
		XhualityExpression:       resolver25,
		RelationalExpression:     resolver26,
		ShiftExpression:          resolver27,
		AdditiveExpression:       resolver28,
		MultiplicativeExpression: resolver29,
		UnDefined:                resolverAbort,
	}
}

func (fr *contractParserRollBuffer) putw(w *contractWord) *contractWord {
	for i := ROLLSIZE - 1; i < 0; i-- {
		fr.buffer[i] = fr.buffer[i-1]
	}
	fr.buffer[0] = w
	return w
}

func (fr *contractParserRollBuffer) achievew() *contractWord {
	if fr.cnt > 0 {
		fr.cnt--
		return fr.buffer[fr.cnt]
	}
	return nil
}

func (fr *contractParserRollBuffer) unachievew() {
	if fr.cnt < ROLLSIZE {
		fr.cnt++
	}
}

func (fp *contractParser) unachievew() {
	fp.rollBuffer.unachievew()
}

func (fp *contractParser) achievew() *contractWord {
	if w := fp.rollBuffer.achievew(); w != nil {
		return w
	}
	switch o := fp.ft.symbol(); o {
	case EOF:
		return nil
	case ERR:
		panic("Abort")
	}
	return fp.rollBuffer.putw(&contractWord{
		typ:   fp.ft.curr.typ,
		name:  fp.ft.curr.name,
		value: fp.ft.curr.value,
	})
}

func (fp *contractParser) peek() int {
	if w := fp.achievew(); w != nil {
		fp.unachievew()
		return w.typ
	}
	return EOF
}

func (fp *contractParser) detect0(typ int) bool {
	return fp.peek() == typ
}

func (fp *contractParser) detect1(grm int) bool {
	if typ := fp.peek(); typ == EOF {
		return false
	} else {
		return firstList[typ]&grm != 0
	}
}

func resolverAbort(fp *contractParser) (*contractNode, error) {
	return nil, errors.New("Abort")
}

func resolver0(fp *contractParser) (*contractNode, error) {
	var err error
	var n1, n2, n3 *contractNode

	n1 = &contractNode{
		op: O_CLASS,
	}
	switch {
	case fp.detect0(LET):
		n1.op |= O_RAM << SUB_OFF
	case fp.detect0(SET):
		n1.op |= O_FLASH << SUB_OFF
	}
	fp.achievew()
	if !fp.detect1(Declarator) {
		return nil, errors.New("Expect Declarator")
	}
	if n2, err = resolverRegistry[Declarator](fp); err != nil {
		return nil, err
	}
	if !fp.detect1(TypeSpecifier) {
		return nil, errors.New("Expect TypeSpecifier")
	}
	if n3, err = resolverRegistry[TypeSpecifier](fp); err != nil {
		return nil, err
	}
	n2.right = n3
	n2.middle = n1
	n2.nameList = fp.nst
	name, _ := n2.value.(string)
	if err = fp.nst.insert(name, &contractSymbol{}); err != nil {
		return nil, err
	}
	return n2, nil

}

func resolver1(fp *contractParser) (*contractNode, error) {
	var err error
	var n1, n2, n3 *contractNode

	if fp.nst.up != nil {
		return nil, errors.New("Nested function definition")
	}
	if !fp.detect0(IDENTIFIER) {
		return nil, errors.New("Expect IDENTIFIER")
	}
	if state, _ := fp.nst.lookUp(name); state == EXT_CUR {
		return nil, fmt.Errorf("Function %s Exist", name)
	}
	if !fp.detect0(BRA) {
		return nil, errors.New("Expect (")
	}
	np := fp.nst
	if fp.detect1(ParameterList) {
		fp.nst = fp.nst.recentNL()
		if n1, err = resolverRegistry[ParameterList](fp); err != nil {
			return nil, err
		}
	}
	if !fp.detect0(KET) {
		return nil, errors.New("Expect )")
	}
	if !fp.detect1(TypeSpecifier) {
		return nil, errors.New("Expect TypeSpecifier")
	}
	if n2, err = resolverRegistry[TypeSpecifier](fp); err != nil {
		return nil, err
	}
	if !fp.detect1(CompoundStatement) {
		return nil, errors.New("Expect {")
	}
	if n3, err = resolverRegistry[CompoundStatement](fp); err != nil {
		return nil, err
	}
	fp.nst = np
	if err = fp.nst.insert(name, &contractSymbol{typ: S_FUNC}); err != nil {
		return nil, err
	}
	return &contractNode{
		left:   n1,
		right:  n2,
		middle: n3,
		value:  name,
		op:     O_FUNC,
	}, nil
}

func resolver2(fp *contractParser) (*contractNode, error) {
	if state, _ := fp.nst.lookUp(name); state == EXT_CUR {
		return nil, fmt.Errorf("Symbol '%s' Exist", name)
	}
	node := &contractNode{
		value: name,
		op:    O_VAR,
	}
	if !fp.detect0(SBR) {
		return node, nil
	}

	if fp.detect0(SBR) {
		var err error
		var n1 *contractNode
		node.op |= O_MAP << SUB_OFF
		fp.achievew() // skip [
		if !fp.detect1(TypeSpecifier) {
			return nil, errors.New("Expect Type Specifier")
		}
		if n1, err = resolverRegistry[TypeSpecifier](fp); err != nil {
			return nil, err
		}
		node.left = n1
		if !fp.detect0(SKT) {
			return nil, errors.New("Expect ]")
		}
	}
	return node, nil
}

func resolver3(fp *contractParser) (*contractNode, error) {
	typ := 0
	switch fp.achievew().typ {
	case INT8:
		typ = motor.INT8
	case INT16:
		typ = motor.INT16
	case INT32:
		typ = motor.INT32
	case INT64:
		typ = motor.INT64
	case UINT8:
		typ = motor.UINT8
	case UINT16:
		typ = motor.UINT16
	case UINT32:
		typ = motor.UINT32
	case UINT64:
		typ = motor.UINT64
	case FLOAT32:
		typ = motor.FLOAT32
	case FLOAT64:
		typ = motor.FLOAT64
	case BOOL:
		typ = motor.BOOL
	case CHAR:
		typ = motor.CHAR
	case STRING:
		typ = motor.STRING
	default:
		return nil, errors.New("Unsupport Type")
	}
	return &contractNode{
		value: typ,
		op:    O_TYPE,
	}, nil
}

func resolver4(fp *contractParser) (*contractNode, error) {
	var err error
	var n1, node *contractNode

	fp.nst = fp.nst.recentNL()
	if fp.detect1(BlockItem) {
		if n1, err = resolverRegistry[BlockItem](fp); err != nil {
			goto OUT
		}
	}
	if !fp.detect0(CKT) {
		err = errors.New("Expect }")
		goto OUT
	}
	node = &contractNode{
		left:     n1,
		op:       O_CPDS,
		nameList: fp.nst,
	}
OUT:
	fp.nst = fp.nst.up
	return node, err
}

func resolver5(fp *contractParser) (*contractNode, error) {
	var err error
	var n1 *contractNode

	node := &contractNode{
		op: O_BLOCK,
	}
	for np := &node; ; np = &(*np).middle {
		if *np == nil {
			*np = &contractNode{
				op: O_BLOCK,
			}
		}
		switch {
		case fp.detect1(Statement):
			if n1, err = resolverRegistry[Statement](fp); err != nil {
				goto OUT
			}
			(*np).left = n1
		case fp.detect1(VariableDefinition):
			if n1, err = resolverRegistry[VariableDefinition](fp); err != nil {
				goto OUT
			}
			(*np).left = n1
		default:
			goto OUT
		}
	}
OUT:
	return node, err
}

func resolver6(fp *contractParser) (*contractNode, error) {
	var err error
	var n1, n2, n3 *contractNode

	if n1, err = resolverRegistry[Declarator](fp); err != nil {
		return nil, err
	}
	if !fp.detect1(TypeSpecifier) {
		return nil, errors.New("Expect Type Specifier")
	}
	if n2, err = resolverRegistry[TypeSpecifier](fp); err != nil {
		return nil, err
	}
	name, _ := n1.value.(string)
	if err = fp.nst.insert(name, &contractSymbol{}); err != nil {
		return nil, err
	}
	if !fp.detect0(COM) {
		return &contractNode{
			left:     n1,
			right:    n2,
			nameList: fp.nst,
			op:       O_PMLST,
		}, nil
	}
	fp.achievew()
	if n3, err = resolverRegistry[ParameterList](fp); err != nil {
		return nil, err
	}
	return &contractNode{
		left:     n1,
		right:    n2,
		middle:   n3,
		nameList: fp.nst,
		op:       O_PMLST,
	}, nil
}

func resolver7(fp *contractParser) (*contractNode, error) {
	switch {
	case fp.detect1(JumpStatement):
		return resolverRegistry[JumpStatement](fp)
	case fp.detect1(CompoundStatement):
		return resolverRegistry[CompoundStatement](fp)
	case fp.detect1(IterationStatement):
		return resolverRegistry[IterationStatement](fp)
	case fp.detect1(SelectionStatement):
		return resolverRegistry[SelectionStatement](fp)
	case fp.detect1(ExpressionStatement):
		return resolverRegistry[ExpressionStatement](fp)
	default:
		return nil, errors.New("Expect Statement")
	}
}

func resolver8(fp *contractParser) (*contractNode, error) {
	switch {
	case fp.detect0(BREAK):
		fp.achievew()
		if !fp.detect0(SEM) {
			return nil, errors.New("Expect ;")
		}
		fp.achievew()
		return &contractNode{
			op: O_BREAK,
		}, nil
	case fp.detect0(CONTINUE):
		fp.achievew()
		if !fp.detect0(SEM) {
			return nil, errors.New("Expect ;")
		}
		fp.achievew()
		return &contractNode{
			op: O_CONT,
		}, nil
	case fp.detect0(RETURN):
		var err error
		var n1 *contractNode
		fp.achievew()
		if !fp.detect1(Expression) {
			return nil, errors.New("Expect Expression")
		}
		if n1, err = resolverRegistry[Expression](fp); err != nil {
			return nil, err
		}
		if !fp.detect0(SEM) {
			return nil, errors.New("Expect ;")
		}
		fp.achievew()
		return &contractNode{
			left: n1,
			op:   O_RTN,
		}, nil
	default:
		return nil, errors.New("Expect JumpStatement")
	}
}

func resolver9(fp *contractParser) (*contractNode, error) {
	var err error
	var n1, n2 *contractNode

	fp.achievew()
	if !fp.detect0(BRA) {
		return nil, errors.New("Expect (")
	}
	fp.achievew()
	if !fp.detect1(Expression) {
		return nil, errors.New("Expect Expression")
	}
	if n1, err = resolverRegistry[Expression](fp); err != nil {
		return nil, err
	}
	if !fp.detect0(KET) {
		return nil, errors.New("Expect )")
	}
	fp.achievew()
	if !fp.detect1(Statement) {
		return nil, errors.New("Expect Statement")
	}
	if n2, err = resolverRegistry[Statement](fp); err != nil {
		return nil, err
	}
	return &contractNode{
		left:   n2,
		middle: n1,
		op:     O_WHILE,
	}, nil
}

func resolver10(fp *contractParser) (*contractNode, error) {
	var err error
	var n1, n2, n3 *contractNode

	fp.achievew()
	if !fp.detect0(BRA) {
		return nil, errors.New("Expect (")
	}
	fp.achievew()
	if !fp.detect1(Expression) {
		return nil, errors.New("Expect Expression")
	}
	if n1, err = resolverRegistry[Expression](fp); err != nil {
		return nil, err
	}
	if !fp.detect0(KET) {
		return nil, errors.New("Expect )")
	}
	fp.achievew()
	if !fp.detect1(Statement) {
		return nil, errors.New("Expect Statement")
	}
	if n2, err = resolverRegistry[Statement](fp); err != nil {
		return nil, err
	}
	if fp.detect0(ELSE) {
		fp.achievew()
		if !fp.detect1(Statement) {
			return nil, errors.New("Expect Statement")
		}
		if n3, err = resolverRegistry[Statement](fp); err != nil {
			return nil, err
		}
	}
	return &contractNode{
		left:   n2,
		right:  n3,
		middle: n1,
		op:     O_IF,
	}, nil
}

func resolver11(fp *contractParser) (*contractNode, error) {
	var err error
	var n1 *contractNode

	if fp.detect1(Expression) {
		if n1, err = resolverRegistry[Expression](fp); err != nil {
			return nil, err
		}
	}
	if !fp.detect0(SEM) {
		return nil, errors.New("Expect ;")
	}
	return &contractNode{
		left: n1,
		op:   O_EXPST,
	}, nil
}

func resolver12(fp *contractParser) (*contractNode, error) {
	var err error
	var n1, n2 *contractNode

	if n1, err = resolverRegistry[AssignmentExpression](fp); err != nil {
		return nil, err
	}
	if !fp.detect0(COM) {
		return n1, nil
	}
	if !fp.detect1(Expression) {
		return nil, errors.New("Expect Expression")
	}
	if n2, err = resolverRegistry[Expression](fp); err != nil {
		return nil, err
	}
	return &contractNode{
		left:   n1,
		middle: n2,
		op:     O_ELIST,
	}, nil
}

func resolver13(fp *contractParser) (*contractNode, error) {
	var err error
	var n1, n2, n3 *contractNode

	if n1, err = resolverRegistry[LogicalOrExpression](fp); err != nil {
		return nil, err
	}
	switch n1.op & TYP_MASK {
	case O_SIZEOF, O_UNARY, O_MACC, O_CALL, O_INC, O_DEC, O_NAME, O_CONST, O_TUPLE:
	default:
		return n1, nil
	}
	if !fp.detect1(AssignmentOperator) {
		return n1, nil
	}
	if n2, err = resolverRegistry[AssignmentOperator](fp); err != nil {
		return nil, err
	}
	if !fp.detect1(Expression) {
		return nil, errors.New("Expect Assignment Expression")
	}
	if n3, err = resolverRegistry[AssignmentExpression](fp); err != nil {
		return nil, err
	}
	n2.left = n1
	n2.right = n3
	return n2, nil
}

func resolver14(fp *contractParser) (*contractNode, error) {
	var err error
	var n1 *contractNode

	switch {
	case fp.detect0(SIZEOF):
		fp.achievew()
		if n1, err = resolverRegistry[UnaryExpression](fp); err != nil {
			return nil, err
		}
		return &contractNode{
			left: n1,
			op:   O_SIZEOF,
		}, nil
	case fp.detect1(UnaryOperator):
		var n2 *contractNode
		if n1, err = resolverRegistry[UnaryOperator](fp); err != nil {
			return nil, err
		}
		if n2, err = resolverRegistry[UnaryExpression](fp); err != nil {
			return nil, err
		}
		n1.left = n2
		return n1, nil
	case fp.detect1(PrimaryExpression):
		return resolverRegistry[PostfixExpression](fp)
	default:
		return nil, errors.New("Expect UnaryExpression")
	}
}
func resolver15(fp *contractParser) (*contractNode, error) {
	return &contractNode{
		op:    O_ASSIGN,
		value: fp.achievew().typ,
	}, nil
}

func resolver16(fp *contractParser) (*contractNode, error) {
	return &contractNode{
		op:    O_UNARY,
		value: fp.achievew().typ,
	}, nil
}

func resolver17(fp *contractParser) (*contractNode, error) {
	var err error
	var n1, n2 *contractNode

	if n1, err = resolverRegistry[PrimaryExpression](fp); err != nil {
		return nil, err
	}
	switch {
	case fp.detect0(SBR):
		fp.achievew()
		if !fp.detect1(Expression) {
			return nil, errors.New("Expect Expression")
		}
		if n2, err = resolverRegistry[Expression](fp); err != nil {
			return nil, err
		}
		if !fp.detect0(SKT) {
			return nil, errors.New("Expect ]")
		}
		fp.achievew()
		return &contractNode{
			left:  n1,
			right: n2,
			op:    O_MACC,
		}, nil
	case fp.detect0(BRA):
		fp.achievew()
		if !fp.detect0(KET) {
			if n2, err = resolverRegistry[ArgumentExpressionList](fp); err != nil {
				return nil, err
			}
		}
		if !fp.detect0(KET) {
			return nil, errors.New("Expect )")
		}
		fp.achievew()
		return &contractNode{
			left:  n1,
			right: n2,
			op:    O_CALL,
		}, nil
	case fp.detect0(INC_OP):
		fp.achievew()
		return &contractNode{
			left: n1,
			op:   O_INC,
		}, nil
	case fp.detect0(DEC_OP):
		fp.achievew()
		return &contractNode{
			left: n1,
			op:   O_DEC,
		}, nil
	default:
		return n1, nil
	}
}

func resolver18(fp *contractParser) (*contractNode, error) {
	switch {
	case fp.detect0(BRA):
		var err error
		var n1 *contractNode
		fp.achievew()
		if !fp.detect1(Expression) {
			return nil, errors.New("Expect Expression")
		}
		if n1, err = resolverRegistry[Expression](fp); err != nil {
			return nil, err
		}
		if !fp.detect0(KET) {
			return nil, errors.New("Expect )")
		}
		fp.achievew()
		return &contractNode{
			left: n1,
			op:   O_TUPLE,
		}, nil
	case fp.detect0(IDENTIFIER):
		name := fp.achievew().name
		if state, _ := fp.nst.lookUp(name); state == NOT_EXT {
			return nil, fmt.Errorf("Cannot Find Symbol %s", name)
		}
		return &contractNode{
			value:    name,
			op:       O_NAME,
			nameList: fp.nst,
		}, nil
	case fp.detect0(INT_CONSTANT):
		return &contractNode{
			value: fp.achievew().value,
			op:    O_CONST | O_INT<<SUB_OFF,
		}, nil
	case fp.detect0(BOOL_CONSTANT):
		return &contractNode{
			value: fp.achievew().value,
			op:    O_CONST | O_BOOL<<SUB_OFF,
		}, nil
	case fp.detect0(CHAR_CONSTANT):
		return &contractNode{
			value: fp.achievew().value,
			op:    O_CONST | O_CHAR<<SUB_OFF,
		}, nil
	case fp.detect0(FLOAT_CONSTANT):
		return &contractNode{
			value: fp.achievew().value,
			op:    O_CONST | O_FLOAT<<SUB_OFF,
		}, nil
	case fp.detect0(STRING_CONSTANT):
		return &contractNode{
			value: fp.achievew().value,
			op:    O_CONST | O_STRING<<SUB_OFF,
		}, nil
	default:
		return nil, errors.New("Expect PrimaryExpression")
	}
}

func resolver19(fp *contractParser) (*contractNode, error) {
	var err error
	var n1, n2 *contractNode

	if n1, err = resolverRegistry[AssignmentExpression](fp); err != nil {
		return nil, err
	}
	if !fp.detect0(COM) {
		return n1, nil
	}
	fp.achievew()
	if !fp.detect1(Expression) {
		return nil, errors.New("Expect Expression")
	}
	if n2, err = resolverRegistry[ArgumentExpressionList](fp); err != nil {
		return nil, err
	}
	return &contractNode{
		left:   n1,
		middle: n2,
		op:     O_ELIST,
	}, nil
}

func resolver20(fp *contractParser) (*contractNode, error) {
	var err error
	var n1 *contractNode

	if n1, err = resolverRegistry[LogicalAndExpression](fp); err != nil {
		return nil, err
	}
	if !fp.detect0(OR_OP) {
		return n1, nil
	}
	fp.achievew()
	node := &contractNode{
		left: n1,
		op:   O_LOR,
	}
	if !fp.detect1(UnaryExpression) {
		return nil, errors.New("Expect LogicalOrExpression")
	}
	if node.right, err = resolverRegistry[LogicalOrExpression](fp); err != nil {
		return nil, err
	}
	return node, nil
}

func resolver21(fp *contractParser) (*contractNode, error) {
	var err error
	var n1 *contractNode

	if n1, err = resolverRegistry[InclusiveOrExpression](fp); err != nil {
		return nil, err
	}
	if !fp.detect0(AND_OP) {
		return n1, nil
	}
	fp.achievew()
	node := &contractNode{
		left: n1,
		op:   O_LAND,
	}
	if !fp.detect1(UnaryExpression) {
		return nil, errors.New("Expect LogicalAndExpression")
	}
	if node.right, err = resolverRegistry[LogicalAndExpression](fp); err != nil {
		return nil, err
	}
	return node, nil
}

func resolver22(fp *contractParser) (*contractNode, error) {
	var err error
	var n1 *contractNode

	if n1, err = resolverRegistry[ExclusiveOrExpression](fp); err != nil {
		return nil, err
	}
	if !fp.detect0(OR) {
		return n1, nil
	}
	fp.achievew()
	node := &contractNode{
		left: n1,
		op:   O_OR,
	}
	if !fp.detect1(UnaryExpression) {
		return nil, errors.New("Expect InclusiveOrExpression")
	}
	if node.right, err = resolverRegistry[InclusiveOrExpression](fp); err != nil {
		return nil, err
	}
	return node, nil
}

func resolver23(fp *contractParser) (*contractNode, error) {
	var err error
	var n1 *contractNode

	if n1, err = resolverRegistry[AndExpression](fp); err != nil {
		return nil, err
	}
	if !fp.detect0(XOR) {
		return n1, nil
	}
	fp.achievew()
	node := &contractNode{
		left: n1,
		op:   O_XOR,
	}
	if !fp.detect1(UnaryExpression) {
		return nil, errors.New("Expect ExclusiveOrExpression")
	}
	if node.right, err = resolverRegistry[ExclusiveOrExpression](fp); err != nil {
		return nil, err
	}
	return node, nil
}

func resolver24(fp *contractParser) (*contractNode, error) {
	var err error
	var n1 *contractNode

	if n1, err = resolverRegistry[XhualityExpression](fp); err != nil {
		return nil, err
	}
	if !fp.detect0(AND) {
		return n1, nil
	}
	fp.achievew()
	node := &contractNode{
		left: n1,
		op:   O_AND,
	}
	if !fp.detect1(UnaryExpression) {
		return nil, errors.New("Expect AndExpression")
	}
	if node.right, err = resolverRegistry[AndExpression](fp); err != nil {
		return nil, err
	}
	return node, nil
}

func resolver25(fp *contractParser) (*contractNode, error) {
	var err error
	var n1 *contractNode

	if n1, err = resolverRegistry[RelationalExpression](fp); err != nil {
		return nil, err
	}
	if !(fp.detect0(EQ_OP) || fp.detect0(NE_OP)) {
		return n1, nil
	}
	node := &contractNode{
		left: n1,
	}
	switch fp.achievew().typ {
	case EQ_OP:
		node.op = O_EQ
	case NE_OP:
		node.op = O_NE
	}
	if !fp.detect1(UnaryExpression) {
		return nil, errors.New("Expect XhualityExpression")
	}
	if node.right, err = resolverRegistry[XhualityExpression](fp); err != nil {
		return nil, err
	}
	return node, nil
}

func resolver26(fp *contractParser) (*contractNode, error) {
	var err error
	var n1 *contractNode

	if n1, err = resolverRegistry[ShiftExpression](fp); err != nil {
		return nil, err
	}
	if !(fp.detect0(LT) || fp.detect0(GT) || fp.detect0(LE_OP) || fp.detect0(GE_OP)) {
		return n1, nil
	}
	node := &contractNode{
		left: n1,
	}
	switch fp.achievew().typ {
	case LT:
		node.op = O_LT
	case GT:
		node.op = O_GT
	case LE_OP:
		node.op = O_LE
	case GE_OP:
		node.op = O_GE
	}
	if !fp.detect1(UnaryExpression) {
		return nil, errors.New("Expect RelationalExpression")
	}
	if node.right, err = resolverRegistry[RelationalExpression](fp); err != nil {
		return nil, err
	}
	return node, nil
}

func resolver27(fp *contractParser) (*contractNode, error) {
	var err error
	var n1 *contractNode

	if n1, err = resolverRegistry[AdditiveExpression](fp); err != nil {
		return nil, err
	}
	if !(fp.detect0(LEFT_OP) || fp.detect0(RIGHT_OP)) {
		return n1, nil
	}
	node := &contractNode{
		left: n1,
	}
	switch fp.achievew().typ {
	case LEFT_OP:
		node.op = O_LEFT
	case RIGHT_OP:
		node.op = O_RIGHT
	}
	if !fp.detect1(UnaryExpression) {
		return nil, errors.New("Expect ShiftExpression")
	}
	if node.right, err = resolverRegistry[ShiftExpression](fp); err != nil {
		return nil, err
	}
	return node, nil
}

func resolver28(fp *contractParser) (*contractNode, error) {
	var err error
	var n1 *contractNode

	if n1, err = resolverRegistry[MultiplicativeExpression](fp); err != nil {
		return nil, err
	}
	if !(fp.detect0(ADD) || fp.detect0(SUB)) {
		return n1, nil
	}
	node := &contractNode{
		left: n1,
	}
	switch fp.achievew().typ {
	case ADD:
		node.op = O_ADD
	case SUB:
		node.op = O_SUB
	}
	if !fp.detect1(UnaryExpression) {
		return nil, errors.New("Expect AdditiveExpression")
	}
	if node.right, err = resolverRegistry[AdditiveExpression](fp); err != nil {
		return nil, err
	}
	return node, nil
}

func resolver29(fp *contractParser) (*contractNode, error) {
	var err error
	var n1 *contractNode

	if n1, err = resolverRegistry[UnaryExpression](fp); err != nil {
		return nil, err
	}
	if !(fp.detect0(MUL) || fp.detect0(DIV) || fp.detect0(MOD)) {
		return n1, nil
	}
	node := &contractNode{
		left: n1,
	}
	switch fp.achievew().typ {
	case MUL:
		node.op = O_MUL
	case DIV:
		node.op = O_DIV
	case MOD:
		node.op = O_MOD
	}
	if !fp.detect1(UnaryExpression) {
		return nil, errors.New("Expect MultiplicativeExpression")
	}
	if node.right, err = resolverRegistry[MultiplicativeExpression](fp); err != nil {
		return nil, err
	}
	return node, nil
}

// it's too low
func (fp *contractParser) loadLibrary() error {
	fp.nst.insert("time", &contractSymbol{typ: S_FUNC})
	fp.nst.insert("delete", &contractSymbol{typ: S_FUNC})
	fp.nst.insert("append", &contractSymbol{typ: S_FUNC})
	fp.nst.insert("sm3Hash", &contractSymbol{typ: S_FUNC})
	fp.nst.insert("sm2Verify", &contractSymbol{typ: S_FUNC})
	fp.nst.insert("elem", &contractSymbol{typ: S_FUNC})
	return nil
}

func NewParser(fd io.Reader) *contractParser {
	ft := NewLex(fd)
	if ft == nil {
		return nil
	}
	return &contractParser{
		ft:  ft,
		nst: recentNL(nil),
		ns:  []*contractNode{},
		rollBuffer: &contractParserRollBuffer{
			cnt:    0,
			buffer: make([]*contractWord, ROLLSIZE),
		},
	}
}

func (fp *contractParser) Parser() error {
	var err error
	var node *contractNode

	if err = fp.loadLibrary(); err != nil {
		return err
	}
	for {
		switch typ := fp.peek(); typ {
		case EOF:
			return nil
		case FUNC:
			node, err = resolverRegistry[FunctionDefinition](fp)
		case LET, SET:
			node, err = resolverRegistry[VariableDefinition](fp)
		default:
			return errors.New("Expect Function Definition or Variable Definition")
		}
		if err != nil {
			return err
		}
		fp.ns = append(fp.ns, node)
	}
	return nil
}
