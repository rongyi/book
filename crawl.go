package main

import (
	"bufio"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"os"
	"strings"
	"time"
)

func main() {
	f, err := os.OpenFile("./books.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	w := bufio.NewWriter(f)

	c := colly.NewCollector()
	c.Limit(&colly.LimitRule{
		RandomDelay: time.Second * 1,
	})

	// Visit only domains: hackerspaces.org, wiki.hackerspaces.org
	c.AllowedDomains = []string{"mebook.cc"}
	c.CacheDir = "./cache"

	// On every a element which has href attribute call callback
	c.OnHTML("div[id=primary]", func(e *colly.HTMLElement) {
		e.DOM.Find("ul.list li").Each(func(_ int, s *goquery.Selection) {
			v, ok := s.Find("div h2 a").Attr("href")
			if ok {
				books := e.Request.Ctx.GetAny(e.Request.URL.String()).([]string)
				books = append(books, v)
				// we need to put it back
				e.Request.Ctx.Put(e.Request.URL.String(), books)
			}
		})
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(req *colly.Request) {
		fmt.Println("Visiting", req.URL.String())
		// every page has 10 books
		req.Ctx.Put(req.URL.String(), make([]string, 0, 10))
	})

	c.OnScraped(func(resp *colly.Response) {
		fmt.Printf("Success: %s\n", resp.Request.URL.String())
		books := resp.Request.Ctx.GetAny(resp.Request.URL.String()).([]string)
		lines := strings.Join(books, "\n")

		w.Write([]byte(lines))
		w.Write([]byte("\n"))
	})

	c.OnError(func(resp *colly.Response, e error) {
		fmt.Printf("Fail: %s\n", resp.Request.URL.String())
		fmt.Println(e)
	})

	// Start scraping on https://hackerspaces.org
	for i := 1; i <= 565; i++ {
		todo := fmt.Sprintf("http://mebook.cc/page/%d", i)
		c.Visit(todo)
	}
	c.Wait()

	w.Flush()
	f.Close()
}
