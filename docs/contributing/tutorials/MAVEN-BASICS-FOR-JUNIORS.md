# Maven - Basics for Junior Developers

## NetBeans and Maven Integration

**Good news!** You don't need to use command line for Maven in NetBeans. Everything is integrated with buttons and menus.

### Working with Maven Projects in NetBeans

**Opening Maven Project:**
1. File → Open Project
2. Navigate to folder with `pom.xml`
   
   - **CassandraGargoyle team standard**: `D:\Dev\CassandraGargoyle\portunix\<repository-name>`
   - Example: `D:\Dev\CassandraGargoyle\portunix\portunix-plugins`
3. NetBeans shows Maven icon on project folder
4. Click Open Project


## What is Maven?
Maven is a build automation tool for managing Java projects. Think of it as your project's assistant that handles all the tedious tasks you'd otherwise do manually.

### What Maven solves

**Without Maven (The Hard Way):**
- Manually download each JAR library from websites
- Figure out which version works with other libraries
- Copy JARs to correct folders
- Write complex compile commands
- Manually track what libraries each developer needs
- Different setup on each developer's machine

**With Maven (The Easy Way):**
- Declare what you need in pom.xml
- Maven downloads everything automatically
- Handles version compatibility
- Same project setup for entire team
- One command builds everything
- Dependencies of dependencies handled automatically

### What Maven Does For You:
- **Dependency Management**: Downloads libraries and their dependencies from internet
- **Standard Project Structure**: Everyone knows where to find source code, tests, resources
- **Build Automation**: Compile, test, package with simple commands
- **Version Control**: Manages library versions and compatibility
- **Team Consistency**: Same build process on every machine
- **IDE Integration**: NetBeans understands Maven projects automatically

## Basic Project Structure (NetBeans Standard)

**CassandraGargoyle Team Convention:**
- **Windows**: `D:\Dev\CassandraGargoyle\portunix\<repository-name>\`
- **Linux**: `~/DEV/CassandraGargoyle/portunix/<repository-name>/`

```
D:\Dev\CassandraGargoyle\portunix\<repository-name>\   # Windows
~/DEV/CassandraGargoyle/portunix/<repository-name>/    # Linux
├── pom.xml                 # Configuration file (heart of the project)
├── src/
│   ├── main/
│   │   ├── java/           # Source code (.java files)
│   │   │   └── org/cassandragargoyle/portunix/  # Package structure
│   │   └── resources/      # Configuration files, properties, etc.
│   └── test/
│       ├── java/           # Test classes
│       └── resources/      # Test resources
├── target/                 # Compiled code (auto-generated)
└── nb-configuration.xml    # NetBeans specific settings (optional)
```

### NetBeans Maven Project Structure Requirements
- **src/main/java**: All Java source files (NetBeans expects this exact path)
- **src/main/resources**: Properties, XML configs, images (included in JAR)
- **src/test/java**: JUnit test classes (same package structure as main)
- **src/test/resources**: Test data files
- **target/**: NetBeans generates all build output here (don't commit to Git)

### Why Separate Test Directory?

**Tests MUST be in src/test/ because:**
1. **Not shipped to users** - Test code never goes into production JAR
2. **Saves space** - Production JAR stays small (no test frameworks)
3. **Security** - Test code might contain sensitive test data
4. **Clean separation** - Clear boundary between app and tests

**Example - What happens with scope:test:**
```xml
<!-- This 5MB library is ONLY for testing -->
<dependency>
    <groupId>org.junit.jupiter</groupId>
    <artifactId>junit-jupiter</artifactId>
    <version>5.10.1</version>
    <scope>test</scope>  <!-- NOT in final JAR! -->
</dependency>
```

**Result:**
- `mvn package` creates JAR WITHOUT JUnit (production-ready)
- `mvn test` uses JUnit (only during testing)
- Your JAR: 500KB instead of 5.5MB
- Users don't get unnecessary test libraries

## Essential Commands
```bash
mvn clean          # Delete target/ folder (uses maven-clean-plugin)
mvn compile        # Compile code (uses maven-compiler-plugin)
mvn test           # Run tests (uses maven-surefire-plugin)
mvn package        # Create JAR file (uses maven-jar-plugin)
mvn install        # Install JAR to local repository (~/.m2)
mvn clean package  # Clean and rebuild JAR

