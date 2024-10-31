docker pull bandprotocol/bothan-api:latest
docker run --log-opt max-size=10m --log-opt max-file=3 -d \
    --name bothan -v "$(pwd)/scripts/bothan/bothan-config.toml:/root/.bothan/config.toml" \
    -p 50051:50051 bandprotocol/bothan-api:latest
