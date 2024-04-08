package executor

import (
	"strings"

	"github.com/go-cmd/cmd"
	"github.com/go-logr/zapr"
	"github.com/jlewi/hydros/pkg/util"
	"github.com/pkg/errors"
	"github.com/timtadh/lexmachine"
	"github.com/timtadh/lexmachine/machines"
	"go.uber.org/zap"
)

type TokenType string

const (
	PipeToken       TokenType = "PIPE"
	QuoteToken      TokenType = "QUOTE"
	UnmatchedToken  TokenType = "UNMATCHED"
	TextToken       TokenType = "TEXT"
	WhiteSpaceToken TokenType = "WHITESPACE"
)

// BashishParser is a parser for the bashish language.
// Bashish is a language that is a very simple subset of bash. It is basically
// shell commands plus the ability to do things like pipe the output of one command to another.
type BashishParser struct {
	l *lexmachine.Lexer
}

// NewBashishParser creates a new parser for the bashish language.
func NewBashishParser() (*BashishParser, error) {
	// We need to construct a lexer for the bashish language.
	l := lexmachine.NewLexer()

	// Here's a couple important details about the how the lexer works. Keep these in mind when constructing the rules.
	//
	// 1. The lexer prefers lower precedence matches that are longer. So be careful about having
	//    matches that are overly broad.
	// 2. Lexer compiles regular expressions to a DFA; it doesn't use GoLang's regexp library.
	//    As a result, the full regexp syntax is not supported. Notably, not all character classes are supported.
	//    For a list of supported classes see https://github.com/timtadh/lexmachine#built-in-character-classes.
	// 3. Another major limitation is https://github.com/timtadh/lexmachine/issues/34' lexmachine can't expand
	//    character classes within character classes. As an example `[\w?=]` won't work. A work around is to expand \w
	//    manually e.g. `[A-Za-z0-9_?=]`

	l.Add([]byte(`\w+`), NewTokenAction(TextToken))
	l.Add([]byte(`\s+`), NewTokenAction(WhiteSpaceToken))
	l.Add([]byte(`['"]`), NewTokenAction(QuoteToken))
	l.Add([]byte(`\|`), NewTokenAction(PipeToken))

	// We rely on the toTokens function to turn unmatched characters into UnmatchedTexToken
	if err := l.Compile(); err != nil {
		return nil, errors.Wrapf(err, "Failed to compile the lexer")
	}
	return &BashishParser{
		l: l,
	}, nil
}

// Parse a multline string into a sequence of commands
func (p *BashishParser) Parse(doc string) ([]Instruction, error) {
	lines := strings.Split(doc, "\n")

	instructions := make([]Instruction, 0, 10)
	for _, l := range lines {
		l := strings.TrimSpace(l)
		tokens, err := p.toTokens([]byte(l))
		if err != nil {
			return nil, err
		}

		iParser := instructionParser{
			insideQuote: false,
			fields:      make([]string, 0, len(tokens)),
			quoteChar:   "",
			newField:    "",
		}

		newInstructions, err := iParser.parse(tokens)
		if err != nil {
			return nil, err
		}
		instructions = append(instructions, newInstructions...)
	}
	return instructions, nil
}

