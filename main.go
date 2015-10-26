package main

import (
	"archive/tar"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	"compress/gzip"
	"log"

	"os/exec"
	"path"
	"path/filepath"
)

type config struct {
	args []string
}

func parseFlags() config {

	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] <input directories>\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	//templateFolder := flag.String("t", "template", "Set this to the directory containing the template go expander file")

	flag.Parse()

	//*templateFolder = path.Clean(*templateFolder)

	if flag.NArg() < 2 {
		flag.Usage()
		fmt.Println("Must provide project template folder and output name")
		os.Exit(1)
	}

	ret := config{flag.Args()}

	return ret
}

func buildTar(searchPath string) bytes.Buffer {
	newTar := new(bytes.Buffer)
	gzWriter := gzip.NewWriter(newTar)
	tarWriter := tar.NewWriter(gzWriter)

	filepath.Walk(searchPath, func(searchPath string, fileHeader os.FileInfo, err error) error {
		typeFlag := byte('0')
		size := int64(0)

		if fileHeader.IsDir() {
			typeFlag = byte('5')
		} else {
			size = fileHeader.Size()
		}

		hdr := tar.Header{
			Name:     searchPath,
			Mode:     int64(fileHeader.Mode().Perm()),
			Size:     size,
			ModTime:  fileHeader.ModTime(),
			Typeflag: typeFlag,
		}

		if err := tarWriter.WriteHeader(&hdr); err != nil {
			log.Fatalln(err)
		}

		readFile, _ := os.Open(searchPath)
		if _, err = io.Copy(tarWriter, readFile); err != nil {
			return err
		}

		return nil
	})

	if err := tarWriter.Close(); err != nil {
		log.Fatalln(err)
	}

	if err := gzWriter.Close(); err != nil {
		log.Fatalln(err)
	}

	return *newTar
}

func main() {
	config := parseFlags()

	newTar := buildTar(path.Clean(config.args[0]))

	goTemplate, err := Asset("main.go")

	if err != nil {
		log.Fatalln(err)
	}

	newTar = convertToByteString(newTar)

	r := bytes.Replace(goTemplate, []byte("{{{.}}}"), newTar.Bytes(), -1)

	fi, _ := os.Create("output.go")
	fi.Write(r)
	fi.Close()

	exec.Command("go", "build", "-o", config.args[1], "output.go").Run()

	os.Remove("output.go")
}

const lowerHex = "0123456789abcdef"

func convertToByteString(tar bytes.Buffer) bytes.Buffer {
	if tar.Len() == 0 {
		log.Fatal("no data passed")
	}
	var ret bytes.Buffer

	buf := []byte(`\x00`)

	for _, b := range tar.Bytes() {

		buf[2] = lowerHex[b/16]
		buf[3] = lowerHex[b%16]
		ret.Write(buf)
	}
	return ret
}
