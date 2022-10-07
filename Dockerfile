FROM golang:1.19-buster

WORKDIR /chain
COPY . /chain

COPY docker-config/run.sh .

RUN make install && make faucet

CMD bandd start --rpc.laddr tcp://0.0.0.0:26657
