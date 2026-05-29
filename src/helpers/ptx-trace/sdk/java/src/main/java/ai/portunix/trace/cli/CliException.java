/*
 * This file is part of CassandraGargoyle Community Project
 * Licensed under the MIT License - see LICENSE file for details
 */
package ai.portunix.trace.cli;

/**
 * Exception thrown when CLI command fails.
 */
public class CliException extends Exception
{
	private final int exitCode;

	public CliException(String message, int exitCode)
	{
		super(message);
		this.exitCode = exitCode;
	}

	public int getExitCode()
	{
		return exitCode;
	}
}
