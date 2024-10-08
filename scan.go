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
	status := fmt.Sprintf("PDF Count: %d, XLS Count: %d", links.PDFCount, links.XLSCount)
	fmt.Println(status)
	// Save files & generate zip files
	pdfZip, err := saveFiles(links.PDF, ".pdf")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return c.String(http.StatusInternalServerError, "error saving pdf files")
	}

	xlsZip, err := saveFiles(links.XLS, ".xls")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return c.String(http.StatusInternalServerError, "error saving xls files")
	}

	//fmt.Printf("Zip File: %s,%s\n", pdfZip, xlsZip)
	return Render(c, http.StatusOK, views.ResultsComponent(status, pdfZip, xlsZip))
}

type Document struct {
	Name string
	URL  string
}

type Links struct {
	PDF      []Document
	PDFCount int
	XLS      []Document
	XLSCount int
}

func getLinks(url string) (Links, error) {
	links := Links{
		PDF:      []Document{},
		PDFCount: 0,
		XLS:      []Document{},
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
		document := Document{}
		if exists {
			trimmedLink := strings.TrimSpace(href)
			trimmedName := strings.TrimSpace(s.Text())
			if strings.HasSuffix(href, ".pdf") {
				links.PDFCount++
				// Check if link is relative
				if strings.HasPrefix(href, "/") {
					// Get domain from url
					url := strings.Split(url, "/")[2]
					document.URL = "https://" + url + trimmedLink
					document.Name = trimmedName
					links.PDF = append(links.PDF, document)
				} else {
					document.URL = trimmedLink
					document.Name = trimmedName
					links.PDF = append(links.PDF, document)
				}
			}
			if strings.HasSuffix(href, ".xlsx") {
				links.XLSCount++
				// Check if link is relative
				if strings.HasPrefix(href, "/") {
					// Get domain from url
					document.URL = "https://" + url + trimmedLink
					document.Name = trimmedName
					links.XLS = append(links.PDF, document)
				} else {
					document.URL = trimmedLink
					document.Name = trimmedName
					links.XLS = append(links.PDF, document)
				}
			}
		}
	})

	return links, nil
}

func saveFiles(links []Document, ext string) (string, error) {
	if len(links) == 0 {
		return "", nil
	}
	rand := utils.GenRandomString(10)
	path := "temp-" + rand + "/"
	for _, link := range links {
		//fmt.Println("LINK:" + link)
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		resp, err := client.Get(link.URL)
		if err != nil {
			fmt.Printf("error downloading file: %s", err)
			return "", err
		}
		defer resp.Body.Close()
		// Create a new file and check a temp directory exists
		tmpName := link.Name + ext
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
