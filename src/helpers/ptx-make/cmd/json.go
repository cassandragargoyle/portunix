package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RunJson executes the json command
func RunJson(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: json <key>=<value> [key=value...]")
	}

	data := make(map[string]string)

	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid key=value pair: %s", arg)
		}
		key := parts[0]
		value := parts[1]
		data[key] = value
	}

	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to generate JSON: %w", err)
	}

	fmt.Println(string(output))
	return nil
}
