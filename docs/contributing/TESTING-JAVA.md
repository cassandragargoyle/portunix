# Java Testing Guidelines

## Purpose
This document defines the testing strategy, structure, and best practices for Java projects in the CassandraGargoyle ecosystem.

## Testing Philosophy

### Testing Pyramid
1. **Unit Tests** (70%) - Fast, isolated tests for individual classes/methods
2. **Integration Tests** (20%) - Test component interactions and external systems
3. **End-to-End Tests** (10%) - Full application workflow tests

### Key Principles
- Follow Test-Driven Development (TDD) when possible
- Tests should be fast, reliable, and independent
- Each test should verify one specific behavior
- Use descriptive test names that explain the scenario
- Mock external dependencies and system interactions
- **Container Testing**: Use Podman for all container-based testing (preferred over Docker)

## Project Structure

### Maven Project Layout
```
project/
├── src/
│   ├── main/
│   │   ├── java/
│   │   │   └── com/cassandragargoyle/project/
│   │   │       ├── config/
│   │   │       ├── service/
│   │   │       └── util/
│   │   └── resources/
│   │       ├── application.properties
│   │       └── logback.xml
│   └── test/
│       ├── java/
│       │   └── com/cassandragargoyle/project/
│       │       ├── config/
│       │       │   ├── ConfigManagerTest.java
│       │       │   └── ConfigManagerIntegrationTest.java
│       │       ├── service/
│       │       │   ├── PackageInstallerTest.java
│       │       │   └── PackageInstallerIntegrationTest.java
│       │       └── util/
│       │           └── TestUtils.java
│       └── resources/
│           ├── application-test.properties
│           ├── test-data/
│           │   ├── valid-config.xml
│           │   └── sample-packages.json
│           └── fixtures/
├── target/                    # Build output
├── scripts/                   # Test scripts
│   ├── test.sh               # Main test runner
│   ├── test-unit.sh          # Unit tests only
│   ├── test-integration.sh   # Integration tests
│   └── coverage.sh           # Coverage reporting
└── pom.xml
```

## Test Dependencies

### Maven Dependencies
```xml
<dependencies>
    <!-- Main application dependencies -->
    <dependency>
        <groupId>org.slf4j</groupId>
        <artifactId>slf4j-api</artifactId>
        <version>1.7.36</version>
    </dependency>
    
    <!-- Test dependencies -->
    <dependency>
        <groupId>org.junit.jupiter</groupId>
        <artifactId>junit-jupiter</artifactId>
        <version>5.9.2</version>
        <scope>test</scope>
    </dependency>
    
    <dependency>
        <groupId>org.mockito</groupId>
        <artifactId>mockito-core</artifactId>
        <version>5.1.1</version>
        <scope>test</scope>
    </dependency>
    
    <dependency>
        <groupId>org.mockito</groupId>
        <artifactId>mockito-junit-jupiter</artifactId>
        <version>5.1.1</version>
        <scope>test</scope>
    </dependency>
    
    <dependency>
        <groupId>org.assertj</groupId>
        <artifactId>assertj-core</artifactId>
        <version>3.24.2</version>
        <scope>test</scope>
    </dependency>
    
    <dependency>
        <groupId>org.testcontainers</groupId>
        <artifactId>junit-jupiter</artifactId>
        <version>1.17.6</version>
        <scope>test</scope>
    </dependency>
    
    <dependency>
        <groupId>org.awaitility</groupId>
        <artifactId>awaitility</artifactId>
        <version>4.2.0</version>
        <scope>test</scope>
    </dependency>
</dependencies>
```

## Test Naming Conventions

### Test Classes
- Unit tests: `ClassNameTest.java`
- Integration tests: `ClassNameIntegrationTest.java`
- End-to-end tests: `ClassNameE2ETest.java`

