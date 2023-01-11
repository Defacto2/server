package meta

import (
	"context"
	"log"

	"github.com/bengarrett/df2023/postgres/models"

	"github.com/bengarrett/df2023/postgres"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Tag int

const (
	Announcement Tag = iota
	ANSIEditor
	AppleII
	AtariST
	BBS
	Logo
)

type URI map[Tag]string

type Name map[Tag]string

type Info map[Tag]string

type Count map[Tag]int

var URIs = URI{
	Announcement: "announcements",
	ANSIEditor:   "ansieditor",
	AppleII:      "appleii",
	AtariST:      "atarist",
	BBS:          "bbs",
	Logo:         "logo",
}

var Names = URI{
	Announcement: "Announcement",
	ANSIEditor:   "ANSI editor",
	AppleII:      "Apple II",
	AtariST:      "Atari ST",
	BBS:          "BBS",
	Logo:         "Brand art or logo",
}

var Infos = Info{
	Announcement: "Public announcements by Scene groups and organisations.",
	ANSIEditor:   "Programs that enable you to create and edit ANSI and ASCII art.",
	AppleII:      "Files pertaining to the Scene on the Apple II computer platform.",
	AtariST:      "Files pertaining to the Scene on the Atari ST computer platform.",
	BBS:          "Files pertaining to the Scene operating over telephone based BBS (Bulletin Board System) systems.",
	Logo:         "Branding logos used by scene groups and organisations.",
}

var Counts = Count{
	Announcement: 0,
	ANSIEditor:   0,
	AppleII:      0,
	AtariST:      0,
	BBS:          0,
	Logo:         0,
}

type Meta struct {
	URI   string
	Name  string
	Info  string
	Count int
}

var Categories []Meta = New()

func New() []Meta {
	var m = make([]Meta, len(URIs))
	i := -1
	for key, val := range URIs {
		i++
		count := Counts[key]
		m[i] = Meta{
			URI:   val,
			Name:  Names[key],
			Info:  Infos[key],
			Count: count,
		}
		// TODO: cache the results and move the function / cache to /models/custom.go
		// https://stackoverflow.com/questions/67788292/add-a-cache-to-a-go-function-as-if-it-were-a-static-member
		if count == 0 {
			t := key
			defer func(i int, t Tag) {
				ctx := context.Background()
				db, err := postgres.ConnectDB()
				if err != nil {
					log.Fatalln(err)
				}
				val, err := models.Files(
					Where("section = ?", URIs[t])).Count(ctx, db)
				if err != nil {
					log.Fatalln(err)
				}
				m[i].Count = int(val)
			}(i, t)
		}
	}
	return m
}
