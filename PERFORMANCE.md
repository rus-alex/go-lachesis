# Result:

go-lachesis import of 20 epoches:

 * 11:52 - MPT (branch develop2)
 * 11:43 - no MPT (branch flatten-evm-storage)

# How to reproduce:

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
