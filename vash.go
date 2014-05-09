package main

import (
	"os"
	"io"
	"fmt"
	"regexp"
	"bytes"
	"strings"
)

const (
	UNDEF = iota
	AT
	ASSIGN_OPERATOR
	AT_COLON
	AT_STAR_CLOSE
	AT_STAR_OPEN
	BACKSLASH
	BRACE_CLOSE
	BRACE_OPEN
	CONTENT
	EMAIL
	ESCAPED_QUOTE
	FORWARD_SLASH
	FUNCTION
	HARD_PAREN_CLOSE
	HARD_PAREN_OPEN
	HTML_TAG_OPEN
	HTML_TAG_CLOSE
	HTML_TAG_VOID_CLOSE
	IDENTIFIER
	KEYWORD
	LOGICAL
	NEWLINE
	NUMERIC_CONTENT
	OPERATOR
	PAREN_CLOSE
	PAREN_OPEN
	PERIOD
	SINGLE_QUOTE
	DOUBLE_QUOTE
	TEXT_TAG_CLOSE
	TEXT_TAG_OPEN
	WHITESPACE
)

type TokenMatch struct {
	Type  int
	Text  string
	Regex *regexp.Regexp
}

func rec(reg string) (*regexp.Regexp) {
	res, err := regexp.Compile("^" + reg)
	if err != nil {
		panic(err)
	}
	return res
}

// The order is important
var Tests = []TokenMatch{
	TokenMatch{EMAIL, "EMAIL", rec(`([a-zA-Z0-9.%]+@[a-zA-Z0-9.\-]+\.(?:ca|co\.uk|com|edu|net|org))\\b`)},
        TokenMatch{AT_STAR_OPEN, "AT_STAR_OPEN", rec(`@\*`)},
        TokenMatch{AT_STAR_CLOSE, "AT_STAR_CLOSE", rec(`(\*@)`)},
        TokenMatch{AT_COLON, "AT_COLON", rec(`(@\:)`)},
        TokenMatch{AT, "AT", rec(`(@)`)},
        TokenMatch{PAREN_OPEN, "PAREN_OPEN", rec(`(\()`)},
        TokenMatch{PAREN_CLOSE, "PAREN_CLOSE", rec(`(\))`)},
        TokenMatch{HARD_PAREN_OPEN, "HARD_PAREN_OPEN", rec(`(\[)`)},
        TokenMatch{HARD_PAREN_CLOSE, "HARD_PAREN_CLOSE", rec(`(\])`)},
        TokenMatch{BRACE_OPEN, "BRACE_OPEN", rec(`(\{)`)},
        TokenMatch{BRACE_CLOSE, "BRACE_CLOSE", rec(`(\})`)},
        TokenMatch{TEXT_TAG_OPEN, "TEXT_TAG_OPEN", rec(`(<text>)`)},
        TokenMatch{TEXT_TAG_CLOSE, "TEXT_TAG_CLOSE", rec(`(<\/text>)`)},
        TokenMatch{HTML_TAG_OPEN, "HTML_TAG_OPEN", rec(`(<[a-zA-Z@]+?[^>]*?["a-zA-Z]*>)`)},
        TokenMatch{HTML_TAG_CLOSE, "HTML_TAG_CLOSE", rec(`(</[^>@]+?>)`)},
        TokenMatch{HTML_TAG_VOID_CLOSE, "HTML_TAG_VOID_CLOSE", rec(`(\/\s*>)`)},
        TokenMatch{PERIOD, "PERIOD", rec(`(\.)`)},
        TokenMatch{NEWLINE, "NEWLINE", rec(`(\n)`)},
        TokenMatch{WHITESPACE, "WHITESPACE", rec(`(\s)`)},
        TokenMatch{FUNCTION, "FUNCTION", rec(`(function)([\D\W])`)},
        TokenMatch{KEYWORD, "KEYWORD", rec(`(case|do|else|section|for|func|goto|if|return|switch|try|var|while|with)([\D\W\S]$)`)},
        TokenMatch{IDENTIFIER, "IDENTIFIER", rec(`([_$a-zA-Z][_$a-zA-Z0-9]*)`)}, //need verify
        TokenMatch{FORWARD_SLASH, "FORWARD_SLASH", rec(`(\/)`)},
        TokenMatch{OPERATOR, "OPERATOR", rec(`(===|!==|==|!==|>>>|<<|>>|>=|<=|>|<|\+|-|\/|\*|\^|%|\:|\?)`)},
        TokenMatch{ASSIGN_OPERATOR, "ASSIGN_OPERATOR", rec(`(\|=|\^=|&=|>>>=|>>=|<<=|-=|\+=|%=|\/=|\*=|=)`)},
        TokenMatch{LOGICAL, "LOGICAL", rec(`(&&|\|\||&|\||\^)`)},
        TokenMatch{ESCAPED_QUOTE, "ESCAPED_QUOTE", rec(`(\\+['\"])`)},
        TokenMatch{BACKSLASH, "BACKSLASH", rec(`(\\)`)},
        TokenMatch{DOUBLE_QUOTE, "DOUBLE_QUOTE", rec(`(")`)},
        TokenMatch{SINGLE_QUOTE, "SINGLE_QUOTE", rec(`(')`)},
        TokenMatch{NUMERIC_CONTENT, "NUMERIC_CONTENT", rec(`([0-9]+)`)},
        TokenMatch{CONTENT, "CONTENT", rec(`([^\s})@.]+?)`)},
}

