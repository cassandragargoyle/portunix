package app

import (
	"fmt"
)

func Help(arguments []string) error {
	fmt.Println("The install command installs software on the computer.")
	fmt.Println("parameters:")
	fmt.Println("  --software name: name of the desired software (java, python, daemon, service, all)")
	fmt.Println()

	fmt.Println("The unzip command extracts the ZIP file to the specified target directory.")
	fmt.Println("parameters:")
	fmt.Println("  --path: path to the ZIP file to be extracted")
	fmt.Println("  --destinationpath: path to the target directory where the files should be extracted")
	fmt.Println()

	return nil
}
