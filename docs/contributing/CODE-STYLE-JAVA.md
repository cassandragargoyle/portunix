# Java Code Style Guidelines

## Purpose

This document defines the Java coding standards for CassandraGargoyle projects, based on industry best practices and established Java conventions.

## General Principles

### 1. Follow Java Conventions

- Use Oracle's Java Code Conventions as baseline
- Follow naming conventions strictly
- Use modern Java features (Java 8+ idioms)
- Embrace object-oriented design principles

### 2. Code Quality

- Write self-documenting code
- Prefer composition over inheritance
- Use interfaces for contracts
- Apply SOLID principles

## File and Package Structure

### File Naming

- Use PascalCase matching class name: `ConfigurationManager.java`
- One public class per file
- File name must match public class name

### Package Naming

Use reverse domain notation with lowercase:

```java
✅ org.cassandragargoyle.portunix.server
✅ org.cassandragargoyle.portunix.agent
❌ org.cassandragargoyle.Portunix.server
❌ org.cassandragargoyle.portunix.Agent
```

### Directory Structure

```text
src/
├── main/
│   ├── java/
│   │   └── org/cassandragargoyle/project/
│   │       ├── Application.java
│   │       ├── config/
│   │       │   ├── ConfigurationManager.java
│   │       │   └── DatabaseConfig.java
│   │       ├── service/
│   │       │   ├── PackageInstaller.java
│   │       │   └── SystemDetector.java
│   │       └── util/
│   │           └── FileUtils.java
│   └── resources/
│       ├── application.properties
│       └── logback.xml
└── test/
    └── java/
        └── org/cassandragargoyle/project/
            ├── config/
            └── service/
```

## Naming Conventions

### Classes and Interfaces

- **Classes**: PascalCase (`ConfigurationManager`, `PackageInstaller`)
- **Interfaces**: PascalCase, often with `-able` suffix or clear contract names
- **Abstract classes**: PascalCase, consider `Abstract` prefix if needed

```java
// Classes
public class PackageInstaller { }
public class DatabaseConnectionPool { }

// Interfaces
public interface Configurable { }
public interface PackageManager { }
public interface Runnable { }

// Abstract classes
public abstract class AbstractPackageManager { }
```

### Methods

Use camelCase with verbs describing actions:
```java
// Good method names
public void installPackage(String packageName) { }
public boolean isConfigurationValid() { }
public Configuration loadConfiguration() { }
public List<Package> getInstalledPackages() { }

// Avoid
public void package() { }  // Not descriptive
public boolean valid() { }  // Missing context
```

### Variables and Fields

- **Instance/Local variables**: camelCase
- **Constants**: ALL_CAPS with underscores
- **Static final fields**: ALL_CAPS with underscores

```java

public class ConfigurationManager
{
	// Constants
	private static final String DEFAULT_CONFIG_PATH = "/etc/config.xml";
	private static final int MAX_RETRY_ATTEMPTS = 3;

	// Instance fields
	private String configurationPath;
	private boolean isInitialized;
	private List<ConfigurationListener> listeners;

	public void processConfiguration()
	{
		// Local variables
		String currentPath = getCurrentPath();
		boolean validationResult = validateConfiguration(currentPath);
	}
}
```

## Code Formatting

### Indentation and Spacing

- Use tabs for indentation (not spaces)
- Maximum line length: 120 characters
- Use blank lines to separate logical blocks

### Braces and Line Breaks

Follow Allman style (braces on new line):

```java
// Method declarations
public void methodName()
{
    if (condition)
    {
        // statements
    }
    else
    {
        // statements
    }

    for (String item : items)
    {
        processItem(item);
    }
}

// Class declarations
public class ClassName
{
    // content
}
```

### Multi-line Method Calls

> **TODO: PENDING VALIDATION** - Continuation line alignment needs to be validated by authority.

When a method call spans multiple lines, align continuation arguments under the first argument:

