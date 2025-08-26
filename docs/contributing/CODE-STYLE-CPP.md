# C++ Code Style Guidelines

## Purpose
This document defines the C++ coding standards for CassandraGargoyle projects, based on modern C++ best practices and industry standards.

## General Principles

### 1. Modern C++ Standards
- Use C++17 or later features when available
- Prefer standard library over custom implementations
- Follow RAII (Resource Acquisition Is Initialization) principle
- Embrace value semantics and move semantics

### 2. Code Quality
- Write self-documenting code
- Use const-correctness throughout
- Prefer stack allocation over heap allocation
- Use smart pointers instead of raw pointers

## File Organization

### File Naming
- Header files: `.hpp` extension (preferred) or `.h`
- Implementation files: `.cpp` extension
- Use snake_case for file names

```
✅ package_installer.hpp
✅ config_manager.cpp
✅ system_detector.hpp
❌ PackageInstaller.hpp
❌ configManager.cpp
```

### Directory Structure
```
project/
├── include/
│   └── cassandragargoyle/
│       └── project/
│           ├── config/
│           │   ├── configuration_manager.hpp
│           │   └── database_config.hpp
│           ├── install/
│           │   ├── package_installer.hpp
│           │   └── system_detector.hpp
│           └── util/
│               └── file_utils.hpp
├── src/
│   ├── config/
│   │   ├── configuration_manager.cpp
│   │   └── database_config.cpp
│   ├── install/
│   │   ├── package_installer.cpp
│   │   └── system_detector.cpp
│   └── util/
│       └── file_utils.cpp
├── test/
│   ├── config/
│   └── install/
└── CMakeLists.txt
```

### Header File Structure
```cpp
// package_installer.hpp
#pragma once

// System headers
#include <memory>
#include <string>
#include <vector>

// Third-party headers
#include <boost/filesystem.hpp>

// Project headers  
#include \"cassandragargoyle/project/config/configuration.hpp\"

namespace cassandragargoyle::project::install {

class PackageInstaller {
public:
    // Public interface
    
private:
    // Private implementation
};

} // namespace cassandragargoyle::project::install
```

## Naming Conventions

### Namespaces
Use lowercase with underscores, follow logical hierarchy:
```cpp
namespace cassandragargoyle::project::install {
    // implementation
}

namespace cassandragargoyle::project::config::detail {
    // implementation details
}
```

### Classes and Types
Use PascalCase for types:
```cpp
class PackageInstaller;
class ConfigurationManager;
struct InstallationResult;
enum class PackageStatus;
using PackageList = std::vector<std::string>;
```

### Functions and Variables
Use snake_case for functions and variables:
```cpp
class PackageInstaller {
public:
    void install_package(const std::string& package_name);
    bool is_package_installed(const std::string& package_name) const;
    
private:
    std::string config_path_;
    bool is_initialized_;
    std::unique_ptr<Configuration> config_;
};
```

### Constants and Enumerations
- Constants: Use UPPER_CASE with underscores
- Enum values: Use PascalCase

```cpp
// Constants
constexpr int MAX_RETRY_ATTEMPTS = 3;
constexpr const char* DEFAULT_CONFIG_PATH = \"/etc/config.xml\";

// Enumerations
enum class PackageStatus {
    NotInstalled,
    Installed,
    UpdateAvailable,
    Error
};
```

### Member Variables
Use trailing underscore for private member variables:
```cpp
class ConfigurationManager {
private:
    std::string config_path_;
    bool is_loaded_;
    mutable std::mutex config_mutex_;
    static const std::string DEFAULT_CONFIG_NAME;
};
```

## Code Formatting

### Indentation and Spacing
- Use 4 spaces for indentation (no tabs)
- Maximum line length: 100 characters
- Use blank lines to separate logical sections

### Braces and Line Breaks
Use Allman style (braces on new line):
```cpp
class PackageInstaller
{
public:
    void install_package(const std::string& package_name)
    {
        if (package_name.empty())
        {
            throw std::invalid_argument(\"Package name cannot be empty\");
        }
        
        for (const auto& dependency : get_dependencies(package_name))
        {
            install_dependency(dependency);
        }
    }
    
private:
    void install_dependency(const std::string& dependency)
    {
        // Implementation
    }
};
```

