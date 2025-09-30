import os
from google.cloud import pubsub_v1

os.environ["PUBSUB_EMULATOR_HOST"] = "localhost:8085"

project_id = "gcplocal-emulator"
subscription_id = "source-sub"
topic_id = "source"

subscriber = pubsub_v1.SubscriberClient()
topic_path = f"projects/{project_id}/topics/{topic_id}"
subscription_path = f"projects/{project_id}/subscriptions/{subscription_id}"

def callback(message):
    print(f"Received message: {message.data.decode('utf-8')}, attributes: {message.attributes}")
    message.ack()

streaming_pull_future = subscriber.subscribe(subscription_path, callback=callback)
print(f"Listening to {topic_id}...")

try:
    streaming_pull_future.result()
except KeyboardInterrupt:
    streaming_pull_future.cancel()

