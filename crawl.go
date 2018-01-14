package main

import (
	"bufio"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"os"
	"strings"
	// "time"
)

type Book struct {
	Name             string
	Passwd           string
	DowloadResources []*Resource
}

type Resource struct {
	Name string
	URL  string
}

func (r *Resource) String() string {
	return r.Name + " " + r.URL
}

func (book *Book) String() string {
	resStr := make([]string, 0, len(book.DowloadResources))
	for _, r := range book.DowloadResources {
		resStr = append(resStr, r.String())
	}
	return fmt.Sprintf("| %s | %s | %s", book.Name, book.Passwd, strings.Join(resStr, " "))
}

func NewBook() *Book {
	return &Book{
		DowloadResources: make([]*Resource, 0, 2),
	}
}

func main() {
	f, err := os.OpenFile("./download.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	w := bufio.NewWriter(f)

	urlFile, _ := os.Open("./books.txt")
	urlBF := bufio.NewReader(urlFile)

	c := colly.NewCollector()

	// Visit only domains: hackerspaces.org, wiki.hackerspaces.org
	c.AllowedDomains = []string{"mebook.cc"}
	c.CacheDir = "./cache"

	// On every a element which has href attribute call callback
	c.OnHTML("div.desc", func(e *colly.HTMLElement) {
		e.DOM.Find("p").Each(func(_ int, s *goquery.Selection) {
			line := s.Text()
			if strings.Contains(line, "网盘密码：") {
				book := e.Request.Ctx.GetAny(e.Request.URL.String()).(*Book)
				book.Passwd = line[len("网盘密码："):]
			} else if strings.Contains(line, "文件名称：") {
				book := e.Request.Ctx.GetAny(e.Request.URL.String()).(*Book)
				book.Name = line[len("文件名称："):]
			}
		})
	})

	c.OnHTML("div.list", func(e *colly.HTMLElement) {
		e.DOM.Find("a").Each(func(_ int, s *goquery.Selection) {
			panName := s.Text()
			dl, ok := s.Attr("href")
			if ok {
				newRes := &Resource{
					Name: panName,
					URL:  dl,
				}
				book := e.Request.Ctx.GetAny(e.Request.URL.String()).(*Book)
				book.DowloadResources = append(book.DowloadResources, newRes)
			}
		})
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(req *colly.Request) {
		fmt.Println("Visiting", req.URL.String())
		req.Ctx.Put(req.URL.String(), NewBook())
	})

	c.OnScraped(func(resp *colly.Response) {
		fmt.Printf("Success: %s\n", resp.Request.URL.String())
		book := resp.Request.Ctx.GetAny(resp.Request.URL.String()).(*Book)
		w.Write([]byte(book.String()))
		w.Write([]byte("\n"))
	})

	c.OnError(func(resp *colly.Response, e error) {
		fmt.Printf("Fail: %s\n", resp.Request.URL.String())
		fmt.Println(e)
	})

	// Start scraping on https://hackerspaces.org
	for {
		line, err := urlBF.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimRight(line, "\n")
		if line == "" {
			continue
		}
		todo := fmt.Sprintf("http://mebook.cc/download.php?id=%s", getid(line))
		c.Visit(todo)
	}

	c.Wait()

	w.Flush()
	f.Close()
}

func getid(s string) string {
	sections := strings.Split(s, "/")
	last := sections[len(sections)-1]
	return strings.TrimRight(last, ".html")
}
