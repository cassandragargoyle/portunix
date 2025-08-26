# C++ Testing Guidelines

## Purpose
This document defines the testing strategy, structure, and best practices for C++ projects in the CassandraGargoyle ecosystem.

## Testing Philosophy

### Testing Pyramid
1. **Unit Tests** (70%) - Fast, isolated tests for individual classes/functions
2. **Integration Tests** (20%) - Test component interactions and system integration
3. **System Tests** (10%) - Full application workflow and performance tests

### Key Principles
- Follow Test-Driven Development (TDD) when applicable
- Tests should be fast, reliable, and deterministic
- Each test should verify one specific behavior
- Use RAII for resource management in tests
- Mock system dependencies and external services
- **Container Testing**: Use Podman for all container-based testing (preferred over Docker)

## Project Structure

### Directory Layout
```
project/
├── include/
│   └── cassandragargoyle/
│       └── project/
│           ├── config/
│           ├── install/
│           └── util/
├── src/
│   ├── config/
│   ├── install/
│   └── util/
├── test/
│   ├── unit/
│   │   ├── config/
│   │   │   ├── test_configuration_manager.cpp
│   │   │   └── test_database_config.cpp
│   │   ├── install/
│   │   │   ├── test_package_installer.cpp
│   │   │   └── test_system_detector.cpp
│   │   └── util/
│   │       └── test_file_utils.cpp
│   ├── integration/
│   │   ├── test_install_workflow.cpp
│   │   └── test_config_integration.cpp
│   ├── fixtures/
│   │   ├── sample_config.xml
│   │   ├── test_packages.json
│   │   └── mock_responses/
│   ├── mocks/
│   │   ├── mock_package_manager.hpp
│   │   └── mock_system_interface.hpp
│   └── utils/
│       ├── test_helpers.hpp
│       └── test_helpers.cpp
├── scripts/
│   ├── test.sh
│   ├── test-unit.sh
│   ├── test-integration.sh
│   └── coverage.sh
├── CMakeLists.txt
└── conanfile.txt
```

## Test Dependencies

### CMake Configuration
```cmake
# CMakeLists.txt
cmake_minimum_required(VERSION 3.20)
project(CassandraGargoyleProject)

# Set C++ standard
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Find packages
find_package(Catch2 3 REQUIRED)
find_package(trompeloeil REQUIRED)

# Main library
add_library(cassandragargoyle_project
    src/config/configuration_manager.cpp
    src/install/package_installer.cpp
    src/util/file_utils.cpp
)

target_include_directories(cassandragargoyle_project
    PUBLIC include
)

# Test configuration
if(BUILD_TESTING)
    enable_testing()
    
    # Unit tests
    add_executable(unit_tests
        test/unit/config/test_configuration_manager.cpp
        test/unit/install/test_package_installer.cpp
        test/unit/util/test_file_utils.cpp
        test/utils/test_helpers.cpp
    )
    
    target_link_libraries(unit_tests
        PRIVATE
            cassandragargoyle_project
            Catch2::Catch2WithMain
            trompeloeil::trompeloeil
    )
    
    # Integration tests
    add_executable(integration_tests
        test/integration/test_install_workflow.cpp
        test/integration/test_config_integration.cpp
        test/utils/test_helpers.cpp
    )
    
    target_link_libraries(integration_tests
        PRIVATE
            cassandragargoyle_project
            Catch2::Catch2WithMain
    )
    
    # Register tests with CTest
    include(CTest)
    include(Catch)
    catch_discover_tests(unit_tests)
    catch_discover_tests(integration_tests)
endif()
```

### Conan Dependencies
```ini
# conanfile.txt
[requires]
catch2/3.4.0
trompeloeil/46

[generators]
CMakeDeps
CMakeToolchain

[options]
catch2:with_main=True

[imports]
., *.dll -> ./bin # Copies all dll files from packages bin folder to my project bin folder
., *.dylib -> ./bin # Copies all dylib files from packages bin folder to my project bin folder
```

## Test Naming Conventions

### Test Files
- Unit tests: `test_class_name.cpp`
- Integration tests: `test_feature_integration.cpp`
- System tests: `test_system_workflow.cpp`

