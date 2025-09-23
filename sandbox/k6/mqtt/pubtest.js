import { sleep } from "k6";
import { Trend, Counter } from "k6/metrics";
import { Client } from "k6/x/mqtt";

const brokerAddress = "mqtt://localhost:1883";
const topic = "dev";

export const options = {
  vus: 100,
  duration: "10s",
  thresholds: {
    mqtt_message_duration: ["p(50)<200", "p(95)<500", "p(99)<1000"],
  },
};

const messagesSent = new Counter("mqtt_messages_sent");
const messageDuration = new Trend("mqtt_message_duration");
const messageBytes = new Trend("mqtt_message_bytes");
const mqttConnections = new Counter("mqtt_concurrent_connections")

export default function () {
  const client = new Client();
  client.connect(brokerAddress);
  mqttConnections.add(1);

  const durationMs = (__ENV.TEST_DURATION * 1000);
  const start = Date.now();
  const stopAt = start + durationMs; 

  while (Date.now() < stopAt) {
    const payload = `Hello from k6 client ${__VU}`;
    const t0 = Date.now();
    client.publish(topic, payload);
    const t1 = Date.now();

    messagesSent.add(1);
    messageDuration.add(t1 - t0);
    messageBytes.add(payload.length);

    //1-2s delay
    sleep(Math.random() + 1); 
  }

  client.end();
}
