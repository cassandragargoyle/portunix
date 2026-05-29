/*
 * This file is part of CassandraGargoyle Community Project
 * Licensed under the MIT License - see LICENSE file for details
 */
package ai.portunix.trace.models;

import java.util.HashMap;
import java.util.Map;

/**
 * Data source information.
 */
public class SourceInfo
{
	private final String type;

	private String file;

	private Integer row;

	private String column;

	private String url;

	private String table;

	public SourceInfo(String type)
	{
		this.type = type;
	}

	public String getType()
	{
		return type;
	}

	public String getFile()
	{
		return file;
	}

	public SourceInfo setFile(String file)
	{
		this.file = file;
		return this;
	}

	public Integer getRow()
	{
		return row;
	}

	public SourceInfo setRow(Integer row)
	{
		this.row = row;
		return this;
	}

	public String getColumn()
	{
		return column;
	}

	public SourceInfo setColumn(String column)
	{
		this.column = column;
		return this;
	}

	public String getUrl()
	{
		return url;
	}

	public SourceInfo setUrl(String url)
	{
		this.url = url;
		return this;
	}

	public String getTable()
	{
		return table;
	}

	public SourceInfo setTable(String table)
	{
		this.table = table;
		return this;
	}

	public Map<String, Object> toMap()
	{
		Map<String, Object> map = new HashMap<>();
		map.put("type", type);
		if (file != null)
			map.put("file", file);
		if (row != null)
			map.put("row", row);
		if (column != null)
			map.put("column", column);
		if (url != null)
			map.put("url", url);
		if (table != null)
			map.put("table", table);
		return map;
	}

	/**
	 * Create a CSV file source.
	 */
	public static SourceInfo csv(String file)
	{
		return new SourceInfo("csv").setFile(file);
	}

	/**
	 * Create a CSV file source with row.
	 */
	public static SourceInfo csv(String file, int row)
	{
		return new SourceInfo("csv").setFile(file).setRow(row);
	}

	/**
	 * Create a CSV file source with row and column.
	 */
	public static SourceInfo csv(String file, int row, String column)
	{
		return new SourceInfo("csv").setFile(file).setRow(row).setColumn(column);
	}

	/**
	 * Create a database source.
	 */
	public static SourceInfo database(String url, String table)
	{
		return new SourceInfo("database").setUrl(url).setTable(table);
	}

	/**
	 * Create a database source with row.
	 */
	public static SourceInfo database(String url, String table, int row)
	{
		return new SourceInfo("database").setUrl(url).setTable(table).setRow(row);
	}

	/**
	 * Create a file source.
	 */
	public static SourceInfo file(String file)
	{
		return new SourceInfo("file").setFile(file);
	}

	/**
	 * Create an API source.
	 */
	public static SourceInfo api(String url)
	{
		return new SourceInfo("api").setUrl(url);
	}
}
