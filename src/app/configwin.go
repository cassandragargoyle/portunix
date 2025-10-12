package app

import "fmt"

func ConfigWin() {
	fmt.Println("Running on Windows")
}

func InstallJreWin() {

}

func InstallConfigWin(setupFile string) {
	ExecuteFile(setupFile)
}
