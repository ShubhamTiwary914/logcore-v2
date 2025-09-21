#!/bin/bash

BASE_URL="http://localhost:8085/v1"
PROJECT="gcplocal-emulator"
TOPIC="source"
RELATION="$2"

if [[ -z "$RELATION" ]]; then
  echo "Usage: $0 <relation_name>"
  exit 1
fi

SCHEMA_JSON=$(cat ../schemas.json)

# check if relation exists
if ! echo "$SCHEMA_JSON" | jq -e --arg rel "$RELATION" '.[$rel] != null' >/dev/null; then
  echo "Relation '$RELATION' not found in schemas.json"
  exit 1
fi

for i in $(seq 1 5); do
  MSG="{"
  for row in $(echo "$SCHEMA_JSON" | jq -r --arg rel "$RELATION" '.[$rel][] | @base64'); do
    _jq() { echo "${row}" | base64 --decode | jq -r "${1}"; }
    NAME=$(_jq '.name')
    TYPE=$(_jq '.type')
    VALUE=""
    case $TYPE in
      STRING)
        if [[ "$NAME" == "id" ]]; then
          VALUE="\"$((RANDOM % 1000000))\""
        else
          VALUE="\"val_$i\""
        fi
        ;;
      TIMESTAMP)
        VALUE="\"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\""
        ;;
      FLOAT)
        VALUE=$(awk -v min=0 -v max=100 'BEGIN{srand(); printf "%.2f", min+rand()*(max-min)}')
        ;;
      INTEGER)
        VALUE=$(awk -v min=0 -v max=1000 'BEGIN{srand(); print int(min+rand()*(max-min))}')
        ;;
    esac
    MSG+=\"${NAME}\":${VALUE},
  done
  MSG="${MSG%,}}"
  echo "Published message for relation - $RELATION"
  echo $MSG
  echo ""
  ENCODED_MSG=$(echo -n "$MSG" | base64 | tr -d '\n')
  curl -s -X POST "$BASE_URL/projects/$PROJECT/topics/$TOPIC:publish" \
    -H "Content-Type: application/json" \
    -d "{
          \"messages\": [
            { 
              \"data\": \"$ENCODED_MSG\",
              \"attributes\": { \"relation\": \"$RELATION\" }
            }
          ]
        }" > /dev/null 
done
