### Requirement
- Golang version > 1.11 (using with module)

### Function
- config env
- middleware group api
- mysql
- unitest
- intergration test
- validate input
- redis
- api document

### Install dependency library
```
$ go mod init <name-project>
Ex: $ go mod init start
```

### To run 
```
$ ./start.sh
```

### Build binary
```
$ go build
```

### Run
development without integration test:
```
$ GIN_MODE=release ./start -ENV=development
```

development with integration test: (need config sqlite3: $GOPATH/bin/caro.db)
```
$ GIN_MODE=release ./start -ENV=development -TEST=true
```

production:
```
$ GIN_MODE=release ./start -ENV=production
```

auto refresh: (only use with development ENV)
```
$ gin --a 3001 main.go 
```

### Unit test
```
$ go test -v -count=1 ./...
```

### Build docker
```
$ docker build -t test-scale .
$ docker-compose --compatibility up
```

## Push code to vnpay
```
git checkout master
git pull
git checkout -b feature/sync
git push -f http://@gitsys.vnpay.local/cloud-vnpay-ci/dvtt/vndelivery/golf-cms.git feature/sync
```

## Deploy code on vnpay ()
```
git checkout master
git pull
git checkout -b release/x.x.x
git merge origin/feature/sync

// deploy trên môi trường test
git checkout feature/deploy_test
git merge release/x.x.x
git push

// deploy trên môi trường prod
git checkout master
git merge release/x.x.x
git push
```