type Token struct {
	Text string
	TypeStr string
	Type int
	Line int
	Pos  int
}

func (token Token) P() {
	textStr := strings.Replace(token.Text, "\n", "\\n", -1)
	textStr = strings.Replace(textStr, "\t", "\\t", -1)
	fmt.Printf("Token: %-20s Location:(%-2d %-2d) Value: %s\n",
		token.TypeStr, token.Line, token.Pos, textStr)
}

type Lexer struct {
	Text  string
	Matches []TokenMatch
}

func LineAndPos(src string, pos int) (int, int) {
	lines := strings.Count(src[:pos], "\n")
	p := pos - strings.LastIndex(src[:pos], "\n")
	return lines+1, p
}

func TagOpen(text string) string {
	regs := []string{
		`([a-zA-Z0-9.%]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,4})\b`,
		`(@)`,
		`(\/\s*>)`}
	res := text
	for _, reg := range regs {
		regc, err := regexp.Compile(reg)
		if err != nil {
			panic(err)
		}
		found := regc.FindIndex([]byte(text))
		if found != nil {
			res = res[:found[0]] //BUG?
		}
	}
	return res
}

func (lexer *Lexer) Scan() ([]Token, error) {
	pos := 0
	toks := []Token{}
	text := strings.Replace(lexer.Text, "\r\n", "\n", -1)
	text = strings.Replace(lexer.Text, "\r", "\n", -1)
	for pos < len(text) {
		left := text[pos:]
        	match := false
		length := 0
		for _, m := range lexer.Matches {
			found := m.Regex.FindIndex([]byte(left))
			if found != nil {
				match = true
				line, pos := LineAndPos(text, pos)
				tokenVal := left[found[0]:found[1]]
				if m.Type == HTML_TAG_OPEN {
					tokenVal = TagOpen(tokenVal)
				}
				tok := Token{tokenVal, m.Text, m.Type, line, pos}
				toks = append(toks, tok)
				length = len(tokenVal)
				break
			}
		}
		if !match {
			err_line, err_pos := LineAndPos(text, pos)
			return toks, fmt.Errorf("%d:%d: Illegal character: %s",
				err_line, err_pos, string(text[pos]))
		}
		pos += length
	}
	return toks, nil
}

//------------------------------ Parser ------------------------------//
const (
	UNK = iota
	PRG
	MKP
	BLK
	EXP
)

