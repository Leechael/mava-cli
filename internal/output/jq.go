package output

import (
	"fmt"
	"os"
	"os/exec"
)

// RunJQ pipes jsonData through the jq binary with the given filter expression.
// Requires jq to be installed on the system.
func RunJQ(jsonData []byte, filter string) error {
	cmd := exec.Command("jq", filter)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = nil // we'll use a pipe instead

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("jq: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("jq not found, install it or use --json without --jq: %w", err)
	}
	stdin.Write(jsonData)
	stdin.Close()
	return cmd.Wait()
}
