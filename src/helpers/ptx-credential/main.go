package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var version = "dev"

// Flags
var (
	flagStore    string
	flagLabel    string
	flagPassword bool
	flagQuiet    bool
	flagJSON     bool
)

// rootCmd represents the base command for ptx-credential
var rootCmd = &cobra.Command{
	Use:   "ptx-credential",
	Short: "Secure credential storage and retrieval",
	Long: `PTX-Credential - Secure Credential Management

Securely store and retrieve credentials (API keys, passwords, tokens) with
AES-256-GCM encryption and PBKDF2 key derivation.

Features:
  - AES-256-GCM encryption with PBKDF2-HMAC-SHA256 key derivation
  - Machine-bound encryption (no password needed for default usage)
  - Optional password protection for additional security
  - M365 token compatibility with Java TokenStorage
  - Multiple named credential stores

Storage location: ~/.portunix/credentials/`,
	Version: version,
}

// setCmd handles "portunix credential set" command
var setCmd = &cobra.Command{
	Use:   "set <name> <value>",
	Short: "Store a credential",
	Long: `Store a credential securely.

Examples:
  portunix credential set github-token "ghp_xxxxxxxxxxxx"
  portunix credential set api-key "secret123" --label "Production API Key"
  portunix credential set company-secret "xxx" --store secure --password`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		value := args[1]

		password := ""
		if flagPassword {
			var err error
			password, err = promptPassword("Enter password: ")
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
		}

		storeName := flagStore
		if storeName == "" {
			storeName = defaultStoreName
		}

		storage, err := NewStorage(storeName, password)
		if err != nil {
			return err
		}

		if err := storage.Set(name, value, flagLabel, nil); err != nil {
			return err
		}

		if !flagQuiet {
			fmt.Printf("Credential '%s' stored successfully\n", name)
		}
		return nil
	},
}

// getCmd handles "portunix credential get" command
var getCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Retrieve a credential",
	Long: `Retrieve a credential value.

Examples:
  portunix credential get github-token
  portunix credential get api-key --quiet
  portunix credential get company-secret --store secure --password`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		password := ""
		if flagPassword {
			var err error
			password, err = promptPassword("Enter password: ")
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
		} else {
			// Check if store is password-protected
			storeName := flagStore
			if storeName == "" {
				storeName = defaultStoreName
			}
			isProtected, _ := IsPasswordProtected(storeName)
			if isProtected {
				var err error
				password, err = promptPassword("Enter password: ")
				if err != nil {
					return fmt.Errorf("failed to read password: %w", err)
				}
			}
		}

		storeName := flagStore
		if storeName == "" {
			storeName = defaultStoreName
		}

		storage, err := NewStorage(storeName, password)
		if err != nil {
			return err
		}

		value, err := storage.Get(name)
		if err != nil {
			return err
		}

		if flagQuiet {
			fmt.Print(value)
		} else {
			fmt.Println(value)
		}
		return nil
	},
}

// deleteCmd handles "portunix credential delete" command
var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a credential",
	Long: `Delete a credential from the store.

Examples:
  portunix credential delete github-token
  portunix credential delete api-key --store secure`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		password := ""
		if flagPassword {
			var err error
			password, err = promptPassword("Enter password: ")
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
		} else {
			storeName := flagStore
			if storeName == "" {
				storeName = defaultStoreName
			}
			isProtected, _ := IsPasswordProtected(storeName)
			if isProtected {
				var err error
				password, err = promptPassword("Enter password: ")
				if err != nil {
					return fmt.Errorf("failed to read password: %w", err)
				}
			}
		}

		storeName := flagStore
		if storeName == "" {
			storeName = defaultStoreName
		}

		storage, err := NewStorage(storeName, password)
		if err != nil {
			return err
		}

		if err := storage.Delete(name); err != nil {
			return err
		}

		if !flagQuiet {
			fmt.Printf("Credential '%s' deleted successfully\n", name)
		}
		return nil
	},
}

