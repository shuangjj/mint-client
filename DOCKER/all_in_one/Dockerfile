# Pull base image.
FROM eris/base

ENV DOCKERFILES ./DOCKER/all_in_one

# grab deps (gmp)
RUN apt-get update && \
  apt-get install -y --no-install-recommends \
    libgmp3-dev jq && \
  rm -rf /var/lib/apt/lists/*

# set the repo and install tendermint
ENV repo /go/src/github.com/shuangjj/mint-client
ADD . $repo
WORKDIR $repo
RUN go install ./...

# grab eris-keys
RUN go get github.com/shuangjj/eris-keys

# grab tendermint
RUN go get github.com/shuangjj/tendermint/cmd/tendermint
ENV TMROOT /home/eris/.eris/blockchains/tendermint
ADD $DOCKERFILES/config.toml $TMROOT/config.toml
RUN chown -R $USER:$USER $TMROOT

ADD ./$DOCKERFILES/test.sh /test.sh
RUN chown $USER:$USER /test.sh

# set user
USER $USER
WORKDIR /home/eris

CMD ["/test.sh"]
