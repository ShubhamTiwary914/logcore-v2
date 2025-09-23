#!/bin/bash
TABLE=$1
COUNT=${2:-1}
SCHEMA_FILE="./schemas.json"
HOST="localhost"
PORT="1883"

if ! jq -e --arg table "$TABLE" '.[$table]' "$SCHEMA_FILE" > /dev/null; then
    echo "wrong relation, check schemas.json"
    exit 1
fi
TOPIC="mqtt-source"

generate_random_id() {
    echo "$TABLE-$(head /dev/urandom | tr -dc 0-9 | head -c12)"
}

timestamp() {
    date -u +"%Y-%m-%dT%H:%M:%SZ"
}

generate_float() {
    awk -v min=$1 -v max=$2 'BEGIN{srand(); print min+rand()*(max-min)}'
}

generate_integer() {
    shuf -i $1-$2 -n 1
}

for ((i=0;i<COUNT;i++)); do
    payload=$(jq -r --arg table "$TABLE" '.[$table]' "$SCHEMA_FILE" | jq -c 'reduce .[] as $field ({}; .[$field.name] = ($field.type | if .=="STRING" then "'"$(generate_random_id)"'" elif .=="TIMESTAMP" then "'"$(timestamp)"'" elif .=="FLOAT" then '"$(generate_float 0 100)"' elif .=="INTEGER" then '"$(generate_integer 0 1000)"' else null end))')
    relation="$TABLE"
    payload=$(jq --arg relation "$relation" '. + {relation: $relation}' <<< "$payload") 
    echo "Pushing into verneMQTT (host=$HOST, port=$PORT, topic=$TOPIC), payload:"
    echo $payload
    mqtt pub -h "$HOST" -p "$PORT" -t "$TOPIC" -m "$payload"
done
