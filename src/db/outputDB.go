package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/pkg/errors"
)

type DBConfig struct {
	User     string
	Password string
	Network  string
	Address  string
	DBName   string
}

type NameServerData struct {
	SearchID    int
	NameServer  string
	IP          string
	CountryCode string
	IPType      int
}

type WebIpData struct {
	SearchID    int
	IP          string
	CountryCode string
}

type URLSearchData struct {
	URL    string
	URLCRC int64
}

func CreateTablesIfNotExists(db *sql.DB) error {
	err := CreateNameServerTableIfNotExists(db)
	if err != nil {
		return err
	}
	err = CreateWEbIpTableIfNotExists(db)
	if err != nil {
		return err
	}
	err = CreateUrlSearchTableIfNotExists(db)
	if err != nil {
		return err
	}
	return nil
}

func tableExists(db *sql.DB, tableName string) bool {
	query := fmt.Sprintf("SHOW TABLES LIKE '%s'", tableName)
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	return rows.Next()
}

func CreateNameServerTableIfNotExists(db *sql.DB) error {
	if tableExists(db, "tb_name_server") {
		return nil
	}
	// 테이블 생성 쿼리
	createTableQuery := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			ns_id INT AUTO_INCREMENT PRIMARY KEY,
			search_id INT,
			name_server VARCHAR(255),
			ip VARCHAR(30),
			country_code VARCHAR(2),
			ip_type INT
            insert_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`, "tb_name_server")

	// 테이블 생성
	_, err := db.Exec(createTableQuery)
	if err != nil {
		return errors.Wrap(err, "CreateNameServerTableQuery failed")
	}

	fmt.Printf("Table tb_name_server created successfully.\n")
	return nil
}

func CreateWEbIpTableIfNotExists(db *sql.DB) error {
	if tableExists(db, "tb_web_ip") {
		return nil
	}
	// 테이블 생성 쿼리
	createTableQuery := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			web_ip_id INT AUTO_INCREMENT PRIMARY KEY,
			search_id INT,
			ip VARCHAR(30),
			country_code VARCHAR(2)
            insert_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`, "tb_web_ip")

	// 테이블 생성
	_, err := db.Exec(createTableQuery)
	if err != nil {
		return errors.Wrap(err, "CreateWEbIpTableQuery failed")
	}

	fmt.Printf("Table tb_web_ip created successfully.\n")
	return nil
}

func CreateUrlSearchTableIfNotExists(db *sql.DB) error {
	if tableExists(db, "tb_url_search") {
		return nil
	}
	// 테이블 생성 쿼리
	createTableQuery := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			search_id INT AUTO_INCREMENT PRIMARY KEY,
			url VARCHAR(255),
			url_crc BIGINT,
			insert_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`, "tb_url_search")

	// 테이블 생성
	_, err := db.Exec(createTableQuery)
	if err != nil {
		return errors.Wrap(err, "CreateUrlSearchTableQuery failed")
	}

	fmt.Printf("Table tb_url_search created successfully.\n")
	return nil
}

func InsertURLSearchDataIntoTable(db *sql.DB, data URLSearchData) (int, error) {
	// 데이터 삽입 쿼리
	insertQuery := fmt.Sprintf(`
		INSERT INTO %s (url, url_crc) 
		VALUES (?, ?)`, "tb_url_search")

	// 데이터 삽입
	result, err := db.Exec(insertQuery, data.URL, data.URLCRC)
	if err != nil {
		return 0, errors.Wrap(err, "Insert URLSearchData failed")
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, "LastInsertId failed")
	}

	return int(lastInsertID), nil
}

func InsertWebIPDataIntoTable(db *sql.DB, data WebIpData) error {
	// 데이터 삽입 쿼리
	insertQuery := fmt.Sprintf(`
		INSERT INTO %s (search_id, ip, country_code) 
		VALUES (?, ?, ?)`, "tb_web_ip")

	// 데이터 삽입
	_, err := db.Exec(insertQuery, data.SearchID, data.IP, data.CountryCode)
	if err != nil {
		return errors.Wrap(err, "Insert WebIpData failed")
	}

	return nil
}

func InsertNameServerDataIntoTable(db *sql.DB, data NameServerData) error {
	// 데이터 삽입 쿼리
	insertQuery := fmt.Sprintf(`
		INSERT INTO %s (search_id, name_server, ip, country_code, ip_type) 
		VALUES (?, ?, ?, ?, ?)`, "tb_name_server")

	// 데이터 삽입
	_, err := db.Exec(insertQuery, data.SearchID, data.NameServer, data.IP, data.CountryCode, data.IPType)
	if err != nil {
		return errors.Wrap(err, "Insert NameServerData failed")
	}

	return nil
}

func GetDBConnect(db_name string) (*sql.DB, error) {
	db, err := GetConnector(db_name)
	if err != nil {
		return nil, errors.Wrap(err, "GetConnector failed")
	}

	err = CreateTablesIfNotExists(db)
	if err != nil {
		return nil, err
	}
	return db, nil
}