var PAIRS = map[int]int{
	AT_STAR_OPEN:    AT_STAR_CLOSE,
	BRACE_OPEN:      BRACE_CLOSE,
	DOUBLE_QUOTE:    DOUBLE_QUOTE,
	HARD_PAREN_OPEN: HARD_PAREN_CLOSE,
	PAREN_OPEN:      PAREN_CLOSE,
	SINGLE_QUOTE:    SINGLE_QUOTE,
	AT_COLON:        NEWLINE,
	FORWARD_SLASH:   FORWARD_SLASH,
}


type Ast struct {
	Parent   *Ast
	Children []interface{}
	Mode     int
	TagName  string
}

func (ast *Ast) ModeStr() string{
	switch ast.Mode {
	case PRG: return "PROGRAM"
	case MKP: return "MARKUP"
	case BLK: return "BLOCK"
	case EXP: return "EXP"
	default: return "UNDEF"
	}
	return "UNDEF"
}

func (ast *Ast) check() {
        if len(ast.Children) >= 100000 {
                panic("Maximum number of elements exceeded.")
        }
}

func (ast *Ast) addChild(child interface{}) {
	ast.Children = append(ast.Children, child)
	ast.check()
	if _a, ok := child.(*Ast); ok {
		_a.Parent = ast
	}
}

func (ast *Ast) addChildren(children []Token) { //BUG?
	for _, c := range children {
		ast.addChild(c)
	}
}

func (ast *Ast) addAst(_ast *Ast) {
	c := _ast
	for {
		if len(c.Children) != 1 {
			break
		}
		first := c.Children[0]
		if _, ok := first.(*Ast); !ok {
			break
		}
		c = first.(*Ast)
	}
	if c.Mode != PRG {
		ast.addChild(c)
	} else {
		for _, x := range c.Children {
			ast.addChild(x)
		}
	}
}

func (ast *Ast) popChild() {
	l := len(ast.Children)
	if l == 0 {
		return
	}
	ast.Children = ast.Children[:l-1]
}

func (ast *Ast) root() *Ast {
	p := ast
	pp := ast.Parent
	for {
		if p == pp || pp == nil {
			return p
		}
		b := pp
		pp = p.Parent
		p = b
	}
	return nil
}


func (ast *Ast) beget(mode int, tag string) *Ast {
	child := &Ast{ast, []interface{}{}, mode, tag}
	ast.addChild(child)
	return child
}

func (ast *Ast) closest(mode int, tag string) *Ast {
	p := ast
	for {
		if p.TagName != tag && p.Parent != nil {
			p = p.Parent
		} else {
			break
		}
	}
	return p
}
func (ast *Ast) debug(depth int, max int) {
	if depth >= max {
		return
	}
	for i := 0; i < depth; i++ {
		fmt.Printf("%c", '-')
	}
        fmt.Printf("TagName: %s Mode: %s Children: %d [[ \n", ast.TagName, ast.ModeStr(), len(ast.Children))
	for _, a := range ast.Children {
		//fmt.Printf("(%d)", idx)
		if _, ok := a.(*Ast); ok {
			b := (*Ast)(a.(*Ast))
			b.debug(depth+1, max)
		} else {
			if depth + 1 < max {
				aa := (Token)(a.(Token))
				for i := 0; i < depth+1; i++ {
					fmt.Printf("%c", '-')
				}
				aa.P()
			}
		}
	}
        for i := 0; i < depth; i++ {
                fmt.Printf("%c", '-')
        }

	fmt.Println("]]")
}

type Parser struct {
	ast         *Ast
	tokens      []Token
	preTokens   []Token
	inComment   bool
	saveTextTag bool
	initMode    int
}

func (parser *Parser) prevToken(idx int) (*Token) {
	l := len(parser.preTokens)
	if l < idx + 1 {
		return nil
	}
	return &(parser.preTokens[l - 1 - idx])
}

func (parser *Parser) deferToken(token Token) {
	parser.tokens = append([]Token{token}, parser.tokens...)
	parser.preTokens = parser.preTokens[:len(parser.preTokens)-1]
}

func (parser *Parser) peekToken(idx int) (*Token) {
        if len(parser.tokens) <= idx  {
		return nil
        }
        return &(parser.tokens[idx])
}