### Test Cases
Use descriptive names with underscores:
```cpp
TEST_CASE(\"PackageInstaller should install package successfully with valid input\", \"[unit][installer]\") {
    // Test implementation
}

TEST_CASE(\"PackageInstaller should throw exception when package name is empty\", \"[unit][installer][error]\") {
    // Test implementation
}
```

### Test Structure Pattern
Use **Given/When/Then** sections with comments:
```cpp
TEST_CASE(\"ConfigurationManager should load valid configuration file\", \"[unit][config]\")
{
    // Given
    const std::string config_path = \"test/fixtures/valid_config.xml\";
    ConfigurationManager manager;
    
    // When
    auto config = manager.load_configuration(config_path);
    
    // Then
    REQUIRE(config != nullptr);
    CHECK(config->get_timeout() == 300);
    CHECK(config->get_retry_count() == 3);
}
```

## Unit Testing with Catch2

### Basic Unit Test Example
```cpp
#include <catch2/catch_test_macros.hpp>
#include <catch2/matchers/catch_matchers_string.hpp>

#include \"cassandragargoyle/project/install/package_installer.hpp\"
#include \"test/mocks/mock_package_manager.hpp\"

using namespace cassandragargoyle::project::install;
using namespace Catch::Matchers;

TEST_CASE(\"PackageInstaller unit tests\", \"[unit][installer]\")
{
    auto mock_manager = std::make_unique<MockPackageManager>();
    PackageInstaller installer(std::move(mock_manager));
    
    SECTION(\"should install package successfully with valid input\")
    {
        // Given
        const std::string package_name = \"python3\";
        REQUIRE_CALL(*mock_manager, install(package_name))
            .RETURN(true);
        
        // When
        bool result = installer.install_package(package_name);
        
        // Then
        CHECK(result == true);
    }
    
    SECTION(\"should throw exception when package name is empty\")
    {
        // Given
        const std::string empty_name = \"\";
        
        // When & Then
        CHECK_THROWS_AS(installer.install_package(empty_name), std::invalid_argument);
        CHECK_THROWS_WITH(installer.install_package(empty_name), 
                         ContainsSubstring(\"Package name cannot be empty\"));
    }
    
    SECTION(\"should handle installation failure gracefully\")
    {
        // Given
        const std::string package_name = \"nonexistent-package\";
        REQUIRE_CALL(*mock_manager, install(package_name))
            .RETURN(false);
        
        // When
        bool result = installer.install_package(package_name);
        
        // Then
        CHECK(result == false);
    }
}
```

### Template and Parameterized Tests
```cpp
#include <catch2/catch_template_test_macros.hpp>
#include <catch2/generators/catch_generators.hpp>

// Template tests for different types
TEMPLATE_TEST_CASE(\"Vector operations\", \"[template]\", int, float, double)
{
    std::vector<TestType> vec{1, 2, 3};
    
    REQUIRE(vec.size() == 3);
    REQUIRE(vec[0] == TestType{1});
}

// Parameterized tests with generators
TEST_CASE(\"Package name validation\", \"[unit][validation]\")
{
    auto invalid_name = GENERATE(as<std::string>{}, \"\", \"   \", \"pkg;rm -rf /\", 
                                 std::string(101, 'a')); // Too long
    
    PackageInstaller installer;
    
    CHECK_THROWS_AS(installer.install_package(invalid_name), std::invalid_argument);
}

TEST_CASE(\"OS detection mapping\", \"[unit][system]\")
{
    auto [input_os, expected_manager] = GENERATE(table<std::string, std::string>({
        {\"linux\", \"apt\"},
        {\"windows\", \"chocolatey\"},
        {\"macos\", \"homebrew\"}
    }));
    
    SystemDetector detector;
    auto manager = detector.get_package_manager_for_os(input_os);
    
    CHECK(manager == expected_manager);
}
```

## Mocking with Trompeloeil

### Interface Definition
```cpp
// include/cassandragargoyle/project/install/package_manager.hpp
#pragma once

#include <string>

namespace cassandragargoyle::project::install {

class IPackageManager
{
public:
    virtual ~IPackageManager() = default;
    
    virtual bool install(const std::string& package_name) = 0;
    virtual bool remove(const std::string& package_name) = 0;
    virtual bool is_installed(const std::string& package_name) const = 0;
    virtual std::vector<std::string> list_installed() const = 0;
};

} // namespace
```