### Pointer and Reference Declarations
Attach * and & to the type:
```cpp
// Preferred
std::string* config_ptr;
const Configuration& config_ref;
std::unique_ptr<PackageInstaller> installer;

// Avoid  
std::string *config_ptr;
const Configuration &config_ref;
```

## Header Files and Includes

### Include Guards
Use `#pragma once` instead of traditional include guards:
```cpp
#pragma once

#include <memory>
// ... rest of header
```

### Include Order
Organize includes in the following order with blank lines between groups:
1. Corresponding header (for .cpp files)
2. C standard library
3. C++ standard library  
4. Third-party libraries
5. Project headers

```cpp
// In package_installer.cpp
#include \"cassandragargoyle/project/install/package_installer.hpp\"

#include <cstdlib>
#include <cstring>

#include <algorithm>
#include <memory>
#include <string>
#include <vector>

#include <boost/filesystem.hpp>
#include <boost/algorithm/string.hpp>

#include \"cassandragargoyle/project/config/configuration.hpp\"
#include \"cassandragargoyle/project/util/file_utils.hpp\"
```

### Forward Declarations
Use forward declarations in headers to reduce dependencies:
```cpp
// In header file
namespace cassandragargoyle::project::config {
    class Configuration;  // Forward declaration
}

class PackageInstaller
{
public:
    void set_configuration(std::unique_ptr<config::Configuration> config);
    
private:
    std::unique_ptr<config::Configuration> config_;
};
```

## Modern C++ Features

### Smart Pointers
Always prefer smart pointers over raw pointers:
```cpp
class PackageInstaller
{
public:
    // Factory method returning unique_ptr
    static std::unique_ptr<PackageInstaller> create(const std::string& config_path)
    {
        auto installer = std::make_unique<PackageInstaller>();
        installer->load_configuration(config_path);
        return installer;
    }
    
    // Use shared_ptr when ownership is shared
    void set_logger(std::shared_ptr<Logger> logger)
    {
        logger_ = std::move(logger);
    }
    
private:
    std::shared_ptr<Logger> logger_;
    std::unique_ptr<Configuration> config_;
};
```

### Move Semantics
Implement move operations for performance:
```cpp
class PackageInstaller
{
public:
    // Move constructor
    PackageInstaller(PackageInstaller&& other) noexcept
        : config_(std::move(other.config_))
        , logger_(std::move(other.logger_))
        , is_initialized_(other.is_initialized_)
    {
        other.is_initialized_ = false;
    }
    
    // Move assignment
    PackageInstaller& operator=(PackageInstaller&& other) noexcept
    {
        if (this != &other)
        {
            config_ = std::move(other.config_);
            logger_ = std::move(other.logger_);
            is_initialized_ = other.is_initialized_;
            other.is_initialized_ = false;
        }
        return *this;
    }
    
    // Delete copy operations if move-only
    PackageInstaller(const PackageInstaller&) = delete;
    PackageInstaller& operator=(const PackageInstaller&) = delete;
    
private:
    std::unique_ptr<Configuration> config_;
    std::shared_ptr<Logger> logger_;
    bool is_initialized_{false};
};
```

### Range-Based For Loops
Use range-based loops when appropriate:
```cpp
void install_packages(const std::vector<std::string>& packages)
{
    for (const auto& package : packages)
    {
        install_package(package);
    }
}

// With structured bindings (C++17)
void print_package_info(const std::map<std::string, PackageInfo>& packages)
{
    for (const auto& [name, info] : packages)
    {
        std::cout << \"Package: \" << name << \", Version: \" << info.version << std::endl;
    }
}
```

### Auto Keyword
Use auto judiciously for type deduction:
```cpp
// Good uses of auto
auto config = load_configuration(\"config.xml\");
auto it = packages.find(package_name);
const auto& package_list = get_installed_packages();

// Avoid when type clarity is important
std::string config_path = get_config_path();  // Clear intent
int retry_count = 3;  // Simple type, be explicit
```

