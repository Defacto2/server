package model_test

import (
	"testing"

	_ "github.com/lib/pq"
)

/* TODO:
Error:      	Not equal:
            	expected: "Mooñpeople"
            	actual  : "Moopeople"

				use utf8 lib to detect extended chars?
*/

func TestAllSlugs(t *testing.T) {
	// ctx := context.Background()
	// db, err := postgres.ConnectDB()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer db.Close()

	// fs, err := model.GroupList(ctx, db)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// for _, x := range fs {
	// 	og := sceners.CleanURL(x.GroupBrandFor.String)
	// 	y := model.Slug(og)
	// 	z := sceners.CleanURL(y)
	// 	assert.Equal(t, og, z, "slug is "+y)
	// }
}
