package main

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

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

func updateAvatarIcon() {
	var users []User

	serr := db.Select(&users, "SELECT id, avatar_icon FROM user")

	if serr != nil {
		fmt.Println("select user err:", serr)
		return
	}

	for _, user := range users {
		fmt.Printf("Processing %d\n", user.ID)

		dotPos := strings.LastIndexByte(user.AvatarIcon, '.')
		new_name := fmt.Sprintf("%d%s", user.ID, user.AvatarIcon[dotPos:])

		fmt.Printf("New name: %s\n", new_name)

		_, err := db.Exec("UPDATE user SET avatar_icon = ? WHERE id = ?", new_name, user.ID)
		if err != nil {
			return
		}
	}
}

func main() {
	updateAvatarIcon();
}
