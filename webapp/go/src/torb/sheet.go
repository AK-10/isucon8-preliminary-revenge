package main

import (
	"encoding/json"
	"log"
	
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
)

const (
	sheetKey = "sheet"
)

func initSheets() error {
	rows, err := db.Query("SELECT * FROM sheets")
	if err != nil {
		return err
	}
	defer rows.Close()

	var sheets []Sheet
	for rows.Next() {
		var sheet Sheet
		if err := rows.Scan(&sheet.ID, &sheet.Rank, &sheet.Num, &sheet.Price); err != nil {
			return err
		}
		sheets = append(sheets, sheet)
	}

	setSheetsToRedis(sheets)
	return nil
}

func setSheetsToRedis(sheets []Sheet) {
	conn, err := redis.Dial("tcp", "localhost:6379")
    if err != nil {
        panic(err)
    }
	defer conn.Close()
	// primitive型以外はjson.Marshalする
	serialized, _ := json.Marshal(sheets)
	conn.Do("SET", sheetKey, serialized)
}

func appendSheet(sheet Sheet) bool {
	sheets, found := getAllSheetFromRedis()
	if !found {
		return false
	}
	sheets = append(sheets, sheet)
	setSheetsToRedis(sheets)
	return true
}

func getAllSheetFromRedis() ([]Sheet, bool) {
	items, found := getItemFromRedis(sheetKey)
	if !found {
		return nil, found
	}

	sheets, ok := items.([]Sheet)
	return sheets, ok
}

func getItemFromRedis(key string) (interface{}, bool) {
	conn, err := redis.Dial("tcp", "localhost:6379")
    if err != nil {
        panic(err)
	}
	defer conn.Close()
	bytes, err := redis.Bytes(conn.Do("GET", sheetKey))
	if err == redis.ErrNil {
		log.Println(err)
		return nil, false
	}
	if err != nil {
		log.Println(err)
		return nil, false
	}
	var deserialized interface{}
	var altDeserialized []Sheet
	json.Unmarshal(bytes, &deserialized)
	json.Unmarshal(bytes, &altDeserialized)
	log.Println(altDeserialized)
	log.Println(deserialized.([]Sheet))
	return deserialized, true
}

// key構成を頑張ってredisだけで走査する vs sliceをredisに持ってgo側で走査する
func findSheetWhere(condition func(s Sheet) bool) (Sheet, bool) {
	sheets, found := getAllSheetFromRedis()
	if !found {
		return Sheet{}, false
	}
	for _, v := range sheets {
		if condition(v) {
			return v, true
		}
	}
	return Sheet{}, false
}

// limitつけてあげた方が良い？
func findSheetsWhere(condition func(s Sheet) bool) ([]Sheet, bool) {
	sheets, found := getAllSheetFromRedis()
	if !found {
		return nil, false
	}
	var resSheets []Sheet
	for _, v := range sheets {
		if condition(v) {
			resSheets = append(resSheets, v)
		}
	}

	return resSheets, len(resSheets) > 0
}


func contains(slice []interface{}, condition func(interface{}) bool) bool {
	for _, v := range slice {
		if condition(v) {
			return true
		}
	}
	return false
}