```java
// Correct - continuation aligned under first argument
LOG.log(Level.FINE, "Executing with processor: {0}, verbose: {1}",
		LogFactory.args(processorType, verbose));

// Correct - builder pattern with single tab indent
return ToolSchema.builder()
	.addStringProperty("input", "Input data")
	.addStringProperty("processor", "Processor type")
	.required("input")
	.build();
```

### Multi-line String Concatenation

> **TODO: PENDING VALIDATION** - String concatenation line break style needs to be validated by authority.

When string concatenation spans multiple lines, put `+` operator at the beginning of continuation line:

```java
// Correct - + at beginning of new line, aligned under first character
throw new IllegalArgumentException("Unsupported format: " + format
								   + ". Supported formats: " + String.join(", ", SUPPORTED_FORMATS));
```

## Import Statements

### Import Organization

Group imports in this order with blank lines between groups:

1. Java standard library (`java.*`, `javax.*`)
2. Third-party libraries
3. Local project packages

```java
// Standard library
import java.io.IOException;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.List;
import java.util.Map;

// Third-party libraries
import org.apache.commons.lang3.StringUtils;

// Local packages
import org.cassandragargoyle.project.config.Configuration;
import org.cassandragargoyle.project.service.PackageInstaller;
```

### Import Guidelines

- Avoid wildcard imports (`import java.util.*;`)
- Use static imports sparingly, only for frequently used utility methods
- Organize imports alphabetically within groups

```java
// Acceptable static imports
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.cassandragargoyle.project.util.Constants.DEFAULT_TIMEOUT;
```

## Class Design

### Enums

> **TODO: PENDING VALIDATION** - Empty line before closing brace in enum needs to be validated by authority.

```java
public enum ProcessingMode
{
	SYNC,
	ASYNC,
	BATCH

}
```

### Class Structure

Organize class members in this order:
1. Static constants
2. Static variables
3. Instance variables
4. Enums (if inner)
5. Constructors
6. Static methods
7. Instance methods
8. Nested classes

```java
public class PackageInstaller
{
	// 1. Static constants
	private static final Logger LOG = LogFactory.getLogger(PackageInstaller.class);

	private static final String DEFAULT_PACKAGE_MANAGER = "apt";

	// 2. Static variables
	private static PackageInstaller instance;

	// 3. Instance variables
	private String packageManager;
	private boolean isInitialized;
	private final Map<String, InstallerConfig> configs;

	// 4. Enums (if inner)
	// (none in this example)

	// 5. Constructors
	public PackageInstaller()
	{
		this.configs = new HashMap<>();
		this.packageManager = detectPackageManager();
	}

	public PackageInstaller(String packageManager)
	{
		this.configs = new HashMap<>();
		this.packageManager = packageManager;
	}

	// 6. Static methods
	public static PackageInstaller getInstance()
	{
		if (instance == null)
		{
			instance = new PackageInstaller();
		}
		return instance;
	}

	// 7. Instance methods
	public void installPackage(String packageName)
	{
		// implementation
	}

	private String detectPackageManager()
	{
		// implementation
	}

	// 8. Nested classes
	private static class InstallerConfig
	{
		// nested class implementation
	}
}
```

## Error Handling

### Exception Handling Best Practices

1. Use specific exception types
2. Always include meaningful error messages
3. Log errors appropriately
4. Clean up resources in finally blocks or use try-with-resources

```java
public Configuration loadConfiguration(String configPath) throws ConfigurationException
{
	if (configPath == null || configPath.trim().isEmpty())
	{
		throw new IllegalArgumentException("Configuration path cannot be null or empty");
	}

	try (InputStream inputStream = Files.newInputStream(Paths.get(configPath)))
	{
		return parseConfiguration(inputStream);
	}
	catch (IOException e)
	{
		LOG.log(Level.SEVERE, "Failed to read configuration file: {0}", configPath);
		throw new ConfigurationException("Failed to read configuration file: " + configPath, e);
	}
	catch (ParseException e)
	{
		LOG.log(Level.SEVERE, "Invalid configuration format in file: {0}", configPath);
		throw new ConfigurationException("Invalid configuration format in file: " + configPath, e);
	}
}
```

