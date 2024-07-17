# TODOs and tasks

### Ideas or doubleful tasks

### Recommendations

- [ ] Use DigitalOcean API to display Estimated Droplet Transfer Pool usage and remaining balance. 
		https://pkg.go.dev/github.com/digitalocean/godo https://docs.digitalocean.com/reference/api/api-reference/#operation/droplets_get
- [ ] After a successful demozoo/pouet upload, defer a sync for the data to the artifact record.
- [ ] Missing screenshots and downloads: https://go.defacto2.net/g/millennium-ftp
- [ ] On startup, modify downloads to use database stored, last modified value.
- [ ] On startup, run magic numbers on all records to replace the current value in database.
- [ ] All temp files should be stored in a single unified temp subdirectory, that can be purged during server restarts.
- [ ] Complete `internal/archive/archive.go to support all archive types. need to supprt legacy zip via hwzip and arc.

### Bug fixes

### Emulate on startup fixes

- [ ] Repack zips that contain programs with bad filenames, for example: http://localhost:1323/f/ab252e4

### Templates

- [ ] `artifactfile.tmpl`
- [ ] `artifactzip.tmpl`
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