### Test Methods
Use descriptive names following the pattern: `should_ExpectedBehavior_When_StateUnderTest`
```java
public class PackageInstallerTest {
    
    @Test
    void shouldInstallPackageSuccessfully_WhenValidPackageNameProvided() { }
    
    @Test
    void shouldThrowException_WhenPackageNameIsNull() { }
    
    @Test
    void shouldThrowException_WhenPackageNameIsEmpty() { }
    
    @Test
    void shouldReturnTrue_WhenPackageIsAlreadyInstalled() { }
}
```

### Test Structure Pattern
Use **Given/When/Then** or **Arrange/Act/Assert** pattern:
```java
@Test
void shouldInstallPackageSuccessfully_WhenValidPackageNameProvided() {
    // Given (Arrange)
    String packageName = \"python3\";
    PackageManager mockManager = mock(PackageManager.class);
    when(mockManager.install(packageName)).thenReturn(true);
    PackageInstaller installer = new PackageInstaller(mockManager);
    
    // When (Act)
    boolean result = installer.installPackage(packageName);
    
    // Then (Assert)
    assertThat(result).isTrue();
    verify(mockManager).install(packageName);
}
```

## Unit Testing

### Basic Unit Test Example
```java
package com.cassandragargoyle.project.service;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.assertj.core.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class PackageInstallerTest {
    
    @Mock
    private PackageManager packageManager;
    
    @Mock
    private SystemDetector systemDetector;
    
    private PackageInstaller packageInstaller;
    
    @BeforeEach
    void setUp() {
        packageInstaller = new PackageInstaller(packageManager, systemDetector);
    }
    
    @Test
    void shouldInstallPackageSuccessfully_WhenValidPackageNameProvided() {
        // Given
        String packageName = \"python3\";
        when(systemDetector.detectOS()).thenReturn(\"linux\");
        when(packageManager.install(packageName)).thenReturn(true);
        
        // When
        boolean result = packageInstaller.installPackage(packageName);
        
        // Then
        assertThat(result).isTrue();
        verify(packageManager).install(packageName);
        verify(systemDetector).detectOS();
    }
    
    @Test
    void shouldThrowIllegalArgumentException_WhenPackageNameIsNull() {
        // Given
        String packageName = null;
        
        // When & Then
        assertThatThrownBy(() -> packageInstaller.installPackage(packageName))
            .isInstanceOf(IllegalArgumentException.class)
            .hasMessage(\"Package name cannot be null or empty\");
    }
}
```

### Parameterized Tests
Use JUnit 5 parameterized tests for multiple scenarios:
```java
@ParameterizedTest
@ValueSource(strings = {\"\", \"   \", \"package;rm -rf /\", \"very-long-package-name-that-exceeds-limit\"})
void shouldThrowException_WhenPackageNameIsInvalid(String invalidPackageName) {
    // When & Then
    assertThatThrownBy(() -> packageInstaller.installPackage(invalidPackageName))
        .isInstanceOf(IllegalArgumentException.class);
}

@ParameterizedTest
@CsvSource({
    \"python3, linux, apt\",
    \"python3, windows, chocolatey\",
    \"python3, macos, homebrew\"
})
void shouldSelectCorrectPackageManager_BasedOnOperatingSystem(
        String packageName, String os, String expectedManager) {
    // Given
    when(systemDetector.detectOS()).thenReturn(os);
    when(packageManager.getManagerName()).thenReturn(expectedManager);
    
    // When
    String actualManager = packageInstaller.getPackageManagerForOS(os);
    
    // Then
    assertThat(actualManager).isEqualTo(expectedManager);
}

@ParameterizedTest
@MethodSource(\"providePackageTestData\")
void shouldHandleVariousPackageScenarios(PackageTestData testData) {
    // Given
    when(systemDetector.detectOS()).thenReturn(testData.os);
    when(packageManager.install(testData.packageName)).thenReturn(testData.expectedResult);
    
    // When
    boolean result = packageInstaller.installPackage(testData.packageName);
    
    // Then
    assertThat(result).isEqualTo(testData.expectedResult);
}

private static Stream<PackageTestData> providePackageTestData() {
    return Stream.of(
        new PackageTestData(\"python3\", \"linux\", true),
        new PackageTestData(\"java\", \"windows\", true),
        new PackageTestData(\"nonexistent\", \"linux\", false)
    );
}

static class PackageTestData {
    final String packageName;
    final String os;
    final boolean expectedResult;
    
    PackageTestData(String packageName, String os, boolean expectedResult) {
        this.packageName = packageName;
        this.os = os;
        this.expectedResult = expectedResult;
    }
}
```

