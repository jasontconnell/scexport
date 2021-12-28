module github.com/jasontconnell/scexport

go 1.17

require (
	github.com/google/uuid v1.1.1
	github.com/jasontconnell/conf v1.1.0
	github.com/jasontconnell/sitecore v1.3.2
)

replace github.com/jasontconnell/sitecore => ../sitecore

require (
	github.com/denisenkom/go-mssqldb v0.0.0-20200620013148-b91950f658ec // indirect
	github.com/golang-sql/civil v0.0.0-20190719163853-cb61b32ac6fe // indirect
	github.com/jasontconnell/sqlhelp v1.0.0 // indirect
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de // indirect
)
