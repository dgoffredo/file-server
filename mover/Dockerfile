from golang:1.20rc1-alpine3.17 as build

workdir /usr/local/src/mover
copy main.go /usr/local/src/mover/main.go
run go build -v -o /usr/local/bin/mover /usr/local/src/mover/main.go

from alpine:3.17
copy --from=build /usr/local/bin/mover /usr/local/bin/mover
cmd ["/usr/local/bin/mover"]
