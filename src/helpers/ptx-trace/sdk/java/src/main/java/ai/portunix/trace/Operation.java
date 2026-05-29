/*
 * This file is part of CassandraGargoyle Community Project
 * Licensed under the MIT License - see LICENSE file for details
 */
package ai.portunix.trace;

import ai.portunix.trace.cli.CliException;
import ai.portunix.trace.models.Severity;
import ai.portunix.trace.models.SourceInfo;

import java.util.*;

/**
 * Represents a traced operation.
 *
 * <p>
 * Use with try-with-resources for automatic timing and recording:</p>
 * <pre>{@code
 * try (Operation op = session.start("normalize_phone")) {
 *     op.input("phone", rawPhone);
 *     String result = normalize(rawPhone);
 *     op.output("phone", result);
 *     op.success();
 * }
 * }</pre>
 */
public class Operation implements AutoCloseable
{
	private final Session session;

	private final String name;

	private final String type;

	private final List<String> tags;

	private final Map<String, Object> inputData;

	private final Map<String, Object> outputData;

	private final Map<String, Object> contextData;

	private SourceInfo source;

	private String errorMsg;

	private Severity errorSeverity = Severity.MEDIUM;

	private String status;

	private long startTime;

	private Long durationUs;

	private boolean ended;

	Operation(Session session, String name, String type, List<String> tags)
	{
		this.session = session;
		this.name = name;
		this.type = type;
		this.tags = tags != null ? new ArrayList<>(tags) : new ArrayList<>();
		this.inputData = new LinkedHashMap<>();
		this.outputData = new LinkedHashMap<>();
		this.contextData = new LinkedHashMap<>();
		this.startTime = System.nanoTime();
	}

	/**
	 * Get operation name.
	 */
	public String getName()
	{
		return name;
	}

	/**
	 * Set input data field.
	 */
	public Operation input(String key, Object value)
	{
		inputData.put(key, value);
		return this;
	}

	/**
	 * Set multiple input fields.
	 */
	public Operation inputs(Map<String, Object> data)
	{
		inputData.putAll(data);
		return this;
	}

	/**
	 * Set output data field.
	 */
	public Operation output(String key, Object value)
	{
		outputData.put(key, value);
		return this;
	}

	/**
	 * Set multiple output fields.
	 */
	public Operation outputs(Map<String, Object> data)
	{
		outputData.putAll(data);
		return this;
	}

	/**
	 * Set the data source.
	 */
	public Operation source(SourceInfo source)
	{
		this.source = source;
		return this;
	}

	/**
	 * Add tags.
	 */
	public Operation tag(String... tags)
	{
		this.tags.addAll(Arrays.asList(tags));
		return this;
	}

	/**
	 * Add context information.
	 */
	public Operation context(String key, Object value)
	{
		contextData.put(key, value);
		return this;
	}

	/**
	 * Record an error.
	 */
	public Operation error(String message)
	{
		return error(message, Severity.MEDIUM);
	}

	/**
	 * Record an error with severity.
	 */
	public Operation error(String message, Severity severity)
	{
		this.errorMsg = message;
		this.errorSeverity = severity;
		return this;
	}

	/**
	 * Record an error with code.
	 */
	public Operation errorWithCode(String code, String message, Severity severity)
	{
		this.errorMsg = "[" + code + "] " + message;
		this.errorSeverity = severity;
		return this;
	}

	/**
	 * Record an exception as error.
	 */
	public Operation error(Throwable t)
	{
		return error(t.getMessage(), Severity.HIGH);
	}

	/**
	 * Mark the operation as successful.
	 */
	public Operation success()
	{
		this.status = "success";
		return this;
	}

	/**
	 * Set duration manually in microseconds.
	 */
	public Operation setDuration(long durationUs)
	{
		this.durationUs = durationUs;
		return this;
	}

	/**
	 * End the operation and record it.
	 */
	public void end() throws CliException
	{
		if (ended)
		{
			return;
		}
		ended = true;

		// Calculate duration if not set manually
		if (durationUs == null)
		{
			durationUs = (System.nanoTime() - startTime) / 1000;
		}

		// Record event via CLI
		session.getCli().addEvent(
			name,
			inputData.isEmpty() ? null : inputData,
			outputData.isEmpty() ? null : outputData,
			errorMsg == null ? status : null,
			errorMsg,
			tags.isEmpty() ? null : tags,
			durationUs
		);
	}

	@Override
	public void close() throws CliException
	{
		end();
	}

	Session getSession()
	{
		return session;
	}
}
