package crawlerlib

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

// x
type LinkFilter func(link string) bool
type LinkHandler func(link string)

type Storage interface {
	LinkSeen(link string) bool
	SetLink(link string)
}
type Crawler struct {
	domains     []string
	baseLink    *string
	collector   *colly.Collector
	linkFilter  LinkFilter
	linkHandler LinkHandler
	storage     Storage
}

func New(domains ...string) *Crawler {
	c := colly.NewCollector(
		colly.Async(true),
		colly.IgnoreRobotsTxt(),
		colly.AllowURLRevisit(),
		colly.AllowedDomains(domains...),
	)

	return &Crawler{
		collector: c,
		domains:   domains,
	}
}

func (c *Crawler) SetLinkFilter(filter LinkFilter) {
	c.linkFilter = filter
}

func (c *Crawler) SetOnLink(handler LinkHandler) {
	c.linkHandler = handler
}

func (c *Crawler) Run() {
	c.collector.OnHTML("a", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		fullLink := c.fullLink(href)
		if !c.storage.LinkSeen(fullLink) && c.linkFilter(fullLink) {
			c.storage.SetLink(fullLink)

			c.linkHandler(fullLink)

			err := e.Request.Visit(fullLink)
			if err != nil {
				panic(err)
			}
		}
	})

	c.collector.Visit(c.getBaseLink())
	c.collector.Wait()
}

func (c *Crawler) fullLink(link string) string {
	isPartial := strings.HasPrefix(link, "/")

	if isPartial {
		return fmt.Sprintf("%s%s", *c.baseLink, link)
	} else {
		return link
	}
}

func (c *Crawler) getBaseLink() string {
	if c.baseLink == nil {
		link := fmt.Sprintf("https://%s", c.domains[0])
		c.baseLink = &link
	}

	return *c.baseLink
}
