package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"ytsruh.com/backmeup/utils"
	"ytsruh.com/backmeup/views"
)

type BulkForm struct {
	Urls string `form:"urls" validate:"required,url"`
}

func Bulk(c echo.Context) error {
	var form BulkForm
	err := c.Bind(&form)
	if err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	// Check if form urls is valid and is not empty
	if form.Urls == "" {
		return c.String(http.StatusBadRequest, "urls are empty")
	}
	// Split urls by comma into a slice and trim whitespace
	rawURLs := strings.Split(strings.TrimSpace(form.Urls), ",")
	urls := make([]string, len(rawURLs))
	for i, rawURL := range rawURLs {
		urls[i] = strings.TrimSpace(rawURL)
		// Check if url is valid and is not empty
		if urls[i] == "" {
			return c.String(http.StatusBadRequest, "urls are empty")
		}
		// Check if url is valid
		if !strings.HasPrefix(urls[i], "http") {
			return c.String(http.StatusBadRequest, "url is not valid")
		}
	}
	// Check if urls are more than 25
	//fmt.Printf("Urls = %v\n", len(urls))
	//fmt.Println(urls)
	if len(urls) >= 31 {
		return c.String(http.StatusBadRequest, "too many urls, max 30 allowed")
	}

	// Check if url is a file, get the extension and save the file to temp directory
	tempPath := "temp-" + utils.GenRandomString(10) + "/" // Create a new temp directory outside of range
	for _, link := range urls {
		u, err := url.Parse(link)
		if err != nil {
			return c.String(http.StatusBadRequest, "url could not be parsed")
		}
		extension := filepath.Ext(u.String())
		if extension != ".pdf" && extension != ".xlsx" {
			return c.String(http.StatusBadRequest, "url is not valid file. Only .pdf and .xlsx files are supported")
		}
		// Get file name & extension from url
		filename := path.Base(u.Path)
		//fmt.Println(filename)
		if filename == "." || filename == "/" {
			return c.String(http.StatusBadRequest, "url is not valid file, name could not be found")
		} else {
			extension := path.Ext(filename)
			fileWithoutExt := filename[:len(filename)-len(extension)]
			fmt.Println(fileWithoutExt)
		}
		// Create a new temp directory & get the file
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		resp, err := client.Get(link)
		if err != nil {
			fmt.Printf("error downloading file: %s", err)
			return c.String(http.StatusInternalServerError, "error downloading file")
		}
		defer resp.Body.Close()
		err = os.MkdirAll(tempPath, 0755) // 0755 is the file permission (read and write permission)
		if err != nil {
			fmt.Printf("error creating temp directory: %s", err)
			return c.String(http.StatusInternalServerError, "error creating a temp directory")
		}
		// Create a new file & copy the response body to the file
		file, err := os.Create(tempPath + filename)
		if err != nil {
			fmt.Printf("error creating file: %s", err)
			return c.String(http.StatusInternalServerError, "error creating a file")
		}
		defer file.Close()
		// Copy the response body to the file
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			fmt.Printf("error copying file: %s", err)
			return c.String(http.StatusInternalServerError, "error copying file")
		}
	}

	// Create a new zip file
	zipName := utils.GenRandomString(10)
	dstfile := "zips/" + zipName + ".zip"
	err = utils.Zip(tempPath, dstfile)
	if err != nil {
		fmt.Printf("error creating zip file: %s", err)
		return c.String(http.StatusInternalServerError, "error creating zip file")
	}
	// Remove all files from temp directory
	err = os.RemoveAll(tempPath)
	if err != nil {
		fmt.Printf("error removing temp directory: %s", err)
		return c.String(http.StatusInternalServerError, "error removing temp directory")
	}

	status := fmt.Sprintf("File Count: %d", len(urls))
	return Render(c, http.StatusOK, views.BulkResultsComponent(status, zipName))
}
