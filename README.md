<h1>Migrator:</h1>
<h2>Migrator flags:</h2>
<ul>
  <li>--conn-str="" [connect-string to database]</li>
  <li>--path=""     [path to migration files]</li>
  <li>--table=""    [migration table name]</li>
  <li>--force=0     [forceful migration to clean]</li>
  <li>--down=true   [down migration]</li>
  <li>--version=0   [version choose]</li>
</ul>

<h2>Migrator сommands examples:</h2>
<ul>
  <li>do migrations:                     go run cmd/migrator/main.go --conn-str="postgres://postgres:0@localhost:5432/etog?sslmode=disable" --path="./migrations"</li>
  <li>do migrations and name migr table: go run cmd/migrator/main.go --conn-str="postgres://postgres:0@localhost:5432/etog?sslmode=disable" --path="./migrations" --table="migr-table"</li>
  <li>force dirty version:               go run cmd/migrator/main.go --conn-str="postgres://postgres:0@localhost:5432/etog?sslmode=disable" --path="./migrations" --force=1</li>
  <li>down to version:                   go run cmd/migrator/main.go --conn-str="postgres://postgres:0@localhost:5432/etog?sslmode=disable" --path="./migrations" --version=1</li>
  <li>down for x steps:                  go run cmd/migrator/main.go --conn-str="postgres://postgres:0@localhost:5432/etog?sslmode=disable" --path="./migrations" --down=true --steps=1</li>
</ul>
