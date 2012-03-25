package soy

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Tokens ---------------------------------------------------------------------

// tokenType identifies the type of lexical tokens.
type tokenType uint8

// token represents a text string returned from the lexer.
type token struct {
	t tokenType // type
	v string    // value
}

// All tokens.
const (
	tokenNil tokenType = iota // not used
	tokenEOF                  // EOF
	tokenError                // error occurred; value is text of error
	tokenLeftDelim            // tag left delimiter: {
	tokenRightDelim           // tag right delimiter: }
	tokenText                 // plain text
	// Primitive literals.
	tokenBool
	tokenFloat
	tokenInteger
	tokenList
	tokenMap
	tokenString
	// Commands.
	tokenCommand              // used only to delimit the commands
	tokenCall                 // {call ...}
	tokenCase                 // {case ...}
	tokenCss                  // {css ...}
	tokenDefault              // {default}
	tokenDelcall              // {delcall ...}
	tokenDelpackage           // {delpackage ...}
	tokenDeltemplate          // {deltemplate ...}
	tokenElse                 // {else}
	tokenElseif               // {elseif ...}
	tokenFor                  // {for ...}
	tokenForeach              // {foreach ...}
	tokenIf                   // {if ...}
	tokenIfempty              // {ifempty}
	tokenLiteral              // {literal}
	tokenMsg                  // {msg ...}
	tokenNamespace            // {namespace}
	tokenParam                // {param ...}
	tokenPrint                // {print ...}
	tokenSwitch               // {switch ...}
	tokenTemplate             // {template ...}
	// Close commands.
	tokenCallEnd              // {/call}
	tokenDelcallEnd           // {/delcall}
	tokenDeltemplateEnd       // {/deltemplate}
	tokenForEnd               // {/for}
	tokenForeachEnd           // {/foreach}
	tokenIfEnd                // {/if}
	tokenLiteralEnd           // {/literal}
	tokenMsgEnd               // {/msg}
	tokenParamEnd             // {/param}
	tokenSwitchEnd            // {/switch}
	tokenTemplateEnd          // {/template}
	// Character commands.
	tokenCarriageReturn       // {\r}
	tokenEmptyString          // {nil}
	tokenLeftBrace            // {lb}
	tokenNewline              // {\n}
	tokenRightBrace           // {rb}
	tokenSpace                // {sp}
	tokenTab                  // {\t}
	// These commands are defined in TemplateParser.jj but not in the docs.
	// Apparently they are not available in the open source version of Soy.
	// See http://goo.gl/V0wsd
	// tokenLet                  // {let}{/let}
	// tokenPlural               // {plural}{/plural}
	// tokenSelect               // {select}{/select}
)

// Lexer ----------------------------------------------------------------------

const (
	eof        = -1
	leftDelim  = "{"
	rightDelim = "}"
	decDigits  = "0123456789"
	hexDigits  = "0123456789ABCDEF"
)

// stateFn represents the state of the lexer as a function that returns the
// next state.
type stateFn func(*lexer) stateFn

// newLexer creates a new lexer for the input string.
//
// It is borrowed from the text/template package with minor changes.
func newLexer(name, input string) *lexer {
	// Two tokens of buffering is sufficient for all state functions.
	l := &lexer{
		name:   name,
		input:  input,
		state:  lexText,
		tokens: make(chan token, 2),
	}
	return l
}

// lexer holds the state of the lexical scanning.
//
// Based on the lexer from the "text/template" package.
// See http://www.youtube.com/watch?v=HxaD_trXwRE
type lexer struct {
	name        string     // the name of the input; used only during errors.
	input       string     // the string being scanned.
	state       stateFn    // the next lexing function to enter.
	pos         int        // current position in the input.
	start       int        // start position of this token.
	width       int        // width of last rune read from input.
	tokens      chan token // channel of scanned tokens.
	doubleDelim bool       // flag for tags starting with double braces.
}

// nextToken returns the next token from the input.
func (l *lexer) nextToken() token {
	for {
		select {
		case token := <-l.tokens:
			return token
		default:
			l.state = l.state(l)
		}
	}
	panic("not reached")
}

// next returns the next rune in the input.
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// emit passes an token back to the client.
func (l *lexer) emit(t tokenType) {
	l.tokens <- token{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) bool {
	pos := l.pos
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
	return l.pos > pos
}

// lineNumber reports which line we're on. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.pos], "\n")
}

// columnNumber reports which column in the current line we're on.
func (l *lexer) columnNumber() int {
	n := strings.LastIndex(l.input[:l.pos], "\n")
	if n == -1 {
		n = 0
	}
	return l.pos - n
}

// error returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextToken.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- token{tokenError, fmt.Sprintf(format, args...)}
	return nil
}

// State functions ------------------------------------------------------------

// lexText scans until an opening command delimiter, "{".
func lexText(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], leftDelim) {
			if l.pos > l.start {
				l.emit(tokenText)
			}
			return lexLeftDelim
		}
		if l.next() == eof {
			break
		}
	}
	// Correctly reached EOF.
	if l.pos > l.start {
		l.emit(tokenText)
	}
	l.emit(tokenEOF)
	return nil
}

// lexLeftDelim scans the left template tag delimiter, which is known
// to be present.
//
// If there are brace characters within a template tag, double braces must
// be used, so we differentiate them to match double closing braces later.
// Double braces are also optional for other cases.
func lexLeftDelim(l *lexer) stateFn {
	if strings.HasPrefix(l.input[l.pos:], leftDelim) {
		// Double delimiter.
		l.pos += 1
		l.doubleDelim = true
	} else {
		l.doubleDelim = false
	}
	l.pos += 1
	l.emit(tokenLeftDelim)
	return lexInsideTag
}