func (parser *Parser) nextToken() (Token) {
        t := parser.peekToken(0)
        if t != nil {
                parser.tokens = parser.tokens[1:]
        }
        return *t
}

func (parser *Parser) skipToken() {
	parser.tokens = parser.tokens[1:]
}

func regMatch(reg string, text string) (string, error) {
	regc, err := regexp.Compile(reg)
	if err != nil {
		panic(err)
		return "", err
	}
	found := regc.FindIndex([]byte(text))
	if found != nil {
		return text[found[0]:found[1]], nil
	}
	return "", nil
}

func (parser *Parser) advanceUntilNot(tokenType int) []Token {
	res := []Token{}
	for {
		t := parser.peekToken(0)
		if t != nil && t.Type == tokenType {
			res = append(res, parser.nextToken())
		} else {
			break
		}
	}
	return res
}

func (parser *Parser) advanceUntil(token Token, start, end, startEsc, endEsc int) []Token {
	var prev *Token = nil
	next := &token
	res := []Token{}
	nstart := 0
	nend := 0
	for {
		if next.Type == start {
			if (prev != nil && prev.Type != startEsc && start != end) || prev == nil {
				nstart++
			} else if start == end && prev.Type != startEsc {
				nend++
			}
		} else if next.Type == end {
			nend++
			if prev != nil && prev.Type == endEsc {
				nend--
			}
		}
		res = append(res, *next)
		if nstart == nend {
			break
		}
		prev = next
		next = parser.peekToken(0)
		if next == nil {
			panic("UNMATCHED")
		}
		parser.nextToken()
	}
	return res
}

func (parser *Parser) subParse(token Token, modeOpen int, includeDelim bool) {
	subTokens := parser.advanceUntil(token, token.Type, PAIRS[token.Type], -1, AT)
	//fmt.Printf("-----------\n")
	//for _, t := range subTokens {
	//t.P()
	//}
	//fmt.Printf("++++++++++\n")

	subTokens = subTokens[1:]
	closer := subTokens[len(subTokens)-1]
	subTokens = subTokens[:len(subTokens)-1]
	if !includeDelim {
		parser.ast.addChild(token)

	}
        _parser := &Parser{&Ast{}, subTokens, []Token{}, false, false, modeOpen}
	_parser.Run()
	if includeDelim {
		_parser.ast.Children = append([]interface{}{token}, _parser.ast.Children...)
		_parser.ast.addChild(closer)
	}
	//_parser.ast.debug(0)
	parser.ast.addAst(_parser.ast)
	if !includeDelim {
		parser.ast.addChild(closer)
	}
}

func (parser *Parser) handleMKP(token Token) {
	next  := parser.peekToken(0)
	//nnext := parser.peekToken(1)
	switch token.Type {
	case AT_STAR_OPEN:
		parser.advanceUntil(token, AT_STAR_OPEN, AT_STAR_CLOSE, AT, AT)

	case AT:
		if next != nil {
			switch next.Type {
			case PAREN_OPEN, IDENTIFIER:
				if len(parser.ast.Children) == 0 {
					parser.ast = parser.ast.Parent
					parser.ast.popChild() //remove empty MKP block
				}
				parser.ast = parser.ast.beget(EXP, "")

			case KEYWORD, FUNCTION, BRACE_OPEN: //BLK
				if len(parser.ast.Children) == 0 {
					parser.ast = parser.ast.Parent
					parser.ast.popChild()
				}
				parser.ast = parser.ast.beget(BLK, "")

			case AT, AT_COLON:
				//we want to keep the token, but remove it's special meanning
				next.Type = CONTENT //BUG, modify from a pointer, work?
				parser.ast.addChild(parser.nextToken())
			default:
				parser.ast.addChild(parser.nextToken())
			}
		}

	case TEXT_TAG_OPEN, HTML_TAG_OPEN:
		tagName, _ := regMatch(`(?i)(^<([^\/ >]+))`, token.Text)
		tagName = strings.Replace(tagName, "<", "", -1)
		//TODO
		if parser.ast.TagName != "" {
			parser.ast = parser.ast.beget(MKP, tagName)
		} else {
			parser.ast.TagName = tagName
		}
		if token.Type == HTML_TAG_OPEN || parser.saveTextTag {
			parser.ast.addChild(token)
		}

	case TEXT_TAG_CLOSE, HTML_TAG_CLOSE:
		tagName, _ := regMatch(`(?i)^<\/([^>]+)`, token.Text)
		tagName = strings.Replace(tagName, "</", "", -1)
		//TODO
		opener := parser.ast.closest(MKP, tagName)
		if opener.TagName != tagName { //unmatched
			panic("UNMATCHED!")
		} else {
			parser.ast = opener
		}
		if token.Type == HTML_TAG_CLOSE || parser.saveTextTag {
			parser.ast.addChild(token)
		}

		// vash.js have bug here, we should skip current MKP,
		// so that we can keep in a right hierarchy
		if parser.ast.Parent != nil && parser.ast.Parent.Mode == MKP {
			parser.ast = parser.ast.Parent
		}

	case HTML_TAG_VOID_CLOSE:
		parser.ast.addChild(token)
		parser.ast = parser.ast.Parent

	case BACKSLASH:
		token.Text += "\\"
		parser.ast.addChild(token)
	default:
		parser.ast.addChild(token)
	}
}

