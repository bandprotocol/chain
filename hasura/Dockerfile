FROM hasura/graphql-engine:v2.9.0.cli-migrations-v2
RUN apt-get update && apt-get install -y curl && apt-get clean

ENV ALERT_SERVICE ""
ENV ALERT_KEY ""

COPY alert.sh .
RUN chmod +x alert.sh

HEALTHCHECK --start-period=2m --interval=1m --timeout=20s \
    CMD curl -s --fail http://localhost:8080/healthz || (./alert.sh ${ALERT_SERVICE} ${ALERT_KEY} && kill 1)


