docker run -d -it -p 8085:8085 -p 8086:8086 \
  --network=host \
  gcr.io/google.com/cloudsdktool/google-cloud-cli:emulators \
  bash -c "
    gcloud components install pubsub-emulator bigtable --quiet && \
    gcloud beta emulators pubsub start --host-port=0.0.0.0:8085 &
    gcloud beta emulators bigtable start --host-port=0.0.0.0:8086
  "

