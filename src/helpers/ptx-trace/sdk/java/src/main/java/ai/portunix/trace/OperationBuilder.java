/*
 * This file is part of CassandraGargoyle Community Project
 * Licensed under the MIT License - see LICENSE file for details
 */
package ai.portunix.trace;

import ai.portunix.trace.cli.CliException;
import ai.portunix.trace.models.Severity;
import ai.portunix.trace.models.SourceInfo;

import java.util.*;
import java.util.function.Consumer;

/**
 * Fluent API for building and executing traced operations.
 *
 * <p>
 * Example:</p>
 * <pre>{@code
 * session.trace("validate_email")
 *     .input("email", email)
 *     .tag("validation")
 *     .execute(ctx -> {
 *         boolean valid = validate(email);
 *         ctx.output("valid", valid);
 *     });
 * }</pre>
 */
public class OperationBuilder
{
	private final Session session;

	private final String name;

	private String type = "transform";

	private final List<String> tags = new ArrayList<>();

	private final Map<String, Object> inputData = new LinkedHashMap<>();

	private final Map<String, Object> contextData = new LinkedHashMap<>();

	private SourceInfo source;

	private String ruleId;

	private String ruleVersion;

	OperationBuilder(Session session, String name)
	{
		this.session = session;
		this.name = name;
	}

	/**
	 * Set operation type.
	 */
	public OperationBuilder withType(String type)
	{
		this.type = type;
		return this;
	}

	/**
	 * Set input data field.
	 */
	public OperationBuilder input(String key, Object value)
	{
		inputData.put(key, value);
		return this;
	}

	/**
	 * Set multiple input fields.
	 */
	public OperationBuilder inputs(Map<String, Object> data)
	{
		inputData.putAll(data);
		return this;
	}

	/**
	 * Set the data source.
	 */
	public OperationBuilder source(SourceInfo source)
	{
		this.source = source;
		return this;
	}

	/**
	 * Add tags.
	 */
	public OperationBuilder tag(String... tags)
	{
		this.tags.addAll(Arrays.asList(tags));
		return this;
	}

	/**
	 * Set rule information.
	 */
	public OperationBuilder withRule(String ruleId, String version)
	{
		this.ruleId = ruleId;
		this.ruleVersion = version;
		return this;
	}

	/**
	 * Add context information.
	 */
	public OperationBuilder context(String key, Object value)
	{
		contextData.put(key, value);
		return this;
	}

	/**
	 * Execute the operation with a consumer.
	 *
	 * @param consumer Consumer that receives an OperationContext for recording outputs
	 */
	public void execute(Consumer<OperationContext> consumer) throws CliException
	{
		Operation op = build();
		try
		{
			OperationContext ctx = new OperationContext(op);
			consumer.accept(ctx);
			op.success();
		}
		catch (Exception e)
		{
			op.error(e.getMessage(), Severity.MEDIUM);
			throw e;
		}
		finally
		{
			op.end();
		}
	}

	/**
	 * Build the operation (for use with try-with-resources).
	 *
	 * @return Operation instance
	 */
	public Operation build()
	{
		Operation op = new Operation(session, name, type, tags);

		// Apply builder settings
		for (Map.Entry<String, Object> entry : inputData.entrySet())
		{
			op.input(entry.getKey(), entry.getValue());
		}

		if (source != null)
		{
			op.source(source);
		}

		for (Map.Entry<String, Object> entry : contextData.entrySet())
		{
			op.context(entry.getKey(), entry.getValue());
		}

		if (ruleId != null)
		{
			op.context("rule_id", ruleId);
			if (ruleVersion != null)
			{
				op.context("rule_version", ruleVersion);
			}
		}

		return op;
	}

	/**
	 * Context passed to operation consumers.
	 */
	public static class OperationContext
	{
		private final Operation operation;

		OperationContext(Operation operation)
		{
			this.operation = operation;
		}

		/**
		 * Set output data field.
		 */
		public OperationContext output(String key, Object value)
		{
			operation.output(key, value);
			return this;
		}

		/**
		 * Set multiple output fields.
		 */
		public OperationContext outputs(Map<String, Object> data)
		{
			operation.outputs(data);
			return this;
		}

		/**
		 * Add a tag.
		 */
		public OperationContext tag(String tag)
		{
			operation.tag(tag);
			return this;
		}

		/**
		 * Add context information.
		 */
		public OperationContext context(String key, Object value)
		{
			operation.context(key, value);
			return this;
		}
	}
}
