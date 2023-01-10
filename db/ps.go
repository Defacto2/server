package ps

import (
	"database/sql"
)

const (
	Name = "postgres"
)

func ConnectDB() (*sql.DB, error) {
	conn, err := sql.Open(Name, "postgres://root:example@localhost:5432/defacto2-ps?sslmode=disable")
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func Version() (string, error) {
	conn, err := ConnectDB()
	if err != nil {
		return "", err
	}
	rows, err := conn.Query("SELECT version();")
	if err != nil {
		return "", err
	}
	var s string
	for rows.Next() {
		rows.Scan(&s)
	}
	rows.Close()
	conn.Close()
	return s, nil
}

// arts
// query := Stmt(fmt.Sprintf("WHERE platform = '%s' AND section != '%s'",
//platforms.Img, tags.Bbs))

/*
const (
	Ami   Platform = "textamiga"
	Ansi  Platform = "ansi"
	Db    Platform = "database"
	Dos   Platform = "dos"
	Htm   Platform = "markup"
	Img   Platform = "image"
	Java  Platform = "java"
	Linux Platform = "linux"
	Mac   Platform = "mac10"
	Pack  Platform = "package"
	Pcb   Platform = "pcboard"
	Pdf   Platform = "pdf"
	Php   Platform = "php"
	Snd   Platform = "audio"
	Txt   Platform = "text"
	Video Platform = "video"
	Win   Platform = "windows"
)
*/
