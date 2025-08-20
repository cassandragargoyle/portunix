// Code generated from Portunix.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser

import (
	"fmt"
	"sync"
	"unicode"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = sync.Once{}
var _ = unicode.IsLetter

type PortunixLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var PortunixLexerLexerStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	ChannelNames           []string
	ModeNames              []string
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func portunixlexerLexerInit() {
	staticData := &PortunixLexerLexerStaticData
	staticData.ChannelNames = []string{
		"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
	}
	staticData.ModeNames = []string{
		"DEFAULT_MODE",
	}
	staticData.LiteralNames = []string{
		"", "", "':'", "'='",
	}
	staticData.SymbolicNames = []string{
		"", "LONG_OPTION", "COLON", "EQUAL", "WORD", "STRING", "WS",
	}
	staticData.RuleNames = []string{
		"LONG_OPTION", "COLON", "EQUAL", "WORD", "STRING", "WS",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 6, 46, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 1, 0, 1, 0, 1, 0, 5, 0, 17, 8, 0, 10, 0, 12, 0, 20,
		9, 0, 1, 1, 1, 1, 1, 2, 1, 2, 1, 3, 4, 3, 27, 8, 3, 11, 3, 12, 3, 28, 1,
		4, 1, 4, 5, 4, 33, 8, 4, 10, 4, 12, 4, 36, 9, 4, 1, 4, 1, 4, 1, 5, 4, 5,
		41, 8, 5, 11, 5, 12, 5, 42, 1, 5, 1, 5, 0, 0, 6, 1, 1, 3, 2, 5, 3, 7, 4,
		9, 5, 11, 6, 1, 0, 5, 3, 0, 65, 90, 95, 95, 97, 122, 4, 0, 48, 57, 65,
		90, 95, 95, 97, 122, 4, 0, 46, 57, 65, 90, 95, 95, 97, 122, 3, 0, 10, 10,
		13, 13, 34, 34, 3, 0, 9, 10, 13, 13, 32, 32, 49, 0, 1, 1, 0, 0, 0, 0, 3,
		1, 0, 0, 0, 0, 5, 1, 0, 0, 0, 0, 7, 1, 0, 0, 0, 0, 9, 1, 0, 0, 0, 0, 11,
		1, 0, 0, 0, 1, 13, 1, 0, 0, 0, 3, 21, 1, 0, 0, 0, 5, 23, 1, 0, 0, 0, 7,
		26, 1, 0, 0, 0, 9, 30, 1, 0, 0, 0, 11, 40, 1, 0, 0, 0, 13, 14, 5, 45, 0,
		0, 14, 18, 7, 0, 0, 0, 15, 17, 7, 1, 0, 0, 16, 15, 1, 0, 0, 0, 17, 20,
		1, 0, 0, 0, 18, 16, 1, 0, 0, 0, 18, 19, 1, 0, 0, 0, 19, 2, 1, 0, 0, 0,
		20, 18, 1, 0, 0, 0, 21, 22, 5, 58, 0, 0, 22, 4, 1, 0, 0, 0, 23, 24, 5,
		61, 0, 0, 24, 6, 1, 0, 0, 0, 25, 27, 7, 2, 0, 0, 26, 25, 1, 0, 0, 0, 27,
		28, 1, 0, 0, 0, 28, 26, 1, 0, 0, 0, 28, 29, 1, 0, 0, 0, 29, 8, 1, 0, 0,
		0, 30, 34, 5, 34, 0, 0, 31, 33, 8, 3, 0, 0, 32, 31, 1, 0, 0, 0, 33, 36,
		1, 0, 0, 0, 34, 32, 1, 0, 0, 0, 34, 35, 1, 0, 0, 0, 35, 37, 1, 0, 0, 0,
		36, 34, 1, 0, 0, 0, 37, 38, 5, 34, 0, 0, 38, 10, 1, 0, 0, 0, 39, 41, 7,
		4, 0, 0, 40, 39, 1, 0, 0, 0, 41, 42, 1, 0, 0, 0, 42, 40, 1, 0, 0, 0, 42,
		43, 1, 0, 0, 0, 43, 44, 1, 0, 0, 0, 44, 45, 6, 5, 0, 0, 45, 12, 1, 0, 0,
		0, 5, 0, 18, 28, 34, 42, 1, 6, 0, 0,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// PortunixLexerInit initializes any static state used to implement PortunixLexer. By default the
// static state used to implement the lexer is lazily initialized during the first call to
// NewPortunixLexer(). You can call this function if you wish to initialize the static state ahead
// of time.
func PortunixLexerInit() {
	staticData := &PortunixLexerLexerStaticData
	staticData.once.Do(portunixlexerLexerInit)
}

// NewPortunixLexer produces a new lexer instance for the optional input antlr.CharStream.
func NewPortunixLexer(input antlr.CharStream) *PortunixLexer {
	PortunixLexerInit()
	l := new(PortunixLexer)
	l.BaseLexer = antlr.NewBaseLexer(input)
	staticData := &PortunixLexerLexerStaticData
	l.Interpreter = antlr.NewLexerATNSimulator(l, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	l.channelNames = staticData.ChannelNames
	l.modeNames = staticData.ModeNames
	l.RuleNames = staticData.RuleNames
	l.LiteralNames = staticData.LiteralNames
	l.SymbolicNames = staticData.SymbolicNames
	l.GrammarFileName = "Portunix.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// PortunixLexer tokens.
const (
	PortunixLexerLONG_OPTION = 1
	PortunixLexerCOLON       = 2
	PortunixLexerEQUAL       = 3
	PortunixLexerWORD        = 4
	PortunixLexerSTRING      = 5
	PortunixLexerWS          = 6
)
