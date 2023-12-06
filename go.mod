module example

go 1.21

toolchain go1.21.0

replace rqcache => ./rqcache

require rqcache v0.0.0-00010101000000-000000000000

require (
	github.com/golang/protobuf v1.5.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)
