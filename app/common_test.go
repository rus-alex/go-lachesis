package app

//go:generate mkdir -p solc
// NOTE: assumed that SFC-repo is in the same dir than lachesis-repo
// 1.0.0 (genesis)
//go:generate bash -c "cd ../../fantom-sfc && git checkout 1.0.0 && docker run --rm -v $(pwd):/src -v $(pwd)/../go-lachesis/app:/dst ethereum/solc:0.5.12 -o /dst/solc/ --optimize --optimize-runs=2000 --bin --abi --allow-paths /src/contracts --overwrite /src/contracts/upgradeability/UpgradeabilityProxy.sol"
//go:generate mkdir -p sfcproxy
//go:generate abigen --bin=./solc/UpgradeabilityProxy.bin --abi=./solc/UpgradeabilityProxy.abi --pkg=sfcproxy --type=Contract --out=sfcproxy/contract.go
// 1.1.0-rc1
//go:generate bash -c "cd ../../fantom-sfc && git checkout 1.1.0-rc1 && docker run --rm -v $(pwd):/src -v $(pwd)/../go-lachesis/app:/dst ethereum/solc:0.5.12 -o /dst/solc/ --optimize --optimize-runs=2000 --bin --abi --allow-paths /src/contracts --overwrite /src/contracts/sfc/Staker.sol"
//go:generate mkdir -p sfc110
//go:generate abigen --bin=./solc/Stakers.bin --abi=./solc/Stakers.abi --pkg=sfc110 --type=Contract --out=sfc110/contract.go
// 2.0.1-rc2
//go:generate bash -c "cd ../../fantom-sfc && git checkout 2.0.1-rc2 && docker run --rm -v $(pwd):/src -v $(pwd)/../go-lachesis/app:/dst ethereum/solc:0.5.12 -o /dst/solc/ --optimize --optimize-runs=2000 --bin --abi --allow-paths /src/contracts --overwrite /src/contracts/sfc/Staker.sol"
//go:generate mkdir -p sfc201
//go:generate abigen --bin=./solc/Stakers.bin --abi=./solc/Stakers.abi --pkg=sfc201 --type=Contract --out=sfc201/contract.go
// clean
//go:generate rm -fr ./solc

