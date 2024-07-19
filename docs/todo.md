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
- [ ] Milestone: first arcade game and mention plato network.
- [ ] - https://web.archive.org/web/20050828232530/http://www.thinkofit.com/plato/dwplato.htm
- [ ] - http://crpgaddict.blogspot.com/2021/06/brief-everything-we-know-about-1970s.html
- [ ] - https://web.archive.org/web/20130719204422/http://classicgaming.gamespy.com/View.php?view=Articles.Detail&id=323
- [ ] - https://www.uvlist.net/platforms/games-list/181
- [ ] - https://www.uvlist.net/game-168996-Airfight
- [ ] - https://arstechnica.com/gaming/2016/10/want-to-see-gamings-past-and-future-dive-into-the-educational-world-of-plato/2/
- [ ] - https://en.wikipedia.org/wiki/Massively_multiplayer_online_role-playing_game
- [ ] - https://web.archive.org/web/20021003215012/http://www.geocities.com/jim_bowery/empire2.html
- [ ] - https://digital.library.illinois.edu/items/fb011d70-4969-0135-3563-0050569601ca-f
- [ ] - https://heavyharmonies.com/Avatar/
- [ ] - https://livinginternet.com/r/ri_talk.htm
- [ ] - http://web.archive.org/web/20071203093009/http://www.geocities.com/jim_bowery/spasim.html
- [ ] - https://web.archive.org/web/20090418205738/http://plato.filmteknik.com/
- [ ] - https://web.archive.org/web/20170802041732fw_/http://home.earthlink.net/~gosamgosamgo/plato1.htm
- [ ] - https://www.atarimagazines.com/v3n3/platorising.php
- [ ] - http://www.daleske.com/plato/empire.php ***
- [ ] - https://en.wikipedia.org/wiki/TUTOR
- [ ] - https://distributedmuseum.illinois.edu/exhibit/tutor-programming-language/
- [ ] - http://www.platohistory.org/
- [ ] - https://www.youtube.com/@Cyber1Plato
- [ ] - https://gamingsubculture.weebly.com/empire.html
- [ ] - https://zeitgame.net/archives/2909
- [ ] - https://grainger.illinois.edu/news/magazine/plato
- [ ] - Empire https://en.wikipedia.org/wiki/Empire_(1973_video_game)
- [ ] - first single player arcade games, late 1974, https://en.wikipedia.org/wiki/Speed_Race, https://en.wikipedia.org/wiki/Qwak!
- [ ] - 8BBS https://everything2.com/title/8BBS

Atari 2600 160 x 192 pixels, 

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