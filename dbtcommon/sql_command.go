package dbtcommon

import "os/exec"

// SqlCommand returns the command to be run
func SqlCommand(fileName string) *exec.Cmd {
	return exec.Command(PsqlPath,
		"-v", "ON_ERROR_STOP=1",
		"-q",
		"-d", DbName,
		"-f", fileName)
}
