package contract

const (
	SUB_OFF  = 8
	TYP_MASK = 0xFF
	SUB_MASK = 0xFF00
)

const (
	O_RAM   = 0x01
	O_FLASH = 0x02
)

const (
	O_MAP = 0xFF
)

const (
	O_INT    = 0x01
	O_CHAR   = 0x02
	O_BOOL   = 0x03
	O_FLOAT  = 0x04
	O_STRING = 0x05
)

const (
	O_VAR    = 0x01
	O_SCS    = 0x02
	O_FUNC   = 0x03
	O_TYPE   = 0x04
	O_NAME   = 0x05
	O_CONST  = 0x06
	O_INC    = 0x07
	O_DEC    = 0x08
	O_MACC   = 0x09
	O_CALL   = 0x0A
	O_UNARY  = 0x0B
	O_ASSIGN = 0x0C
	O_SIZEOF = 0x0D
	O_MUL    = 0x0E
	O_DIV    = 0x0F
	O_MOD    = 0x10
	O_ADD    = 0x11
	O_SUB    = 0x12
	O_LEFT   = 0x13
	O_RIGHT  = 0x14
	O_LT     = 0x15
	O_GT     = 0x16
	O_LE     = 0x17
	O_GE     = 0x18
	O_EQ     = 0x19
	O_NE     = 0x1A
	O_AND    = 0x1B
	O_XOR    = 0x1C
	O_OR     = 0x1D
	O_LAND   = 0x1E
	O_LOR    = 0x1F
	O_TUPLE  = 0x20
	O_ELIST  = 0x21
	O_EXPST  = 0x22
	O_IF     = 0x23
	O_WHILE  = 0x24
	O_BREAK  = 0x25
	O_CONT   = 0x26
	O_RTN    = 0x27
	O_BLOCK  = 0x28
	O_CPDS   = 0x29
	O_CLASS  = 0x2A
	O_PMLST  = 0x2B
)

const (
	UnDefined                = 0x00
	VariableDefinition       = 0x01
	FunctionDefinition       = 0x02
	Declarator               = 0x04
	TypeSpecifier            = 0x08
	CompoundStatement        = 0x10
	BlockItem                = 0x20
	ParameterList            = 0x40
	Statement                = 0x80
	JumpStatement            = 0x100
	IterationStatement       = 0x200
	SelectionStatement       = 0x400
	ExpressionStatement      = 0x800
	Expression               = 0x1000
	AssignmentExpression     = 0x2000
	UnaryExpression          = 0x4000
	AssignmentOperator       = 0x8000
	UnaryOperator            = 0x10000
	PostfixExpression        = 0x20000
	PrimaryExpression        = 0x40000
	ArgumentExpressionList   = 0x80000
	LogicalOrExpression      = 0x100000
	LogicalAndExpression     = 0x200000
	InclusiveOrExpression    = 0x400000
	ExclusiveOrExpression    = 0x800000
	AndExpression            = 0x1000000
	XhualityExpression       = 0x2000000
	RelationalExpression     = 0x4000000
	ShiftExpression          = 0x8000000
	AdditiveExpression       = 0x10000000
	MultiplicativeExpression = 0x20000000
)

type contractParser struct {
	ft         *contractLex
	ns         []*contractNode
	nst        *contractNameList
	rollBuffer *contractParserRollBuffer
}

type contractParserRollBuffer struct {
	cnt    int
	buffer []*contractWord
}

type contractNode struct {
	op       int
	left     *contractNode
	right    *contractNode
	middle   *contractNode
	value    interface{}
	nameList *contractNameList
}

type contractParserFunc (func(*contractParser) (*contractNode, error))

var resolverRegistry map[int]contractParserFunc

var firstList = []int{
	BlockItem | Statement | Declarator | ParameterList | Expression | ExpressionStatement |
		UnaryExpression | PrimaryExpression,
	BlockItem | Statement | Expression | ExpressionStatement | UnaryExpression |
		PrimaryExpression,
	BlockItem | Statement | Expression | ExpressionStatement | UnaryExpression |
		PrimaryExpression,
	BlockItem | Statement | Expression | ExpressionStatement | UnaryExpression |
		PrimaryExpression,
	BlockItem | Statement | Expression | ExpressionStatement | UnaryExpression |
		PrimaryExpression,
	BlockItem | Statement | Expression | ExpressionStatement | UnaryExpression |
		PrimaryExpression,
	UnDefined,
	UnDefined,
	UnDefined,
	UnDefined,
	UnDefined,
	UnDefined,
	UnDefined,
	UnDefined,
	UnDefined,
	UnDefined,
	AssignmentOperator,
	AssignmentOperator,
	AssignmentOperator,           
	AssignmentOperator,             
	AssignmentOperator,             
	AssignmentOperator,          
	AssignmentOperator,           
	AssignmentOperator,           
	AssignmentOperator,            
	AssignmentOperator,            
	TypeSpecifier,                  
	TypeSpecifier,                 
	TypeSpecifier,                  
	
	BlockItem | Statement | SelectionStatement, 
	UnDefined, 
	BlockItem | Statement | IterationStatement,                                 
	
	BlockItem | Statement | Expression | ExpressionStatement | UnaryExpression |
	
}
