/* statement*/
package ast

type Stat interface{}

type EmptyStat /**/ struct{}               // ;
type BreakStat /**/ struct{ Line int }     // break
type LabelStat /**/ struct{ Name string }  // `::`Name`::`
type GotoStat /* */ struct{ Name string }  // goto Name
type DoStat /*   */ struct{ Block *Block } // do block end

type FuncCallStat = FuncCallExp // function call, both statement and expression

// EBNF: while exp do block end
type WhileStat struct {
	Exp   Exp
	Block *Block
}

// EBNF: repeat block until exp
type RepeatStat struct {
	Block *Block
	Exp   Exp
}

// simplified EBNF: if exp then block {elseif exp then block} end
type IfStat struct {
	// index 0 contains if-then, others contain elseif-then
	Exps   []Exp
	Blocks []*Block
}

// EBNF: for Name '=' exp ',' exp [',' exp] do block end
type ForNumStat struct {
	LineOfFor int
	LineOfDo  int
	VarName   string

	InitExp  Exp
	LimitExp Exp
	StepExp  Exp

	Block *Block
}

// EBNF:
// for namelist in explist do block end
// namelist::= Name {',' Name}
// explist::= exp {',' exp}
type ForInStat struct {
	LineOfDo int
	NameList []string
	ExpList  []Exp
	Block    *Block
}

// EBNF:
// local namelist ['=' explist]
// namelist::=Name{',' Name}
// explist::=exp {',' exp}
type LocalVarDeclStat struct {
	LastLine int
	NameList []string
	ExpList  []Exp
}

// EBNF:
// varlist '=' explist
// varlist::= var{',' var}
// var::= Name | prefixexp '{' exp '}' | prefixexp '.' Name
// explist::= exp {',' exp}
type AssignStat struct {
	LastLine int
	VarList  []Exp
	ExpList  []Exp
}

// EBNF:
// function funcname funcbody
// funcname::= Name {',' Name} [':' Name]
// funcbody::= '(' [parlist] ')' block end
// parlist::= namelist [',' '...'] | '...'
// namelist::= Name {',' Name}

// EBNF:
// local function Name funcbody
type LocalFuncDefStat struct {
	Name string
	Exp  *FuncDefExp
}
