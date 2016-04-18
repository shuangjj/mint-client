#! /bin/bash

# set the default chain id:
CHAIN_ID="mychain"

usage() {
   cat <<EOF
    This file is a complete demonstration for the following steps in the eris pipeline
    using nothing but "curl" to talk to HTTP servers and standard unix commands for processing:

     0) Install
     1) Start a chain with one validator
     2) Compile a solidity contract
     3) Create and sign transaction to deploy contract
     4) Send the transaction to the blockchain
     5) Wait for a confirmation
     6) Ensure the contract was deployed correctly
     7) Create, sign, broadcast, and wait for transaction that talks to the contract
     8) Query the contract without sending a transaction

    you are expected to:
    - be on a unix platform
    - have golang installed (https://golang.org/doc/install)
    - set \$GOPATH, set GOBIN=\$GOPATH/bin, set PATH=\$GOBIN:\$PATH
    - have jq installed (https://stedolan.github.io/jq/download)

   Usage: $0 -i -c <chain name[$CHAIN_ID]>  -h
          -i  do software installs only and exit;
          -c  set chain id, default is $CHAIN_ID
          -h  display this message
EOF
   exit 1;
}



########################################
## Step 0 - Installs		      ##
########################################
do_installs() {
    #gmp-dev package is needed because eris-keys links to it and needs <gmp.h>
    #And, jq JSON parser is used throughout (https://stedolan.github.io/jq/download)
    #These are probably the right commands and package names for a few OSs, but only yum install was tested on AWS AMI
    if [ `which yum 2>/dev/null` ]; then
      sudo yum install -y gmp-devel jq
    elif [ `which apt-get 2>/dev/null` ]; then
      sudo apt-get install libgmp3-dev jq
    elif [ `which brew 2>/dev/null` ]; then
      brew install gmp jq
    else
      echo "ERROR: unknown OS"
    fi

    # install the keys daemon for signing transactions
    go get github.com/shuangjj/eris-keys

    # install the mint-client tools
    go get github.com/shuangjj/mint-client/...

    # install the erisdb (requires branch 0.10.3)
    git clone https://github.com/shuangjj/eris-db $GOPATH/src/github.com/shuangjj/eris-db
    cd $GOPATH/src/github.com/shuangjj/eris-db
    git checkout 0.10.3
    go install ./cmd/erisdb #TODO: this installs into $GOPATH/bin/erisdb - inconsistent naming convention. Everything else has eris-..

    # install the eris-abi tool (required for formatting transactions to contracts)
    go get github.com/shuangjj/eris-abi/cmd/eris-abi

    exit 0;
}

while getopts "c:ih" opt; do
    case "$opt" in
        c) CHAIN_ID="$OPTARG";;
        i) do_installs;;
        h|*) usage;;
    esac
done


########################################
## Step 1 - Start a chain	      ##
########################################


# create genesis file and validator's private key
# and store them in ~/.eris/blockchains/$CHAIN_ID
# Expected output:
    #Generating accounts ...
    #genesis.json and priv_validator.json files saved in /home/ec2-user/.eris/blockchains/mychain
mintgen random 1 $CHAIN_ID


# create a config file
mintconfig > ~/.eris/blockchains/$CHAIN_ID/config.toml

# start the chain (pipe the log to a file)
erisdb ~/.eris/blockchains/$CHAIN_ID  &> ~/.eris/blockchains/$CHAIN_ID/log &

# start the keys server (for signing transactions)
eris-keys server  &> ~/.eris/blockchains/$CHAIN_ID/eris-keys.log &

# let everything start up
sleep 2

# import the validator's private key into the key server
# NOTE: this converts the tendermint private key format to the eris private key format
# This step will be eliminated in the near future as the validator's come to use the eris key format
ADDRESS=`mintkey eris ~/.eris/blockchains/$CHAIN_ID/priv_validator.json`

echo "OUR ADDRESS:"
echo $ADDRESS
echo ""


