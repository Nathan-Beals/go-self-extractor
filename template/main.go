package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	data := []byte("{{{.}}}")

	r, err := gzip.NewReader(bytes.NewReader(data))

	if err != nil {
		log.Fatal(err)
	}

	tr := tar.NewReader(r)

	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				// end of tar archive
				break
			}
			log.Fatal(err)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			fmt.Println("creating:   " + hdr.Name)

			err = os.MkdirAll(hdr.Name, hdr.FileInfo().Mode()) //Create the directory

			if err != nil {
				log.Fatal(err)
			}

		case tar.TypeReg, tar.TypeRegA:
			fmt.Println("extracting: " + hdr.Name)

			w, err := os.Create(hdr.Name) //Create file in destination

			if err != nil {
				log.Fatal(err)
			}

			_, err = io.Copy(w, tr) //Copy contents to created file

			if err != nil {
				log.Fatal(err)
			}
			w.Close()
		}
	}
}