### Custom Exceptions

Create domain-specific exception classes:
```java
public class ConfigurationException extends Exception
{
	public ConfigurationException(String message)
	{
		super(message);
	}

	public ConfigurationException(String message, Throwable cause)
	{
		super(message, cause);
	}
}

public class PackageInstallationException extends Exception
{
	private final String packageName;

	public PackageInstallationException(String packageName, String message)
	{
		super(String.format("Failed to install package '%s': %s", packageName, message));
		this.packageName = packageName;
	}

	public String getPackageName()
	{
		return packageName;
	}
}
```

## Logging

### Logger Declaration

All classes use consistent logging with `LogFactory` from `org.cassandragargoyle.api.log`:

```java
import org.cassandragargoyle.api.log.LogFactory;
import java.util.logging.Level;
import java.util.logging.Logger;

public class MyService
{
	private static final Logger LOG = LogFactory.getLogger(MyService.class);

	public void doSomething()
	{
		LOG.log(Level.INFO, "Starting operation");
		LOG.log(Level.FINE, "Processing file: {0}", fileName);
		LOG.log(Level.WARNING, "Warning for file: {0}", fileName);
		LOG.log(Level.SEVERE, "Error processing: {0}, cause: {1}", LogFactory.args(fileName, exception.getMessage()));
	}
}
```

### Logging Guidelines

- **Logger name**: Always use `LOG` (not `LOGGER` or other names)
- **Logger factory**: Use `LogFactory.getLogger()` from `org.cassandragargoyle.api.log`
- **Logger type**: Use `java.util.logging.Logger`
- **Logging method**: Use `LOG.log(Level.XXX, message, args)`
- **Parameters**:
  - Single parameter: pass directly as third argument
  - Multiple parameters: use `LogFactory.args(...)`
  - Use `{0}`, `{1}`, etc. as placeholders (java.util.logging format)
- **Log levels** (java.util.logging.Level):
  - `Level.SEVERE` - For errors that need attention
  - `Level.WARNING` - For potential issues
  - `Level.INFO` - For important operational messages
  - `Level.FINE` / `Level.FINER` / `Level.FINEST` - For detailed debugging information

## Documentation and Comments

### Javadoc Comments

Use Javadoc for all public APIs.

> **TODO: PENDING VALIDATION** - The following Javadoc formatting styles need to be validated by authority:
> - `<p>` tag on separate line
> - `<li>` tags without indentation inside `<ul>`

```java
/**
 * Manages package installation across different operating systems.
 *
 * <p>
 * This class provides a unified interface for installing software packages
 * using various package managers including APT, Chocolatey, and Homebrew.
 * It automatically detects the appropriate package manager based on the
 * operating system.</p>
 *
 * <p>
 * Supported features:</p>
 * <ul>
 * <li>Cross-platform support</li>
 * <li>Automatic package manager detection</li>
 * <li>Retry mechanism</li>
 * </ul>
 *
 * @author <OS username>
 * @version 1.0
 * @since 1.0
 */
public class PackageInstaller
{

	/**
	 * Installs the specified package using the appropriate package manager.
	 *
	 * @param packageName the name of the package to install, must not be null or empty
	 * @throws IllegalArgumentException if packageName is null or empty
	 * @throws PackageInstallationException if installation fails
	 */
	public void installPackage(String packageName) throws PackageInstallationException
	{
		// implementation
	}
}
```

### Javadoc @param Alignment

> **TODO: PENDING VALIDATION** - Javadoc @param alignment style needs to be validated by authority.

When method has multiple parameters, align descriptions into columns:

```java
/**
 * Registers a custom transformer.
 *
 * @param name        the transformer name
 * @param transformer the transformation function
 * @return true if registered successfully, false if name already exists
 */
public boolean registerTransformer(String name, Function<String, String> transformer)
{
	// implementation
}
```

### Inline Comments

- Explain complex algorithms and business logic
- Avoid stating the obvious
- Keep comments up to date with code changes

