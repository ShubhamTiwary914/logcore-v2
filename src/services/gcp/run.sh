docker run -d -it \
  --network=host \
  gcr.io/google.com/cloudsdktool/google-cloud-cli:emulators \
  bash -c "
    gcloud components install pubsub-emulator bigtable --quiet && \
    gcloud beta emulators pubsub start --host-port=0.0.0.0:8085 &
    gcloud beta emulators bigtable start --host-port=0.0.0.0:8086
  "