########################################
## Step 2 - Compile Solidity Contract ##
########################################

# simple solidity contract
read -r -d '' CODE << EOM
contract MyContract {

  function add(int a, int b) constant returns (int sum) {
        sum = a + b;
	  }
  }
EOM

# the solidity code needs to be in base64 for the compile server. And at least on RHEL Linux, base64 wraps long line unless --wrap=0 is passed
CODE64=`echo $CODE | base64 --wrap=0`

# json data for the curl request to the compile server
read -r -d '' JSON_DATA << EOM
{
	"name":"mycontract",
	"language":"sol",
	"script":"$CODE64"
}
EOM

# location of compiler
URL="https://compilers.eris.industries:8091/compile"

# compile that baby!
RESULT=`curl --silent -X POST -d "${JSON_DATA}" $URL --header "Content-Type:application/json"`

# the compile server returns the bytecode (in base64) and the abi (json)
BYTECODE=`echo $RESULT | jq .bytecode`
ABI=`echo $RESULT | jq .abi`

# trim quotes
BYTECODE="${BYTECODE%\"}"
BYTECODE="${BYTECODE#\"}"

# convert bytecode to hex
if [ "$(uname)" == "Darwin" ]; then
   BYTECODE=`echo $BYTECODE | base64 -D | hexdump -ve '1/1 "%.2X"'`
elif [ "$(uname)" == "Linux" ]; then
   BYTECODE=`echo $BYTECODE | base64 --decode | hexdump -ve '1/1 "%.2X"'`
else
   echo "ERROR: Uknown OS!!"
fi

# unescape quotes in the json and write the ABI to file
# TODO: fix the lllc-server so this doesn't happen
ABI=`eval echo $ABI`
ABI=`echo $ABI | jq .`

echo "BYTE CODE:"
echo $BYTECODE
echo ""

echo "ABI:"
echo $ABI
echo ""

#################################################################
## Step 3 - Create and Sign Transaction for Deploying Contract ##
#################################################################

# location of the blockchain node's rpc server
ERISDB_HOST="localhost:46657"

# to create the transaction, we need to know the account's nonce, so we fetch from the blockchain using simple HTTP
NONCE=`curl -X GET 'http://'"$ERISDB_HOST"'/get_account?address="'"$ADDRESS"'"' --silent | jq ."result"[1].account.sequence`

echo "NONCE:"
echo $NONCE
echo ""

# some variables for the call tx
CALLTX_TYPE=2 # each tx has a type (they can be found in github.com/shuangjj/tendermint/types/tx.go)
FEE=0
GAS=1000
AMOUNT=1
NONCE=$(($NONCE + 1)) # the nonce in the transaction must be one greater than the account's current nonce

# the string that must be signed is a special, canonical, deterministic json structure
# that includes the chain_id and the transaction, where all fields are alphabetically ordered and there are no spaces
SIGN_BYTES='{"chain_id":"'"$CHAIN_ID"'","tx":['"$CALLTX_TYPE"',{"address":"","data":"'"$BYTECODE"'","fee":'"$FEE"',"gas_limit":'"$GAS"',"input":{"address":"'"$ADDRESS"'","amount":'"$AMOUNT"',"sequence":'"$NONCE"'}}]}'

# we convert the sign bytes to hex to send to the keys server for signing
SIGN_BYTES_HEX=`echo -n $SIGN_BYTES | hexdump -ve '1/1 "%.2X"'`

echo "SIGNBYTES:"
echo $SIGN_BYTES
echo ""

echo "SIGNBYTES HEX:"
echo $SIGN_BYTES_HEX
echo ""


# to sign the SIGN_BYTES, we curl the eris-keys server:
# (we gave it the private key for this address at the beginning - with mintkey)
read -r -d '' REQ << EOM
{
	"msg":"$SIGN_BYTES_HEX",
	"addr":"$ADDRESS"
}
EOM