### Mock Implementation
```cpp
// test/mocks/mock_package_manager.hpp
#pragma once

#include <trompeloeil.hpp>
#include \"cassandragargoyle/project/install/package_manager.hpp\"

class MockPackageManager : public cassandragargoyle::project::install::IPackageManager
{
public:
    MAKE_MOCK1(install, bool(const std::string&), override);
    MAKE_MOCK1(remove, bool(const std::string&), override);
    MAKE_MOCK1(is_installed, bool(const std::string&), const override);
    MAKE_MOCK0(list_installed, std::vector<std::string>(), const override);
};
```

### Using Mocks in Tests
```cpp
#include <trompeloeil.hpp>
#include \"test/mocks/mock_package_manager.hpp\"

TEST_CASE(\"PackageInstaller with mocked dependencies\", \"[unit][installer]\")
{
    auto mock_manager = std::make_unique<MockPackageManager>();
    auto* mock_ptr = mock_manager.get(); // Keep reference for expectations
    
    PackageInstaller installer(std::move(mock_manager));
    
    SECTION(\"should retry installation on failure\")
    {
        // Given
        const std::string package_name = \"unstable-package\";
        
        // Setup expectations - fail twice, then succeed
        REQUIRE_CALL(*mock_ptr, install(package_name))
            .RETURN(false)
            .TIMES(2);
        REQUIRE_CALL(*mock_ptr, install(package_name))
            .RETURN(true)
            .TIMES(1);
        
        // When
        bool result = installer.install_package_with_retry(package_name, 3);
        
        // Then
        CHECK(result == true);
    }
    
    SECTION(\"should capture and validate call arguments\")
    {
        // Given
        std::string captured_package;
        REQUIRE_CALL(*mock_ptr, install(trompeloeil::_))
            .LR_SIDE_EFFECT(captured_package = _1)
            .RETURN(true);
        
        // When
        installer.install_package(\"test-package\");
        
        // Then
        CHECK(captured_package == \"test-package\");
    }
}
```

## Integration Testing

### Integration Test Example
```cpp
// test/integration/test_install_workflow.cpp
#include <catch2/catch_test_macros.hpp>
#include <filesystem>
#include <fstream>

#include \"cassandragargoyle/project/install/package_installer.hpp\"
#include \"cassandragargoyle/project/config/configuration_manager.hpp\"

TEST_CASE(\"Installation workflow integration\", \"[integration][workflow]\")
{
    // Skip if not in integration test environment
    if (!std::getenv(\"INTEGRATION_TESTS\")) {
        SKIP(\"Integration tests disabled\");
    }
    
    SECTION(\"should complete full installation workflow\")
    {
        // Given - Setup test environment
        std::filesystem::path temp_config = std::filesystem::temp_directory_path() / \"test_config.xml\";
        create_test_config_file(temp_config);
        
        ConfigurationManager config_manager;
        auto config = config_manager.load_configuration(temp_config.string());
        
        PackageInstaller installer(config);
        
        // When
        bool result = installer.install_package(\"curl\"); // Safe package to test
        
        // Then
        CHECK(result == true);
        CHECK(installer.is_package_installed(\"curl\") == true);
        
        // Cleanup
        std::filesystem::remove(temp_config);
    }
}

void create_test_config_file(const std::filesystem::path& path)
{
    std::ofstream file(path);
    file << R\"(<?xml version=\"1.0\"?>
<configuration>
    <timeout>300</timeout>
    <retry_count>3</retry_count>
    <package_manager>auto</package_manager>
</configuration>)\";
}
```

