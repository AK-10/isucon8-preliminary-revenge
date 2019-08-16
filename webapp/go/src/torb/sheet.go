package main

import (
	"encoding/json"
	
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

func appendSheet(sheet Sheet) {
	sheets := getAllSheetFromRedis()
	sheets = append(sheets, sheet)
	setSheetsToRedis(sheets)
}

func getAllSheetFromRedis() []Sheet {
	conn, err := redis.Dial("tcp", "localhost:6379")
    if err != nil {
        panic(err)
    }
	defer conn.Close()
	bytes, _ := redis.Bytes(conn.Do("GET", sheetKey))
	var deserialized []Sheet
	json.Unmarshal(bytes, deserialized)
	return deserialized
}

// key構成を頑張ってredisだけで走査する vs sliceをredisに持ってgo側で走査する
func findSheetWhere(condition func(s Sheet) bool) (bool, Sheet) {
	sheets := getAllSheetFromRedis()
	for _, v := range sheets {
		if condition(v) {
			return true, v
		}
	}
	return false, Sheet{}
}

// limitつけてあげた方が良い？
func findSheetsWhere(condition func(s Sheet) bool) []Sheet {
	sheets := getAllSheetFromRedis()
	var resSheets []Sheet
	for _, v := range sheets {
		if condition(v) {
			resSheets = append(resSheets, v)
		}
	}

	return resSheets
}

func getSheetById(id int64) *Sheet {
	sheets := getAllSheetFromRedis()
	for _, v := range sheets {
		if v.ID == id {
			return &v
		}
	}
	return nil
}


func contains(slice []interface{}, condition func(interface{}) bool) bool {
	for _, v := range slice {
		if condition(v) {
			return true
		}
	}
	return false
}