### Lambda Expressions
Use lambdas for local functions and callbacks:
```cpp
void install_packages_async(const std::vector<std::string>& packages)
{
    std::for_each(std::execution::par_unseq, packages.begin(), packages.end(),
        [this](const std::string& package)
        {
            try
            {
                install_package(package);
            }
            catch (const std::exception& e)
            {
                log_error(\"Failed to install package: \" + package + \", error: \" + e.what());
            }
        });
}
```

## Error Handling

### Exception Safety
Write exception-safe code using RAII:
```cpp
class ConfigurationManager
{
public:
    void save_configuration(const Configuration& config)
    {
        // RAII ensures file is properly closed even if exception is thrown
        std::ofstream file(config_path_);
        if (!file.is_open())
        {
            throw std::runtime_error(\"Failed to open config file: \" + config_path_);
        }
        
        // Serialize config - may throw
        serialize_config(file, config);
        
        // file automatically closed by destructor
    }
    
private:
    std::string config_path_;
};
```

### Custom Exceptions
Create domain-specific exception classes:
```cpp
namespace cassandragargoyle::project::install {

class InstallationException : public std::runtime_error
{
public:
    explicit InstallationException(const std::string& message)
        : std::runtime_error(message)
    {
    }
    
    InstallationException(const std::string& package_name, const std::string& reason)
        : std::runtime_error(\"Failed to install package '\" + package_name + \"': \" + reason)
        , package_name_(package_name)
    {
    }
    
    const std::string& package_name() const noexcept
    {
        return package_name_;
    }
    
private:
    std::string package_name_;
};

} // namespace cassandragargoyle::project::install
```

## Class Design

### Interface Design
Use pure virtual interfaces for contracts:
```cpp
namespace cassandragargoyle::project::install {

class IPackageManager
{
public:
    virtual ~IPackageManager() = default;
    
    virtual void install_package(const std::string& package_name) = 0;
    virtual void remove_package(const std::string& package_name) = 0;
    virtual bool is_package_installed(const std::string& package_name) const = 0;
    virtual std::vector<std::string> list_installed_packages() const = 0;
};

class AptPackageManager : public IPackageManager
{
public:
    void install_package(const std::string& package_name) override;
    void remove_package(const std::string& package_name) override;
    bool is_package_installed(const std::string& package_name) const override;
    std::vector<std::string> list_installed_packages() const override;
};

} // namespace cassandragargoyle::project::install
```

### PIMPL Idiom
Use PIMPL for stable ABI and fast compilation:
```cpp
// In header file
class PackageInstaller
{
public:
    PackageInstaller();
    ~PackageInstaller();
    
    // Move operations
    PackageInstaller(PackageInstaller&& other) noexcept;
    PackageInstaller& operator=(PackageInstaller&& other) noexcept;
    
    // Copy operations deleted for simplicity
    PackageInstaller(const PackageInstaller&) = delete;
    PackageInstaller& operator=(const PackageInstaller&) = delete;
    
    void install_package(const std::string& package_name);
    bool is_package_installed(const std::string& package_name) const;
    
private:
    class Impl;
    std::unique_ptr<Impl> pimpl_;
};

// In implementation file
class PackageInstaller::Impl
{
public:
    void install_package(const std::string& package_name)
    {
        // Implementation details
    }
    
    bool is_package_installed(const std::string& package_name) const
    {
        // Implementation details
        return false;
    }
    
private:
    std::string config_path_;
    std::unique_ptr<Configuration> config_;
};

PackageInstaller::PackageInstaller()
    : pimpl_(std::make_unique<Impl>())
{
}

PackageInstaller::~PackageInstaller() = default;

PackageInstaller::PackageInstaller(PackageInstaller&& other) noexcept = default;
PackageInstaller& PackageInstaller::operator=(PackageInstaller&& other) noexcept = default;

void PackageInstaller::install_package(const std::string& package_name)
{
    pimpl_->install_package(package_name);
}
```

