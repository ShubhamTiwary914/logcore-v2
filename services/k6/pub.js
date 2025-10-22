//Required Arguments:
// INTERVAL(sec): gap between publish (for each vus)
// DURATION(sec): after how much time should vus's MQTT connection end?
// RELATION(string): type of schema (from test schema)

import { Trend, Counter } from "k6/metrics";
import { Client } from "k6/x/mqtt";

const schemas = {
  "weather_sensor": [
    {"name": "id", "type": "STRING"},
    {"name": "timestamp", "type": "TIMESTAMP"},
    {"name": "temperature_c", "type": "FLOAT"},
    {"name": "humidity_percent", "type": "FLOAT"},
    {"name": "pressure_hpa", "type": "FLOAT"},
    {"name": "wind_speed_mps", "type": "FLOAT"},
    {"name": "rainfall_mm", "type": "FLOAT"}
  ],
  "machine_metrics": [
    {"name": "id", "type": "STRING"},
    {"name": "timestamp", "type": "TIMESTAMP"},
    {"name": "vibration_mms", "type": "FLOAT"},
    {"name": "motor_rpm", "type": "INTEGER"},
    {"name": "power_kw", "type": "FLOAT"},
    {"name": "oil_temp_c", "type": "FLOAT"},
    {"name": "status_code", "type": "INTEGER"}
  ]
}

const brokerAddress = `mqtt://${__ENV.VERNE_IP}:1883`;
const topic = "mqtt-source";

const QOS = Object.freeze({
  ATMOSTONCE: 0,
  ATLEASTONCE: 1,
  EXACTLYONCE: 2
})

export const options = {
  vus: 1,
  duration: "10s",
  thresholds: {
    mqtt_message_duration: ["p(50)<200", "p(95)<500", "p(99)<1000"],
  },
};

const messagesSent = new Counter("mqtt_messages_sent");
const messagesFailed = new Counter("mqtt_messages_failed");
const messageDuration = new Trend("mqtt_message_duration");
const messageBytes = new Trend("mqtt_message_bytes");
const mqttConnections = new Counter("mqtt_concurrent_connections")


export function setup(){
  console.log(`
    Running K6 Load test with ARGS: 
    Duration: ${__ENV.DURATION}s
    Interval: ${__ENV.INTERVAL}s
    Virtual Users: ${__ENV.VUS}
    Verne Broker Address: ${__ENV.VERNE_IP}
    \n
  `)
}


export default function () {
  const client = new Client()
  const relation = getRelation()

  client.on("connect", () => {
    mqttConnections.add(1)
    console.log(`User(${__VU}): Connected to MQTT broker`)

    const intervalId = setInterval(() => {
      let payload = generatePayload(relation)
      let beforePub = Date.now()
      client.publish(topic, payload, {
        qos: QOS.EXACTLYONCE
      }, (err)=>{
        let afterPub = Date.now()
        if(!err){
          messageDuration.add(afterPub-beforePub)
          messagesSent.add(1)
          messageBytes.add(payload.length*2)
        }
        else
          messagesFailed.add(1)
      })
    }, __ENV.INTERVAL*1000)

    setTimeout(() => {
      clearInterval(intervalId)
      client.end()
      //duration for single vus end connection
    }, __ENV.DURATION*1000)
  })

  client.on("end", () => {
    console.log("Disconnected from MQTT broker")
  })

  client.connect(brokerAddress)
}

//------------------
//region helpers

function randomString(length = 12) {
  const chars = "abcdefghijklmnopqrstuvwxyz0123456789";
  let result = "";
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

function generateValue(type, relation) {
  switch (type) {
    case "STRING":
      return `${relation}-${randomString(12)}`;
    case "TIMESTAMP":
      return new Date().toISOString();
    case "FLOAT":
      return parseFloat((Math.random() * 100).toFixed(2));
    case "INTEGER":
      return Math.floor(Math.random() * 100);
    default:
      return null;
  }
}

function generatePayload(relation) {
  const schema = schemas[relation];
  if (!schema) throw new Error(`Schema for relation '${relation}' not found`);
  const obj = {};
  schema.forEach((col) => {
    obj[col.name] = generateValue(col.type, relation);
  });
  obj["relation"] = relation
  return JSON.stringify(obj);
}

function getRelation() {
  const relArg = __ENV.RELATION;
  if (!relArg) 
    throw new Error("RELATION environment variable is required");
  return relArg;
}
