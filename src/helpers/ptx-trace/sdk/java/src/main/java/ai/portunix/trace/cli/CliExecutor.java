/*
 * This file is part of CassandraGargoyle Community Project
 * Licensed under the MIT License - see LICENSE file for details
 */
package ai.portunix.trace.cli;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.gson.reflect.TypeToken;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.lang.reflect.Type;
import java.util.*;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;

/**
 * Wrapper for executing portunix trace CLI commands.
 */
public class CliExecutor
{
	private final String binaryPath;

	private final Gson gson;

	private final long timeoutSeconds;

	public CliExecutor()
	{
		this("portunix", 30);
	}

	public CliExecutor(String binaryPath)
	{
		this(binaryPath, 30);
	}

	public CliExecutor(String binaryPath, long timeoutSeconds)
	{
		this.binaryPath = binaryPath;
		this.timeoutSeconds = timeoutSeconds;
		this.gson = new GsonBuilder()
			.setDateFormat("yyyy-MM-dd'T'HH:mm:ss")
			.create();
	}

	/**
	 * Execute a CLI command and return raw output.
	 */
	public String execute(String... args) throws CliException
	{
		List<String> command = new ArrayList<>();
		command.add(binaryPath);
		command.add("trace");
		command.addAll(Arrays.asList(args));

		try
		{
			ProcessBuilder pb = new ProcessBuilder(command);
			pb.redirectErrorStream(false);

			Process process = pb.start();

			StringBuilder stdout = new StringBuilder();
			StringBuilder stderr = new StringBuilder();

			// Read stdout
			try (BufferedReader reader = new BufferedReader(
				new InputStreamReader(process.getInputStream())))
			{
				String line;
				while ((line = reader.readLine()) != null)
				{
					stdout.append(line).append("\n");
				}
			}

			// Read stderr
			try (BufferedReader reader = new BufferedReader(
				new InputStreamReader(process.getErrorStream())))
			{
				String line;
				while ((line = reader.readLine()) != null)
				{
					stderr.append(line).append("\n");
				}
			}

			boolean completed = process.waitFor(timeoutSeconds, TimeUnit.SECONDS);
			if (!completed)
			{
				process.destroyForcibly();
				throw new CliException("Command timed out", -1);
			}

			int exitCode = process.exitValue();
			if (exitCode != 0)
			{
				throw new CliException(
					"Command failed: " + stderr.toString().trim(),
					exitCode
				);
			}

			return stdout.toString().trim();

		}
		catch (IOException e)
		{
			throw new CliException("Failed to execute command: " + e.getMessage(), -1);
		}
		catch (InterruptedException e)
		{
			Thread.currentThread().interrupt();
			throw new CliException("Command interrupted", -1);
		}
	}

	/**
	 * Execute a CLI command and parse JSON output.
	 */
	public <T> T executeJson(Type type, String... args) throws CliException
	{
		// Add --format json to args
		String[] jsonArgs = new String[args.length + 2];
		System.arraycopy(args, 0, jsonArgs, 0, args.length);
		jsonArgs[args.length] = "--format";
		jsonArgs[args.length + 1] = "json";

		String output = execute(jsonArgs);
		if (output == null || output.isEmpty())
		{
			return null;
		}

		try
		{
			return gson.fromJson(output, type);
		}
		catch (Exception e)
		{
			throw new CliException("Failed to parse JSON: " + e.getMessage(), -1);
		}
	}

	/**
	 * Start a new trace session.
	 */
	public String startSession(String name, String source, String destination,
							   List<String> tags, double sampling, boolean piiMask)
		throws CliException
	{

		List<String> args = new ArrayList<>();
		args.add("start");
		args.add(name);

		if (source != null && !source.isEmpty())
		{
			args.add("--source");
			args.add(source);
		}

		if (destination != null && !destination.isEmpty())
		{
			args.add("--destination");
			args.add(destination);
		}

		if (tags != null)
		{
			for (String tag : tags)
			{
				args.add("--tag");
				args.add(tag);
			}
		}

		if (sampling < 1.0)
		{
			args.add("--sampling");
			args.add(String.valueOf(sampling));
		}

		if (piiMask)
		{
			args.add("--pii-mask");
		}

		String output = execute(args.toArray(new String[0]));

		// Parse session ID from output like "Session started: ses_2026-01-27_import"
		for (String line : output.split("\n"))
		{
			if (line.startsWith("Session started:"))
			{
				return line.split(":", 2)[1].trim();
			}
		}

		throw new CliException("Failed to get session ID from output", -1);
	}