### Testing Exceptions
```java
@Test
void shouldThrowConfigurationException_WhenConfigFileNotFound() {
    // Given
    String nonExistentPath = \"/path/that/does/not/exist.xml\";
    ConfigurationLoader loader = new ConfigurationLoader();
    
    // When & Then
    assertThatThrownBy(() -> loader.loadConfiguration(nonExistentPath))
        .isInstanceOf(ConfigurationException.class)
        .hasMessage(\"Configuration file not found: \" + nonExistentPath)
        .hasCauseInstanceOf(FileNotFoundException.class);
}

@Test
void shouldHandleInterruptedException_WhenInstallationIsInterrupted() {
    // Given
    String packageName = \"slow-package\";
    when(packageManager.install(packageName)).thenThrow(new InterruptedException());
    
    // When
    assertThatThrownBy(() -> packageInstaller.installPackage(packageName))
        .isInstanceOf(PackageInstallationException.class)
        .hasMessage(\"Installation was interrupted\");
    
    // Then
    assertThat(Thread.currentThread().isInterrupted()).isTrue();
}
```

## Integration Testing

### Integration Test Structure
```java
package com.cassandragargoyle.project.service;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.junit.jupiter.SpringExtension;
import org.testcontainers.junit.jupiter.Testcontainers;

@ExtendWith(SpringExtension.class)
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
@Testcontainers
class PackageInstallerIntegrationTest {
    
    private PackageInstaller packageInstaller;
    
    @BeforeEach
    void setUp() {
        // Real configuration, not mocked
        ConfigurationManager configManager = new ConfigurationManager();
        SystemDetector systemDetector = new SystemDetector();
        PackageManager packageManager = PackageManagerFactory.create(systemDetector.detectOS());
        
        packageInstaller = new PackageInstaller(packageManager, systemDetector);
    }
    
    @Test
    void shouldInstallAndVerifyPackage_InRealEnvironment() {
        // This test might be skipped in CI or require special setup
        assumeTrue(isIntegrationTestEnvironment());
        
        // Given
        String packageName = \"curl\"; // Safe package to install
        
        // When
        boolean installResult = packageInstaller.installPackage(packageName);
        boolean isInstalled = packageInstaller.isPackageInstalled(packageName);
        
        // Then
        assertThat(installResult).isTrue();
        assertThat(isInstalled).isTrue();
    }
    
    private boolean isIntegrationTestEnvironment() {
        return Boolean.parseBoolean(System.getProperty(\"integration.tests.enabled\", \"false\"));
    }
}
```

### TestContainers Integration
```java
@Testcontainers
class DockerBasedIntegrationTest {
    
    @Container
    static GenericContainer<?> ubuntu = new GenericContainer<>(\"ubuntu:22.04\")
            .withCommand(\"tail\", \"-f\", \"/dev/null\")
            .withWorkingDirectory(\"/app\");
    
    @Test
    void shouldInstallPackageInDockerContainer() throws Exception {
        // Given
        String packageName = \"python3\";
        
        // When - Execute installation command in container
        Container.ExecResult result = ubuntu.execInContainer(
            \"bash\", \"-c\", \"apt-get update && apt-get install -y \" + packageName
        );
        
        // Then
        assertThat(result.getExitCode()).isEqualTo(0);
        
        // Verify installation
        Container.ExecResult verifyResult = ubuntu.execInContainer(
            \"bash\", \"-c\", \"which python3\"
        );
        assertThat(verifyResult.getExitCode()).isEqualTo(0);
    }
}
```

## Mocking with Mockito

