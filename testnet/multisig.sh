
export ACCOUNT1=$(lotus wallet new)
export ACCOUNT2=$(lotus wallet new)
export ACCOUNT3=$(lotus wallet new)
export ACCOUNT4=$(lotus wallet new)

export RET=$(lotus msig create $ACCOUNT1 $ACCOUNT2 $ACCOUNT3 $ACCOUNT4)

# should error earlier if no addresses

export MSIG_ADDRESS=$(echo $RET | awk 'NF>1{print $NF}')

export MAIN=$(cat ../localnet2.json | jq -r '.Accounts | .[0] | .Meta .Owner')

lotus send --source $MAIN $ACCOUNT1 5
lotus send --source $MAIN $ACCOUNT2 5
lotus send --source $MAIN $ACCOUNT3 5
lotus send --source $MAIN $ACCOUNT4 5
export TXID=$(lotus send --source $MAIN $MSIG_ADDRESS 5000000)

while [ "5000000" != "$(lotus wallet balance $MSIG_ADDRESS)" ]
do
 sleep 1
 lotus wallet balance $MSIG_ADDRESS
done

lotus wallet balance $MSIG_ADDRESS

export PARAM=$(lotus-shed verifreg add-verifier --dry t01001 100000000000000000000000000000000000000000)

lotus msig propose --source $ACCOUNT1 $MSIG_ADDRESS t080 0 2 $PARAM

lotus msig approve --source $ACCOUNT2 $MSIG_ADDRESS 2 $ACCOUNT1 t080 0 2 $PARAM
lotus msig approve --source $ACCOUNT3 $MSIG_ADDRESS 2 $ACCOUNT1 t080 0 2 $PARAM
lotus msig approve --source $ACCOUNT4 $MSIG_ADDRESS 2 $ACCOUNT1 t080 0 2 $PARAM

