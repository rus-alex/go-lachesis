# State store: Flattened vs MPT 


## Pure DB benchmark:

`$ go test -benchmem -bench=BenchmarkPureDB ./gossip`
```
BenchmarkPureDB/MemoryDB-3                 57889             20900 ns/op            1488 B/op         17 allocs/op
BenchmarkPureDB/LevelDB-3                  35722             34919 ns/op            1893 B/op         21 allocs/op
```
Memory is 41% faster than disk.


## SSTORE/SLOAD benchmark:

`$ go test -bench=BenchmarkStateDB -benchmem ./gossip`
```
BenchmarkStateDB/MemoryDB/OverKVDB-3               37676             33396 ns/op            2584 B/op         32 allocs/op
BenchmarkStateDB/MemoryDB/OverMPT-3                24478             49286 ns/op            5785 B/op         59 allocs/op
BenchmarkStateDB/LevelDB/OverKVDB-3                25848             46883 ns/op            3002 B/op         37 allocs/op
BenchmarkStateDB/LevelDB/OverMPT-3                 25093             46714 ns/op            5833 B/op         60 allocs/op
BenchmarkStateDB/CachedLevelDB/OverKVDB-3          35586             32842 ns/op            2711 B/op         36 allocs/op
BenchmarkStateDB/CachedLevelDB/OverMPT-3           25064             46785 ns/op            5785 B/op         59 allocs/op
```
SSTORE/SLOAD over flattened KVDB is 30% faster than over MPT (obvious because MPT is over KVDB).


## StateDB with Ballot contract benchmark:

`$ go test -bench=BenchmarkStateDbWithBallot -benchmem ./gossip`
```
BenchmarkStateDbWithBallot/overMPT-3                 561           2356892 ns/op          103571 B/op        760 allocs/op
BenchmarkStateDbWithBallot/overKVDB-3                596           2316528 ns/op          102279 B/op        751 allocs/op
```
Using flattened state db instead of MPT gets faster about 3%.


## Real node benchmark:

### go-lachesis import of 20 epoches:

 * 11:52 - MPT (branch develop2)
 * 11:43 - no MPT (branch flatten-evm-storage)

### How to reproduce:

1. Export events:
```
git checkout flatten-evm-storage
N=3 ./docker/local-start.sh
go run ./cmd/tx-storm
sleep 3600
./docker/local-stop.sh
go run ./cmd/lachesis --fakenet=1/3,docker/test_accs.json --datadir=docker/.lachesis0 export fakenet.events 0 20
```

2. Import events:
```
date >> import.log; go run ./cmd/lachesis --fakenet=1/3,docker/test_accs.json --datadir=111 import fakenet.events; date >> import.log

rm -fr 111
git checkout develop2
date >> import.log; go run ./cmd/lachesis --fakenet=1/3,docker/test_accs.json --datadir=111 import fakenet.events; date >> import.log
```

3. Timing:
```
cat import.log
```