```java
public boolean detectWindowsSandbox()
{
	// Windows Sandbox has specific registry entries and limited filesystem access
	// Check for these indicators to determine if we're running in sandbox mode
	if (!isWindows())
	{
		return false;
	}

	try
	{
		// Sandbox environments typically have restricted access to certain registry keys
		String sandboxIndicator = System.getProperty("os.name");
		return sandboxIndicator != null && sandboxIndicator.contains("Sandbox");
	}
	catch (SecurityException e)
	{
		// Limited registry access might indicate sandbox environment
		LOG.log(Level.FINE, "Registry access restricted, might be in sandbox environment: {0}", e.getMessage());
		return true;
	}
}
```

## Modern Java Features

### Streams API

Use streams for collection processing:

```java
public List<String> getInstalledPackageNames()
{
	return installedPackages.stream()
		.filter(Package::isActive)
		.map(Package::getName)
		.sorted()
		.collect(Collectors.toList());
}

public Optional<Package> findPackage(String name)
{
	return installedPackages.stream()
		.filter(pkg -> pkg.getName().equals(name))
		.findFirst();
}
```

### Optional Usage

Use Optional to handle nullable values:

```java
public Optional<Configuration> loadOptionalConfiguration(String path)
{
	try
	{
		Configuration config = loadConfiguration(path);
		return Optional.of(config);
	}
	catch (ConfigurationException e)
	{
		LOG.log(Level.WARNING, "Failed to load optional configuration: {0}", path);
		return Optional.empty();
	}
}

public void processConfiguration(String configPath)
{
	loadOptionalConfiguration(configPath)
		.ifPresentOrElse(
			this::applyConfiguration,
			this::useDefaultConfiguration
		);
}
```

### Lambda Expressions

Use lambdas for functional interfaces:
```java
// Event handling
button.addActionListener(event -> handleInstallation());

// Comparators
packages.sort((p1, p2) -> p1.getName().compareTo(p2.getName()));

// Custom functional interfaces
public interface ConfigurationProcessor
{
	void process(Configuration config) throws ProcessingException;
}

public void withConfiguration(ConfigurationProcessor processor)
{
	try
	{
		Configuration config = loadConfiguration();
		processor.process(config);
	}
	catch (ProcessingException e)
	{
		LOG.log(Level.SEVERE, "Failed to process configuration: {0}", e.getMessage());
	}
}
```

## Testing Guidelines

### Test Class Organization

- Use same package structure as main code
- Name test classes with `Test` suffix
- Group tests logically using nested classes

```java
public class PackageInstallerTest
{

	private PackageInstaller installer;

	@BeforeEach
	void setUp()
	{
		installer = new PackageInstaller();
	}

	@Nested
	@DisplayName("Package Installation Tests")
	class PackageInstallationTests
	{

		@Test
		@DisplayName("Should install valid package successfully")
		void shouldInstallValidPackage()
		{
			// Arrange
			String packageName = "test-package";

			// Act & Assert
			assertDoesNotThrow(() -> installer.installPackage(packageName));
		}

		@Test
		@DisplayName("Should throw exception for null package name")
		void shouldThrowExceptionForNullPackageName()
		{
			// Act & Assert
			IllegalArgumentException exception = assertThrows(
				IllegalArgumentException.class,
				() -> installer.installPackage(null)
			);

			assertEquals("Package name cannot be null or empty", exception.getMessage());
		}
	}
}
```

### Test Naming

Use descriptive test method names:

```java
// Good test names
@Test void shouldReturnTrueWhenPackageExists() { }
@Test void shouldThrowExceptionWhenConfigurationFileNotFound() { }
@Test void shouldInstallPackageSuccessfullyOnLinux() { }

// Avoid generic names
@Test void testInstall() { }
@Test void testConfig() { }
```

## Dependency Management

### Maven Configuration

Use Maven for dependency management with clear grouping:

