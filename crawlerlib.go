package crawlerlib

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/AIMI-Technology/crawlerlib/classifier"
	"github.com/AIMI-Technology/crawlerlib/database"

	"github.com/PuerkitoBio/goquery"
	lru "github.com/hashicorp/golang-lru"
)

type PageData struct {
	Url  string
	Text string
	Date *time.Time
}

type DocumentInterface interface {
	Find(selector string) *Selection
}

type SelectionInterface interface {
	Each(func(idx int, selection *Selection))
}
type Document struct {
	*goquery.Document
}

func (d *Document) Find(selector string) *Selection {
	s := d.Document.Find(selector)
	return &Selection{s}
}

type Selection struct {
	*goquery.Selection
}

func (s *Selection) Each(callback func(idx int, selection *Selection)) {
	s.Selection.Each(func(idx int, sel *goquery.Selection) {
		callback(idx, &Selection{sel})
	})
}

type Crawler struct {
	baseUrl         string
	visited         *lru.Cache
	articleSelector string
	hrefPattern     *regexp.Regexp
	linkPattern     *regexp.Regexp
	onDocument      func(doc DocumentInterface, url string) (*PageData, error)
	onRelevant      func(data *PageData)
	processHref     func(href string) string
	linkChan        chan *PageData
	dateCutoff      time.Time
	sourceCountry   string
	numOfWorkers    int
}

type CrawlerConfig struct {
	BaseUrl string
	// ArticleTextSelector string
	HrefPattern   string
	LinkPattern   string
	DateCutoff    time.Time
	NumOfWorkers  *int
	SourceCountry string
}

func NewCrawler(
	config CrawlerConfig) *Crawler {
	hrefRegex := regexp.MustCompile(config.HrefPattern)
	linkRegex := regexp.MustCompile(config.LinkPattern)
	visited, err := lru.New(10000)
	if err != nil {
		panic(err)
	}

	linkChan := make(chan *PageData, 1000)
	numOfWorkers := 3
	if config.NumOfWorkers != nil {
		numOfWorkers = *config.NumOfWorkers
	}

	dateCutoff, _ := time.Parse(time.DateOnly, "2023-12-31")

	return &Crawler{
		baseUrl:     config.BaseUrl,
		visited:     visited,
		hrefPattern: hrefRegex,
		linkPattern: linkRegex,
		// articleSelector: config.ArticleTextSelector,
		linkChan:      linkChan,
		dateCutoff:    dateCutoff,
		sourceCountry: config.SourceCountry,
		numOfWorkers:  numOfWorkers,
	}
}

func (c *Crawler) Start(url string) {
	for i := 0; i < c.numOfWorkers; i++ {
		go c.worker(c.linkChan, i+1)
	}

	c.visit(url)
}

func (c *Crawler) OnDocument(handler func(doc DocumentInterface, url string) (*PageData, error)) {
	c.onDocument = handler
}

func (c *Crawler) OnRelevant(handler func(data *PageData)) {
	c.onRelevant = handler
}

func (c *Crawler) ProcessHref(handler func(href string) string) {
	c.processHref = handler
}

func (c *Crawler) visit(url string) {
	queue := []string{url}

	for len(queue) > 0 {
		url := queue[0]
		queue = queue[1:]

		log.Println("Visiting", url)

		if _, seen := c.visited.Get(url); seen {
			continue
		}
		c.visited.Add(url, true)

		doc, err := c.get(url)
		if err != nil {
			fmt.Println("err", err)
			continue
		}

		doc.Find("a").Each(func(i int, selection *goquery.Selection) {
			if href, found := selection.Attr("href"); found {
				if c.hrefPattern.MatchString(href) {
					var link string
					if c.processHref != nil {
						link = c.processHref(href)
					} else {
						link = c.baseUrl + href
					}

					if _, seen := c.visited.Get(link); seen {
						return
					}

					queue = append(queue, link)
					if c.linkPattern.MatchString(link) {
						doc, err := c.get(link)
						if err != nil {
							return
						}

						var d Document
						if doc != nil {
							d = Document{doc}
						}

						pageData, err := c.onDocument(&d, link)
						if err != nil {
							log.Println("Page data error", err)
							return
						}

						c.visited.Add(link, true)
						c.linkChan <- pageData
					}
				}
			}
		})
	}
}

func (c *Crawler) get(url string) (*goquery.Document, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}

	return doc, nil
}

func (c *Crawler) worker(pageData chan *PageData, id int) {
	for pageDatum := range pageData {
		if pageDatum == nil {
			log.Println("~~~ skipping processing due to nil data")
			continue
		}

		log.Println("Processing", id, pageDatum.Url)
		if classifier.IsArticleRelevant(pageDatum.Text) {
			ctx := context.Background()
			log.Println("Saving", pageDatum.Url)

			textCollection := strings.TrimSpace(pageDatum.Text)

			var publishedAt time.Time
			if pageDatum.Date != nil {
				publishedAt = *pageDatum.Date
				if publishedAt.Before(c.dateCutoff) {
					return
				}
			}

			err := database.PutItem(ctx, database.ScrapedData{
				Text:          textCollection,
				SourceCountry: c.sourceCountry,
				Url:           pageDatum.Url,
				ScrapedAt:     time.Now().UTC(),
				PublishedAt:   &publishedAt,
			})
			if err != nil {
				panic(err)
			}
		}
	}
}
