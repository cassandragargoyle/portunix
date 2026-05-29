/*
 * This file is part of CassandraGargoyle Community Project
 * Licensed under the MIT License - see LICENSE file for details
 */
package ai.portunix.trace.models;

/**
 * Session status values.
 */
public enum SessionStatus
{
	ACTIVE("active"),
	COMPLETED("completed"),
	FAILED("failed"),
	CANCELLED("cancelled");

	private final String value;

	SessionStatus(String value)
	{
		this.value = value;
	}

	public String getValue()
	{
		return value;
	}

	public static SessionStatus fromString(String text)
	{
		for (SessionStatus s : SessionStatus.values())
		{
			if (s.value.equalsIgnoreCase(text))
			{
				return s;
			}
		}
		return ACTIVE;
	}
}
