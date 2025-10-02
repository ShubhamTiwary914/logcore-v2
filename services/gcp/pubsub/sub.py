import os
from google.cloud import pubsub_v1
from google.api_core.exceptions import AlreadyExists

os.environ["PUBSUB_EMULATOR_HOST"] = "localhost:8085"

project_id = "gcplocal-emulator"
subscription_id = "source-sub"
topic_id = "source"


subscriber = pubsub_v1.SubscriberClient()
topic_path = f"projects/{project_id}/topics/{topic_id}"
subscription_path = f"projects/{project_id}/subscriptions/{subscription_id}"
try:
    subscriber.create_subscription(name=subscription_path, topic=topic_path)
except AlreadyExists:
    pass

def callback(message):
    print(f"Received message: {message.data.decode('utf-8')}, attributes: {message.attributes}")
    message.ack()

streaming_pull_future = subscriber.subscribe(subscription_path, callback=callback)
print(f"Listening to {topic_id}...")


try:
    streaming_pull_future.result()
except KeyboardInterrupt:
    streaming_pull_future.cancel()

