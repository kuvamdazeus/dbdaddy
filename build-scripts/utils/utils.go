package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

func GetCurrentVersion() string {
	r, err := os.ReadFile("version")
	if err != nil {
		panic("couldn't read file 'version'")
	}

	return string(r)
}

func GetOutFilePath(goos string, goarch string) string {
	outFile := path.Join("bin", fmt.Sprintf("dbdaddy-%s-%s", goos, goarch))

	return outFile
}

func Build(goos string, goarch string) {
	fmt.Println("Starting build for", goos, goarch)

	GOOS := strings.ToLower(goos)
	GOARCH := strings.ToLower(goarch)

	outFile := path.Join("bin", fmt.Sprintf("dbdaddy-%s-%s", GOOS, GOARCH))

	cmd := exec.Command("go", "build", "-o", outFile, ".")
	cmd.Env = append(os.Environ(), "GOOS="+GOOS, "GOARCH="+GOARCH)
	cmdErr := cmd.Run()
	if cmdErr != nil {
		fmt.Println("ERR:", cmdErr)
		panic("error occured while running 'go build'")
	}

	fmt.Println("Built binary for", goos, goarch)
	fmt.Println()
}

func Release(version string, dir string) {
	ghCmd := exec.Command("gh", "release", "create", version, dir)
	ghCmd.Stdout = os.Stdout
	ghCmd.Stdin = os.Stdin
	ghCmd.Stderr = os.Stderr

	ghCmdErr := ghCmd.Run()
	if ghCmdErr != nil {
		panic("error occured while running 'gh release command'")
	}
}