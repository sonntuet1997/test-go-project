### Docker dev env

`docker-compose up -d`

`./terminal.sh`

###Test 1 func: quick debugging purpose
https://labix.org/gocheck

`/app/Chapter06/linkgraph/store/memory# go test -v -check.f "InMemoryGraphTestSuite.TestUpsertLink"`

exact match: `go test -check.f "\bInMemoryBleveTestSuite.TestMatchSearch\b" -v
`

###Trick 1: cache lib for fast debug by running the following command in Docker container

`go mod vendor`
