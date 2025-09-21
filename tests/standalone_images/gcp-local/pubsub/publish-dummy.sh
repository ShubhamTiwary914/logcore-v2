BASE_URL="http://localhost:8085/v1"
PROJECT="gcplocal-emulator"
TOPIC="source"

MSG='{
  "id": "sensor1",
  "timestamp": "2025-09-20T21:00:00Z",
  "temperature_c": 25.0,
  "humidity_percent": 50,
  "pressure_hpa": 1013
}'

ENCODED_MSG=$(echo -n "$MSG" | base64 | tr -d '\n')

curl -X POST "$BASE_URL/projects/$PROJECT/topics/$TOPIC:publish" \
  -H "Content-Type: application/json" \
  -d "{
        \"messages\": [
          { 
            \"data\": \"$ENCODED_MSG\",
            \"attributes\": { 
              \"relation\": \"weather_sensor\"
            }
          }
        ]
      }"
