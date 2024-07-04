# TODOs and tasks

### Ideas or doubleful tasks

- [ ] Implememnt a scheduling library for Go
- - [gocron](https://github.com/go-co-op/gocron)
- - [Go Quartz](https://github.com/reugn/go-quartz)
- [ ] Print the running user and group on startup.
- [ ] When running rename to extras, if first error is a permission, abort the range track.
	  `defer repair file rename: rename /mnt/volume_sfo3_01/assets/downloads/c709d160-c0ed-4acc-8d4b-327eaa1d344e.txt /mnt/volume_sfo3_01/assets/extra/c709d160-c0ed-4acc-8d4b-327eaa1d344e.txt: permission denied`
- [ ] Enforce 664 permissions on all files within the assets directories.
- [ ] http://localhost:1323/f/ac1946e not working is Cascadia Code PL.


### Recommendations

- [X] _Unique ID_ and _Record Key_ buttons should copy the value to the clipboard.
- [ ] Use DigitalOcean API to display Estimated Droplet Transfer Pool usage and remaining balance. 
		https://pkg.go.dev/github.com/digitalocean/godo https://docs.digitalocean.com/reference/api/api-reference/#operation/droplets_get
- [ ] js-dos doesn't yet support `extras` zipped files.
- [ ] After a successful demozoo/pouet upload, defer a sync for the data to the artifact record.
- [ ] Find all uuid zip files that use legacy zip encodings, EXPLODE, UNSHRINK, UNREDUCE.
- [ ] Find all msdos apps that use incompatible archives such as LHA, ARJ, ARC.

### Bug fixes

- [ ] htmx search for id or uuid, incomplete (less than 32 chars) uuid should not trigger a search.
- [ ] Reader is not displaying the textfile. `/f/b62716c`
- [ ] Change helper.CookieStore behavour when no value is used, the randomized value must be ASCII compatible, as UTF-8 strings break the noice value.

`https://accounts.google.com/gsi/select?client_id=885513036389-n4uee89egjaph948pbpg7qcesf00gi0g.apps.googleusercontent.com&ux_mode=popup&ui_mode=card&context=signin&nonce=7GkgLTBJRM3rTdqcxsyGXzpZMy7ywR&as=bsyE0ir2A6e3m1Tvc6RPow&channel_id=0a45e0ae69ad153f5661f8e91886c93f928c53c60face22f9147467de6a13d85&origin=https%3A%2F%2Fgo.defacto2.net`

`https://accounts.google.com/v3/signin/identifier?continue=https%3A%2F%2Faccounts.google.com%2Fgsi%2Fselect%3Fclient_id%3D885513036389-n4uee89egjaph948pbpg7qcesf00gi0g.apps.googleusercontent.com%26ux_mode%3Dpopup%26ui_mode%3Dcard%26context%3Dsignin%26nonce%3DsomeRandomValueGoesHere%26as%3DMhdMzsSQuy5xMRNsba5tRw%26channel_id%3D2c14e46a58d9683df03ad5d98a4fb5c9304025771a6932baa78734c1224e58f5%26origin%3Dhttp%3A%2F%2Flocalhost%3A1323&faa=1&ifkv=AS5LTAQL7kn6em8sTJMmfggtQ9NiP1kCBkA6-WUzKcGerVAo2t6x8Dlg1V6DDzk2RbeBDuD213yRDw&flowName=GlifWebSignIn&flowEntry=ServiceLogin&dsh=S671648061%3A1720053481879346&ddm=0`

### Templates



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
