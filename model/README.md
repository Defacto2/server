# placeholder code and info for model data interactions

Example API in Echo.

// type ExampleRequest struct {
// 	FileName string `json:"filename" query:"filename"`
// }

// https://blog.boatswain.io/post/handling-http-request-in-go-echo-framework-1/

```go

	// TODO remove
	got, err := models.Files().One(ctx, db)
	if err != nil {
		log.DPanic(err)
	}
	fmt.Println("got record:", got.ID)
	f := &models.File{ID: 1}
	found, err := models.FindFile(ctx, db, f.ID)
	if err != nil {
		log.DPanic(err)
	}
	fmt.Println("found record:", found.Filename)
	foundAgain, err := models.Files(qm.Where("id = ?", 1)).One(ctx, db)
	if err != nil {
		log.DPanic(err)
	}
	fmt.Println("found again record:", foundAgain.Filename)
	exists, err := models.FileExists(ctx, db, foundAgain.ID)
	if err != nil {
		log.DPanic(err)
	}
	fmt.Println("found again exists?:", exists)
	count, err := models.Files(qm.WithDeleted()).Count(ctx, db)
	if err != nil {
		log.DPanic(err)
	}
	countPub, err := models.Files().Count(ctx, db)
	if err != nil {
		log.DPanic(err)
	}
	fmt.Println("total files vs total files that are public:", count, "vs", countPub)

	var users []*models.File
	rows, err := db.QueryContext(context.Background(), `select * from files;`)
	if err != nil {
		log.Fatal(err)
	}
	err = queries.Bind(rows, &users)
	if err != nil {
		log.Fatal(err)
	}
	rows.Close()

	users = nil
	err = models.Files(qm.SQL(`select * from files;`)).Bind(ctx, db, &users)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print("raw files:")
	for i, u := range users {
		if i != 0 {
			fmt.Print(", ")
		}
		fmt.Print(u.Filename.String)
	}
	fmt.Println()
```