# These commands work even without any <build> section in pom.xml!
# Maven uses default plugins automatically
```

## pom.xml - Full Example

**⚠️ IMPORTANT for CassandraGargoyle Team:**

- **Never modify pom.xml without approval** - All dependency changes must be reviewed
- **Check with team lead** before adding new libraries
- **Document why** you need a new dependency in commit message
- **Security review required** for external dependencies
- **Version changes** must be tested thoroughly

```xml
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 
         http://maven.apache.org/xsd/maven-4.0.0.xsd">
    
    <modelVersion>4.0.0</modelVersion>
    
    <groupId>org.cassandragargoyle.portunix</groupId>
    <artifactId>my-project</artifactId>
    <version>1.0.0</version>
    <packaging>jar</packaging>  <!-- Creates JAR file -->
    
    <name>My Project</name>
    <description>Example project for learning Maven</description>
    
    <properties>
        <!-- Java version -->
        <maven.compiler.source>21</maven.compiler.source>
        <maven.compiler.target>21</maven.compiler.target>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
        
        <!-- Dependency versions - manage all versions in one place -->
        <spring.version>5.3.20</spring.version>
        <junit.version>5.10.1</junit.version>
        <slf4j.version>2.0.9</slf4j.version>
        
        <!-- Custom properties for your project -->
        <app.name>${project.artifactId}</app.name>
        <app.version>${project.version}</app.version>
    </properties>
    
    <dependencies>
        <!-- Spring Framework - using ${spring.version} property -->
        <dependency>
            <groupId>org.springframework</groupId>
            <artifactId>spring-core</artifactId>
            <version>${spring.version}</version>  <!-- Uses property defined above -->
        </dependency>
        
        <!-- Logging - using ${slf4j.version} property -->
        <dependency>
            <groupId>org.slf4j</groupId>
            <artifactId>slf4j-api</artifactId>
            <version>${slf4j.version}</version>  <!-- Uses property defined above -->
        </dependency>
        
        <!-- Testing (only for test phase) - using ${junit.version} property -->
        <dependency>
            <groupId>org.junit.jupiter</groupId>
            <artifactId>junit-jupiter-api</artifactId>
            <version>${junit.version}</version>  <!-- Uses property defined above -->
            <scope>test</scope>
        </dependency>
    </dependencies>
    
    <build>
        <plugins>
            <!-- Compiler plugin for Java version -->
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-compiler-plugin</artifactId>
                <version>3.12.1</version>
                <configuration>
                    <release>21</release>
                </configuration>
            </plugin>
            
            <!-- Surefire plugin for running tests -->
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-surefire-plugin</artifactId>
                <version>3.2.3</version>
            </plugin>
        </plugins>
    </build>
</project>
```

## Practical Tips
- **Dependency Hell**: Maven automatically resolves version conflicts
- **Repository**: Libraries are downloaded from Maven Central (central repository)
- **Lifecycle**: clean → compile → test → package → install → deploy
- **IDE Integration**: NetBeans, IntelliJ IDEA and Eclipse have built-in Maven support

## Key pom.xml Sections Explained

### Properties
- Define variables for reuse (versions, encoding, Java version)
- Access with `${property.name}` syntax
- Keeps versions consistent across dependencies
- Can reference other properties: `${project.version}`, `${project.artifactId}`

**Perfect for Multiple Artifacts from Same Vendor:**
```xml
<properties>
    <!-- One version for all Spring modules -->
    <spring.version>5.3.20</spring.version>
</properties>

<dependencies>
    <!-- All Spring modules use same version -->
    <dependency>
        <groupId>org.springframework</groupId>
        <artifactId>spring-core</artifactId>
        <version>${spring.version}</version>
    </dependency>
    <dependency>
        <groupId>org.springframework</groupId>
        <artifactId>spring-context</artifactId>
        <version>${spring.version}</version>
    </dependency>
    <dependency>
        <groupId>org.springframework</groupId>
        <artifactId>spring-web</artifactId>
        <version>${spring.version}</version>
    </dependency>
    <!-- Update spring.version once = all 3 dependencies updated! -->