### Docker-based Integration Tests
```cpp
#include <catch2/catch_test_macros.hpp>
#include <cstdlib>
#include <memory>
#include <stdexcept>
#include <string>

class DockerContainer
{
public:
    DockerContainer(const std::string& image) : image_(image)
    {
        container_id_ = start_container();
    }
    
    ~DockerContainer()
    {
        if (!container_id_.empty()) {
            stop_container();
        }
    }
    
    std::string exec(const std::string& command) const
    {
        std::string full_command = \"docker exec \" + container_id_ + \" \" + command;
        return execute_command(full_command);
    }
    
private:
    std::string image_;
    std::string container_id_;
    
    std::string start_container()
    {
        std::string command = \"docker run -d \" + image_ + \" tail -f /dev/null\";
        return execute_command(command);
    }
    
    void stop_container()
    {
        execute_command(\"docker rm -f \" + container_id_);
    }
    
    std::string execute_command(const std::string& command) const
    {
        std::unique_ptr<FILE, decltype(&pclose)> pipe(popen(command.c_str(), \"r\"), pclose);
        if (!pipe) {
            throw std::runtime_error(\"popen() failed!\");
        }
        
        char buffer[128];
        std::string result;
        while (fgets(buffer, sizeof(buffer), pipe.get()) != nullptr) {
            result += buffer;
        }
        
        // Remove trailing newline
        if (!result.empty() && result.back() == '\\n') {
            result.pop_back();
        }
        
        return result;
    }
};

TEST_CASE(\"Docker-based integration tests\", \"[integration][docker]\")
{
    SECTION(\"should install package in Ubuntu container\")
    {
        // Given
        DockerContainer container(\"ubuntu:22.04\");
        
        // When - Update package list and install
        container.exec(\"apt-get update -qq\");
        auto install_result = container.exec(\"apt-get install -y python3 2>/dev/null; echo $?\");
        
        // Then
        CHECK(install_result == \"0\"); // Exit code 0 means success
        
        // Verify installation
        auto which_result = container.exec(\"which python3\");
        CHECK(!which_result.empty());
    }
}
```

## Memory Management and RAII Testing

### Resource Management Tests
```cpp
#include <catch2/catch_test_macros.hpp>
#include <memory>

class ResourceManager
{
public:
    ResourceManager(const std::string& resource_name) 
        : resource_name_(resource_name), is_acquired_(false)
    {
        acquire_resource();
    }
    
    ~ResourceManager()
    {
        if (is_acquired_) {
            release_resource();
        }
    }
    
    // Move constructor
    ResourceManager(ResourceManager&& other) noexcept
        : resource_name_(std::move(other.resource_name_))
        , is_acquired_(other.is_acquired_)
    {
        other.is_acquired_ = false;
    }
    
    // Move assignment
    ResourceManager& operator=(ResourceManager&& other) noexcept
    {
        if (this != &other) {
            if (is_acquired_) {
                release_resource();
            }
            resource_name_ = std::move(other.resource_name_);
            is_acquired_ = other.is_acquired_;
            other.is_acquired_ = false;
        }
        return *this;
    }
    
    // Delete copy operations
    ResourceManager(const ResourceManager&) = delete;
    ResourceManager& operator=(const ResourceManager&) = delete;
    
    bool is_acquired() const { return is_acquired_; }
    
private:
    std::string resource_name_;
    bool is_acquired_;
    
    void acquire_resource() { is_acquired_ = true; }
    void release_resource() { is_acquired_ = false; }
};

TEST_CASE(\"RAII resource management\", \"[unit][raii]\")
{
    SECTION(\"should automatically release resources on scope exit\")
    {
        bool resource_acquired = false;
        
        {
            ResourceManager manager(\"test-resource\");
            resource_acquired = manager.is_acquired();
            CHECK(resource_acquired == true);
        } // ResourceManager destructor called here
        
        // Resource should be released automatically
        CHECK(resource_acquired == true); // Still true because we captured the value
    }
    
    SECTION(\"should support move semantics\")
    {
        auto create_manager = []() {
            return ResourceManager(\"moveable-resource\");
        };
        
        ResourceManager manager = create_manager(); // Move constructor
        CHECK(manager.is_acquired() == true);
        
        ResourceManager another_manager = std::move(manager); // Move assignment
        CHECK(another_manager.is_acquired() == true);
        CHECK(manager.is_acquired() == false); // Moved-from object
    }
}
```

