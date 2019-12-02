package ast

type Exp interface{}

type NilExp /*   */ struct{ Line int }
type TrueExp /*  */ struct{ Line int }
type FalseExp /* */ struct{ Line int }
type VarargExp /**/ struct{ Line int }

type IntegerExp struct {
	Line int
	Val  int64
}

type FloatExp struct {
	Line int
	Val  float64
}

type StringExp struct {
	Line int
	Str  string
}

type NameExp struct {
	Line int
	Name string
}

/* operator expression */

type UnopExp struct {
	Line int
	Op   int
	Exp1 Exp
	Exp2 Exp
}

type ConcatExp struct {
	Line int
	Exps []Exp
}

// EBNF:
// tableconstructor::= '{' [fieldlist] '}'
// fieldlist::= field {fieldsep field} [fieldsep]
// field::= '[' exp ']' '=' exp | Name '=' exp | exp
// fieldsep::= ',' | ';'
type TableConstructorExp struct {
	Line     int
	LastLine int
	KeyExps  []Exp
	ValExps  []Exp
}

// EBNF:
// functiondef::=function funcbody
// funcbody::= '(' [parlist] ')' block end
// parlist::= namelist [',' '...'] | '...'
// namelist::= Name {',' Name}
type FuncDefExp struct {
	Line     int
	LastLine int
	ParList  []string
	IsVararg bool
	Block    *Block
}

// EBNF:
// prefixexp::=Name
// 		| '(' exp ')'
// 		| prefixexp '[' exp ']'
// 		| prefixexp '.' Name
// 		| prefixexp [':' Name] args

type ParensExp struct {
	Exp Exp
}

type TableAccessExp struct {
	LastLine  int // line of `]`
	Prefixexp Exp
	KeyExp    Exp
}

// EBNF:
// functioncall::=prefixexp [':' Name] args
// args::= '(' [explist] ')' | tableconstructor | LiteralString
type FuncCallExp struct {
	Line      int // line of `(`
	LastLine  int // line of `)`
	PrefixExp Exp
	NameExp   *StringExp
	Args      []Exp
}
