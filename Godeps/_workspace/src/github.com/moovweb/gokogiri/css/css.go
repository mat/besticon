package css

// package main

import (
	"fmt"
	"github.com/moovweb/rubex"
	"strings"
)

type Lexeme int

const (
	SPACES Lexeme = iota
	COMMA
	UNIVERSAL
	TYPE
	ELEMENT
	CLASS
	ID
	LBRACKET
	RBRACKET
	ATTR_NAME
	ATTR_VALUE
	EQUALS
	CONTAINS_CLASS
	DASH_PREFIXED
	STARTS_WITH
	ENDS_WITH
	CONTAINS
	MATCH_OP
	PSEUDO_CLASS
	FIRST_CHILD
	FIRST_OF_TYPE
	NTH_CHILD
	NTH_OF_TYPE
	ONLY_CHILD
	ONLY_OF_TYPE
	LAST_CHILD
	LAST_OF_TYPE
	NOT
	LPAREN
	RPAREN
	COEFFICIENT
	SIGNED
	UNSIGNED
	ODD
	EVEN
	N
	OPERATOR
	PLUS
	MINUS
	BINOMIAL
	ADJACENT_TO
	PRECEDES
	PARENT_OF
	ANCESTOR_OF
	// and a counter ... I can't believe I didn't think of this sooner
	NUM_LEXEMES
)

var pattern [NUM_LEXEMES]string

func init() {
	pattern[SPACES] = `\s+`
	pattern[COMMA] = `\s*,`
	pattern[UNIVERSAL] = `\*`
	pattern[TYPE] = `[_a-zA-Z]\w*`
	pattern[ELEMENT] = `(\*|[_a-zA-Z]\w*)`
	pattern[CLASS] = `\.[-\w]+`
	pattern[ID] = `\#[-\w]+`
	pattern[LBRACKET] = `\[`
	pattern[RBRACKET] = `\]`
	pattern[ATTR_NAME] = `[-_:a-zA-Z][-\w:.]*`
	pattern[ATTR_VALUE] = `("(\\.|[^"\\])*"|'(\\.|[^'\\])*')`
	pattern[EQUALS] = `=`
	pattern[CONTAINS_CLASS] = `~=`
	pattern[DASH_PREFIXED] = `\|=`
	pattern[STARTS_WITH] = `\^=`
	pattern[ENDS_WITH] = `\$=`
	pattern[CONTAINS] = `\*=`
	pattern[MATCH_OP] = "(" + strings.Join([]string{pattern[EQUALS], pattern[CONTAINS_CLASS], pattern[DASH_PREFIXED], pattern[STARTS_WITH], pattern[ENDS_WITH], pattern[CONTAINS]}, "|") + ")"
	pattern[PSEUDO_CLASS] = `:[-a-z]+`
	pattern[FIRST_CHILD] = `:first-child`
	pattern[FIRST_OF_TYPE] = `:first-of-type`
	pattern[NTH_CHILD] = `:nth-child`
	pattern[NTH_OF_TYPE] = `:nth-of-type`
	pattern[ONLY_CHILD] = `:only-child`
	pattern[ONLY_OF_TYPE] = `:only-of-type`
	pattern[LAST_CHILD] = `:last-child`
	pattern[LAST_OF_TYPE] = `:last-of-type`
	pattern[NOT] = `:not`
	pattern[LPAREN] = `\s*\(`
	pattern[RPAREN] = `\s*\)`
	pattern[COEFFICIENT] = `[-+]?(\d+)?`
	pattern[SIGNED] = `[-+]?\d+`
	pattern[UNSIGNED] = `\d+`
	pattern[ODD] = `odd`
	pattern[EVEN] = `even`
	pattern[N] = `[nN]`
	pattern[OPERATOR] = `[-+]`
	pattern[PLUS] = `\+`
	pattern[MINUS] = `-`
	pattern[BINOMIAL] = strings.Join([]string{pattern[COEFFICIENT], pattern[N], `\s*`, pattern[OPERATOR], `\s*`, pattern[UNSIGNED]}, "")
	pattern[ADJACENT_TO] = `\s*\+`
	pattern[PRECEDES] = `\s*~`
	pattern[PARENT_OF] = `\s*>`
	pattern[ANCESTOR_OF] = `\s+`
}

type Scope int

const (
	GLOBAL Scope = iota
	LOCAL
)

func Convert(css string, scope Scope) string {
	matchers := allocate()
	defer deallocate(matchers)
	xpath, _ := selectors(matchers, []byte(css), scope)
	return xpath
}

func allocate() []*rubex.Regexp {
	// some overlap in here, but it'll make the parsing functions clearer
	matchers := make([]*rubex.Regexp, 0, NUM_LEXEMES)
	for _, p := range pattern {
		matchers = append(matchers, rubex.MustCompile(`\A`+p))
	}
	return matchers
}

func deallocate(matchers []*rubex.Regexp) {
	for _, m := range matchers {
		m.Free()
	}
}

