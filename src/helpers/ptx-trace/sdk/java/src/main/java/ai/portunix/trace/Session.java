/*
 * This file is part of CassandraGargoyle Community Project
 * Licensed under the MIT License - see LICENSE file for details
 */
package ai.portunix.trace;

import ai.portunix.trace.cli.CliException;
import ai.portunix.trace.cli.CliExecutor;
import ai.portunix.trace.models.SessionStatus;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

/**
 * Manages a PTX-TRACE session.
 *
 * <p>
 * Use with try-with-resources for automatic session lifecycle:</p>
 * <pre>{@code
 * try (Session session = Trace.newSession("import-customers")
 *         .withPIIMasking(true)
 *         .build()) {
 *
 *     try (Operation op = session.start("normalize_phone")) {
 *         op.input("phone", rawPhone);
 *         String result = normalize(rawPhone);
 *         op.output("phone", result);
 *         op.success();
 *     }
 * }
 * }</pre>
 */
public class Session implements AutoCloseable
{
	private final CliExecutor cli;

	private final String name;

	private final String source;

	private final String destination;

	private final List<String> tags;

	private final double sampling;

	private final boolean piiMasking;

	private String sessionId;

	private boolean started;

	private boolean closed;

	Session(CliExecutor cli, String name, String source, String destination,
			List<String> tags, double sampling, boolean piiMasking, String sessionId)
	{
		this.cli = cli;
		this.name = name;
		this.source = source;
		this.destination = destination;
		this.tags = tags != null ? new ArrayList<>(tags) : new ArrayList<>();
		this.sampling = sampling;
		this.piiMasking = piiMasking;
		this.sessionId = sessionId;
		this.started = sessionId != null;
	}

	/**
	 * Get session ID.
	 */
	public String getId()
	{
		return sessionId;
	}

	/**
	 * Get session name.
	 */
	public String getName()
	{
		return name;
	}

	/**
	 * Check if session is active.
	 */
	public boolean isActive()
	{
		return started && !closed;
	}

	/**
	 * Start the session.
	 *
	 * @return Session ID
	 */
	public String start() throws CliException
	{
		if (started)
		{
			return sessionId;
		}

		sessionId = cli.startSession(name, source, destination, tags, sampling, piiMasking);
		started = true;
		return sessionId;
	}

	/**
	 * End the session with a specific status.
	 */
	public void end(SessionStatus status) throws CliException
	{
		if (closed)
		{
			return;
		}

		cli.endSession(status.getValue(), false);
		closed = true;
	}

	/**
	 * End the session successfully.
	 */
	@Override
	public void close() throws CliException
	{
		end(SessionStatus.COMPLETED);
	}

	/**
	 * End the session as failed.
	 */
	public void fail() throws CliException
	{
		end(SessionStatus.FAILED);
	}

	/**
	 * End the session as cancelled.
	 */
	public void cancel() throws CliException
	{
		end(SessionStatus.CANCELLED);
	}

	/**
	 * Create a traced operation builder.
	 *
	 * @param operationName Name of the operation
	 * @return OperationBuilder for fluent configuration
	 */
	public OperationBuilder trace(String operationName)
	{
		return new OperationBuilder(this, operationName);
	}

	/**
	 * Start a traced operation directly.
	 *
	 * @param operationName Name of the operation
	 * @return Operation instance for try-with-resources
	 */
	public Operation start(String operationName)
	{
		return new Operation(this, operationName, "transform", null);
	}

	/**
	 * Start a traced operation with type.
	 *
	 * @param type          Operation type
	 * @param operationName Name of the operation
	 * @return Operation instance
	 */
	public Operation start(String type, String operationName)
	{
		return new Operation(this, operationName, type, null);
	}

	/**
	 * Get session statistics.
	 */
	public Map<String, Object> stats() throws CliException
	{
		return cli.getSessionStats(sessionId);
	}

	/**
	 * Get events from this session.
	 */
	public List<Map<String, Object>> events(String operation, String status,
											String level, String tag, int limit)
		throws CliException
	{
		return cli.viewEvents(sessionId, operation, status, level, tag, limit);
	}

	/**
	 * Get events from this session.
	 */
	public List<Map<String, Object>> events(int limit) throws CliException
	{
		return events(null, null, null, null, limit);
	}

	/**
	 * Get events from this session.
	 */
	public List<Map<String, Object>> events() throws CliException
	{
		return events(100);
	}

	/**
	 * List all sessions.
	 */
	public static List<Map<String, Object>> list(CliExecutor cli, Integer limit, String status)
		throws CliException
	{
		return cli.listSessions(limit, status);
	}

	/**
	 * Load an existing session.
	 */
	public static Session load(CliExecutor cli, String sessionId) throws CliException
	{
		Map<String, Object> data = cli.getSessionStats(sessionId);
		String name = (String) data.get("name");
		@SuppressWarnings("unchecked")
		List<String> tags = (List<String>) data.get("tags");

		return new Session(cli, name, null, null, tags, 1.0, false, sessionId);
	}

	CliExecutor getCli()
	{
		return cli;
	}
}
