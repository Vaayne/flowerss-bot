package model

import (
	"encoding/csv"
	"log"
	"net/http"

	_ "github.com/jinzhu/gorm/dialects/mysql" //mysql driver
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// const sqliteFulltextTable = "CREATE VIRTUAL TABLE feed_fts USING FTS5(name,bizid,description)"
const mysqlFulltextIndex = "CREATE FULLTEXT INDEX fts ON feeds(name, bizid, description)"
const wechatAccountSource = "https://raw.githubusercontent.com/hellodword/wechat-feeds/main/list.csv"

type Feed struct {
	ID          int    `gorm:"primaryKey,autoIncrement"`
	Name        string `gorm:"class:FULLTEXT"`
	Bizid       string `gorm:"uniqueIndex"`
	Description string
}

func loadData() []Feed {
	resp, err := http.Get(wechatAccountSource)
	if err != nil {
		log.Fatal("Get wechat account from origin source failed. ", err)
		return nil
	}

	defer resp.Body.Close()
	csvReader := csv.NewReader(resp.Body)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse csv file. ", err)
	}
	var feeds []Feed
	for _, record := range records[1:] {
		feed := Feed{
			// ID:          idx,
			Name:        record[0],
			Bizid:       record[1],
			Description: record[2],
		}
		feeds = append(feeds, feed)
	}
	return feeds
}

// LoadWechatAccounts 导入所有的可以订阅的公众号信息，明天更新一次
func LoadWechatAccounts() {

	feeds := loadData()
	db.AutoMigrate(&Feed{})
	db.DropTable(&Feed{})
	db.CreateTable(&Feed{})
	db.Exec(mysqlFulltextIndex)

	tx := db.Begin()
	for _, feed := range feeds {
		tx.Create(&feed)
	}
	tx.Commit()
}

func SearchWechatAccounts(keyword string) []Feed {
	keyword = "%" + keyword + "%"
	rows, err := db.Model(&Feed{}).Where("name like ? or bizid like ? or description like ? ", keyword, keyword, keyword).Rows()
	if err != nil {
		log.Fatal("Unable to search query with keyword: "+keyword+".\t", err)
	}
	defer rows.Close()

	var feeds []Feed
	var feed Feed
	for rows.Next() {
		db.ScanRows(rows, &feed)
		feeds = append(feeds, feed)
	}
	return feeds
}