ERIS_KEYS_HOST="localhost:4767"
SIGN_URL="$ERIS_KEYS_HOST/sign"
SIGNATURE=`curl --silent -X POST -d "${REQ}" $SIGN_URL --header "Content-Type:application/json"`

SIGNATURE=`echo $SIGNATURE | jq .Response`

echo "SIGNATURE:"
echo $SIGNATURE
echo ""

# we're going to need the pubkey
# (the pubkey can also be fetched via a curl request to $ERIS_KEYS_HOST/pub with post body {"addr:"$ADDRESS"}
PUBKEY=`eris-keys pub --addr=$ADDRESS`

# now we can actually construct the transaction (it's just the sign bytes plus the pubkey and signature!)
# since it's a CallTx with an empty address, a new contract will be created from the data (the bytecode)
read -r -d '' CREATE_CONTRACT_TX << EOM
[$CALLTX_TYPE, {
	"input":{
		"address":"$ADDRESS",
		"amount":$AMOUNT,
		"sequence":$NONCE,
		"signature":[1,$SIGNATURE],
		"pub_key":[1,"$PUBKEY"]
	},
	"address":"",
	"gas_limit":$GAS,
	"fee":$FEE,
	"data":"$BYTECODE"
}]
EOM

echo "CREATE CONTRACT TX:"
echo $CREATE_CONTRACT_TX
echo ""

#############################################
## Step 4 - Broadcast tx to the blockchain ##
#############################################

# package the jsonrpc request for sending the transaction to the blockchain
JSON_DATA='{"jsonrpc":"2.0","id":"","method":"broadcast_tx","params":['"$CREATE_CONTRACT_TX"']}'

echo "JSON DATA:"
echo $JSON_DATA
echo ""

# broadcast the transaction to the chain!
CONTRACT_ADDRESS=`curl --silent -X POST -d "${JSON_DATA}" "$ERISDB_HOST" --header "Content-Type:application/json" | jq .result[1].receipt.contract_addr`

CONTRACT_ADDRESS="${CONTRACT_ADDRESS%\"}"
CONTRACT_ADDRESS="${CONTRACT_ADDRESS#\"}"

echo "CONTRACT ADDRESS:"
echo $CONTRACT_ADDRESS
echo ""

#############################################
## Step 5 - Wait for a confirmation	   ##
#############################################

# now we wait for a block to be confirmed by polling the status endpoint until the block_height increases

BLOCKHEIGHT_START=`curl -X GET 'http://'"$ERISDB_HOST"'/status' --silent | jq ."result"[1]."latest_block_height"`

BLOCKHEIGHT=$BLOCKHEIGHT_START

while [[ "$BLOCKHEIGHT_START" == "$BLOCKHEIGHT" ]]; do
	BLOCKHEIGHT=`curl -X GET 'http://'"$ERISDB_HOST"'/status' --silent | jq ."result"[1]."latest_block_height"`
done

echo "BLOCKHEIGHT"
echo "$BLOCKHEIGHT"
echo ""

# Note we could also set up a websocket connection and subscribe to NewBlock events
# (eg. subscribeAndWait in github.com/shuangjj/mint-client/mintx/core/core.go )

#############################################
## Step 6 - Verify the contract's bytecode ##
#############################################

CODE=`curl -X GET 'http://'"$ERISDB_HOST"'/get_account?address="'"$CONTRACT_ADDRESS"'"' --silent | jq ."result"[1].account.code`

# strip quotes
CODE="${CODE%\"}"
CODE="${CODE#\"}"

echo "CODE AT CONTRACT:"
echo $CODE
echo ""

# NOTE: CODE won't be exactly equal to BYTECODE
# because BYTECODE contains additional code for the actual deployment (the init/constructor sequence of a contract)
# so we only ensure that BYTECODE contains CODE
if [[ "$BYTECODE" == *"$CODE"* ]]; then
	echo 'THE CODE WAS DEPLOYED CORRECTLY!'
else
	echo 'THE CODE AT THE CONTRACT ADDRESS IS NOT WHAT WE DEPLOYED!'
	echo "Deployed: $BYTECODE"
	echo "Got: $CODE"