## Memory Management

### RAII Pattern
Always use RAII for resource management:
```cpp
class FileHandler
{
public:
    explicit FileHandler(const std::string& filename)
        : file_(filename)
    {
        if (!file_.is_open())
        {
            throw std::runtime_error(\"Failed to open file: \" + filename);
        }
    }
    
    ~FileHandler()
    {
        // File automatically closed by destructor
    }
    
    // Move-only class
    FileHandler(FileHandler&& other) noexcept = default;
    FileHandler& operator=(FileHandler&& other) noexcept = default;
    FileHandler(const FileHandler&) = delete;
    FileHandler& operator=(const FileHandler&) = delete;
    
    void write(const std::string& data)
    {
        file_ << data;
    }
    
private:
    std::ofstream file_;
};
```

### Container Usage
Choose appropriate containers for your use case:
```cpp
class PackageRegistry
{
private:
    // Fast lookups
    std::unordered_map<std::string, PackageInfo> packages_;
    
    // Ordered iteration
    std::map<std::string, std::string> sorted_packages_;
    
    // Unique items
    std::unordered_set<std::string> installed_packages_;
    
    // Sequential access
    std::vector<std::string> installation_order_;
    
    // Thread-safe operations (if needed)
    mutable std::shared_mutex packages_mutex_;
};
```

## Concurrency

### Thread Safety
Use appropriate synchronization primitives:
```cpp
class ThreadSafePackageInstaller
{
public:
    void install_package(const std::string& package_name)
    {
        std::unique_lock<std::shared_mutex> lock(packages_mutex_);
        // Modify shared state
        installed_packages_.insert(package_name);
    }
    
    bool is_package_installed(const std::string& package_name) const
    {
        std::shared_lock<std::shared_mutex> lock(packages_mutex_);
        // Read-only access allows concurrent reads
        return installed_packages_.find(package_name) != installed_packages_.end();
    }
    
private:
    std::unordered_set<std::string> installed_packages_;
    mutable std::shared_mutex packages_mutex_;
};
```

### Async Programming
Use std::future and std::async for asynchronous operations:
```cpp
class AsyncPackageInstaller
{
public:
    std::future<void> install_package_async(const std::string& package_name)
    {
        return std::async(std::launch::async, [this, package_name]()
        {
            install_package_impl(package_name);
        });
    }
    
    std::future<std::vector<std::string>> get_installed_packages_async() const
    {
        return std::async(std::launch::async, [this]()
        {
            return get_installed_packages_impl();
        });
    }
    
private:
    void install_package_impl(const std::string& package_name)
    {
        // Implementation
    }
    
    std::vector<std::string> get_installed_packages_impl() const
    {
        // Implementation
        return {};
    }
};
```

## Documentation

### Class Documentation
Document classes and their purpose:
```cpp
/**
 * @brief Manages cross-platform package installation.
 * 
 * The PackageInstaller provides a unified interface for installing software
 * packages across different operating systems. It automatically detects the
 * appropriate package manager (APT, Chocolatey, Homebrew) and handles
 * platform-specific installation procedures.
 * 
 * @note This class is not thread-safe. Use external synchronization
 *       for concurrent access.
 * 
 * Example usage:
 * @code
 * auto installer = PackageInstaller::create(\"/etc/config.xml\");
 * installer->install_package(\"python3\");
 * @endcode
 */
class PackageInstaller
{
public:
    /**
     * @brief Installs the specified package.
     * 
     * @param package_name The name of the package to install
     * @throws InstallationException if installation fails
     * @throws std::invalid_argument if package_name is empty
     */
    void install_package(const std::string& package_name);
    
    /**
     * @brief Checks if a package is currently installed.
     * 
     * @param package_name The name of the package to check
     * @return true if the package is installed, false otherwise
     * @throws std::invalid_argument if package_name is empty
     */
    bool is_package_installed(const std::string& package_name) const;
};
```

## Testing