func (parser *Parser) handleBLK(token Token) {
	next := parser.peekToken(0)
	switch token.Type {
	case AT:
		if (next.Type != AT) && (!parser.inComment) {
			parser.deferToken(token)
			parser.ast = parser.ast.beget(MKP, "")
		} else {
			next.Type = CONTENT
			parser.ast.addChild(*next)
			parser.skipToken()
		}

	case AT_STAR_OPEN:
        	parser.advanceUntil(token, AT_STAR_OPEN, AT_STAR_CLOSE, AT, AT)

	case AT_COLON:
                parser.subParse(token, MKP, true)

	case TEXT_TAG_OPEN, TEXT_TAG_CLOSE, HTML_TAG_OPEN, HTML_TAG_CLOSE:
                parser.ast = parser.ast.beget(MKP, "")
		parser.deferToken(token)

	case FORWARD_SLASH, SINGLE_QUOTE, DOUBLE_QUOTE:
                if token.Type == FORWARD_SLASH && next != nil && next.Type == FORWARD_SLASH {
			parser.inComment = true
		}
		if !parser.inComment {
			subTokens := parser.advanceUntil(token, token.Type,
				PAIRS[token.Type],
				BACKSLASH,
				BACKSLASH)
			for idx, _ := range subTokens {
				if subTokens[idx].Type == AT {
					subTokens[idx].Type = CONTENT
				}
			}
			parser.ast.addChildren(subTokens)
		} else {
			parser.ast.addChild(token)
		}

	case NEWLINE:
                if parser.inComment {
			parser.inComment = false
		}
		parser.ast.addChild(token)

	case BRACE_OPEN, PAREN_OPEN:
		subMode := BLK
		if false && token.Type == BRACE_OPEN {  //TODO
			subMode = MKP
		}
		parser.subParse(token, subMode, false)
		subTokens := parser.advanceUntilNot(WHITESPACE)
		next := parser.peekToken(0)
		if next != nil && next.Type != KEYWORD &&
			next.Type != FUNCTION && next.Type != BRACE_OPEN &&
			token.Type != PAREN_OPEN {
			parser.tokens = append(parser.tokens, subTokens...)
			parser.ast = parser.ast.Parent
		} else {
			parser.ast.addChildren(subTokens)
		}
	default:
		parser.ast.addChild(token)
	}
}


