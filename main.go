package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var bankSchema = `
CREATE TABLE jp_bank( 
	id integer primary key autoincrement,
	kanji_name text not null, 
	kata_name text, 
	bank_code char(10));
`

var branchSchema = `
CREATE TABLE jp_bank_branch(
	id integer primary key autoincrement,
	kanji_name text not null,
	kata_name text,
	branch_code char(10),
	bank_code char(10), 
	address text, 
	phone char(100));
`

func main() {
	db, err := sqlx.Connect("sqlite3", "jpbank.db")
	if err != nil {
		log.Fatalln(err)
	}

	db.Exec(bankSchema)
	db.Exec(branchSchema)

	itemCh := make(chan interface{})
	doneCh := make(chan struct{})
	jpBankPipeline := NewJPBankPipeline(db, itemCh, doneCh)

	go jpBankPipeline.Run()

	baseURL := "https://bkichiran.hikak.com/"
	c := colly.NewCollector()

	branchListCollector := c.Clone()
	branchDetailCollector := c.Clone()

	c.OnHTML("tr", func(e *colly.HTMLElement) {
		classAttrs := []string{"e31", "e32"}
		for _, attr := range classAttrs {
			condition := fmt.Sprintf(`[class="%s"]`, attr)
			e.DOM.Find(condition).Each(func(i int, s *goquery.Selection) {
				url, _ := s.Find("a").Attr("href")
				url = baseURL + url
				bankCode := strings.TrimSuffix(url, "/")
				kanjiName := s.Text()
				bank := JPBank{
					KanjiName: kanjiName,
					BankCode:  bankCode,
				}

				itemCh <- bank
				branchListCollector.Visit(url)
			})
		}
	})

	branchListCollector.OnHTML("tr", func(e *colly.HTMLElement) {
		classAttrs := []string{"e31", "e32"}
		for _, attr := range classAttrs {
			condition := fmt.Sprintf(`[class="%s"]`, attr)
			e.DOM.Find(condition).Each(func(i int, s *goquery.Selection) {
				url, _ := s.Find("a").Attr("href")
				log.Printf("Parse branch url %s\n", url)
				branchDetailCollector.Visit(url)
			})
		}
	})

	branchDetailCollector.OnHTML(`table[class="tbl1 ft11"]`, func(e *colly.HTMLElement) {
		branch := JPBankBranch{}

		e.DOM.Find("tr").Each(func(i int, s *goquery.Selection) {
			field := s.Find(".b53").Text()
			value := s.Find(".b54").Text()

			setBranchAttribute(field, value, &branch)
		})

		itemCh <- branch
	})

	c.Visit(baseURL)
	close(itemCh)
	<-doneCh
}
