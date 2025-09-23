import { Trend, Counter } from "k6/metrics";
import { Client } from "k6/x/mqtt";
import exec from 'k6/execution';

const brokerAddress = "mqtt://localhost:1883";
const topic = "dev";

export const options = {
  vus: 1,
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
  const client = new Client()

  client.on("connect", async () => {
    mqttConnections.add(1)
    console.log("Connected to MQTT broker")
    console.log(`Interval: ${__ENV.INTERVAL} && Duration: ${__ENV.DURATION}`)

    const intervalId = setInterval(() => {
      let payload = `hello from k6 user: ${__VU}`
      let beforePub = Date.now()
      client.publish(topic, payload)
      let afterPub = Date.now()
      messageDuration.add(afterPub-beforePub)
      messagesSent.add(1)
      messageBytes.add(payload.length*2)
    }, __ENV.INTERVAL*1000)

    setTimeout(() => {
      clearInterval(intervalId)
      client.end()
    }, __ENV.DURATION*1000)
  })

  client.on("end", () => {
    console.log("Disconnected from MQTT broker")
  })

  client.connect(brokerAddress)
}
