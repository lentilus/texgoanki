package compile

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// compile pdf from tex file and return pdfs path
func Latexmk(source string, outdir string) (string, error) {
	cmd := exec.Command(
		"latexmk",
		"-pdf",
		"-f",
		"-norc",
		"-lualatex",
		"-interaction=batchmode",
		fmt.Sprintf("-outdir=%s", outdir),
		"-cd",
		source,
	)

	cmd.Run()
	// latexmk really likes to error so
	// we just check if a pdf was produced

	// turn .tex path into .pdf path
	_, file := path.Split(source)
	pdf_file := strings.Split(file, ".")[0]
	pdf_file += ".pdf"

	pdf_path := path.Join(outdir, pdf_file)

	_, statErr := os.Stat(pdf_path)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return "", errors.New("Compilation failed")
		}
	}
	return pdf_path, nil
}

func Pdf2svg(source string, target string) error {
	cmd := exec.Command(
		"pdf2svg",
		source,
		target,
	)
	err := cmd.Run()
	return err
}

func ResizeSvg(source string) error {
	cmd := exec.Command(
		"inkscape",
		"--actions",
		"select-all;fit-canvas-to-selection",
		"--export-overwrite",
		source,
	)
	err := cmd.Run()
	return err
}

// walks directory and returns cummulated size of files in bytes
func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func Src2svg(
	src string,
	root string,
	context []string,
	srcTarget string,
	ramdir string,
) (string, error) {
	var size int64
	size = 0
	for _, c := range context {
		oc := path.Join(root, c) // original context
		s, err := dirSize(oc)
		if err != nil {
			return "", err
		}
		size += s
	}

	if size > 1024*1024 {
		return "", errors.New(fmt.Sprintf("Context too large: %db", size))
	}

	base := path.Base(root)
	rcontext := path.Join(ramdir, base)
	rmain := path.Join(rcontext, srcTarget)

	// create external dir
	err := os.MkdirAll(rcontext, os.ModePerm)
	if err != nil {
		return "", err
	}

	// make sure we remove our dir in ram again
	defer func() {
		err := os.RemoveAll(rcontext)
		if err != nil {
			panic(err)
		}
	}()

	// copy all context over
	for _, c := range context {
		rc := path.Join(rcontext, c) // in ram context
		oc := path.Join(root, c)     // original context

		info, err := os.Stat(oc)
		if err != nil {
			return "", err
		}

		if info.IsDir() {
			CreateIfNotExists(rc, os.ModePerm)
			err = CopyDirectory(oc, rc)
		} else {
			CreateIfNotExists(path.Dir(rc), os.ModePerm)
			err = Copy(oc, rc)
		}
		if err != nil {
			return "", err
		}
	}

	// write src to file
	err = os.WriteFile(rmain, []byte(src), os.ModePerm)
	if err != nil {
		return "", err
	}

	// compile pdf
	pdf, err := Latexmk(rmain, rcontext)
	if err != nil {
		return "", err
	}

	// convert to svg
	svg := path.Join(rcontext, "converted.svg")
	err = Pdf2svg(pdf, svg)
	if err != nil {
		return "", err
	}

	// try to rezise (we dont care about errors)
	ResizeSvg(svg)

	b, err := os.ReadFile(svg)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}