// toTokens turns the provided input into a stream of tokens.
func (p *BashishParser) toTokens(inBytes []byte) ([]*Token, error) {
	scanner, err := p.l.Scanner(inBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to initialize the scanner")
	}
	tokens := make([]*Token, 0, 50)
	log := zapr.NewLogger(zap.L())
	for tok, err, eos := scanner.Next(); !eos; tok, err, eos = scanner.Next() {
		unErr := &machines.UnconsumedInput{}
		isUncosumed := errors.As(err, &unErr)
		if isUncosumed {
			// If the scanner returns an unconsumed input we need to handle it.
			// We rely on this to turn unmatched characters into UnmatchedToken
			scanner.TC = unErr.FailTC
			text := unErr.Text[unErr.StartTC:unErr.FailTC]
			log.V(util.Debug).Info("lexer returned unconsumed token", "text", string(text))

			newToken := &Token{
				TokenType: UnmatchedToken,
				Lexeme:    string(unErr.Text[unErr.StartTC:unErr.FailTC]),
				Match: &machines.Match{
					Bytes: text,
				},
			}
			tokens = append(tokens, newToken)
			continue
		} else if err != nil {
			return nil, err
		}

		token, ok := tok.(*Token)
		if !ok {
			return nil, errors.New("token isn't of type token")
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}

// instructionParser is a state machine for parsing a string of tokens.
type instructionParser struct {
	insideQuote bool
	fields      []string
	newField    string
	// To handle nested quotes we need to keep track of the quote character
	quoteChar string
}

// parse parses a sequence of tokens into a sequence of instructions.
func (p *instructionParser) parse(tokens []*Token) ([]Instruction, error) {
	instructions := make([]Instruction, 0, len(tokens))
	for _, token := range tokens {
		val := string(token.Match.Bytes)
		switch token.TokenType {
		case PipeToken:
			if !p.insideQuote && len(p.fields) > 0 {
				// Complete the instruction
				i := Instruction{
					Command: cmd.NewCmd(p.fields[0], p.fields[1:]...),
					Piped:   true,
				}
				instructions = append(instructions, i)
				p.fields = make([]string, 0, len(tokens))
			} else {
				p.newField += string(token.Match.Bytes)
			}
		case QuoteToken:
			p.handleQuoteToken(val)
		case TextToken:
			p.newField += val
		case UnmatchedToken:
			p.newField += string(token.Match.Bytes)
		case WhiteSpaceToken:
			if !p.insideQuote {
				if len(p.newField) > 0 {
					p.fields = append(p.fields, p.newField)
					p.newField = ""
				}
			} else {
				p.newField += string(token.Match.Bytes)
			}
		default:
			return nil, errors.Errorf("parse encoutered unknown token type %v", token.TokenType)
		}
	}

	// Any remaining fields should be rolled up into a final instruction.s
	if len(p.newField) > 0 {
		p.fields = append(p.fields, p.newField)
	}
	if len(p.fields) > 0 {
		i := Instruction{
			Command: cmd.NewCmd(p.fields[0], p.fields[1:]...),
			Piped:   false,
		}
		instructions = append(instructions, i)
	}
	return instructions, nil
}

func (p *instructionParser) handleQuoteToken(val string) {
	lastChar := ""
	if len(p.newField) > 0 {
		lastChar = string(p.newField[len(p.newField)-1])
	}

	if lastChar == "\\" {
		// Since slash is an escape character we remove it and add the quote
		p.newField = p.newField[:len(p.newField)-1]
		p.newField += val
		return
	}
	if p.insideQuote && p.quoteChar != val {
		// We encountered a quote within a quote but it is a different quote character
		// so we aren't closing the quotation. So just add it to the field
		p.newField += val
		return
	}
	// We emulate the shell behavior. In particular, we don't include the quotes in the field.
	// For example, suppose we have the shell command
	// echo "hello world"
	// This is equal to []string{"echo", "hello world"} not
	// []string{"echo", "\"hello world\""}

	if p.insideQuote {
		// Close the quote by adding the field
		p.fields = append(p.fields, p.newField)
		p.newField = ""
		p.quoteChar = ""
		p.insideQuote = false
	} else {
		// Start a quotation
		p.quoteChar = val
		p.insideQuote = true
	}
}

type Token struct {
	TokenType TokenType
	Lexeme    string
	Match     *machines.Match
}

func NewToken(tokenType TokenType, m *machines.Match) *Token {
	return &Token{
		TokenType: tokenType,
		Lexeme:    string(m.Bytes),
		Match:     m,
	}
}

// NewTokenAction creates a lexmachine action for the given tokentype
func NewTokenAction(t TokenType) lexmachine.Action {
	return func(scan *lexmachine.Scanner, match *machines.Match) (interface{}, error) {
		return NewToken(t, match), nil
	}
}

// Instruction represents one instruction in the bashish language.
// This is typically a command that should be executed. In addition it contains information about
// how that command should be executed; e.g. if the output of this command should be piped to the next command.
type Instruction struct {
	Command *cmd.Cmd

	// Piped should be set to true if the output of this command should be piped to the next instruction.
	Piped bool
}