// listCmd handles "portunix credential list" command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all credentials",
	Long: `List all credentials in a store (names and labels only, never values).

Examples:
  portunix credential list
  portunix credential list --store secure
  portunix credential list --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		password := ""
		if flagPassword {
			var err error
			password, err = promptPassword("Enter password: ")
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
		} else {
			storeName := flagStore
			if storeName == "" {
				storeName = defaultStoreName
			}
			isProtected, _ := IsPasswordProtected(storeName)
			if isProtected {
				var err error
				password, err = promptPassword("Enter password: ")
				if err != nil {
					return fmt.Errorf("failed to read password: %w", err)
				}
			}
		}

		storeName := flagStore
		if storeName == "" {
			storeName = defaultStoreName
		}

		storage, err := NewStorage(storeName, password)
		if err != nil {
			return err
		}

		credentials, err := storage.List()
		if err != nil {
			return err
		}

		if flagJSON {
			data, err := json.MarshalIndent(credentials, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		if len(credentials) == 0 {
			fmt.Println("No credentials stored")
			return nil
		}

		// Table output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tLABEL\tUPDATED")
		for _, cred := range credentials {
			label := cred.Label
			if label == "" {
				label = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", cred.Name, label, cred.Updated.Format("2006-01-02"))
		}
		w.Flush()
		return nil
	},
}

// storeCmd is the parent command for store management
var storeCmd = &cobra.Command{
	Use:   "store",
	Short: "Manage credential stores",
	Long:  `Manage credential stores (create, list, delete).`,
}

// storeCreateCmd handles "portunix credential store create" command
var storeCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new credential store",
	Long: `Create a new credential store.

Examples:
  portunix credential store create mystore
  portunix credential store create secure --password`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		storeName := args[0]

		password := ""
		if flagPassword {
			var err error
			password, err = promptPassword("Enter password: ")
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			confirmPassword, err := promptPassword("Confirm password: ")
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			if password != confirmPassword {
				return fmt.Errorf("passwords do not match")
			}
		}

		if err := CreateStore(storeName, password); err != nil {
			return err
		}

		if !flagQuiet {
			if password != "" {
				fmt.Printf("Password-protected store '%s' created successfully\n", storeName)
			} else {
				fmt.Printf("Store '%s' created successfully\n", storeName)
			}
		}
		return nil
	},
}

// storeListCmd handles "portunix credential store list" command
var storeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all credential stores",
	Long:  `List all available credential stores.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		stores, err := ListStores()
		if err != nil {
			return err
		}

		if flagJSON {
			data, err := json.MarshalIndent(stores, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		if len(stores) == 0 {
			fmt.Println("No credential stores found")
			return nil
		}

		fmt.Println("Available stores:")
		for _, store := range stores {
			isProtected, _ := IsPasswordProtected(store)
			if isProtected {
				fmt.Printf("  %s (password-protected)\n", store)
			} else {
				fmt.Printf("  %s\n", store)
			}
		}
		return nil
	},
}

// storeDeleteCmd handles "portunix credential store delete" command
var storeDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a credential store",
	Long: `Delete a credential store and all its credentials.

Examples:
  portunix credential store delete mystore`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		storeName := args[0]

		if storeName == defaultStoreName {
			return fmt.Errorf("cannot delete the default store")
		}

		// Confirm deletion
		fmt.Printf("Are you sure you want to delete store '%s'? This action cannot be undone. [y/N]: ", storeName)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Deletion cancelled")
			return nil
		}

		if err := DeleteStoreByName(storeName); err != nil {
			return err
		}

		if !flagQuiet {
			fmt.Printf("Store '%s' deleted successfully\n", storeName)
		}
		return nil
	},
}

// m365Cmd is the parent command for M365 compatibility
var m365Cmd = &cobra.Command{
	Use:   "m365",
	Short: "M365 token compatibility mode",
	Long: `M365 token compatibility mode for Java TokenStorage compatibility.

These commands allow reading and writing M365 tokens in a format compatible
with the Java TokenStorage implementation used by m365-extractor plugin.`,
}

// m365GetCmd handles "portunix credential m365 get" command
var m365GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get M365 tokens",
	Long: `Get M365 tokens in JSON format.

This command retrieves M365 tokens from the legacy token file
(~/.portunix/.portunix-m365-tokens.enc) in a format compatible
with Java TokenStorage.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		password := ""
		if flagPassword {
			var err error
			password, err = promptPassword("Enter password: ")
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
		}

		storage, err := NewM365Storage(password)
		if err != nil {
			return err
		}

		data, err := storage.GetRawM365Data()
		if err != nil {
			return err
		}

		fmt.Println(data)
		return nil
	},
}

// m365SetCmd handles "portunix credential m365 set" command
var m365SetCmd = &cobra.Command{
	Use:   "set <json>",
	Short: "Set M365 tokens",
	Long: `Set M365 tokens from JSON.

