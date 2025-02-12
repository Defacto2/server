# Paths to infrequent functions

#### The `server fix` command

`internal/config/fixer.go` 

- Config.Fixer()

##### Fix the database

`model/fix/fix.go`

- [Repair.Run()](https://pkg.go.dev/github.com/Defacto2/server/model/fix#Repair.Run)

##### Fix assets such as generated texts, images, etc.

`internal/config/repair.go`

- Config.RepairAssets()
- Config.TextFiles() 