### Memory Leak Testing
```cpp
#include <catch2/catch_test_macros.hpp>
#include <memory>
#include <vector>

TEST_CASE(\"Memory management\", \"[unit][memory]\")
{
    SECTION(\"should not leak memory with smart pointers\")
    {
        std::vector<std::unique_ptr<PackageInstaller>> installers;
        
        // Create many installers
        for (int i = 0; i < 1000; ++i) {
            auto mock_manager = std::make_unique<MockPackageManager>();
            installers.push_back(std::make_unique<PackageInstaller>(std::move(mock_manager)));
        }
        
        // They should all be automatically cleaned up when vector is destroyed
        CHECK(installers.size() == 1000);
    }
    
    SECTION(\"should handle circular references with weak_ptr\")
    {
        struct Parent;
        struct Child {
            std::weak_ptr<Parent> parent;
        };
        
        struct Parent {
            std::vector<std::shared_ptr<Child>> children;
        };
        
        auto parent = std::make_shared<Parent>();
        auto child = std::make_shared<Child>();
        child->parent = parent;
        parent->children.push_back(child);
        
        CHECK(parent.use_count() == 1); // Not counting weak_ptr
        CHECK(child.use_count() == 2);  // Shared by parent and our variable
    }
}
```

## Performance and Benchmark Testing

### Basic Performance Tests
```cpp
#include <catch2/catch_test_macros.hpp>
#include <catch2/benchmark/catch_benchmark.hpp>
#include <chrono>

TEST_CASE(\"Performance tests\", \"[performance]\")
{
    PackageInstaller installer;
    
    SECTION(\"installation should complete within time limit\")
    {
        auto start = std::chrono::high_resolution_clock::now();
        
        // Simulate quick installation
        bool result = installer.install_package(\"quick-package\");
        
        auto end = std::chrono::high_resolution_clock::now();
        auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(end - start);
        
        CHECK(result == true);
        CHECK(duration.count() < 5000); // Should complete in less than 5 seconds
    }
    
    BENCHMARK(\"package installation benchmark\") {
        return installer.install_package(\"benchmark-package\");
    };
    
    BENCHMARK_ADVANCED(\"advanced installation benchmark\")(Catch::Benchmark::Chronometer meter) {
        // Setup
        std::vector<std::string> packages = {\"pkg1\", \"pkg2\", \"pkg3\"};
        
        // Measure only the actual work
        meter.measure([&] {
            for (const auto& package : packages) {
                installer.install_package(package);
            }
        });
    };
}
```

### Stress Testing
```cpp
TEST_CASE(\"Stress tests\", \"[stress]\")
{
    SECTION(\"should handle concurrent installations\")
    {
        const int num_threads = 10;
        const int installations_per_thread = 100;
        
        std::vector<std::thread> threads;
        std::atomic<int> successful_installations{0};
        
        for (int i = 0; i < num_threads; ++i) {
            threads.emplace_back([&, i]() {
                PackageInstaller installer;
                for (int j = 0; j < installations_per_thread; ++j) {
                    if (installer.install_package(\"thread-\" + std::to_string(i) + \"-pkg-\" + std::to_string(j))) {
                        successful_installations++;
                    }
                }
            });
        }
        
        for (auto& thread : threads) {
            thread.join();
        }
        
        CHECK(successful_installations.load() == num_threads * installations_per_thread);
    }
}
```

## Test Fixtures and Utilities

### Test Helper Classes
```cpp
// test/utils/test_helpers.hpp
#pragma once

#include <filesystem>
#include <string>
#include <fstream>

class TemporaryDirectory
{
public:
    TemporaryDirectory();
    ~TemporaryDirectory();
    
    // Non-copyable, moveable
    TemporaryDirectory(const TemporaryDirectory&) = delete;
    TemporaryDirectory& operator=(const TemporaryDirectory&) = delete;
    TemporaryDirectory(TemporaryDirectory&&) noexcept;
    TemporaryDirectory& operator=(TemporaryDirectory&&) noexcept;
    
    std::filesystem::path path() const { return path_; }
    std::filesystem::path create_file(const std::string& filename, const std::string& content = \"\");
    
private:
    std::filesystem::path path_;
};

class ConfigFileBuilder
{
public:
    ConfigFileBuilder& set_timeout(int timeout);
    ConfigFileBuilder& set_retry_count(int count);
    ConfigFileBuilder& set_package_manager(const std::string& manager);
    
    std::string build_xml() const;
    std::filesystem::path write_to_temp_file() const;
    
private:
    int timeout_ = 300;
    int retry_count_ = 3;
    std::string package_manager_ = \"auto\";
};
```

