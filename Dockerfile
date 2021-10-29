FROM golang:1.16.9-buster

WORKDIR /chain
COPY . /chain

RUN make install

CMD bandd start --rpc.laddr tcp://0.0.0.0:26657
