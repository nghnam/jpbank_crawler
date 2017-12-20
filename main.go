package main

import (
	"database/sql"
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
	jpBankPipeline := NewJPBankPipeline(db, itemCh)

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
				bankCode := strings.TrimSuffix(url, "/")
				url = baseURL + url
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
		count := 0
		e.DOM.Find("tr").Each(func(i int, s *goquery.Selection) {
			field := s.Find(".b53").Text()
			value := s.Find(".b54").Text()

			if field == "支店名" {
				branch.KanjiName = value
			}
			if field == "フリガナ" && count == 0 {
				count++ // Kata Bank Name
			}
			if field == "フリガナ" && count == 1 {
				branch.KataName = value
			}
			if field == "金融機関コード" {
				branch.BankCode = value
			}
			if field == "支店コード" {
				branch.BranchCode = value
			}
			if field == "住所" {
				address := strings.Split(value, " ")[0]
				branch.Address = sql.NullString{
					String: address,
					Valid:  true,
				}
			}
			if field == "電話番号" {
				branch.Phone = sql.NullString{
					String: value,
					Valid:  true,
				}
			}
		})

		itemCh <- branch
	})

	c.Visit(baseURL)

	c.Wait()
}
