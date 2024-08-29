docker pull bandprotocol/bothan-api:latest
docker run --log-opt max-size=10m --log-opt max-file=3 -d \
    --name bothan -v $(pwd)/scripts/bothan/api-config.toml:/app/config.toml \
    -p 50051:50051 bandprotocol/bothan-api:latest
