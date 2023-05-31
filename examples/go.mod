module example

go 1.20

require (
	github.com/adoublef-go/version v0.0.0
	github.com/go-chi/chi/v5 v5.0.8
	github.com/stretchr/testify v1.8.4
)

replace github.com/adoublef-go/version => ../

require (
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
