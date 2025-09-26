<h1>Migrator</h1>:
Migrator flags:
--conn-str="" [connect-string to database]
--path=""     [path to migration files]
--table=""    [migration table name]
--force=0     [forceful migration to clean]
--down=true   [down migration]
--version=0   [version choose]

Migrator сommands examples:
do migrations:                     go run cmd/migrator/main.go --conn-str="postgres://postgres:0@localhost:5432/etog?sslmode=disable" --path="./migrations"
do migrations and name migr table: go run cmd/migrator/main.go --conn-str="postgres://postgres:0@localhost:5432/etog?sslmode=disable" --path="./migrations" --table="migr-table"
force dirty version:               go run cmd/migrator/main.go --conn-str="postgres://postgres:0@localhost:5432/etog?sslmode=disable" --path="./migrations" --force=1
down to version:                   go run cmd/migrator/main.go --conn-str="postgres://postgres:0@localhost:5432/etog?sslmode=disable" --path="./migrations" --version=1
down for x steps:                  go run cmd/migrator/main.go --conn-str="postgres://postgres:0@localhost:5432/etog?sslmode=disable" --path="./migrations" --down=true --steps=1
