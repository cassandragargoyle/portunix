package app

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/pbnjay/memory"
)

func GetOSName() (string, error) {
	osRelease, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "", errors.New("Error reading /etc/os-release:" + err.Error())
	}
	// Convert the file content to a string.
	osReleaseStr := string(osRelease)
	if strings.Contains(osReleaseStr, "openHABian") {
		return "openHABian", nil
	} else if strings.Contains(osReleaseStr, "Debian") {
		return "Debian", nil
	} else if strings.Contains(osReleaseStr, "Raspbian") {
		return "Raspbian", nil
	}

	return "", errors.New("Unknown operating system")
}

func PrintLine(character string) {
	// Define the character for the horizontal line
	if len(character) == 0 {
		character = "-"
	}

	// Define the width (number of characters) of the line
	width := 40

	// Create a horizontal line by repeating the character
	horizontalLine := ""
	for i := 0; i < width; i++ {
		horizontalLine += character
	}
	// Print the horizontal line
	fmt.Println(horizontalLine)
}

func PrintSystemInfo() {
	PrintLine("-")
	fmt.Printf("OS: %s", runtime.GOOS)            // Operating system name (e.g., "linux", "windows", "darwin")
	fmt.Printf(" ARCH: %s", runtime.GOARCH)       // Architecture (e.g., "amd64", "386", "arm")
	fmt.Printf(" NumCPU: %d\n", runtime.NumCPU()) // Number of logical CPUs
	printNetworkInfo()
	printMemoryInfo()
	PrintLine("-")
}

func PrintLogo() {
	println(" ----------------------------------")
	println(" ------------ PORTUNIX ------------")
	println(" ----------------------------------")
}

func println(a ...any) {
	fmt.Println(a...)
}

func printMemoryInfo() {
	var m runtime.MemStats

	// Read memory statistics
	runtime.ReadMemStats(&m)

	// Total memory allocated by the Go program (in bytes)
	totalAllocated := m.TotalAlloc

	// Convert bytes to a more human-readable format
	totalAllocatedMB := float64(totalAllocated) / 1024 / 1024

	// Convert bytes to a more human-readable format
	totalMB := memory.TotalMemory() / 1024 / 1024

	fmt.Printf("Total memory allocated: %.2f MB\n", totalAllocatedMB)
	fmt.Printf("Total system memory: %d MB\n", totalMB)
}

func printNetworkInfo() {
	// Get a list of network interfaces and their addresses
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Network interface error:", err)
		return
	}

	fmt.Println("Network interfaces:")
	// Iterate through each network interface
	for _, intf := range interfaces {
		fmt.Print(intf.Name)

		// Get the addresses for the current interface
		addrs, err := intf.Addrs()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		// Print each address associated with the interface
		for _, addr := range addrs {
			fmt.Print(" 	 ", addr)
		}
		fmt.Println()
	}
}

func ExecuteFile(fileName string) {

	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting working directory:", err)
		return
	}

	executablePath := wd + string(os.PathSeparator) + fileName

	// Check if the executable file exists
	if _, err := os.Stat(executablePath); os.IsNotExist(err) {
		fmt.Printf("Executable not found at %s\n", executablePath)
		return
	}

	// Create the command to run the executable
	cmd := exec.Command(executablePath)

	// Run the executable and check for errors
	errCmd := cmd.Run()
	if errCmd != nil {
		fmt.Printf("Error running the executable: %s\n", errCmd)
		return
	}
}

func DownloadFile(url, filepath string) error {
	// Create a file to store the downloaded contents
	outFile, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Perform the HTTP GET request
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Check if the response status code indicates success
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed with status code: %d", response.StatusCode)
	}

	// Copy the response body to the output file
	_, err = io.Copy(outFile, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func IsHtml(text string) bool {
	return strings.HasPrefix(strings.TrimSpace(strings.ToLower(text)), "<!doctype html")
}

func GetValueOrDefault(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func ReplaceTemplateVariables(template string, replacements map[string]string) string {
	for key, value := range replacements {
		template = strings.ReplaceAll(template, "$" + "{" + key + "}", value)
	}
	return template
}

func ProcessArguments(arguments []string, enabledArguments []string) (map[string]string, []string) {
	argsMap := map[string]string{}
	other := []string{}
	skip := 0
	founded := false
	for i, str := range arguments {
		if skip > 0 {
			skip--
			continue
		}
		founded = false
		for _, ea := range enabledArguments {
			if strings.HasPrefix(str, ea+"=") {
				version := strings.SplitN(str, "=", 2)[1]
				argsMap[ea] = version
				founded = true
				break
			} else if strings.HasPrefix(str, "-"+ea) || strings.HasPrefix(str, "--"+ea) {
				if i+1 < len(arguments) {
					argsMap[ea] = arguments[i+1]
					skip = 1
					founded = true
					break
				}
			}
		}
		if !founded {
			other = append(other, str)
		}
	}
	return argsMap, other
}

func FileExist(path string) bool {
	_, err := os.Stat(path)
	// Check if the error is because the file does not exist
	return !os.IsNotExist(err)
}