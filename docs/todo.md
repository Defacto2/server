# TODOs and tasks

### Ideas or doubleful tasks

### Recommendations

- [ ] Use DigitalOcean API to display Estimated Droplet Transfer Pool usage and remaining balance. 
		https://pkg.go.dev/github.com/digitalocean/godo https://docs.digitalocean.com/reference/api/api-reference/#operation/droplets_get
- [ ] After a successful demozoo/pouet upload, defer a sync for the data to the artifact record.

### Bug fixes

- [ ] http://localhost:1323/f/a61e5a2 ~ `error: could not link group`
- [ ] http://localhost:1323/g/defacto2.net - Something crashed! 
- -   500 internal error for the URL, "releasers page for,defacto2.net": namer.Humanize: the path contains invalid characters vs https://defacto2.net/g/defacto2net
- -   todo: do a manual fix, use `defacto2net` but display _Defacto2 .net_ or _Defacto2 website_ in the UI.

### Emulate on startup fixes

- [ ] Delete emulate no aspect ratio from the database, as it isn't used, `dosee_no_aspect_ratio_fix`.
- [ ] Repack zips that contain programs with bad filenames, for example: http://localhost:1323/f/ab252e4
- [ ] Create a DRIVER.ZIP archive to load in js-dos, it should contain GUS drivers and DOS4GW.EXE etc.

> `dosee_load_utilities`

```js  
    Dos(canvas).ready((fs, main) => {
        fs.extractAll([
            { url: "ultrsound.zip", mountPoint: "/snd" },
            { url: "game.zip", mountPoint: "/game" },
        }).then(() => {
            main(["-c", "cd game", "-c", "game.exe"])
        });
    });

	Dos(canvas).ready((fs, main) => {
        fs.createFile("dosbox.conf", `
            [autoexec]
            mount c .
            c:
            type dosbox~1.con
        `);
        main(["-conf", "dosbox.conf"]);
    });
```

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