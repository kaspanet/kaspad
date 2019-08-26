package main

import (
	"fmt"
	"github.com/daglabs/btcd/apiserver/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func main() {
	db, err := gorm.Open("mysql", "root:1@tcp(localhost:3306)/apiserver?parseTime=true")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()
	var block models.Block
	db.Preload("AcceptingBlock").Preload("ParentBlocks").First(&block, 2)
	fmt.Println("Hi")
	fmt.Println(block)
	fmt.Println("Wazzup")
}
