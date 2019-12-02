package lexer

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var reOpeningLongBracket = regexp.MustCompile(`^\[=*\[`)
var reNewLine = regexp.MustCompile("\r\n|\n\r|\n|\r")
var reShortStr = regexp.MustCompile(`(?s)(^'(\\\\|\\'|\\\n|\\z\s*|[^'\n])*')|(^"(\\\\|\\"|\\\n|\\z\s*|[^"\n])*")`)
var reIdentifier = regexp.MustCompile(`^[_\d\w]+`)
var reNumber = regexp.MustCompile(`^0[xX][0-9a-fA-F]*(\.[0-9a-fA-F]*)?([pP][+\-]?[0-9]+)?|^[0-9]*(\.[0-9]*)?([eE][+\-]?[0-9]+)?`)

var reDecEscapeSeq = regexp.MustCompile(`^\\[0-9]{1,3}`)
var reHexEscapeSeq = regexp.MustCompile(`^\\x[0-9a-fA-F]{2}`)
var reUnicodeEscapeSeq = regexp.MustCompile(`^\\u\{[0-9a-fA-F]+\}`)

type Lexer struct {
	chunk     string // remaining unparsed part of source code
	chunkName string // source file name
	line      int    // current index of line

	// cache
	nextToken     string
	nextTokenKind int
	nextTokenLine int
}

// constructor
func NewLexer(chunk, chunkName string) *Lexer {
	return &Lexer{chunk, chunkName, 1, "", 0, 0}
}

func (l *Lexer) Line() int {
	return l.line
}

// Scan next token and update cache, returns nextTokenKind.
func (l *Lexer) LookAhead() int {
	if l.nextTokenLine > 0 {
		return l.nextTokenKind
	}

	currentLine := l.line // backup current line
	line, kind, token := l.NextToken()
	l.line = currentLine // restore line index

	l.nextTokenLine, l.nextTokenKind, l.nextToken = line, kind, token

	return kind
}

func (l *Lexer) NextIdentifier() (line int, token string) {
	return l.NextTokenOfKind(TOKEN_IDENTIFIER)
}

// Returns line and token of next token if next token matches the given
// expected kind, or a syntax error will be thrown.
func (l *Lexer) NextTokenOfKind(kind int) (line int, token string) {
	line, _kind, token := l.NextToken()
	if kind != _kind {
		l.error("syntax error near '%s'", token)
	}

	return line, token
}

