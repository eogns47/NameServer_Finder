package db

import (
	"database/sql"

	"github.com/pkg/errors"
)

type RootDomain struct {
	RootID        int
	RootDomain    string
	RootDomainCRC int64
	LastWhoisTime sql.NullString
	InsertTime    sql.NullString
	UpdateTime    sql.NullString
}

type URLData struct {
	URLId  int
	URL    string
	URLCRC int64
}

func ReadDomainTable(db *sql.DB, tableName string) ([]URLData, error) {
	// tb_root_domain의 모든 데이터를 읽고, 그 중 root_id와 roo_domain, root_domain_crc를 읽어서 출력
	query := "SELECT * FROM " + tableName + ";"
	rows, err := db.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, "Read Table "+tableName+" failed")
	}
	defer rows.Close()

	var urlDatas []URLData

	for rows.Next() {
		var rootDomain RootDomain

		if err := rows.Scan(&rootDomain.RootID, &rootDomain.RootDomain, &rootDomain.RootDomainCRC, &rootDomain.LastWhoisTime, &rootDomain.InsertTime, &rootDomain.UpdateTime); err != nil {
			return nil, errors.Wrap(err, "Table attribute must be matched with int, string, int64, string, string, string")
		}

		urlData := URLData{URLId: rootDomain.RootID, URL: rootDomain.RootDomain, URLCRC: rootDomain.RootDomainCRC}

		urlDatas = append(urlDatas, urlData)
	}

	return urlDatas, nil
}