func selectors(matchers []*rubex.Regexp, input []byte, scope Scope) (string, []byte) {
	x, input := selector(matchers, input, scope)
	xs := []string{x}
	for peek(matchers, COMMA, input) {
		_, input = token(matchers, COMMA, input)
		x, input = selector(matchers, input, scope)
		xs = append(xs, x)
	}
	return strings.Join(xs, " | "), input
}

func selector(matchers []*rubex.Regexp, input []byte, scope Scope) (string, []byte) {
	var combinator Lexeme
	var xs []string
	if scope == LOCAL {
		xs = []string{"."}
	}
	if matched, remainder := token(matchers, PARENT_OF, input); matched != nil {
		combinator, input = PARENT_OF, remainder
	} else {
		combinator = ANCESTOR_OF
	}
	x, input := sequence(matchers, input, combinator)
	xs = append(xs, x)
	for {
		if matched, remainder := token(matchers, ADJACENT_TO, input); matched != nil {
			combinator, input = ADJACENT_TO, remainder
		} else if matched, remainder := token(matchers, PRECEDES, input); matched != nil {
			combinator, input = PRECEDES, remainder
		} else if matched, remainder := token(matchers, PARENT_OF, input); matched != nil {
			combinator, input = PARENT_OF, remainder
		} else if matched, remainder := token(matchers, ANCESTOR_OF, input); matched != nil {
			combinator, input = ANCESTOR_OF, remainder
		} else {
			break
		}
		x, input = sequence(matchers, input, combinator)
		xs = append(xs, x)
	}
	return strings.Join(xs, ""), input
}

func sequence(matchers []*rubex.Regexp, input []byte, combinator Lexeme) (string, []byte) {
	_, input = token(matchers, SPACES, input)
	x, ps := "", []string{}

	switch combinator {
	case ANCESTOR_OF:
		x = "/descendant-or-self::*/*"
	case PARENT_OF:
		x = "/child::*"
	case PRECEDES:
		x = "/following-sibling::*"
	case ADJACENT_TO:
		x = "/following-sibling::*"
		ps = append(ps, "position()=1")
	}

	if e, remainder := token(matchers, ELEMENT, input); e != nil {
		input = remainder
		if len(ps) > 0 {
			ps = append(ps, " and ")
		}
		ps = append(ps, "self::"+string(e))
		if !(peek(matchers, ID, input) || peek(matchers, CLASS, input) || peek(matchers, PSEUDO_CLASS, input) || peek(matchers, LBRACKET, input)) {
			pstr := strings.Join(ps, "")
			if pstr != "" {
				pstr = fmt.Sprintf("[%s]", pstr)
			}
			return x + pstr, input
		}
	}
	q, input, connective := qualifier(matchers, input)
	if q == "" {
		panic("Invalid CSS selector")
	}
	if len(ps) > 0 {
		ps = append(ps, connective)
	}
	ps = append(ps, q)
	for q, r, c := qualifier(matchers, input); q != ""; q, r, c = qualifier(matchers, input) {
		ps, input = append(ps, c, q), r
	}
	pstr := strings.Join(ps, "")
	if combinator != NOT {
		pstr = fmt.Sprintf("[%s]", pstr)
	}
	return x + pstr, input
}

func qualifier(matchers []*rubex.Regexp, input []byte) (string, []byte, string) {
	p, connective := "", ""
	if t, remainder := token(matchers, CLASS, input); t != nil {
		p = fmt.Sprintf(`contains(concat(" ", @class, " "), " %s ")`, string(t[1:]))
		input = remainder
		connective = " and "
	} else if t, remainder := token(matchers, ID, input); t != nil {
		p, input, connective = fmt.Sprintf(`@id="%s"`, string(t[1:])), remainder, " and "
	} else if peek(matchers, PSEUDO_CLASS, input) {
		p, input, connective = pseudoClass(matchers, input)
	} else if peek(matchers, LBRACKET, input) {
		p, input = attribute(matchers, input)
		connective = " and "
	}
	return p, input, connective
}

func pseudoClass(matchers []*rubex.Regexp, input []byte) (string, []byte, string) {
	class, input := token(matchers, PSEUDO_CLASS, input)
	var p, connective string
	switch string(class) {
	case ":first-child":
		p, connective = "position()=1", " and "
	case ":first-of-type":
		p, connective = "position()=1", "]["
	case ":last-child":
		p, connective = "position()=last()", " and "
	case ":last-of-type":
		p, connective = "position()=last()", "]["
	case ":only-child":
		p, connective = "position() = 1 and position() = last()", " and "
	case ":only-of-type":
		p, connective = "position() = 1 and position() = last()", "]["
	case ":nth-child":
		p, input = nth(matchers, input)
		connective = " and "
	case ":nth-of-type":
		p, input = nth(matchers, input)
		connective = "]["
	case ":not":
		p, input = negate(matchers, input)
		connective = " and "
	default:
		panic(`Cannot convert CSS pseudo-class "` + string(class) + `" to XPath.`)
	}
	return p, input, connective
}