func (l *Lexer) NextToken() (line, kind int, token string) {
	if l.nextTokenLine > 0 {
		line, kind, token = l.nextTokenLine, l.nextTokenKind, l.nextToken
		l.line = l.nextTokenLine
		l.nextTokenLine = 0 // clear cache
		return
	}

	l.skipWhiteSpaces()
	if len(l.chunk) == 0 {
		return l.line, TOKEN_EOF, "EOF"
	}

	switch l.chunk[0] {
	case ';':
		l.next(1)
		return l.line, TOKEN_SEP_SEMI, ";"
	case ',':
		l.next(1)
		return l.line, TOKEN_SEP_COMMA, ","
	case '(':
		l.next(1)
		return l.line, TOKEN_SEP_LPAREN, "("
	case ')':
		l.next(1)
		return l.line, TOKEN_SEP_RPAREN, ")"
	case ']':
		l.next(1)
		return l.line, TOKEN_SEP_RBRACK, "]"
	case '{':
		l.next(1)
		return l.line, TOKEN_SEP_LCURLY, "{"
	case '}':
		l.next(1)
		return l.line, TOKEN_SEP_RCURLY, "}"
	case '+':
		l.next(1)
		return l.line, TOKEN_OP_ADD, "+"
	case '-':
		l.next(1)
		return l.line, TOKEN_OP_MINUS, "-"
	case '*':
		l.next(1)
		return l.line, TOKEN_OP_MUL, "*"
	case '^':
		l.next(1)
		return l.line, TOKEN_OP_POW, "^"
	case '%':
		l.next(1)
		return l.line, TOKEN_OP_MOD, "%"
	case '&':
		l.next(1)
		return l.line, TOKEN_OP_BAND, "&"
	case '|':
		l.next(1)
		return l.line, TOKEN_OP_BOR, "|"
	case '#':
		l.next(1)
		return l.line, TOKEN_OP_LEN, "#"
	case ':':
		if l.test("::") {
			l.next(2)
			return l.line, TOKEN_SEP_LABEL, "::"
		} else {
			l.next(1)
			return l.line, TOKEN_SEP_COLON, ":"
		}
	case '/':
		if l.test("//") {
			l.next(2)
			return l.line, TOKEN_OP_IDIV, "//"
		} else {
			l.next(1)
			return l.line, TOKEN_OP_DIV, "/"
		}
	case '~':
		if l.test("~=") {
			l.next(2)
			return l.line, TOKEN_OP_NE, "~="
		} else {
			l.next(1)
			return l.line, TOKEN_OP_WAVE, "~"
		}
	case '=':
		if l.test("==") {
			l.next(2)
			return l.line, TOKEN_OP_EQ, "=="
		} else {
			l.next(1)
			return l.line, TOKEN_OP_ASSIGN, "="
		}
	case '<':
		if l.test("<<") {
			l.next(2)
			return l.line, TOKEN_OP_SHL, "<<"
		} else if l.test("<=") {
			l.next(2)
			return l.line, TOKEN_OP_LE, "<="
		} else {
			l.next(1)
			return l.line, TOKEN_OP_LT, "<"
		}
	case '>':
		if l.test(">>") {
			l.next(2)
			return l.line, TOKEN_OP_SHR, ">>"
		} else if l.test(">=") {
			l.next(2)
			return l.line, TOKEN_OP_GE, ">="
		} else {
			l.next(1)
			return l.line, TOKEN_OP_GT, ">"
		}
	case '.':
		if l.test("...") {
			l.next(3)
			return l.line, TOKEN_VARARG, "..."
		} else if l.test("..") {
			l.next(2)
			return l.line, TOKEN_OP_CONCAT, ".."
		} else if len(l.chunk) == 1 || !isDigit(l.chunk[1]) {
			l.next(1)
			return l.line, TOKEN_SEP_DOT, "."
		}
	case '[':
		if l.test("[[") || l.test("[=") {
			return l.line, TOKEN_STRING, l.scanLongString()
		} else {
			l.next(1)
			return l.line, TOKEN_SEP_LBRACK, "["
		}
	case '\'', '"':
		return l.line, TOKEN_STRING, l.scanShortString()
	}

	c := l.chunk[0]
	if c == '.' || isDigit(c) {
		token := l.scanNumber()
		return l.line, TOKEN_NUMBER, token
	}
	if c == '_' || isLetter(c) {
		token := l.scanIdentifier()
		if kind, found := keywords[token]; found {
			return l.line, kind, token // keyword
		} else {
			return l.line, TOKEN_IDENTIFIER, token
		}
	}

	l.error("unexpected symbol near %q", c)
	return
}

// Extract a long string from the chunk.
func (l *Lexer) scanLongString() string {
	openingLongBracket := reOpeningLongBracket.FindString(l.chunk)
	if openingLongBracket == "" {
		l.error("invalid long string delimiter near '%s'", l.chunk[0:2])
	}

	closingLongBracket := strings.Replace(openingLongBracket, "[", "]", -1)
	closingLongBracketIdx := strings.Index(l.chunk, closingLongBracket)
	if closingLongBracketIdx < 0 {
		l.error("unfinished long string or comment")
	}

	str := l.chunk[len(openingLongBracket):closingLongBracketIdx]
	l.next(closingLongBracketIdx + len(closingLongBracket))
	str = reNewLine.ReplaceAllString(str, "\n") // replace all line breaks with "\n"
	l.line += strings.Count(str, "\n")
	if len(str) > 0 && str[0] == '\n' { // first line break in the string is ignored
		str = str[1:]
	}

	return str
}

