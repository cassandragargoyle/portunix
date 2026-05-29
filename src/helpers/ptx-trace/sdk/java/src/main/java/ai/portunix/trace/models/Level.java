/*
 * This file is part of CassandraGargoyle Community Project
 * Licensed under the MIT License - see LICENSE file for details
 */
package ai.portunix.trace.models;

/**
 * Log levels.
 */
public enum Level
{
	DEBUG("debug"),
	INFO("info"),
	WARNING("warning"),
	ERROR("error");

	private final String value;

	Level(String value)
	{
		this.value = value;
	}

	public String getValue()
	{
		return value;
	}

	public static Level fromString(String text)
	{
		for (Level l : Level.values())
		{
			if (l.value.equalsIgnoreCase(text))
			{
				return l;
			}
		}
		return INFO;
	}
}