func (parser *Parser) handleEXP(token Token) {
	switch token.Type {
	case KEYWORD, FUNCTION:
		parser.ast = parser.ast.beget(BLK, "")
		parser.deferToken(token)

	case WHITESPACE, LOGICAL, ASSIGN_OPERATOR, OPERATOR, NUMERIC_CONTENT:
		if parser.ast.Parent != nil && parser.ast.Parent.Mode == EXP {
			parser.ast.addChild(token)
		} else {
			parser.ast = parser.ast.Parent
			parser.deferToken(token)
		}
	case IDENTIFIER:
		parser.ast.addChild(token)

	case SINGLE_QUOTE, DOUBLE_QUOTE:
		//TODO
		if parser.ast.Parent != nil && parser.ast.Parent.Mode == EXP {
			subTokens := parser.advanceUntil(token, token.Type,
				                         PAIRS[token.Type], BACKSLASH, BACKSLASH)
			parser.ast.addChildren(subTokens)
		} else {
			parser.ast = parser.ast.Parent
			parser.deferToken(token)
		}

	case HARD_PAREN_OPEN, PAREN_OPEN:
		prev := parser.prevToken(0)
		next := parser.peekToken(0)
		if token.Type == HARD_PAREN_OPEN && next.Type == HARD_PAREN_CLOSE {
			// likely just [], which is not likely valid outside of EXP
			parser.deferToken(token)
			parser.ast = parser.ast.Parent
			break
		}
		parser.subParse(token, EXP, false)
		if (prev != nil && prev.Type == AT) || (next != nil && next.Type == IDENTIFIER) {
			parser.ast = parser.ast.Parent
		}

	case BRACE_OPEN:
		parser.deferToken(token)
		parser.ast = parser.ast.beget(BLK, "")

	case PERIOD:
		next := parser.peekToken(0)
		if next != nil && (next.Type == IDENTIFIER || next.Type == KEYWORD ||
			next.Type == FUNCTION || next.Type == PERIOD ||
			(parser.ast.Parent != nil && parser.ast.Parent.Mode == EXP)) {
			parser.ast.addChild(token)
		} else {
			parser.ast = parser.ast.Parent
			parser.deferToken(token)
		}
	default:
		if parser.ast.Parent != nil && parser.ast.Parent.Mode != EXP {
			parser.ast = parser.ast.Parent
			parser.deferToken(token)
		} else {
			parser.ast.addChild(token)
		}
	}
}

func (parser *Parser) Run() (err error) {
        //fmt.Println("---------------BEGIN------------------------")
	curr := Token{"UNDEF", "UNDEF", UNDEF, 0, 0}
	parser.ast.Mode = PRG
	for {
		if len(parser.tokens) == 0 {
			break
		}
                parser.preTokens = append(parser.preTokens, curr)
        	curr = parser.nextToken()
		if parser.ast.Mode == PRG {
			init := parser.initMode
			if init == UNK {
				init = MKP
			}
			parser.ast = parser.ast.beget(init, "")
			if parser.initMode == EXP {
				parser.ast = parser.ast.beget(EXP, "")
			}
		}
		//fmt.Println("curr: ")
		//curr.P()
		//fmt.Printf(" mode: %s\n", parser.ast.ModeStr())
		switch parser.ast.Mode {
		case MKP:
			parser.handleMKP(curr)
		case BLK:
			parser.handleBLK(curr)
		case EXP:
			parser.handleEXP(curr)
		}
	}

	parser.ast = parser.ast.root()
	//fmt.Println("---------------END---------------------")
	return nil
}

type Complier struct {
	ast *Ast
	buf  string
}

func (cp *Complier) visitBlock(token Token) {
	cp.buf += "BLK(" + token.Text + ")BLK"
}

// func (cp *Compiler) visitExp(token Token,  parent Token, index int, isHomo bool) {
// 	start := ""
// 	end   := ""
// 	//parentIsNotExp = true
// 	//TODO

// 	if cp.options[htmlEsc] {
// 		if parentIsNotExp && index == 0 && isHomo {
// 			if token.Text == "helper" || token.Text == "raw" ||
// 				cp.options['package'] = "layout" {
// 				start += "("
// 			} else {
// 				start += "gorazor.HTMLEscape("
// 			}
// 		}
// 		if parentIsNotExp && index == token.parent - 1 && isHomo {
// 			end += ")"
// 		}
// 	}

