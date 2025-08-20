package app

func CreateVm(vmtype string, vmname string, iso string, basefolder string) ([]byte, error) {
	exist, error := PreprocessorCheckExists()
	if error != nil {
		return nil, error
	}

	if exist {
		var cmd []string
		cmd = append(cmd, vmtype)
		cmd = append(cmd, "--create")
		cmd = append(cmd, "--type")
		cmd = append(cmd, "demo")
		cmd = append(cmd, "--vmname")
		cmd = append(cmd, vmname)
		cmd = append(cmd, "--iso")
		cmd = append(cmd, iso)
		cmd = append(cmd, "--basefolder")
		cmd = append(cmd, basefolder)

		output, error := PreprocessorExecute(cmd)
		return output, error
	}
	return nil, nil
}