### Test Helper Implementation
```cpp
// test/utils/test_helpers.cpp
#include \"test_helpers.hpp\"
#include <random>
#include <sstream>

TemporaryDirectory::TemporaryDirectory()
{
    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_int_distribution<> dis(100000, 999999);
    
    auto temp_dir = std::filesystem::temp_directory_path();
    path_ = temp_dir / (\"test_\" + std::to_string(dis(gen)));
    
    std::filesystem::create_directories(path_);
}

TemporaryDirectory::~TemporaryDirectory()
{
    if (std::filesystem::exists(path_)) {
        std::filesystem::remove_all(path_);
    }
}

TemporaryDirectory::TemporaryDirectory(TemporaryDirectory&& other) noexcept
    : path_(std::move(other.path_))
{
    other.path_.clear();
}

TemporaryDirectory& TemporaryDirectory::operator=(TemporaryDirectory&& other) noexcept
{
    if (this != &other) {
        if (std::filesystem::exists(path_)) {
            std::filesystem::remove_all(path_);
        }
        path_ = std::move(other.path_);
        other.path_.clear();
    }
    return *this;
}

std::filesystem::path TemporaryDirectory::create_file(const std::string& filename, const std::string& content)
{
    auto file_path = path_ / filename;
    std::ofstream file(file_path);
    file << content;
    return file_path;
}

ConfigFileBuilder& ConfigFileBuilder::set_timeout(int timeout)
{
    timeout_ = timeout;
    return *this;
}

ConfigFileBuilder& ConfigFileBuilder::set_retry_count(int count)
{
    retry_count_ = count;
    return *this;
}

ConfigFileBuilder& ConfigFileBuilder::set_package_manager(const std::string& manager)
{
    package_manager_ = manager;
    return *this;
}

std::string ConfigFileBuilder::build_xml() const
{
    std::ostringstream oss;
    oss << \"<?xml version=\\\"1.0\\\"?>\\n\"
        << \"<configuration>\\n\"
        << \"    <timeout>\" << timeout_ << \"</timeout>\\n\"
        << \"    <retry_count>\" << retry_count_ << \"</retry_count>\\n\"
        << \"    <package_manager>\" << package_manager_ << \"</package_manager>\\n\"
        << \"</configuration>\\n\";
    return oss.str();
}

std::filesystem::path ConfigFileBuilder::write_to_temp_file() const
{
    TemporaryDirectory temp_dir;
    return temp_dir.create_file(\"config.xml\", build_xml());
}
```

### Using Test Helpers
```cpp
#include \"test/utils/test_helpers.hpp\"

TEST_CASE(\"Configuration loading with helpers\", \"[unit][config]\")
{
    SECTION(\"should load configuration with custom values\")
    {
        // Given
        auto config_file = ConfigFileBuilder()
            .set_timeout(600)
            .set_retry_count(5)
            .set_package_manager(\"apt\")
            .write_to_temp_file();
        
        ConfigurationManager manager;
        
        // When
        auto config = manager.load_configuration(config_file.string());
        
        // Then
        CHECK(config->get_timeout() == 600);
        CHECK(config->get_retry_count() == 5);
        CHECK(config->get_package_manager() == \"apt\");
    }
    
    SECTION(\"should work with temporary directories\")
    {
        // Given
        TemporaryDirectory temp_dir;
        auto test_file = temp_dir.create_file(\"test.txt\", \"test content\");
        
        // When
        std::ifstream file(test_file);
        std::string content((std::istreambuf_iterator<char>(file)),
                           std::istreambuf_iterator<char>());
        
        // Then
        CHECK(content == \"test content\");
        CHECK(std::filesystem::exists(test_file));
    }
    // temp_dir automatically cleaned up here
}
```

## Coverage and Quality

### Code Coverage with gcov/lcov
```cmake
# CMakeLists.txt - Coverage configuration
option(ENABLE_COVERAGE \"Enable coverage reporting\" OFF)

if(ENABLE_COVERAGE)
    if(CMAKE_CXX_COMPILER_ID STREQUAL \"GNU\" OR CMAKE_CXX_COMPILER_ID STREQUAL \"Clang\")
        set(CMAKE_CXX_FLAGS \"${CMAKE_CXX_FLAGS} --coverage -O0 -g\")
        set(CMAKE_EXE_LINKER_FLAGS \"${CMAKE_EXE_LINKER_FLAGS} --coverage\")
    endif()
endif()
```