func (l *Lexer) scanShortString() string {
	if str := reShortStr.FindString(l.chunk); str != "" {
		l.next(len(str))
		str = str[1 : len(str)-1]
		if strings.Index(str, `\`) >= 0 {
			l.line += len(reNewLine.FindAllString(str, -1))
			str = l.escape(str)
		}

		return str
	}

	l.error("unfinished string")
	return ""
}

func (l *Lexer) scanNumber() string {
	return l.scan(reNumber)
}

func (l *Lexer) scanIdentifier() string {
	return l.scan(reIdentifier)
}

func (l *Lexer) scan(re *regexp.Regexp) string {
	if token := re.FindString(l.chunk); token != "" {
		l.next(len(token))
		return token
	}

	panic("unreachable !")
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// Skip following white spaces, line breaks and comments.
func (l *Lexer) skipWhiteSpaces() {
	for len(l.chunk) > 0 {
		if l.test("--") {
			l.skipComment()
		} else if l.test("\r\n") || l.test("\n\r") {
			l.next(2)
			l.line += 1
		} else if isNewLine(l.chunk[0]) {
			l.next(1)
			l.line += 1
		} else if isWhiteSpace(l.chunk[0]) {
			l.next(1)
		} else {
			break
		}
	}
}

func (l *Lexer) skipComment() {
	l.next(2)        // skip "--"
	if l.test("[") { // long comment
		if reOpeningLongBracket.FindString(l.chunk) != "" {
			l.scanLongString()
			return
		}
	}

	// short comment
	for len(l.chunk) > 0 && !isNewLine(l.chunk[0]) {
		l.next(1)
	}
}

// Whether following chunk starts with given string.
func (l *Lexer) test(s string) bool {
	return strings.HasPrefix(l.chunk, s)
}

// Skip given number of rune in the chunk.
func (l *Lexer) next(n int) {
	l.chunk = l.chunk[n:]
}

func isWhiteSpace(c byte) bool {
	switch c {
	case ' ', '\n', '\r', '\t', '\v', '\f':
		return true
	}

	return false
}

func isNewLine(c byte) bool {
	return c == '\r' || c == '\n'
}

func (l *Lexer) escape(str string) string {
	var buf bytes.Buffer

	for len(str) > 0 {
		if str[0] != '\\' {
			buf.WriteByte(str[0])
			str = str[1:]
			continue
		}

		if len(str) == 1 {
			l.error("unfinished string")
		}

		switch str[1] {
		case 'a':
			buf.WriteByte('\a')
			str = str[2:]
			continue
		case 'b':
			buf.WriteByte('\b')
			str = str[2:]
			continue
		case 'f':
			buf.WriteByte('\f')
			str = str[2:]
			continue
		case 'n', '\n':
			buf.WriteByte('\n')
			str = str[2:]
			continue
		case 'r':
			buf.WriteByte('\r')
			str = str[2:]
			continue
		case 't':
			buf.WriteByte('\t')
			str = str[2:]
			continue
		case 'v':
			buf.WriteByte('\v')
			str = str[2:]
			continue
		case '"':
			buf.WriteByte('"')
			str = str[2:]
			continue
		case '\'':
			buf.WriteByte('\'')
			str = str[2:]
			continue
		case '\\':
			buf.WriteByte('\\')
			str = str[2:]
			continue
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // \ddd
			if found := reDecEscapeSeq.FindString(str); found != "" {
				d, _ := strconv.ParseInt(found[1:], 10, 32)
				if d <= 0xFF {
					buf.WriteByte(byte(d))
					str = str[len(found):]
					continue
				}
				l.error("decimal escape too large near '%s'", found)
			}
		case 'x': // \xXX
			if found := reHexEscapeSeq.FindString(str); found != "" {
				d, _ := strconv.ParseInt(found[2:], 16, 32)
				buf.WriteByte(byte(d))
				str = str[len(found):]
				continue
			}
		case 'u': // \u{XXX}
			if found := reUnicodeEscapeSeq.FindString(str); found != "" {
				d, err := strconv.ParseInt(found[3:len(found)-1], 16, 32)
				if err == nil && d <= 0x10FFFF {
					buf.WriteRune(rune(d))
					str = str[len(found):]
					continue
				}
				l.error("UTF-8 value too large near '%s'", found)
			}
		case 'z':
			str = str[2:]
			for len(str) > 0 && isWhiteSpace(str[0]) { // todo
				str = str[1:]
			}
			continue
		}
		l.error("invalid escape sequence near '\\%c'", str[1])
	}

	return buf.String()
}

// Panic with given params and chunk file name, line index.
func (l *Lexer) error(f string, a ...interface{}) {
	err := fmt.Sprintf(f, a...)
	err = fmt.Sprintf("%s:%d:    %s", l.chunkName, l.line, err)
	panic(err)
}
