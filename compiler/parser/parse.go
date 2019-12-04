package parser

import (
	."github.com/gonearewe/lua-compiler/compiler/lexer"
	."github.com/gonearewe/lua-compiler/compiler/ast"
)

// Parse given source whose file name is also given.
func Parse(chunk,chunkName string)*Block{
	lexer:=NewLexer(chunk, chunkName)
	block:=parseBlock(lexer)
	lexer.NextTokenOfKind(TOKEN_EOF) // make sure all source is parsed

	return block
}