</dependencies>
```

**Benefits:**

- Change version in one place, updates everywhere
- No typos from copying version numbers
- Easy to see all versions at a glance
- **Prevents version mismatches** between related libraries
- **Essential for frameworks** like Spring, Apache Commons, Jackson

### Dependencies
- **groupId**: Organization/company identifier
- **artifactId**: Project/library name
- **version**: Specific version or version range
- **scope**: When dependency is needed (compile, test, runtime, provided)

### Build Plugins

**Important:** Maven has default plugins that work automatically! 
You don't need to add them to pom.xml unless you want to customize their behavior.

**Default Plugins (work without configuration):**
- **maven-compiler-plugin**: Compiles Java code (uses Java 8 by default)
- **maven-surefire-plugin**: Runs tests automatically
- **maven-jar-plugin**: Creates JAR file
- **maven-clean-plugin**: Cleans target directory
- **maven-resources-plugin**: Copies resources to JAR

**When to Add Plugin Configuration:**
```xml
<!-- Only add this if you need Java 21 (not default Java 8) -->
<plugin>
    <groupId>org.apache.maven.plugins</groupId>
    <artifactId>maven-compiler-plugin</artifactId>
    <version>3.12.1</version>
    <configuration>
        <release>21</release>  <!-- Custom Java version -->
    </configuration>
</plugin>
```

**Common Custom Plugins (must be added manually):**

- **maven-assembly-plugin**: Creates executable JARs with all dependencies
- **maven-shade-plugin**: Alternative for uber-JARs
- **maven-javadoc-plugin**: Generates documentation

## Most Common pom.xml Mistakes (and How to Fix Them)

### 1. Wrong Property Syntax
```xml
<!-- WRONG - missing ${} -->
<version>spring.version</version>

<!-- CORRECT -->
<version>${spring.version}</version>
```

### 2. Property Not Defined
```xml
<!-- WRONG - using undefined property -->
<version>${undefined.version}</version>  <!-- Will literally use "${undefined.version}" as version -->

<!-- CORRECT - define property first -->
<properties>
    <my.version>1.0.0</my.version>
</properties>
<version>${my.version}</version>
```

### 3. Wrong XML Structure
```xml
<!-- WRONG - dependencies outside of <dependencies> tag -->
<project>
    <dependency>...</dependency>  <!-- ERROR! -->
</project>

<!-- CORRECT -->
<project>
    <dependencies>
        <dependency>...</dependency>
    </dependencies>
</project>
```

### 4. Missing Scope for Test Dependencies
```xml
<!-- WRONG - JUnit in production code -->
<dependency>
    <groupId>org.junit.jupiter</groupId>
    <artifactId>junit-jupiter-api</artifactId>
    <version>5.10.1</version>
    <!-- Missing scope! -->
</dependency>

<!-- CORRECT - only for tests -->
<dependency>
    <groupId>org.junit.jupiter</groupId>
    <artifactId>junit-jupiter-api</artifactId>
    <version>5.10.1</version>
    <scope>test</scope>  <!-- Important! -->
</dependency>
```

### 5. Version Conflicts
```xml
<!-- WRONG - different versions of same library -->
<dependency>
    <groupId>org.springframework</groupId>
    <artifactId>spring-core</artifactId>
    <version>5.3.20</version>
</dependency>
<dependency>
    <groupId>org.springframework</groupId>
    <artifactId>spring-context</artifactId>
    <version>5.2.15</version>  <!-- Different version! -->
</dependency>

<!-- CORRECT - use property for consistent versions -->
<properties>
    <spring.version>5.3.20</spring.version>
</properties>
<dependencies>
    <dependency>
        <groupId>org.springframework</groupId>
        <artifactId>spring-core</artifactId>
        <version>${spring.version}</version>
    </dependency>
    <dependency>
        <groupId>org.springframework</groupId>
        <artifactId>spring-context</artifactId>
        <version>${spring.version}</version>  <!-- Same version -->
    </dependency>
