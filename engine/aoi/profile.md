```go
go test dreamcity/engine/aoi -count 1 -v -bench=. -benchmem -cpuprofile cpu.out -heapprofile heap.out
go tool pprof cpu.out

// https://golang.org/cmd/go/#hdr-Test_packages
go test dreamcity/engine/aoi -count 1 -bench Enter  -benchmem -cpuprofile cpu.out -memprofile mem.out 
go tool pprof -http=:8081 cpu.out   // web中显示
```

> if you want to run only benchmark and not all the tests of your program you can also do the following command
> `go test -bench=. -run=^$ . -cpuprofile profile.out`
> using -run=^$ means =>run all test method that the name follow the following regexp, but ^$ select no test. So it only run benchmark ;)

``https://github.com/golang/go/wiki/CompilerOptimizations#function-inlining
>Use -gcflags -m to observe the result of escape analysis and inlining decisions for the gc toolchain
>build的时候打印出逃逸分析和会内联的函数列表
> https://dave.cheney.net/2020/04/25/inlining-optimisations-in-go
> https://github.com/golang/perf/blob/master/README.md  `go get golang.org/x/perf/cmd/.... `
>`go build -gcflags -m main.go`

>playerA in Region Ri  Si  在边界做出一个动作
>playerB in Region Rj  Sj
>Si 决定这个动作的结果，将结果转发给 Sj
>Si 的职责是 模拟Pi的动作，计算各种状态，发给周围的Server Sj
>就好像发送给自己的客户端一样，用一个字段表示，hasGhost，如果有
>ghost，则各种状态，动作都需要转发给周围的server。
>同样的，当Sj 收到Pi的状态，将该状态发给Sj中关心此玩家的
> --- 翻译自<a Novel Interest Management Algorithm for Distributed Simultaions>


一个城市 15万个静态资源，30000个土地，每个土地算有4个基建，则30000*4 = 12w 。 15w+12w+3w = 30w 其中18万是不动的，12万是很少移动的。
算 1000个玩家在里面移动，每个玩家的aoi距离为100

#### 测试结果
##### 0个静态实体，1000个玩家，aoi玩家为5
Benchmark_Empty
    Benchmark_Empty: benchmark_test.go:95: enter 0 entities, time:145ns, coord count 0 
    Benchmark_Empty: benchmark_test.go:108: enter 1000 avatars, time:75.338858ms, coord count 3000 
Benchmark_Empty/move
    Benchmark_Empty/move: benchmark_test.go:118: 1000 avatars, move time:31.109037ms,
Benchmark_Empty/move-12                       90          11474942 ns/op
Benchmark_Empty/leave
    Benchmark_Empty: benchmark_test.go:129: 1000 avatars, leave time:763.447µs,

##### 10000个静态实体，1000个玩家，aoi半径为5
Benchmark_Opt10000
    Benchmark_Opt10000: benchmark_test.go:95: enter 10000 entities, time:36.289354ms, coord count 10000 
    Benchmark_Opt10000: benchmark_test.go:108: enter 1000 avatars, time:484.493482ms, coord count 13000 
Benchmark_Opt10000/move
    Benchmark_Opt10000/move: benchmark_test.go:118: 1000 avatars, move time:186.014529ms,
Benchmark_Opt10000/move-12                     6         177589785 ns/op
Benchmark_Opt10000/leave
    Benchmark_Opt10000: benchmark_test.go:129: 1000 avatars, leave time:24.162738ms,

##### 20000个静态实体，1000个玩家，aoi半径为5
Benchmark_Opt20000
    Benchmark_Opt20000: benchmark_test.go:95: enter 20164 entities, time:85.793267ms, coord count 20164 
    Benchmark_Opt20000: benchmark_test.go:108: enter 1000 avatars, time:1.485092705s, coord count 23164 
Benchmark_Opt20000/move
    Benchmark_Opt20000/move: benchmark_test.go:118: 1000 avatars, move time:524.643408ms,
Benchmark_Opt20000/move-12                     3         482789802 ns/op
Benchmark_Opt20000/leave
    Benchmark_Opt20000: benchmark_test.go:129: 1000 avatars, leave time:48.998368ms,

##### 50000个静态实体，1000个玩家，aoi半径为5
Benchmark_Opt50000
    Benchmark_Opt50000: benchmark_test.go:95: enter 50176 entities, time:333.157535ms, coord count 50176 
    Benchmark_Opt50000: benchmark_test.go:108: enter 1000 avatars, time:2.60219052s, coord count 53176 
Benchmark_Opt50000/move
    Benchmark_Opt50000/move: benchmark_test.go:118: 1000 avatars, move time:829.750681ms,
Benchmark_Opt50000/move-12                     2         823318219 ns/op
Benchmark_Opt50000/leave
    Benchmark_Opt50000: benchmark_test.go:129: 1000 avatars, leave time:52.06475ms,

##### 200000个静态实体，1000个玩家，aoi半径为5
Benchmark_Opt200000
    Benchmark_Opt200000: benchmark_test.go:95: enter 200704 entities, time:3.126838156s, coord count 200704 
    Benchmark_Opt200000: benchmark_test.go:108: enter 1000 avatars, time:6.331643664s, coord count 203704 
Benchmark_Opt200000/move
    Benchmark_Opt200000/move: benchmark_test.go:118: 1000 avatars, move time:1.748632024s,
Benchmark_Opt200000/move-12                    1        1748675261 ns/op
Benchmark_Opt200000/leave
    Benchmark_Opt200000: benchmark_test.go:129: 1000 avatars, leave time:52.156795ms,


### 209 内网测试机
加载 6w4的静态资源 耗时 "res enter aoi finish","count":64085,"consume":"12m38.30115872s"}



### TODO
1. [x] 有格子属性的，插入时加入基准点，进行插入。比如 插入A时，不直接从头插入，而是找一个已经在map中的点，插入，然后进行排序