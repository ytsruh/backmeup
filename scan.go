package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/labstack/echo/v4"
	"ytsruh.com/backmeup/utils"
	"ytsruh.com/backmeup/views"
)

type Form struct {
	Url string `form:"url" validate:"required,url"`
}

func Scan(c echo.Context) error {
	var form Form
	err := c.Bind(&form)
	if err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	if form.Url == "" || !strings.HasPrefix(form.Url, "http") {
		return c.String(http.StatusBadRequest, "url is not valid")
	}
	links, err := getLinks(form.Url)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return c.String(http.StatusInternalServerError, "internal server error")
	}
	status := fmt.Sprintf("XLS Count: %d, PDF Count: %d", links.XLSCount, links.PDFCount)
	// Save files & generate a zip file
	zipFile, err := saveFiles(links.PDF, ".pdf")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return c.String(http.StatusInternalServerError, "error saving pdf files")
	}
	//fmt.Printf("Zip File: %s\n", zipFile)
	return Render(c, http.StatusNotFound, views.ResultsComponent(status, zipFile))
}

type Links struct {
	PDF      []string
	PDFCount int
	XLS      []string
	XLSCount int
}

func getLinks(url string) (Links, error) {
	links := Links{
		PDF:      []string{},
		PDFCount: 0,
		XLS:      []string{},
		XLSCount: 0,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return links, err
	}
	// Create a Transport that doesn't do any SSL verification
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	// Now create a new client with that Transport
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return links, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return links, err
	}
	// Get all links from page
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			trimmed := strings.TrimSpace(href)
			if strings.HasSuffix(href, ".pdf") {
				links.PDFCount++
				// Check if link is relative
				if strings.HasPrefix(href, "/") {
					// Get domain from url
					url := strings.Split(url, "/")[2]
					links.PDF = append(links.PDF, "https://"+url+trimmed)
				} else {
					links.PDF = append(links.PDF, trimmed)
				}
			}
			if strings.HasSuffix(href, ".xlsx") {
				links.XLSCount++
				// Check if link is relative
				if strings.HasPrefix(href, "/") {
					// Get domain from url
					url := strings.Split(url, "/")[2]
					links.XLS = append(links.XLS, url+trimmed)
				} else {
					links.XLS = append(links.XLS, trimmed)
				}
			}
		}
	})

	return links, nil
}

func saveFiles(links []string, ext string) (string, error) {
	rand := utils.GenRandomString(10)
	path := "temp" + rand + "/"
	for _, link := range links {
		//fmt.Println("LINK:" + link)
		resp, err := http.Get(link)
		if err != nil {
			fmt.Printf("error downloading file: %s", err)
			return "", err
		}
		defer resp.Body.Close()
		// Create a new file and check a temp directory exists
		tmpName := utils.GenRandomString(10) + ext
		err = os.MkdirAll(path, 0755) // 0755 is the file permission (read and write permission)
		if err != nil {
			fmt.Printf("error creating temp directory: %s", err)
			return "", err
		}
		file, err := os.Create(path + tmpName)
		if err != nil {
			fmt.Printf("error creating file: %s", err)
			return "", err
		}
		defer file.Close()
		// Copy the response body to the file
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			fmt.Printf("error copying file: %s", err)
			return "", err
		}
	}
	// Create a new zip file
	var dstfile = "zips/" + rand + ".zip"
	err := utils.Zip(path, dstfile)
	if err != nil {
		fmt.Printf("error creating zip file: %s", err)
		return "", err
	}
	// Remove all files from temp directory
	err = os.RemoveAll(path)
	if err != nil {
		fmt.Printf("error removing temp directory: %s", err)
		return "", err
	}
	return rand, nil
}
