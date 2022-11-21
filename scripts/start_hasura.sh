HASURA_CONSOLE="${HASURA_CONSOLE:-false}"

: "${HASURA_DB_URL?Need to set HASURA_DB_URL env}"
: "${HASURA_METADATA_DB_URL?Need to set HASURA_METADATA_DB_URL env}"

echo HASURA_DB_URL=$HASURA_DB_URL
echo HASURA_METADATA_DB_URL=$HASURA_METADATA_DB_URL
echo HASURA_CONSOLE=$HASURA_CONSOLE

docker rm -f hasura
docker build -t hasura:latest $(pwd)/hasura
docker run -d -p 80:8080 \
    --name hasura --restart=always \
    -v $(pwd)/hasura/hasura-metadata:/hasura-metadata \
    --health-cmd "curl -s --fail http://localhost:8080/healthz || kill 1" --health-start-period=2m --health-interval=1m --health-timeout=20s \
    --env HASURA_GRAPHQL_DATABASE_URL=$HASURA_DB_URL \
    --env HASURA_GRAPHQL_METADATA_DATABASE_URL=$HASURA_METADATA_DB_URL \
    --env HASURA_GRAPHQL_ENABLE_CONSOLE=$HASURA_CONSOLE \
    --env HASURA_GRAPHQL_SERVER_HOST=0.0.0.0 \
    --env HASURA_GRAPHQL_ENABLED_LOG_TYPES="startup, http-log, webhook-log, websocket-log, query-log" \
    --env HASURA_GRAPHQL_STRINGIFY_NUMERIC_TYPES=true  \
    hasura:latest
