module example

go 1.20

require (
	github.com/adoublef-go/version v0.0.0
	github.com/go-chi/chi/v5 v5.0.8
)

replace github.com/adoublef-go/version => ../

require github.com/Masterminds/semver/v3 v3.2.1 // indirect
