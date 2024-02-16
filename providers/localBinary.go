package providers

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func LocalBinary(name, path string) ProviderFactory {
	factory := func(ctx context.Context, opts ProviderOptions) (Port, error) {
		return startLocalBinary(ctx, path, name, opts.WorkDir)
	}
	return factory
}

func startLocalBinary(ctx context.Context, path, name, cwd string) (Port, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	if stat.IsDir() {
		binaryName := "pulumi-resource-" + name
		path = filepath.Join(path, binaryName)
	}
	cmd := exec.CommandContext(ctx, path)
	cmd.Dir = cwd
	reader, err := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr
	if err != nil {
		return 0, err
	}
	err = cmd.Start()
	if err != nil {
		return 0, err
	}
	return readPortNumber(reader)
}

func readPortNumber(reader io.Reader) (Port, error) {
	// Now that we have a process, we expect it to write a single line to STDOUT: the port it's listening on.  We only
	// read a byte at a time so that STDOUT contains everything after the first newline.
	var portString string
	b := make([]byte, 1)
	for {
		n, err := reader.Read(b)
		if err != nil {
			return 0, fmt.Errorf("failed to read port number from provider: %v", err)
		}
		if n > 0 && b[0] == '\n' {
			break
		}
		portString += string(b[:n])
	}
	// Trim any whitespace from the first line (this is to handle things like windows that will write
	// "1234\r\n", or slightly odd providers that might add whitespace like "1234 ")
	portString = strings.TrimSpace(portString)

	var port int
	var err error
	if port, err = strconv.Atoi(portString); err != nil {
		return 0, fmt.Errorf("failed to parse port number from provider: %v", err)
	}
	return Port(port), nil
}