</dependencies>
```

## Common Issues and Solutions
- **Issue**: "Cannot resolve dependency" → Check internet connection and verify groupId/artifactId
- **Issue**: Slow downloads → Maven caches libraries in `~/.m2/repository`
- **Issue**: Version conflicts → Use `mvn dependency:tree` to display dependency tree
- **Issue**: "Plugin not found" → Update Maven version or add plugin repository
- **Issue**: Tests failing during build → Skip tests temporarily with `mvn package -DskipTests`
- **Issue**: "Invalid POM" → Check XML syntax, ensure all tags are properly closed
- **Issue**: Properties not working → Ensure property is defined before use

## JAR Files - What You Need to Know

### Types of JAR Files
1. **Standard JAR**: Contains only your compiled code (target/my-project-1.0.0.jar)
2. **Executable JAR with dependencies**: All-in-one JAR file (target/my-project-1.0.0-jar-with-dependencies.jar)

### Running JAR Files
```bash
# Standard JAR (requires classpath)
java -cp "lib/*:my-project.jar" org.cassandragargoyle.portunix.Main

# Executable JAR with dependencies
java -jar my-project-jar-with-dependencies.jar
```

### NetBeans Development Tasks

**Daily Development (No Command Line Needed!):**

| Task | NetBeans Action | Keyboard Shortcut | Maven Command Behind Scenes |
|------|-----------------|-------------------|------------------------------|
| **Build** | Right-click project → Build | F11 | `mvn compile` |
| **Run** | Right-click project → Run | F6 | `mvn exec:java` |
| **Debug** | Right-click project → Debug | Ctrl+F5 | Starts debugger |
| **Test** | Right-click project → Test | Alt+F6 | `mvn test` |
| **Clean & Build** | Right-click project → Clean and Build | Shift+F11 | `mvn clean install` |
| **Add Library** | Right-click Dependencies → Add Dependency ⚠️ | - | Updates pom.xml (requires approval!) |

**Additional NetBeans Features:**

- **Import Project**: File → Open Project → Select folder with pom.xml
- **Update Dependencies**: Right-click project → Reload POM
- **View Dependency Tree**: Right-click Dependencies → Show Dependency Graph
- **Project Properties**: Right-click → Properties → Sources (verify Java version)

**Debugging in NetBeans:**

1. Set breakpoint: Click left margin next to line number (red square appears)
2. Debug project: Right-click → Debug or press Ctrl+F5
3. Step through code:
   - F7: Step into method
   - F8: Step over line
   - Ctrl+F7: Step out of method
   - F5: Continue execution

**Visual Indicators:**

- **Red underlines**: Compilation errors (hover for details)
- **Yellow bulb**: Quick fixes available (click to fix)
- **Dependencies node**: Shows all libraries (double-click to see JAR contents)
- **Test Results window**: Shows passed/failed tests with details
- **Output window**: Shows Maven build output

### Creating Executable JAR
Add maven-assembly-plugin to create a JAR with all dependencies:
```xml
<plugin>
    <groupId>org.apache.maven.plugins</groupId>
    <artifactId>maven-assembly-plugin</artifactId>
    <version>3.6.0</version>
    <configuration>
        <descriptorRefs>
            <descriptorRef>jar-with-dependencies</descriptorRef>
        </descriptorRefs>
        <archive>
            <manifest>
                <mainClass>org.cassandragargoyle.portunix.Main</mainClass>
            </manifest>
        </archive>
    </configuration>
    <executions>
        <execution>
            <id>make-assembly</id>
            <phase>package</phase>
            <goals>
                <goal>single</goal>
            </goals>
        </execution>
    </executions>
</plugin>
```

## Team Rules for pom.xml Management

### CassandraGargoyle Development Process
1. **Before Adding Dependencies:**
   - Check if similar library already exists in project
   - Research security vulnerabilities (CVE database)
   - Verify license compatibility (Apache, MIT preferred)
   - Get approval from team lead

2. **Commit Rules:**
   ```
   git commit -m "deps: add Apache Commons IO 2.15.1 for file operations
   
   Reason: Needed for robust file copying in backup module
   Security: No known CVEs
   License: Apache 2.0
   Approved by: @zdenek"
   ```

3. **Review Process:**
   - All pom.xml changes go through code review
   - CI/CD pipeline checks for security issues
   - Dependency updates tested in staging first

## Advanced Tips for Growth
1. **Shade plugin**: Alternative to assembly plugin for creating uber-JARs
2. **Manifest customization**: Add metadata to JAR files
3. **Multi-module projects**: Parent POM can manage versions for child modules
4. **Repository management**: Deploy JARs to Nexus or Artifactory