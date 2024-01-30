package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"

	"github.com/go-sql-driver/mysql"
)

func GetConnector(db_name string) (*sql.DB, error) {
	os.Clearenv()

	dbPath := "./Config/" + db_name + "Config.env"
	fmt.Println(dbPath)
	err := godotenv.Load(dbPath)
	if err != nil {
		return nil, errors.Wrap(err, "Error loading DB config file: "+err.Error())
	}
	fmt.Println(os.Getenv("DB_NAME"))
	cfg := mysql.Config{
		User:                 os.Getenv("DB_USER"),
		Passwd:               os.Getenv("DB_PASSWORD"),
		Net:                  os.Getenv("DB_NETWORK"),
		Addr:                 os.Getenv("DB_ADDRESS"),
		Collation:            "utf8mb4_general_ci",
		Loc:                  time.UTC,
		MaxAllowedPacket:     4 << 20.,
		AllowNativePasswords: true,
		CheckConnLiveness:    true,
		DBName:               os.Getenv("DB_NAME"),
	}
	os.Clearenv()
	connector, err := mysql.NewConnector(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "mysql.NewConnector failed")
	}
	db := sql.OpenDB(connector)
	return db, err
}
