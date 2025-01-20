package libUtils

import (
	"fmt"
	"os"
	"os/exec"
)

func OpenFileInEditor(filePath string) error {
	vimOsCmd := exec.Command("vim", filePath)
	vimOsCmd.Stdin = os.Stdin
	vimOsCmd.Stdout = os.Stdout
	vimOsCmd.Stderr = os.Stderr

	vimErr := vimOsCmd.Run()
	if vimErr != nil {
		nanoOsCmd := exec.Command("nano", filePath)
		nanoOsCmd.Stdin = os.Stdin
		nanoOsCmd.Stdout = os.Stdout
		nanoOsCmd.Stderr = os.Stderr

		nanoErr := nanoOsCmd.Run()
		if nanoErr != nil {
			return fmt.Errorf(
				"Tried opening file '%s' in nano & vim, both failed!\n%s\n%s",
				filePath,
				vimErr.Error(),
				nanoErr.Error(),
			)
		}
	}

	return nil
}
