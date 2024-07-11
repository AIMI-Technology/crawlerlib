package database

import "time"

type ScrapedData struct {
	ID             string    `db:"id"`
	Text           string    `db:"text"`
	ScrapedAt      time.Time `db:"scraped_at"`
	PublishedAt    time.Time `db:"published_at"`
	Url            string    `db:"url"`
	SourceCountry  string    `db:"source_country"`
	ContentCountry string    `db:"content_country"`
}
