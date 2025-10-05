FROM python:3.12-slim AS builder
WORKDIR /app

COPY requirements.txt .
RUN pip wheel --no-cache-dir --wheel-dir /wheels -r requirements.txt


FROM python:3.12-slim
WORKDIR /app
COPY --from=builder /wheels /wheels
RUN pip install --no-cache /wheels/*

COPY pipeline.py .

ENV project=gcplocal-emulator
ENV topic=projects/gcplocal-emulator/topics/source
ENV pubsubhost=gcp-emulators:8085
ENV bigtablehost=gcp-emulators:8086

ENTRYPOINT ["sh", "-c", "python pipeline.py --project \"$project\" --topic \"$topic\" --pubsubhost \"$pubsubhost\" --bigtablehost \"$bigtablehost\""]