// 	if parentIsNotExp && index == 0 {
// 		start  = "_buffer.WriteString(" + start
// 	}
// 	if parentIsNotExp && index == token.parent - 1 {
// 		end += ")\n"
// 	}
// 	if token.Text == "raw" {
// 		cp.buf += start + end
// 	} else {
// 		cp.buf += start + token.Text + end
// 	}
// }

// func (cp *Complier) visitMKP(parent *Ast, ast *Ast, idx int) {
// 	start := ""
// 	end   := ""

// 	if index == 0 {
// 		start = "_buffer.WriteString(" + start
// 	}
// 	if index == len(parent.Children) - 1 {
// 		end += ")\n"
// 	}
// 	if

// }

func (cp *Complier) visitMKP(child interface{}, ast *Ast) {
	switch v := child.(type) {
	case *Ast:
		cp.buf += "MKP(" + v.TagName + ")MKP"
	case Token:
		cp.buf += "MKP(" + v.Text + ")MKP"
	}
}

func (cp *Complier) visitBLK(child interface{}, ast *Ast) {
	switch v := child.(type) {
	case *Ast:
		cp.buf += "BLK(" + v.TagName + ")BLK"
	case Token:
		cp.buf += "BLK(" + v.Text + ")BLK"
	}
}

func (cp *Complier) visitAst(ast *Ast) {
	switch ast.Mode {
	case MKP:
		cp.buf += "MKP(" + ast.TagName + ")MKP"
		for _, c := range ast.Children {
			cp.visitMKP(c, ast)
		}
	case PRG:
		cp.buf += "PRG(" + ast.TagName + ")PRG"
                for _, c := range ast.Children {
			cp.visitNode(c)
		}
	case BLK:
		cp.buf += "BLK(" + ast.TagName + ")BLK"
                for _, c := range ast.Children {
                        cp.visitBLK(c, ast)
                }
        }

        //for _, c := range ast.Children {
	//cp.visitNode(c)
//}
}

 func (cp *Complier) visitToken(token Token) {
// 	cp.buf += token.Text
// 	switch token.Type {
// 		case
 }

func (cp *Complier) visit() {

	cp.visitNode(cp.ast)
	fmt.Println(cp.buf)
        cp.buf = strings.Replace(cp.buf, "\n", "\\n", -1)
        cp.buf = strings.Replace(cp.buf, "\t", "\\t", -1)
        cp.buf = strings.Replace(cp.buf, ")MKPMKP(", "", -1)
	cp.buf = strings.Replace(cp.buf, "MKP(", "\n_buffer.WriteString(\"", -1)
	cp.buf = strings.Replace(cp.buf, ")MKP", "\")\n", -1)
        cp.buf = "var _buffer bytes.Buffer\n" + cp.buf
        cp.buf += "\nreturn _buffer.String()"
}

func (cp *Complier) visitNode(node interface{}) {
	switch v := node.(type) {
	case *Ast:
		cp.visitAst(v)
		//fmt.Println("visitAST:", v)
	case Token:
		//fmt.Println("visitToken:", v)
		cp.visitToken(v)
	}
}

//------------------------------ Compiler ------------------------------ //
func main() {
	buf := bytes.NewBuffer(nil)
	f , err := os.Open("./now/bug.gohtml")
	if err != nil {
		panic(err)
	}
	io.Copy(buf, f)
	f.Close()

        text := string(buf.Bytes())
	lex := &Lexer{text, Tests}
        fmt.Println("buf:", text)
	res, err := lex.Scan()

	if err != nil {
		panic(err)
	}

	for _, elem := range res {
		elem.P()
	}

	parser := &Parser{&Ast{}, res, []Token{}, false, false, UNK}
	err = parser.Run()

        parser.ast.debug(0, 3)
	if parser.ast.Mode != PRG {
		panic("TYPE")
	}

	cp := &Complier{parser.ast, ""}
	cp.visit()

	fmt.Println("---------------------------------------")
	fmt.Println(cp.buf)
}
