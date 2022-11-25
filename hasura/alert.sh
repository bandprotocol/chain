#!/bin/bash

SERVICE=$1
KEY=$2

case $SERVICE in
  "pagerduty")
        curl -s -X POST 'https://events.pagerduty.com/v2/enqueue' --header 'Content-Type: application/json' --data "{ \"routing_key\":  \"$KEY\", \"event_action\": \"trigger\", \"payload\": { \"summary\": \"Hasura: Unhealthy\", \"source\": \"hasura\", \"severity\": \"critical\"}}";
    ;;
esac
