package compile

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
)

func Base64Encode(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

func Latexmk(source string, outdir string) error {
	bin, err := exec.LookPath("latexmk")
	if err != nil {
		return err
	}

	args := []string{
		"latexmk",
		"-pdf",
		"-f",
		"-norc",
		"-lualatex",
		"-interaction=batchmode",
		fmt.Sprintf("-outdir=%s", outdir),
		"-cd",
		source,
	}

	execErr := syscall.Exec(bin, args, os.Environ())
	if execErr != nil {
		return execErr
	}
	return nil
}

func SVG(source string) error {
	// carefull here: we read the entire file into ram
	b, err := os.ReadFile(source)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(base64.StdEncoding.EncodeToString(b))

	return nil
}

// func Pdf(context string, source string)