fi

##################################################################
## Step 7 - Create and Sign Transaction for Talking to Contract ##
##################################################################

# some variables for the call tx
FEE=0
GAS=1000
AMOUNT=1

# we are going to call the "add" function of our contract
# and use it to add two numbers
FUNCTION="add"
ARG1="25"
ARG2="37"
SUM_EXPECTED=$(( $ARG1 + $ARG2 ))

# we need to format the data for the abi properly
# this part is tricky because we need to get the function identifier for the function we are trying to call from the contract
# the function identifier is the first 4 bytes of the sha3 hash of a canonical form of the function signature
# details are here: https://github.com/ethereum/wiki/wiki/Ethereum-Contract-ABI
# we use the eris-abi tool to make this simple:

# TODO: info messages make it to the output the first time around(if ~/.eris/abi does not exist) They need to go to stderr:
# for that reason we run the command the second time, until this is fixed. Example: echo $DATA
# Abi directory tree incomplete... Creating it... Directory tree built! a5f3c23b00000000000000000000000000000000000000000000000000000000000000190000000000000000000000000000000000000000000000000000000000000025

DATA=`eris-abi pack --input file <(echo $ABI) $FUNCTION $ARG1 $ARG2`
DATA=`eris-abi pack --input file <(echo $ABI) $FUNCTION $ARG1 $ARG2`

echo "DATA FOR CONTRACT CALL:"
echo $DATA
echo ""

# since we've already shown how to create a transaction, sign it, and send it to the blockchain using curl,
# now we do it simply using the mintx tool.
# the --sign and --broadcast flags ensure the transaction is signed (by the private key associated with --pubkey)
# and broadcast to the chain. the --wait flag waits until the transaction is confirmed
RESULT=`mintx call --node-addr=$ERISDB_HOST --chainID=$CHAIN_ID --to=$CONTRACT_ADDRESS --amt=$AMOUNT --fee=$FEE --gas=$GAS --data=$DATA --pubkey=$PUBKEY --sign --broadcast --wait`

echo "$RESULT"
echo ""

# grab the return value
SUM_GOT=`echo "$RESULT" | grep "Return Value:" | awk '{print $3}' | sed 's/^0*//'`

# convert it from hex to int
SUM_GOT=`echo $((16#$SUM_GOT))`


if [[ "$SUM_GOT" != "$SUM_EXPECTED" ]]; then
	echo "SMART CONTRACT ADDITION TX FAILED"
	echo "GOT $SUM_GOT"
	echo "EXPECTED $SUM_EXPECTED"
else
	echo 'SMART CONTRACT ADDITION TX SUCCEEDED!'
	echo "$ARG1 + $ARG2 = $SUM_GOT"
fi
echo ""


##################################################################
## Step 8 - Talk to Contracts Without Creating Transactions     ##
##################################################################

# It is possible to "query" contracts using the /call endpoint.
# Such queries are only "simulated calls", in that there is no transaction (or signature)
# required, and hence they have no effect on the blockchain state.

SUM_GOT=`curl -X GET 'http://'"$ERISDB_HOST"'/call?fromAddress="'"$ADDRESS"'"&toAddress="'"$CONTRACT_ADDRESS"'"&data="'"$DATA"'"' --silent | jq ."result"[1].return`

# strip quotes
SUM_GOT="${SUM_GOT%\"}"
SUM_GOT="${SUM_GOT#\"}"

# convert it from hex to int
SUM_GOT=`echo $((16#$SUM_GOT))`

if [[ "$SUM_GOT" != "$SUM_EXPECTED" ]]; then
	echo "SMART CONTRACT ADDITION QUERY FAILED"
	echo "GOT $SUM_GOT"
	echo "EXPECTED $SUM_EXPECTED"
else
	echo 'SMART CONTRACT ADDITION QUERY SUCCEEDED!'
	echo "$ARG1 + $ARG2 = $SUM_GOT"
fi
