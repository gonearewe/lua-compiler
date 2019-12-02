package parser

import (
	. "github.com/gonearewe/lua-compiler/compiler/ast"
	. "github.com/gonearewe/lua-compiler/compiler/lexer"
)

/*
stat ::=  ‘;’
	| break
	| ‘::’ Name ‘::’
	| goto Name
	| do block end
	| while exp do block end
	| repeat block until exp
	| if exp then block {elseif exp then block} [else block] end
	| for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end
	| for namelist in explist do block end
	| function funcname funcbody
	| local function Name funcbody
	| local namelist [‘=’ explist]
	| varlist ‘=’ explist
	| functioncall
*/
func parseStat(lexer *Lexer) Stat {
	switch lexer.LookAhead() {
	case TOKEN_SEP_SEMI:
		return parseEmptyStat(lexer)
	case TOKEN_KW_BREAK:
		return parseBreakStat(lexer)
	case TOKEN_SEP_LABEL:
		return parseLabelStat(lexer)
	case TOKEN_KW_GOTO:
		return parseGotoStat(lexer)
	case TOKEN_KW_DO:
		return parseDoStat(lexer)
	case TOKEN_KW_WHILE:
		return parseWhileStat(lexer)
	case TOKEN_KW_REPEAT:
		return parseRepeatStat(lexer)
	case TOKEN_KW_IF:
		return parseIfStat(lexer)
	case TOKEN_KW_FOR:
		return parseForStat(lexer)
	case TOKEN_KW_FUNCTION:
		return parseFuncDefStat(lexer)
	case TOKEN_KW_LOCAL:
		return parseLocalAssignOrFuncDefStat(lexer)
	default:
		return parseAssignOrFuncCallStat(lexer)
	}
}

// `;`
func parseEmptyStat(lexer *Lexer) *EmptyStat {
	lexer.NextTokenOfKind(TOKEN_SEP_SEMI)
	return &EmptyStat{}
}

// `break`
func parseBreakStat(lexer *Lexer) *BreakStat {
	lexer.NextTokenOfKind(TOKEN_KW_BREAK)
	return &BreakStat{lexer.Line()}
}

// `::label_name::`
func parseLableStat(lexer *Lexer) *LabelStat {
	lexer.NextTokenOfKind(TOKEN_SEP_LABEL) // `::`
	_, name := lexer.NextIdentifier()      // Name
	lexer.NextTokenOfKind(TOKEN_SEP_LABEL) // `::`

	return &LabelStat{name}
}

// `goto label_name`
func parseGotoStat(lexer *Lexer) *GotoStat {
	lexer.NextTokenOfKind(TOKEN_KW_GOTO) // `goto`
	_, name := lexer.NextIdentifier()    // Name

	return &GotoStat{name}
}

// `do block end`
func parseDoStat(lexer *Lexer) *DoStat {
	lexer.NextTokenOfKind(TOKEN_KW_DO)  // `do`
	block := parseBlock(lexer)          // block
	lexer.NextTokenOfKind(TOKEN_KW_END) // `end`

	return &DoStat{block}
}

// `while exp do block end`
func parseWhileStat(lexer *Lexer) *WhileStat {
	lexer.NextTokenOfKind(TOKEN_KW_WHILE) // `while`
	exp := parseExp(lexer)                // exp
	lexer.NextTokenOfKind(TOKEN_KW_DO)    // `do`
	block := parseBlock(lexer)            // block
	lexer.NextTokenOfKind(TOKEN_KW_END)   // `end`

	return &RepeatStat(block, exp)

}

// `if exp then block {elseif exp then block} [else block] end`
func parseIfStat(lexer *Lexer) *IfStat {
	exps := make([]Exp, 0, 4)
	blocks := make([]*Block, 0, 4)

	lexer.NextTokenOfKind(TOKEN_KW_IF)         // `if`
	exps = append(exps, parseExp(lexer))       // exp
	lexer.NextTokenOfKind(TOKEN_KW_THEN)       // `then`
	blocks = append(blocks, parseBlock(lexer)) // block

	for lexer.LookAhead() == TOKEN_KW_ELSEIF {
		lexer.NextToken()                          // `else if`
		exps = append(exps, parseExp(lexer))       // exp
		lexer.NextTokenOfKind(TOKEN_KW_THEN)       // `then`
		blocks = append(blocks, parseBlock(lexer)) // block
	}

	// else block => elseif true then block
	if lexer.LookAhead() == TOKEN_KW_ELSE {
		lexer.NextToken() // else
		exps = append(exps, &TrueExp{lexer.Line()})
		blocks = append(blocks, parseBlock(lexer)) // block
	}

	lexer.NextTokenOfKind(TOKEN_KW_END) // end

	return &IfStat{exps, blocks}
}

func parseForStat(lexer *Lexer) Stat {
	lineOfFor, _ := lexer.NextTokenOfKind(TOKEN_KW_FOR)
	_, name := lexer.NextIdentifier()
	if lexer.LookAhead() == TOKEN_OP_ASSIGN {
		return _finishForNumStat(lexer, lineOfFor, name)
	} else {
		return _finishForInStat(lexer, name)
	}
}

func _finishForNumStat(lexer *Lexer, lineOfFor int, name string) *ForNumStat {
	lexer.NextTokenOfKind(TOKEN_OP_ASSIGN) // `=`
	initExp := parseExp(lexer)             // exp
	lexer.NextTokenOfKind(TOKEN_SEP_COMMA) // `,`
	limitExp := parseExp(lexer)            // exp

	var stepExp Exp
	if lexer.LookAhead() == TOKEN_SEP_COMMA {
		lexer.NextToken() // `,`
		stepExp = parseExp(lexer)
	} else {
		stepExp = &IntegerExp{lexer.Line(), 1} // default step is 1
	}

	lineOfDo, _ := lexer.NextTokenOfKind(TOKEN_KW_DO) // `do`
	block := parseBlock(lexer)                        // block
	lexer.NextTokenOfKind(TOKEN_KW_END)               // `end`

	return &ForNumStat{lineOfFor, lineOfDo, varName, initExp, limitExp, stepExp, block}
}

func _finishForInStat(lexer *Lexer, name0 string) *ForInStat {
	nameList := _finishNameList(lexer, name0)         // namelist
	lexer.NextTokenOfKind(TOKEN_KW_IN)                // `in`
	explist := parseExpList(lexer)                    // explist
	lineOfDo, _ := lexer.NextTokenOfKind(TOKEN_KW_DO) // do
	block := parseBlock(lexer)                        // block
	lexer.NextTokenOfKind(TOKEN_KW_END)               // `end`

	return &ForInStat{lineOfDo, nameList, explist, block}
}

func _finishNameList(lexer *Lexer, name0 string) []string {
	names := []string{name0} // Name
	for lexer.LookAhead() == TOKEN_SEP_COMMA {
		lexer.NextToken()                 // `,`
		_, name := lexer.NextIdentifier() // Name
		names = append(names, name)
	}

	return names

}
