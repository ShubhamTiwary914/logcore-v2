#ENVs needed: (PROJECT, TOPIC, INSTANCE, TABLE)
#Assumes ADC(app def credentials) / GKE KSA set for GCP auth
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r /app/requirements.txt
COPY pipeline.py .

ENTRYPOINT ["sh", "-c", "exec python /app/pipeline.py --project \"$PROJECT\" --topic \"$TOPIC\" --instance \"$INSTANCE\" --table \"$TABLE\""]