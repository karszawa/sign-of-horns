package main

import (
	crand "crypto/rand"
	"database/sql"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
	"bytes"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
)

var (
	db            *sqlx.DB
)

func init() {
	seedBuf := make([]byte, 8)
	crand.Read(seedBuf)
	rand.Seed(int64(binary.LittleEndian.Uint64(seedBuf)))

	db_host := os.Getenv("ISUBATA_DB_HOST")
	if db_host == "" {
		db_host = "127.0.0.1"
	}
	db_port := os.Getenv("ISUBATA_DB_PORT")
	if db_port == "" {
		db_port = "3306"
	}
	db_user := os.Getenv("ISUBATA_DB_USER")
	if db_user == "" {
		db_user = "root"
	}
	db_password := os.Getenv("ISUBATA_DB_PASSWORD")
	if db_password != "" {
		db_password = ":" + db_password
	}

	dsn := fmt.Sprintf("%s%s@tcp(%s:%s)/isubata?parseTime=true&loc=Local&charset=utf8mb4",
		db_user, db_password, db_host, db_port)

	fmt.Printf("Connecting to db: %q", dsn)
	db, _ = sqlx.Open("mysql", dsn)

	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(5 * time.Minute)
	log.Printf("Succeeded to connect db.")
}

type User struct {
	ID          int64     `json:"-" db:"id"`
	Name        string    `json:"name" db:"name"`
	Salt        string    `json:"-" db:"salt"`
	Password    string    `json:"-" db:"password"`
	DisplayName string    `json:"display_name" db:"display_name"`
	AvatarIcon  string    `json:"avatar_icon" db:"avatar_icon"`
	CreatedAt   time.Time `json:"-" db:"created_at"`
}

func exportImages() {
	var users []User

	serr := db.Select(&users, "SELECT id, avatar_icon FROM user")

	if serr != nil {
		fmt.Println("select user err:", serr)
		return
	}

	for _, user := range users {
		fmt.Printf("Processing %d\n", user.ID)
		var name string
		var data []byte

		err := db.QueryRow("SELECT name, data FROM image WHERE name = ?", user.AvatarIcon).Scan(&name, &data)
		if err == sql.ErrNoRows {
			fmt.Println(err)
		}

		dotPos := strings.LastIndexByte(name, '.')
		if dotPos < 0 {
			fmt.Println("Err: can not know ext")
		}

		filename := fmt.Sprintf("./public/%d%s", user.ID, name[dotPos:])

		file, ferr := os.Create(filename)

		if ferr != nil {
			fmt.Println("file create error:", ferr)
			return
		}

		buf := new(bytes.Buffer)
		err2 := binary.Write(buf, binary.BigEndian, data)
		if err2 != nil {
			fmt.Println("bynary write error:", err2)
			return
		}

		_, err3 := file.Write(buf.Bytes())
		if err3 != nil {
			fmt.Println("file write err:", err3)
			return
		}
	}

	fmt.Println("Done")
}

func main() {
	exportImages();
}
