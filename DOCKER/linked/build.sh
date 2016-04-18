#! /bin/bash
set -e

# mct = mint client test

cd $GOPATH/src/github.com/shuangjj/mint-client
echo "********** BUILDING TENDERMINT ********"

docker build -t mct_tendermint -f ./DOCKER/linked/DockerfileTendermint . 
echo "********** BUILDING ERIS-KEYS ********"
docker build -t mct_keys -f ./DOCKER/linked/DockerfileKeys . 
echo "********** BUILDING MINT-CLIENT ********"
docker build -t mct_client -f ./DOCKER/linked/DockerfileClient . 

cd ./DOCKER/linked
bash run.sh
