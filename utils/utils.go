package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenRandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func Zip(srcDir string, dstfile string) error {
	zf, err := os.Create(dstfile)
	if err != nil {
		fmt.Printf("error creating zip file: %s", err)
		return err
	}

	w := zip.NewWriter(zf)
	defer w.Close()

	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				fmt.Printf("error opening file: %s", err)
				return err
			}
			defer file.Close()

			f, err := w.Create(path)
			if err != nil {
				fmt.Printf("error creating entry for filename in zip: %s", err)
				return err
			}

			_, err = io.Copy(f, file)
			if err != nil {
				fmt.Printf("error copying file to zip: %s", err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
