import http from 'k6/http';
import { expect } from "https://jslib.k6.io/k6-testing/0.5.0/index.js";

export const options = {
  vus: 10,
  duration: '30s',
};

export default function() {
  let res = http.get('http://localhost:9080');
  expect.soft(res.status).toBe(200); 
}