```xml
<dependencies>
    <!-- Core dependencies -->
    <dependency>
        <groupId>org.apache.commons</groupId>
        <artifactId>commons-lang3</artifactId>
        <version>3.12.0</version>
    </dependency>
    
    <!-- Testing dependencies -->
    <dependency>
        <groupId>org.junit.jupiter</groupId>
        <artifactId>junit-jupiter</artifactId>
        <version>5.9.2</version>
        <scope>test</scope>
    </dependency>
</dependencies>
```

## Security Best Practices

### Input Validation

Always validate external input:

```java
public void installPackage(String packageName) throws PackageInstallationException
{
	validatePackageName(packageName);
	// ... rest of implementation
}

private void validatePackageName(String packageName)
{
	if (packageName == null || packageName.trim().isEmpty())
	{
		throw new IllegalArgumentException("Package name cannot be null or empty");
	}

	// Prevent command injection
	if (packageName.contains(";") || packageName.contains("&") || packageName.contains("|"))
	{
		throw new IllegalArgumentException("Package name contains invalid characters");
	}

	// Length validation
	if (packageName.length() > 100)
	{
		throw new IllegalArgumentException("Package name is too long");
	}
}
```

### Resource Management

Use try-with-resources for automatic resource cleanup:

```java
public String readConfigurationFile(String path) throws IOException
{
	try (BufferedReader reader = Files.newBufferedReader(Paths.get(path), StandardCharsets.UTF_8))
	{
		return reader.lines()
			.collect(Collectors.joining(System.lineSeparator()));
	}
}
```

## Performance Guidelines

### Collection Usage

Choose appropriate collection types:

```java
// For frequent lookups
private final Map<String, PackageInfo> packageCache = new HashMap<>();

// For ordered data
private final List<String> installationOrder = new ArrayList<>();

// For unique items
private final Set<String> installedPackages = new HashSet<>();

// For thread-safe access
private final Map<String, PackageInfo> threadSafeCache = new ConcurrentHashMap<>();
```

### String Handling

Use StringBuilder for string concatenation in loops:

```java
public String buildInstallCommand(List<String> packages)
{
	StringBuilder command = new StringBuilder("install");
	for (String packageName : packages)
	{
		command.append(" ").append(packageName);
	}
	return command.toString();
}
```

## Build Configuration

### Maven Plugins

Essential plugins for code quality:

```xml
<build>
    <plugins>
        <!-- Compiler plugin -->
        <plugin>
            <groupId>org.apache.maven.plugins</groupId>
            <artifactId>maven-compiler-plugin</artifactId>
            <version>3.11.0</version>
            <configuration>
                <source>11</source>
                <target>11</target>
                <encoding>UTF-8</encoding>
            </configuration>
        </plugin>
        
        <!-- Code formatting -->
        <plugin>
            <groupId>com.spotify.fmt</groupId>
            <artifactId>fmt-maven-plugin</artifactId>
            <version>2.21.1</version>
            <executions>
                <execution>
                    <goals>
                        <goal>format</goal>
                    </goals>
                </execution>
            </executions>
        </plugin>
        
        <!-- Static analysis -->
        <plugin>
            <groupId>com.github.spotbugs</groupId>
            <artifactId>spotbugs-maven-plugin</artifactId>
            <version>4.7.3.6</version>
        </plugin>
    </plugins>
</build>
```

## Code Review Checklist

### Before Submitting Code

- [ ] Code follows naming conventions
- [ ] All public APIs have Javadoc
- [ ] Exception handling is appropriate
- [ ] Tests cover new functionality
- [ ] No code smells or warnings
- [ ] Imports are organized correctly
- [ ] Resource cleanup is handled properly

### Design Review Points

- [ ] Classes have single responsibility
- [ ] Methods are focused and concise
- [ ] Appropriate design patterns used
- [ ] Error handling strategy is consistent
- [ ] Performance considerations addressed
- [ ] Security implications considered

---

**Note**: These guidelines should be adapted based on specific project requirements and team preferences. Regular review and updates ensure alignment with evolving best practices.

*Created: 2025-08-23*
*Last updated: 2026-01-20*