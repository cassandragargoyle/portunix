// Code generated from Portunix.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // Portunix

import "github.com/antlr4-go/antlr/v4"

// BasePortunixListener is a complete listener for a parse tree produced by PortunixParser.
type BasePortunixListener struct{}

var _ PortunixListener = &BasePortunixListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BasePortunixListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BasePortunixListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BasePortunixListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BasePortunixListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterProgram is called when production program is entered.
func (s *BasePortunixListener) EnterProgram(ctx *ProgramContext) {}

// ExitProgram is called when production program is exited.
func (s *BasePortunixListener) ExitProgram(ctx *ProgramContext) {}

// EnterCommand is called when production command is entered.
func (s *BasePortunixListener) EnterCommand(ctx *CommandContext) {}

// ExitCommand is called when production command is exited.
func (s *BasePortunixListener) ExitCommand(ctx *CommandContext) {}

// EnterParameters is called when production parameters is entered.
func (s *BasePortunixListener) EnterParameters(ctx *ParametersContext) {}

// ExitParameters is called when production parameters is exited.
func (s *BasePortunixListener) ExitParameters(ctx *ParametersContext) {}

// EnterArguments is called when production arguments is entered.
func (s *BasePortunixListener) EnterArguments(ctx *ArgumentsContext) {}

// ExitArguments is called when production arguments is exited.
func (s *BasePortunixListener) ExitArguments(ctx *ArgumentsContext) {}