### Unit Test Structure
Use a modern testing framework like Catch2:
```cpp
#include <catch2/catch_test_macros.hpp>
#include \"cassandragargoyle/project/install/package_installer.hpp\"

using namespace cassandragargoyle::project::install;

TEST_CASE(\"PackageInstaller basic functionality\", \"[package_installer]\")
{
    SECTION(\"should create installer successfully\")
    {
        auto installer = PackageInstaller::create(\"test_config.xml\");
        REQUIRE(installer != nullptr);
    }
    
    SECTION(\"should throw exception for empty package name\")
    {
        auto installer = PackageInstaller::create(\"test_config.xml\");
        REQUIRE_THROWS_AS(installer->install_package(\"\"), std::invalid_argument);
    }
    
    SECTION(\"should install valid package\")
    {
        auto installer = PackageInstaller::create(\"test_config.xml\");
        REQUIRE_NOTHROW(installer->install_package(\"test-package\"));
    }
}
```

## Build Configuration

### CMake Configuration
Use modern CMake practices:
```cmake
cmake_minimum_required(VERSION 3.20)

project(CassandraGargoyleProject
    VERSION 1.0.0
    DESCRIPTION \"CassandraGargoyle project utilities\"
    LANGUAGES CXX
)

# Set C++ standard
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# Compiler flags
if(MSVC)
    add_compile_options(/W4 /WX)
else()
    add_compile_options(-Wall -Wextra -Werror -pedantic)
endif()

# Find packages
find_package(Boost REQUIRED COMPONENTS filesystem system)

# Main library
add_library(cassandragargoyle_project
    src/install/package_installer.cpp
    src/config/configuration_manager.cpp
    src/util/file_utils.cpp
)

target_include_directories(cassandragargoyle_project
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
        $<INSTALL_INTERFACE:include>
)

target_link_libraries(cassandragargoyle_project
    PUBLIC
        Boost::filesystem
        Boost::system
)

# Tests
if(BUILD_TESTING)
    find_package(Catch2 REQUIRED)
    
    add_executable(tests
        test/install/test_package_installer.cpp
        test/config/test_configuration_manager.cpp
    )
    
    target_link_libraries(tests
        PRIVATE
            cassandragargoyle_project
            Catch2::Catch2WithMain
    )
    
    include(CTest)
    include(Catch)
    catch_discover_tests(tests)
endif()
```

## Code Quality Tools

### Static Analysis
Use tools for code quality checking:
- **clang-tidy**: Static analysis and code linting
- **cppcheck**: Static analysis for bugs
- **AddressSanitizer**: Runtime memory error detection
- **Valgrind**: Memory debugging and profiling

### Formatting
Use consistent code formatting:
```bash
# .clang-format
BasedOnStyle: Allman
IndentWidth: 4
ColumnLimit: 100
PointerAlignment: Left
ReferenceAlignment: Left
```

## Security Best Practices

### Input Validation
```cpp
void install_package(const std::string& package_name)
{
    // Validate input
    if (package_name.empty())
    {
        throw std::invalid_argument(\"Package name cannot be empty\");
    }
    
    // Prevent command injection
    if (package_name.find_first_of(\";|&\") != std::string::npos)
    {
        throw std::invalid_argument(\"Package name contains invalid characters\");
    }
    
    // Length validation
    if (package_name.length() > MAX_PACKAGE_NAME_LENGTH)
    {
        throw std::invalid_argument(\"Package name too long\");
    }
    
    // Proceed with installation
    install_package_impl(package_name);
}
```

### Safe String Operations
```cpp
#include <string_view>

// Use string_view for read-only string parameters
bool validate_package_name(std::string_view package_name)
{
    return !package_name.empty() && 
           package_name.length() <= MAX_PACKAGE_NAME_LENGTH &&
           package_name.find_first_of(\";|&\") == std::string_view::npos;
}
```

---

**Note**: These guidelines should be adapted based on specific project requirements and evolve with C++ standards. Regular code reviews ensure adherence to these practices.

*Created: 2025-08-23*
*Last updated: 2025-08-23*