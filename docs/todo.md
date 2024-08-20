# TODOs and tasks

### Recommendations

* On startup, modify downloads to use database stored, last modified value.
* On startup, run magic numbers on all records to replace the current value in database, lookup `application/octet-stream` or `application/...` and empty vales.
* If artifact is a text file displayed in readme, then delete image preview, these are often super long, large and not needed.

* After a successful demozoo/pouet upload, defer a sync for the data to the artifact record.

* Handle magazines title on artifact page, http://localhost:1323/f/a55ed, this is hard to read, "Issue 4\nThe Pirate Syndicate +\nThe Pirate World"

* Use DigitalOcean API to display Estimated Droplet Transfer Pool usage and remaining balance. 
		https://pkg.go.dev/github.com/digitalocean/godo https://docs.digitalocean.com/reference/api/api-reference/#operation/droplets_get

- [ ] - http://www.platohistory.org/
- [ ] - https://portcommodore.com/dokuwiki/doku.php?id=larry:comp:bbs:about_cbbs
- [ ] - 8BBS https://everything2.com/title/8BBS

### Locations

#### On startup tasks

 - `server.go` 
 - - checks()
 - - repairs() ~ 
 - - repairDatabase() ~ /model/fix/fix.go ~ Artifacts.Run()

### Live go.defacto2.net issues.

- [ ] Missing screenshots and downloads: https://go.defacto2.net/g/millennium-ftp

### Bug fixes

### Emulate on startup fixes

- [ ] Repack zips that contain programs with bad filenames, for example: http://localhost:1323/f/ab252e4

### Templates

- [ ] `artifactfile.tmpl`
- [ ] `layoutjs.tmpl`
 
---

#### Upload tests

24 June.

- [X] Demozoo Prod.
- [X] Demozoo Graphic.
- [X] Intros.
- [X] Trainer.
- [X] Installer.
- [X] PC and Amiga text.
- [X] Image brand, logo or proof.
- [X] Text, DOS and Windows magazines.
- [X] Advanced.

#### Software libraries

#####  Scheduling library for Go

- [gocron](https://github.com/go-co-op/gocron)
- [Go Quartz](https://github.com/reugn/go-quartz)