	/**
	 * End the active trace session.
	 */
	public void endSession(String status, boolean summary) throws CliException
	{
		List<String> args = new ArrayList<>();
		args.add("end");
		args.add("--status");
		args.add(status);

		if (summary)
		{
			args.add("--summary");
		}

		execute(args.toArray(new String[0]));
	}

	/**
	 * Add a trace event.
	 */
	public void addEvent(String operation, Map<String, Object> inputData,
						 Map<String, Object> outputData, String status,
						 String error, List<String> tags, Long duration)
		throws CliException
	{

		List<String> args = new ArrayList<>();
		args.add("event");
		args.add(operation);

		if (inputData != null && !inputData.isEmpty())
		{
			String inputStr = inputData.entrySet().stream()
				.map(e -> e.getKey() + "=" + e.getValue())
				.collect(Collectors.joining(","));
			args.add("--input");
			args.add(inputStr);
		}

		if (outputData != null && !outputData.isEmpty())
		{
			String outputStr = outputData.entrySet().stream()
				.map(e -> e.getKey() + "=" + e.getValue())
				.collect(Collectors.joining(","));
			args.add("--output");
			args.add(outputStr);
		}

		if (error != null && !error.isEmpty())
		{
			args.add("--error");
			args.add(error);
		}
		else if (status != null && !status.isEmpty())
		{
			args.add("--status");
			args.add(status);
		}

		if (tags != null)
		{
			for (String tag : tags)
			{
				args.add("--tag");
				args.add(tag);
			}
		}

		if (duration != null)
		{
			args.add("--duration");
			args.add(String.valueOf(duration));
		}

		execute(args.toArray(new String[0]));
	}

	/**
	 * List sessions.
	 */
	public List<Map<String, Object>> listSessions(Integer limit, String status)
		throws CliException
	{

		List<String> args = new ArrayList<>();
		args.add("sessions");

		if (limit != null)
		{
			args.add("--limit");
			args.add(String.valueOf(limit));
		}

		if (status != null && !status.isEmpty())
		{
			args.add("--status");
			args.add(status);
		}

		Type listType = new TypeToken<List<Map<String, Object>>>()
		{
		}.getType();
		List<Map<String, Object>> result = executeJson(listType, args.toArray(new String[0]));
		return result != null ? result : new ArrayList<>();
	}

	/**
	 * Get session statistics.
	 */
	public Map<String, Object> getSessionStats(String sessionId) throws CliException
	{
		List<String> args = new ArrayList<>();
		args.add("stats");

		if (sessionId != null && !sessionId.isEmpty())
		{
			args.add(sessionId);
		}

		Type mapType = new TypeToken<Map<String, Object>>()
		{
		}.getType();
		return executeJson(mapType, args.toArray(new String[0]));
	}

	/**
	 * View events.
	 */
	public List<Map<String, Object>> viewEvents(String sessionId, String operation,
												String status, String level,
												String tag, int limit)
		throws CliException
	{

		List<String> args = new ArrayList<>();
		args.add("view");

		if (sessionId != null && !sessionId.isEmpty())
		{
			args.add(sessionId);
		}

		if (operation != null && !operation.isEmpty())
		{
			args.add("--operation");
			args.add(operation);
		}

		if (status != null && !status.isEmpty())
		{
			args.add("--status");
			args.add(status);
		}

		if (level != null && !level.isEmpty())
		{
			args.add("--level");
			args.add(level);
		}

		if (tag != null && !tag.isEmpty())
		{
			args.add("--tag");
			args.add(tag);
		}

		args.add("--limit");
		args.add(String.valueOf(limit));

		Type listType = new TypeToken<List<Map<String, Object>>>()
		{
		}.getType();
		List<Map<String, Object>> result = executeJson(listType, args.toArray(new String[0]));
		return result != null ? result : new ArrayList<>();
	}

	public Gson getGson()
	{
		return gson;
	}
}