### Basic Mocking
```java
@ExtendWith(MockitoExtension.class)
class ServiceTest {
    
    @Mock
    private ExternalService externalService;
    
    @InjectMocks
    private BusinessService businessService;
    
    @Test
    void shouldProcessData_WhenExternalServiceReturnsData() {
        // Given
        String inputData = \"test-data\";
        String externalResult = \"processed-data\";
        when(externalService.process(inputData)).thenReturn(externalResult);
        
        // When
        String result = businessService.handleData(inputData);
        
        // Then
        assertThat(result).contains(\"processed-data\");
        verify(externalService).process(inputData);
    }
}
```

### Advanced Mocking Patterns
```java
@Test
void shouldRetryOnFailure_AndEventuallySucceed() {
    // Given
    String packageName = \"unstable-package\";
    when(packageManager.install(packageName))
        .thenThrow(new RuntimeException(\"First attempt fails\"))
        .thenThrow(new RuntimeException(\"Second attempt fails\"))
        .thenReturn(true); // Third attempt succeeds
    
    // When
    boolean result = packageInstaller.installPackageWithRetry(packageName);
    
    // Then
    assertThat(result).isTrue();
    verify(packageManager, times(3)).install(packageName);
}

@Test
void shouldCaptureArgumentsPassedToMock() {
    // Given
    ArgumentCaptor<InstallationRequest> requestCaptor = ArgumentCaptor.forClass(InstallationRequest.class);
    
    // When
    packageInstaller.installPackage(\"python3\");
    
    // Then
    verify(packageManager).install(requestCaptor.capture());
    InstallationRequest capturedRequest = requestCaptor.getValue();
    assertThat(capturedRequest.getPackageName()).isEqualTo(\"python3\");
    assertThat(capturedRequest.getTimeout()).isPositive();
}
```

## Test Fixtures and Data

### Using Test Resources
```java
class ConfigurationLoaderTest {
    
    @Test
    void shouldLoadValidConfiguration_FromResourceFile() throws Exception {
        // Given
        String configPath = getClass().getResource(\"/test-data/valid-config.xml\").getPath();
        ConfigurationLoader loader = new ConfigurationLoader();
        
        // When
        Configuration config = loader.loadConfiguration(configPath);
        
        // Then
        assertThat(config).isNotNull();
        assertThat(config.getTimeout()).isEqualTo(300);
        assertThat(config.getRetryCount()).isEqualTo(3);
    }
    
    @Test
    void shouldHandleInvalidConfiguration_Gracefully() {
        // Given
        String invalidConfigPath = getClass().getResource(\"/test-data/invalid-config.xml\").getPath();
        ConfigurationLoader loader = new ConfigurationLoader();
        
        // When & Then
        assertThatThrownBy(() -> loader.loadConfiguration(invalidConfigPath))
            .isInstanceOf(ConfigurationException.class)
            .hasMessageContaining(\"Invalid XML format\");
    }
}
```

### Temporary Files and Directories
```java
class FileProcessorTest {
    
    @Test
    void shouldProcessTemporaryFile(@TempDir Path tempDir) throws IOException {
        // Given
        Path testFile = tempDir.resolve(\"test-file.txt\");
        Files.writeString(testFile, \"test content\");
        FileProcessor processor = new FileProcessor();
        
        // When
        ProcessingResult result = processor.processFile(testFile.toString());
        
        // Then
        assertThat(result.isSuccess()).isTrue();
        assertThat(result.getProcessedLines()).isEqualTo(1);
    }
}
```

## Asynchronous Testing

