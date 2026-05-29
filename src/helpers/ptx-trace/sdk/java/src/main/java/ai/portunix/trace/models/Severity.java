/*
 * This file is part of CassandraGargoyle Community Project
 * Licensed under the MIT License - see LICENSE file for details
 */
package ai.portunix.trace.models;

/**
 * Error severity levels.
 */
public enum Severity
{
	LOW("low"),
	MEDIUM("medium"),
	HIGH("high"),
	CRITICAL("critical");

	private final String value;

	Severity(String value)
	{
		this.value = value;
	}

	public String getValue()
	{
		return value;
	}

	public static Severity fromString(String text)
	{
		for (Severity s : Severity.values())
		{
			if (s.value.equalsIgnoreCase(text))
			{
				return s;
			}
		}
		return MEDIUM;
	}
}
