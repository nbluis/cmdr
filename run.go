package cmdr

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

func getPATH() []string {
	return strings.Split(os.Getenv("PATH"), ":")
}

func findInPath(cmd string) (found bool) {

	for _, dir := range getPATH() {

		fullPath := fmt.Sprintf("%s/%s", dir, cmd)

		if fileExist(fullPath) {
			found = true
			break
		}
	}

	return
}

func fileExist(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

// RunCmd runs a command in the operating system
func RunCmd(c Command) ([]byte, error) {
	return runCmd(c)
}

func makeCmd(c Command) (cmd *exec.Cmd) {
	if c.Options.UseShell {
		cmd = exec.Command("bash", "-c", fmt.Sprintf("%s %s", c.Command, strings.Join(c.Args, " ")))
	} else {
		cmd = exec.Command(c.Command, c.Args...)
	}

	return
}

func runCmd(c Command) (output []byte, err error) {

	err = validateCmd(c)
	if err != nil {
		return
	}

	cmd := makeCmd(c)

	outReader, _ := cmd.StdoutPipe()
	err = cmd.Start()

	if err != nil {
		err = fmt.Errorf("Error starting a command: %v", err)
		return
	}

	var timer *time.Timer

	if c.Options.Timeout > 0 {

		execLimit := time.Duration(c.Options.Timeout) * time.Second

		timer = time.AfterFunc(execLimit, func() {
			cmd.Process.Kill()
		})
	}

	output, err = ioutil.ReadAll(outReader)

	if err != nil {
		err = fmt.Errorf("Error reading output: %v", err)
	}

	err = cmd.Wait()
	if err != nil {
		err = fmt.Errorf("Error running a command: %v", err)
	}

	if c.Options.Timeout > 0 {
		timer.Stop()
	}

	return
}

func validateCmd(c Command) (err error) {

	if c.Command == "" {
		err = fmt.Errorf("Missing command name")
		return
	}

	if !findInPath(c.Command) {
		err = fmt.Errorf("Command not found in PATH")
		return
	}

	return
}