### Coverage Script
```bash
#!/bin/bash
# scripts/coverage.sh

set -e

echo \"Generating C++ coverage report...\"

# Clean previous coverage data
find . -name \"*.gcda\" -delete
find . -name \"*.gcno\" -delete

# Build with coverage
mkdir -p build-coverage
cd build-coverage
cmake -DENABLE_COVERAGE=ON -DCMAKE_BUILD_TYPE=Debug ..
make -j$(nproc)

# Run tests
ctest --output-on-failure

# Generate coverage report
lcov --capture --directory . --output-file coverage.info
lcov --remove coverage.info '/usr/*' '*/test/*' '*/build/*' --output-file coverage_filtered.info
genhtml coverage_filtered.info --output-directory coverage_html

echo \"Coverage report generated in coverage_html/index.html\"

# Show summary
lcov --summary coverage_filtered.info

# Open report if possible
if command -v xdg-open > /dev/null; then
    xdg-open coverage_html/index.html
elif command -v open > /dev/null; then
    open coverage_html/index.html
fi
```

## CI/CD Integration

### GitHub Actions Example
```yaml
name: C++ Test Suite

on: [push, pull_request]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        compiler: [gcc, clang]
        exclude:
          - os: windows-latest
            compiler: clang
    
    runs-on: ${{ matrix.os }}
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Install dependencies (Ubuntu)
      if: matrix.os == 'ubuntu-latest'
      run: |
        sudo apt-get update
        sudo apt-get install -y build-essential cmake lcov
    
    - name: Install dependencies (Windows)
      if: matrix.os == 'windows-latest'
      run: |
        choco install cmake
    
    - name: Install dependencies (macOS)
      if: matrix.os == 'macos-latest'
      run: |
        brew install cmake lcov
    
    - name: Configure CMake
      run: cmake -B build -DCMAKE_BUILD_TYPE=Release
    
    - name: Build
      run: cmake --build build --config Release
    
    - name: Test
      run: |
        cd build
        ctest --output-on-failure
    
    - name: Coverage (Ubuntu only)
      if: matrix.os == 'ubuntu-latest' && matrix.compiler == 'gcc'
      run: |
        cmake -B build-coverage -DENABLE_COVERAGE=ON -DCMAKE_BUILD_TYPE=Debug
        cmake --build build-coverage
        cd build-coverage
        ctest
        lcov --capture --directory . --output-file coverage.info
        lcov --remove coverage.info '/usr/*' '*/test/*' --output-file coverage_filtered.info
    
    - name: Upload coverage
      if: matrix.os == 'ubuntu-latest' && matrix.compiler == 'gcc'
      uses: codecov/codecov-action@v3
      with:
        file: build-coverage/coverage_filtered.info
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

echo -e \"${GREEN}Running C++ Test Suite${NC}\"

# Create build directory
mkdir -p build
cd build

# Configure
echo -e \"${YELLOW}Configuring project...${NC}\"
cmake -DCMAKE_BUILD_TYPE=Debug -DBUILD_TESTING=ON ..

# Build
echo -e \"${YELLOW}Building project...${NC}\"
make -j$(nproc)

# Run tests
echo -e \"${YELLOW}Running tests...${NC}\"
ctest --output-on-failure

echo -e \"${GREEN}All tests passed!${NC}\"
```

## Best Practices Summary

### Test Organization
1. Use clear directory structure separating unit and integration tests
2. Follow consistent naming conventions for test files and functions
3. Use descriptive test names that explain the scenario
4. Group related tests using Catch2 sections

### Test Quality
1. Each test should verify one specific behavior
2. Use RAII for resource management in tests
3. Mock external dependencies appropriately
4. Write tests that are fast, reliable, and independent

### Modern C++ Practices
1. Use smart pointers for memory management
2. Leverage RAII for resource cleanup
3. Use move semantics where appropriate
4. Follow const-correctness principles

### Performance and Maintainability
1. Keep unit tests fast (< 100ms each)
2. Use appropriate build configurations for testing
3. Monitor test execution times and resource usage
4. Keep test dependencies minimal and up-to-date

---

**Note**: These guidelines should be adapted based on specific project requirements and C++ standard used. Regular review ensures tests remain effective and maintainable.

*Created: 2025-08-23*
*Last updated: 2025-08-23*