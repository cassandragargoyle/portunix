package app

import (
	"log"
	"os/exec"
)

func PreprocessorCheckExists() (bool, error) {
	exists, err := CheckIfExecutableExists("tp.cli")
	if err != nil {
		return false, err
	}

	if !exists {
		log.Fatalf("The program '%s' is not available.\n", "tp.cli")
		return false, nil
	}
	return true, nil
}

func PreprocessorExecute(args []string) ([]byte, error) {
	exists, err := PreprocessorCheckExists()
	if err != nil {
		return nil, err
	}

	if exists {
		cmd := exec.Command("tp.cli", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("Error running the command: %v\n", err)
		}
		return output, nil
	}
	return nil, nil
}

func CheckIfExecutableExists(program string) (bool, error) {
	_, err := exec.LookPath(program)
	if err != nil {
		return false, err
	}
	return true, nil
}
