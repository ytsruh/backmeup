package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/labstack/echo/v4"
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
	fmt.Println(form.Url)
	if form.Url == "" || !strings.HasPrefix(form.Url, "http") {
		return c.String(http.StatusBadRequest, "url is not valid")
	}
	links, err := getLinks(form.Url)
	if err != nil {
		return c.String(http.StatusInternalServerError, "internal server error")
	}
	fmt.Println(links.PDFCount)
	fmt.Println(links.XLSCount)
	return c.String(http.StatusOK, fmt.Sprintf("XLS Count: %d, PDF Count: %d", links.XLSCount, links.PDFCount))
}

type Links struct {
	PDF      map[string]string
	PDFCount int
	XLS      map[string]string
	XLSCount int
}

func getLinks(url string) (Links, error) {
	links := Links{
		PDF:      make(map[string]string),
		PDFCount: 0,
		XLS:      make(map[string]string),
		XLSCount: 0,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return links, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return links, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return links, err
	}

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			if strings.HasSuffix(href, ".pdf") {
				links.PDFCount++
				links.PDF[href] = s.Text()
			}
			if strings.HasSuffix(href, ".xlsx") {
				links.XLSCount++
				links.XLS[href] = s.Text()
			}
		}
	})

	return links, nil
}
