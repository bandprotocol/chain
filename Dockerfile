FROM golang:1.24.2-bookworm

WORKDIR /chain
COPY . /chain

RUN make install

CMD ["bandd", "start", "--rpc.laddr", "tcp://0.0.0.0:26657"]
