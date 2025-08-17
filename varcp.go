package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s SOURCE DEST\n", os.Args[0])
		os.Exit(1)
	}

	source := os.Args[1]
	dest := os.Args[2]

	if err := varcp(source, dest); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func varcp(src, dst string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	content := string(in)

	// Find all variables in the file
	matches := re.FindAllStringSubmatch(content, -1)
	vars := make(map[string]string)

	for _, m := range matches {
		name := m[1]
		if _, ok := vars[name]; ok {
			continue // already handled
		}

		if val, ok := os.LookupEnv(name); ok {
			vars[name] = val
		} else {
			val, err := prompt(fmt.Sprintf("Enter value for %s: ", name))
			if err != nil {
				return err
			}
			vars[name] = val
		}
	}

	// Replace variables
	out := re.ReplaceAllStringFunc(content, func(s string) string {
		name := re.FindStringSubmatch(s)[1]
		return vars[name]
	})

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Write file
	return os.WriteFile(dst, []byte(out), 0644)
}

func prompt(msg string) (string, error) {
	fmt.Fprint(os.Stdout, msg)
	reader := bufio.NewReader(os.Stdin)
	resp, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(resp), nil
}
