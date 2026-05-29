/*
 * This file is part of CassandraGargoyle Community Project
 * Licensed under the MIT License - see LICENSE file for details
 */
package ai.portunix.trace;

import ai.portunix.trace.cli.CliException;
import ai.portunix.trace.cli.CliExecutor;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

/**
 * Factory and builder for creating PTX-TRACE sessions.
 *
 * <p>
 * Example:</p>
 * <pre>{@code
 * try (Session session = Trace.newSession("import-customers")
 *         .withSource("customers.csv")
 *         .withTags("production", "daily")
 *         .withPIIMasking(true)
 *         .withSampling(0.5)
 *         .build()) {
 *
 *     // Trace operations...
 * }
 * }</pre>
 */
public class Trace
{
	private static String defaultBinaryPath = "portunix";

	private static long defaultTimeout = 30;

	/**
	 * Set the default binary path for all new sessions.
	 */
	public static void setDefaultBinaryPath(String path)
	{
		defaultBinaryPath = path;
	}

	/**
	 * Set the default timeout for CLI commands.
	 */
	public static void setDefaultTimeout(long timeoutSeconds)
	{
		defaultTimeout = timeoutSeconds;
	}

	/**
	 * Create a new session builder.
	 *
	 * @param name Session name
	 * @return SessionBuilder for fluent configuration
	 */
	public static SessionBuilder newSession(String name)
	{
		return new SessionBuilder(name);
	}

	/**
	 * Create and start a session with default settings.
	 *
	 * @param name Session name
	 * @return Started session
	 */
	public static Session createSession(String name) throws CliException
	{
		return newSession(name).build();
	}

	/**
	 * Load an existing session.
	 *
	 * @param sessionId Session ID
	 * @return Loaded session
	 */
	public static Session loadSession(String sessionId) throws CliException
	{
		CliExecutor cli = new CliExecutor(defaultBinaryPath, defaultTimeout);
		return Session.load(cli, sessionId);
	}

	/**
	 * List all sessions.
	 */
	public static List<Map<String, Object>> listSessions() throws CliException
	{
		return listSessions(null, null);
	}

	/**
	 * List sessions with filters.
	 */
	public static List<Map<String, Object>> listSessions(Integer limit, String status)
		throws CliException
	{
		CliExecutor cli = new CliExecutor(defaultBinaryPath, defaultTimeout);
		return Session.list(cli, limit, status);
	}

	/**
	 * Builder for creating sessions.
	 */
	public static class SessionBuilder
	{
		private final String name;

		private String binaryPath = defaultBinaryPath;

		private long timeout = defaultTimeout;

		private String source;

		private String destination;

		private final List<String> tags = new ArrayList<>();

		private double sampling = 1.0;

		private boolean piiMasking = false;

		SessionBuilder(String name)
		{
			this.name = name;
		}

		/**
		 * Set the portunix binary path.
		 */
		public SessionBuilder withBinaryPath(String path)
		{
			this.binaryPath = path;
			return this;
		}

		/**
		 * Set the CLI timeout.
		 */
		public SessionBuilder withTimeout(long timeoutSeconds)
		{
			this.timeout = timeoutSeconds;
			return this;
		}

		/**
		 * Set the data source.
		 */
		public SessionBuilder withSource(String source)
		{
			this.source = source;
			return this;
		}

		/**
		 * Set the data destination.
		 */
		public SessionBuilder withDestination(String destination)
		{
			this.destination = destination;
			return this;
		}

		/**
		 * Add tags.
		 */
		public SessionBuilder withTags(String... tags)
		{
			for (String tag : tags)
			{
				this.tags.add(tag);
			}
			return this;
		}

		/**
		 * Set the sampling rate (0.0 to 1.0).
		 */
		public SessionBuilder withSampling(double rate)
		{
			this.sampling = rate;
			return this;
		}

		/**
		 * Enable PII masking.
		 */
		public SessionBuilder withPIIMasking(boolean enabled)
		{
			this.piiMasking = enabled;
			return this;
		}

		/**
		 * Build and start the session.
		 *
		 * @return Started session
		 */
		public Session build() throws CliException
		{
			CliExecutor cli = new CliExecutor(binaryPath, timeout);
			Session session = new Session(cli, name, source, destination,
										  tags, sampling, piiMasking, null);
			session.start();
			return session;
		}
	}
}