// lexRightDelim scans the right template tag delimiter, which is known
// to be present.
func lexRightDelim(l *lexer) stateFn {
	if l.doubleDelim {
		if strings.HasPrefix(l.input[l.pos:], rightDelim) {
			l.pos += 1
		} else {
			return l.errorf("expected double closing braces in tag")
		}
	}
	l.pos += 1
	l.emit(tokenRightDelim)
	return lexText
}

// lexInsideTag scans the elements inside a template tag.
//
// The first token within a tag is the command name (the print command is
// implied if no command is specified), and the rest of the text within the
// tag (if any) is referred to as the command text.
//
// Soy tag format:
//
//     - Can be delimited by single braces "{...}" or double braces "{{...}}".
//     - Soy tags delimited by double braces are allowed to contain single
//       braces within.
//     - Some Soy tags are allowed to end in "/}" or "/}}" to denote immediate
//       ending of a block.
//     - It is an error to use "/}" or "/}}" when it's not applicable to the
//       command.
//     - If there is a command name, it must come immediately after the
//       opening delimiter.
//     - The command name must be followed by either the closing delimiter
//       (if the command does not take any command text) or a whitespace (if
//       the command takes command text).
//     - It is an error to provide command text when it's not applicable,
//       and vice versa.
//
// Commands without closing tag (can't end in "/}" or "/}}"):
//
//     - {delpackage ...}
//     - {namespace ...}
//     - {print ...}
//     - {...} (implicit print)
//     - {\r}
//     - {nil}
//     - {lb}
//     - {\n}
//     - {rb}
//     - {sp}
//     - {\t}
//     - {elseif ...}
//     - {else ...}
//     - {case ...}
//     - {default}
//     - {ifempty}
//     - {css ...}
//
// Commands with optional closing tag:
//
//     - {call ... /} or {call ...}...{/call}
//     - {delcall ... /} or {delcall ...}...{/delcall}
//     - {param ... /} or {param ...}...{/param}
//
// Commands with required closing tag:
//
//     - {deltemplate ...}...{/deltemplate}
//     - {for ...}...{/for}
//     - {foreach ...}...{/foreach}
//     - {if ...}...{/if}
//     - {literal}...{/literal}
//     - {msg ...}...{/msg}
//     - {switch ...}...{/switch}
//     - {template ...}...{/template}
func lexInsideTag(l *lexer) stateFn {
	// TODO
	if strings.HasPrefix(l.input[l.pos:], rightDelim) {
		return lexRightDelim
	}
	return lexText
}

// lexLiteral scans until a closing literal delimiter, "{\literal}".
// It emits the literal text and the closing tag.
//
// A literal section contains raw text and may include braces.
func lexLiteral(l *lexer) stateFn {
	var end bool
	var pos int
	for {
		if strings.HasPrefix(l.input[l.pos:], "{/literal}") {
			end, pos = true, 10
		} else if strings.HasPrefix(l.input[l.pos:], "{{/literal}}") {
			end, pos = true, 12
		}
		if end {
			if l.pos > l.start {
				l.emit(tokenText)
			}
			l.pos += pos
			l.emit(tokenLiteralEnd)
		}
		if l.next() == eof {
			return l.errorf("unclosed literal")
		}
	}
	return lexText
}

// lexNumber scans a number: a float or integer (which can be decimal or hex).
func lexNumber(l *lexer) stateFn {
	typ, ok := scanNumber(l)
	if !ok {
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}
	// Emits tokenFloat or tokenInteger.
	l.emit(typ)
	return lexInsideTag
}

// scanNumber scans a number according to Soy's specification.
//
// It returns the scanned tokenType (tokenFloat or tokenInteger) and a flag
// indicating if an error was found.
//
// Floats must be in decimal and must either:
//
//     - Have digits both before and after the decimal point (both can be
//       a single 0), e.g. 0.5, -100.0, or
//     - Have a lower-case e that represents scientific notation,
//       e.g. -3e-3, 6.02e23.
//
// Integers can be:
//
//     - decimal (e.g. -827)
//     - hexadecimal (must begin with 0x and must use capital A-F,
//       e.g. 0x1A2B).
func scanNumber(l *lexer) (typ tokenType, ok bool) {
	typ = tokenInteger
	// Optional leading sign.
	hasSign := l.accept("+-")
	if l.input[l.pos:l.pos+2] == "0x" {
		// Hexadecimal.
		if hasSign {
			// No signs for hexadecimals.
			return
		}
		l.acceptRun("0x")
		if !l.acceptRun(hexDigits) {
			// Requires at least one digit.
			return
		}
		if l.accept(".") {
			// No dots for hexadecimals.
			return
		}
	} else {
		// Decimal.
		if !l.acceptRun(decDigits) {
			// Requires at least one digit.
			return
		}
		if l.accept(".") {
			// Float.
			if !l.acceptRun(decDigits) {
				// Requires a digit after the dot.
				return
			}
			typ = tokenFloat
		} else {
			if (!hasSign && l.input[l.start] == '0') ||
				(hasSign && l.input[l.start+1] == '0') {
				// Integers can't start with 0.
				return
			}
		}
		if l.accept("e") {
			l.accept("+-")
			if !l.acceptRun(decDigits) {
				// A digit is required after the scientific notation.
				return
			}
			typ = tokenFloat
		}
	}
	// Next thing must not be alphanumeric.
	if isAlphaNumeric(l.peek()) {
		l.next()
		return
	}
	ok = true
	return
}

// Helpers --------------------------------------------------------------------

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}