### Testing CompletableFuture
```java
@Test
void shouldHandleAsynchronousInstallation() throws Exception {
    // Given
    String packageName = \"async-package\";
    when(packageManager.installAsync(packageName))
        .thenReturn(CompletableFuture.completedFuture(true));
    
    // When
    CompletableFuture<Boolean> future = packageInstaller.installPackageAsync(packageName);
    
    // Then
    assertThat(future.get(5, TimeUnit.SECONDS)).isTrue();
    verify(packageManager).installAsync(packageName);
}

@Test
void shouldTimeoutAsynchronousOperation() {
    // Given
    String packageName = \"slow-package\";
    CompletableFuture<Boolean> neverCompletingFuture = new CompletableFuture<>();
    when(packageManager.installAsync(packageName)).thenReturn(neverCompletingFuture);
    
    // When & Then
    assertThatThrownBy(() -> 
        packageInstaller.installPackageAsync(packageName).get(1, TimeUnit.SECONDS))
        .isInstanceOf(TimeoutException.class);
}
```

### Testing with Awaitility
```java
@Test
void shouldEventuallyCompleteInstallation() {
    // Given
    String packageName = \"eventually-installed-package\";
    AtomicBoolean installationComplete = new AtomicBoolean(false);
    
    // Start async installation
    CompletableFuture.runAsync(() -> {
        try {
            Thread.sleep(2000); // Simulate slow installation
            installationComplete.set(true);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    });
    
    // Then
    await().atMost(5, TimeUnit.SECONDS)
           .untilAsserted(() -> assertThat(installationComplete).isTrue());
}
```

## Performance Testing

### Simple Performance Tests
```java
@Test
@Timeout(value = 5, unit = TimeUnit.SECONDS)
void shouldCompleteInstallationWithinTimeLimit() {
    // Given
    String packageName = \"quick-package\";
    when(packageManager.install(packageName)).thenReturn(true);
    
    // When
    boolean result = packageInstaller.installPackage(packageName);
    
    // Then
    assertThat(result).isTrue();
}

@RepeatedTest(10)
void shouldConsistentlyInstallPackage() {
    // This test runs 10 times to check for consistency
    String packageName = \"reliable-package\";
    when(packageManager.install(packageName)).thenReturn(true);
    
    boolean result = packageInstaller.installPackage(packageName);
    
    assertThat(result).isTrue();
}
```

### JMH Benchmarking
```java
@BenchmarkMode(Mode.AverageTime)
@OutputTimeUnit(TimeUnit.MILLISECONDS)
@State(Scope.Benchmark)
public class PackageInstallerBenchmark {
    
    private PackageInstaller installer;
    
    @Setup
    public void setup() {
        PackageManager mockManager = mock(PackageManager.class);
        when(mockManager.install(anyString())).thenReturn(true);
        installer = new PackageInstaller(mockManager);
    }
    
    @Benchmark
    public boolean benchmarkPackageInstallation() {
        return installer.installPackage(\"benchmark-package\");
    }
}
```

## Test Categories and Profiles

### Maven Profiles for Test Categories
```xml
<profiles>
    <profile>
        <id>unit-tests</id>
        <activation>
            <activeByDefault>true</activeByDefault>
        </activation>
        <build>
            <plugins>
                <plugin>
                    <groupId>org.apache.maven.plugins</groupId>
                    <artifactId>maven-surefire-plugin</artifactId>
                    <configuration>
                        <excludes>
                            <exclude>**/*IntegrationTest.java</exclude>
                            <exclude>**/*E2ETest.java</exclude>
                        </excludes>
                    </configuration>
                </plugin>
            </plugins>
        </build>
    </profile>
    
    <profile>
        <id>integration-tests</id>
        <build>
            <plugins>
                <plugin>
                    <groupId>org.apache.maven.plugins</groupId>
                    <artifactId>maven-failsafe-plugin</artifactId>
                    <configuration>
                        <includes>
                            <include>**/*IntegrationTest.java</include>
                        </includes>
                    </configuration>
                </plugin>
            </plugins>
        </build>
    </profile>
</profiles>
```

### Running Different Test Categories
```bash
# Unit tests only (default)
mvn test

# Integration tests
mvn test -P integration-tests

# All tests
mvn test -P unit-tests,integration-tests

# Specific test class
mvn test -Dtest=PackageInstallerTest

# Specific test method
mvn test -Dtest=PackageInstallerTest#shouldInstallPackageSuccessfully_WhenValidPackageNameProvided
```

