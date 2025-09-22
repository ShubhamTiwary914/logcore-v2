from locust import User, task, between, events
import paho.mqtt.client as mqtt
import time, json 
import logging

logger : logging.Logger = logging.getLogger()
logger.setLevel(logging.INFO)


broker_host : str = "10.43.79.115" 


class MqttUser(User):
    wait_time = between(1, 1.5)
    def on_start(self):
        self.client = mqtt.Client()
        self.client.connect(broker_host, 1883, 60)
        self.client.loop_start()
    
    @task
    def publish(self):
        ts = time.time()
        payload = json.dumps({"ts": ts, "msg": "payload"})
        start_time = time.time()
        try:
            result = self.client.publish("test/topic", payload)
            result.wait_for_publish()
            total_time = int((time.time() - start_time) * 1000)
            events.request.fire(
                request_type="MQTT",
                name="publish",
                response_time=total_time,
                response_length=len(payload),
                exception=None,
            )
        except Exception as e:
            events.request.fire(
                request_type="MQTT",
                name="publish",
                response_time=int((time.time() - start_time) * 1000),
                response_length=0,
                exception=e,
            )
            