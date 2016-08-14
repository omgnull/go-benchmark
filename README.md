# Golang benchmarks

## Byte buffers

[Bytes Buffer results](https://omgnull.github.io/go-benchmark/buffer/)

#### Buffer libs used
* [sync](https://golang.org/pkg/sync/)
* [bpool](https://github.com/oxtoacart/bpool)
* [bytebufferpool](https://github.com/valyala/bytebufferpool)


#### Launch bench
```sh
$ go test ./... -bench=. -benchmem
```

##### Buffer tests
```sh
$ cd buffer/main
$ go build
$ ./main
```

Flags:
* `duration` Test duration in seconds (default 60)
* `method` Function to run; allowed: "generic" "stack" "alloc" "sync" "bpool" "bbpool" (default "generic")
* `out` Filename to write report; Prints into stdout by default
* `queue` Number of goroutines; default is NumCPU (default 8)
