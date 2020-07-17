#!/bin/bash

sleep 20

lotus wait-api

lotus chain head

export MAIN=$(cat ../localnet.json | jq -r '.Accounts | .[0] | .Meta .Owner')

export ROOT1=t1macdda4f5osjhobtg6ycqbdlomsedjt2zijdtuy
export ROOT2=t1dce6ymzvvl3m3nk3o74ztkg4izmvzarky3uqtza

# Send funds to root key
lotus send --source $MAIN $ROOT2 5000000
lotus send --source $MAIN $ROOT1 5000000

export VERIFIER=$(lotus wallet new)
export CLIENT=$(lotus wallet new)

# Send funds to verifier
lotus send --source $MAIN $VERIFIER 5000000

# Send funds to client
lotus send --source $MAIN $CLIENT 5000000

while [ "5000000 FIL" != "$(lotus wallet balance $ROOT1)" ]
do
 sleep 1
 lotus wallet balance $ROOT1
done


# lotus-shed verifreg add-verifier t080 100000000000000000000000000000000000000000
# lotus-shed verifreg add-verifier t1fj2s6phuwkn32t3ocilhcpd2vwuu2zdcngdcqhy 100000000000000000000000000000000000000000

export PARAM=$(lotus-shed verifreg add-verifier --dry t01003 100000000000000000000000000000000000000000)

# lotus-shed verifreg add-verifier --from $ROOT t01003 100000000000000000000000000000000000000000
lotus msig propose --source $ROOT1 t0101 t06 0 2 $PARAM
lotus msig inspect t0101

lotus msig approve --source $ROOT2 t0101 0 $ROOT1 t06 0 2 $PARAM

lotus-shed verifreg list-verifiers

lotus-shed verifreg verify-client --from $VERIFIER $CLIENT 10000000000000000000000000000000000000000

lotus-shed verifreg list-clients

export DATA=$(lotus client import dddd | awk '{print $NF}')

lotus client local

lotus client deal --verified-deal --from $CLIENT $DATA t01000 0.005 100000

while [ "3" != "$(lotus-storage-miner sectors list | wc -l)" ]
do
 sleep 10
 lotus-storage-miner sectors list
done

# hmm
curl -H "Content-Type: application/json" -H "Authorization: Bearer $(cat ~/.lotusstorage/token)" -d '{"id": 1, "method": "Filecoin.SectorStartSealing", "params": [2]}' localhost:2345/rpc/v0

lotus-storage-miner info

lotus-storage-miner sectors list

while [ "3" != "$(lotus-storage-miner sectors list | grep Proving | wc -l)" ]
do
 sleep 5
 lotus-storage-miner sectors list | tail -n 1
 lotus-storage-miner info | grep "Actual Power"
done

sleep 300000