func nth(matchers []*rubex.Regexp, input []byte) (string, []byte) {
	lparen, input := token(matchers, LPAREN, input)
	if lparen == nil {
		panic(":nth-child and :nth-of-type require an parenthesized argument")
	}
	_, input = token(matchers, SPACES, input)
	var expr string
	if e, rem := token(matchers, EVEN, input); e != nil {
		expr, input = "position() mod 2 = 0", rem
	} else if e, rem := token(matchers, ODD, input); e != nil {
		expr, input = "position() mod 2 = 1", rem
	} else if e, _ := token(matchers, BINOMIAL, input); e != nil {
		var coefficient, operator, constant []byte
		coefficient, input = token(matchers, COEFFICIENT, input)
		switch string(coefficient) {
		case "", "+":
			coefficient = []byte("1")
		case "-":
			coefficient = []byte("-1")
		}
		_, input = token(matchers, N, input)
		_, input = token(matchers, SPACES, input)
		operator, input = token(matchers, OPERATOR, input)
		_, input = token(matchers, SPACES, input)
		constant, input = token(matchers, UNSIGNED, input)
		expr = fmt.Sprintf("(position() %s %s) mod %s = 0", invert(string(operator)), string(constant), string(coefficient))
	} else if e, rem := token(matchers, SIGNED, input); e != nil {
		expr, input = "position() = "+string(e), rem
	} else {
		panic("Invalid argument to :nth-child or :nth-of-type.")
	}
	fmt.Println(string(input))
	_, input = token(matchers, SPACES, input)
	rparen, input := token(matchers, RPAREN, input)
	if rparen == nil {
		panic("Unterminated argument to :nth-child or :nth-of-type.")
	}
	return expr, input
}

func invert(op string) string {
	op = strings.TrimSpace(op)
	if op == "+" {
		op = "-"
	} else {
		op = "+"
	}
	return op
}

func negate(matchers []*rubex.Regexp, input []byte) (string, []byte) {
	_, input = token(matchers, SPACES, input)
	lparen, input := token(matchers, LPAREN, input)
	if lparen == nil {
		panic(":not requires a parenthesized argument.")
	}
	_, input = token(matchers, SPACES, input)
	p, input := sequence(matchers, input, NOT)
	_, input = token(matchers, SPACES, input)
	rparen, input := token(matchers, RPAREN, input)
	if rparen == nil {
		panic("Unterminated argument to :not.")
	}
	return fmt.Sprintf("not(%s)", p), input
}

func attribute(matchers []*rubex.Regexp, input []byte) (string, []byte) {
	_, input = token(matchers, LBRACKET, input)
	_, input = token(matchers, SPACES, input)
	name, input := token(matchers, ATTR_NAME, input)
	if name == nil {
		panic("Attribute selector requires an attribute name.")
	}
	_, input = token(matchers, SPACES, input)
	if rbracket, remainder := token(matchers, RBRACKET, input); rbracket != nil {
		return "@" + string(name), remainder
	}
	op, input := token(matchers, MATCH_OP, input)
	if op == nil {
		panic("Missing operator in attribute selector.")
	}
	_, input = token(matchers, SPACES, input)
	val, input := token(matchers, ATTR_VALUE, input)
	if val == nil {
		panic("Missing value in attribute selector.")
	}
	_, input = token(matchers, SPACES, input)
	rbracket, input := token(matchers, RBRACKET, input)
	if rbracket == nil {
		panic("Unterminated attribute selector.")
	}
	var expr string
	n, v := string(name), string(val)
	switch string(op) {
	case "=":
		expr = fmt.Sprintf("@%s=%s", n, v)
	case "~=":
		expr = fmt.Sprintf(`contains(concat(" ", @%s, " "), concat(" ", %s, " "))`, n, v)
	case "|=":
		expr = fmt.Sprintf(`(@%s=%s or starts-with(@%s, concat(%s, "-")))`, n, v, n, v)
	case "^=":
		expr = fmt.Sprintf("starts-with(@%s, %s)", n, v)
	case "$=":
		// oy, libxml doesn't support ends-with
		// generate something like: div[substring(@class, string-length(@class) - string-length('foo') + 1) = 'foo']
		expr = fmt.Sprintf("substring(@%s, string-length(@%s) - string-length(%s) + 1) = %s", n, n, v, v)
	case "*=":
		expr = fmt.Sprintf("contains(@%s, %s)", n, v)
	}
	return expr, input
}

func token(matchers []*rubex.Regexp, lexeme Lexeme, input []byte) ([]byte, []byte) {
	matched := matchers[lexeme].Find(input)
	length := len(matched)
	if length == 0 {
		matched = nil
	}
	return matched, input[length:]
}

func peek(matchers []*rubex.Regexp, lexeme Lexeme, input []byte) bool {
	matched, _ := token(matchers, lexeme, input)
	return matched != nil
}