This command stores M365 tokens in the legacy token file
(~/.portunix/.portunix-m365-tokens.enc) in a format compatible
with Java TokenStorage.

Example:
  portunix credential m365 set '{"accessToken":"...","refreshToken":"...","expiresAt":1234567890}'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonData := args[0]

		password := ""
		if flagPassword {
			var err error
			password, err = promptPassword("Enter password: ")
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
		}

		storage, err := NewM365Storage(password)
		if err != nil {
			return err
		}

		if err := storage.SetRawM365Data(jsonData); err != nil {
			return err
		}

		if !flagQuiet {
			fmt.Println("M365 tokens stored successfully")
		}
		return nil
	},
}

// m365DeleteCmd handles "portunix credential m365 delete" command
var m365DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete M365 tokens",
	Long:  `Delete M365 tokens from the legacy token file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		password := ""
		if flagPassword {
			var err error
			password, err = promptPassword("Enter password: ")
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
		}

		storage, err := NewM365Storage(password)
		if err != nil {
			return err
		}

		if err := storage.DeleteStore(); err != nil {
			return err
		}

		if !flagQuiet {
			fmt.Println("M365 tokens deleted successfully")
		}
		return nil
	},
}

// infoCmd shows system information used for seed generation
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show system information used for key derivation",
	Long: `Show system information used for cryptographic key derivation.

This is useful for debugging compatibility issues with Java TokenStorage.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := GetOSInfo()
		if err != nil {
			return err
		}

		if flagJSON {
			data, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("Hostname: %s\n", info.Hostname)
		fmt.Printf("Username: %s\n", info.Username)
		fmt.Printf("OS Name:  %s\n", info.OSName)
		fmt.Printf("Home Dir: %s\n", info.HomeDir)

		seed, _ := GenerateDefaultSeed()
		fmt.Printf("\nDefault Seed: %s\n", seed)

		m365Seed, _ := GenerateDefaultM365Seed()
		fmt.Printf("M365 Seed:    %s\n", m365Seed)

		return nil
	},
}

// promptPassword prompts for password input without echo
func promptPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	// Check if we're in a terminal
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		password, err := term.ReadPassword(fd)
		fmt.Println() // Print newline after password input
		if err != nil {
			return "", err
		}
		return string(password), nil
	}

	// Fallback for non-terminal input (e.g., piped input)
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(password), nil
}

// checkEnvPassword checks for password in environment variable
func checkEnvPassword() string {
	return os.Getenv("PORTUNIX_CREDENTIAL_PASSWORD")
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&flagStore, "store", "", "Credential store name (default: \"default\")")
	rootCmd.PersistentFlags().BoolVar(&flagPassword, "password", false, "Use password-protected store")
	rootCmd.PersistentFlags().BoolVar(&flagQuiet, "quiet", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output in JSON format")

	// Set command flags
	setCmd.Flags().StringVar(&flagLabel, "label", "", "Human-readable label for the credential")

	// Add commands
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(infoCmd)

	// Store management commands
	storeCmd.AddCommand(storeCreateCmd)
	storeCmd.AddCommand(storeListCmd)
	storeCmd.AddCommand(storeDeleteCmd)
	rootCmd.AddCommand(storeCmd)

	// M365 compatibility commands
	m365Cmd.AddCommand(m365GetCmd)
	m365Cmd.AddCommand(m365SetCmd)
	m365Cmd.AddCommand(m365DeleteCmd)
	rootCmd.AddCommand(m365Cmd)
}

func main() {
	// Handle dispatcher pattern: when called as "portunix credential ...",
	// the dispatcher passes "credential" as the first argument which we need to skip
	if len(os.Args) > 1 && os.Args[1] == "credential" {
		// Remove "credential" from args so Cobra can process subcommands directly
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}

	// Check for password in environment variable
	if envPassword := checkEnvPassword(); envPassword != "" {
		flagPassword = true
	}

	// Check for store in environment variable
	if envStore := os.Getenv("PORTUNIX_CREDENTIAL_STORE"); envStore != "" && flagStore == "" {
		flagStore = envStore
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
