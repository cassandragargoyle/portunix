// Code generated from Portunix.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // Portunix

import "github.com/antlr4-go/antlr/v4"

// PortunixListener is a complete listener for a parse tree produced by PortunixParser.
type PortunixListener interface {
	antlr.ParseTreeListener

	// EnterProgram is called when entering the program production.
	EnterProgram(c *ProgramContext)

	// EnterCommand is called when entering the command production.
	EnterCommand(c *CommandContext)

	// EnterParameters is called when entering the parameters production.
	EnterParameters(c *ParametersContext)

	// EnterArguments is called when entering the arguments production.
	EnterArguments(c *ArgumentsContext)

	// ExitProgram is called when exiting the program production.
	ExitProgram(c *ProgramContext)

	// ExitCommand is called when exiting the command production.
	ExitCommand(c *CommandContext)

	// ExitParameters is called when exiting the parameters production.
	ExitParameters(c *ParametersContext)

	// ExitArguments is called when exiting the arguments production.
	ExitArguments(c *ArgumentsContext)
}
