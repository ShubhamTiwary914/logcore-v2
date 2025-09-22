PROJECT="gcplocal-emulator"
TOPICS=("source")
BASE_URL="http://localhost:8085/v1"

for TOPIC in "${TOPICS[@]}"; do
    curl -X PUT "$BASE_URL/projects/$PROJECT/topics/$TOPIC"
done