## Coverage and Quality

### JaCoCo Coverage Configuration
```xml
<plugin>
    <groupId>org.jacoco</groupId>
    <artifactId>jacoco-maven-plugin</artifactId>
    <version>0.8.8</version>
    <executions>
        <execution>
            <goals>
                <goal>prepare-agent</goal>
            </goals>
        </execution>
        <execution>
            <id>report</id>
            <phase>test</phase>
            <goals>
                <goal>report</goal>
            </goals>
        </execution>
        <execution>
            <id>check</id>
            <goals>
                <goal>check</goal>
            </goals>
            <configuration>
                <rules>
                    <rule>
                        <element>CLASS</element>
                        <limits>
                            <limit>
                                <counter>LINE</counter>
                                <value>COVEREDRATIO</value>
                                <minimum>0.80</minimum>
                            </limit>
                        </limits>
                    </rule>
                </rules>
            </configuration>
        </execution>
    </executions>
</plugin>
```

### Coverage Requirements
- **Minimum coverage**: 80% line coverage
- **Critical classes**: 90% coverage required
- **Integration tests**: 60% additional coverage

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Java Test Suite

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-java@v3
        with:
          java-version: '17'
          distribution: 'temurin'
      
      - name: Cache Maven dependencies
        uses: actions/cache@v3
        with:
          path: ~/.m2
          key: ${{ runner.os }}-m2-${{ hashFiles('**/pom.xml') }}
      
      - name: Run unit tests
        run: mvn test
      
      - name: Generate test report
        run: mvn surefire-report:report
      
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3

  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-java@v3
        with:
          java-version: '17'
          distribution: 'temurin'
      
      - name: Run integration tests
        run: mvn test -P integration-tests
```

## Testing Scripts

### Main Test Script
```bash
#!/bin/bash
# scripts/test.sh

set -e

# Colors
RED='\\033[0;31m'
GREEN='\\033[0;32m'
YELLOW='\\033[1;33m'
NC='\\033[0m'

echo -e \"${GREEN}Running Java Test Suite${NC}\"

# Unit tests
echo -e \"${YELLOW}Running unit tests...${NC}\"
mvn clean test

# Integration tests
echo -e \"${YELLOW}Running integration tests...${NC}\"
mvn test -P integration-tests

# Coverage report
echo -e \"${YELLOW}Generating coverage report...${NC}\"
mvn jacoco:report

# Check coverage threshold
mvn jacoco:check

echo -e \"${GREEN}All tests passed!${NC}\"
```

### Coverage Script
```bash
#!/bin/bash
# scripts/coverage.sh

set -e

echo \"Generating detailed coverage report...\"

# Run tests with coverage
mvn clean test jacoco:report

# Generate HTML report
echo \"Coverage report generated at: target/site/jacoco/index.html\"

# Show coverage summary
mvn jacoco:report | grep -A 10 \"Coverage Summary\"

# Open report if possible
if command -v xdg-open > /dev/null; then
    xdg-open target/site/jacoco/index.html
elif command -v open > /dev/null; then
    open target/site/jacoco/index.html
fi
```

## Best Practices Summary

### Test Organization
1. Follow Maven standard directory structure
2. Use descriptive test method names
3. Group related tests in nested classes
4. Separate unit, integration, and E2E tests

### Test Quality
1. Each test should verify one specific behavior
2. Use meaningful assertions with custom messages
3. Mock external dependencies properly
4. Clean up resources in `@AfterEach` methods

### Performance and Reliability
1. Keep unit tests fast (< 100ms each)
2. Make tests deterministic and independent
3. Use appropriate timeouts for async operations
4. Handle test data cleanup properly

### Maintenance
1. Update tests when code changes
2. Remove obsolete tests regularly
3. Keep test dependencies up to date
4. Monitor test execution times

---

**Note**: These guidelines should be adapted based on specific project requirements and frameworks used. Regular review ensures tests remain effective and maintainable.

*Created: 2025-08-23*
*Last updated: 2025-08-23*