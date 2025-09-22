FROM gcr.io/google.com/cloudsdktool/google-cloud-cli:emulators

RUN gcloud components install pubsub-emulator --quiet
RUN gcloud components install bigtable --quiet

CMD ["/bin/bash"]
