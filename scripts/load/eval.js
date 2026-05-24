// k6 load test for the /eval endpoints.
//
// Usage:
//   k6 run scripts/load/eval.js
//   k6 run -e BASE_URL=http://localhost:8080 -e FLAG_KEY=my-flag scripts/load/eval.js
//   k6 run -e SCENARIO=ramp scripts/load/eval.js
//
// Required setup: a flag must already exist (POST /api/v1/flags) whose key
// matches FLAG_KEY (default "bench-flag"). The script does not create it.

import http from 'k6/http';
import { check } from 'k6';
import { Trend } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const FLAG_KEY = __ENV.FLAG_KEY || 'bench-flag';
const SCENARIO = __ENV.SCENARIO || 'steady';

const evalSingleLatency = new Trend('eval_single_ms', true);
const evalAllLatency = new Trend('eval_all_ms', true);

const scenarios = {
  steady: {
    executor: 'constant-vus',
    vus: Number(__ENV.VUS || 50),
    duration: __ENV.DURATION || '30s',
  },
  ramp: {
    executor: 'ramping-vus',
    startVUs: 0,
    stages: [
      { duration: '15s', target: 50 },
      { duration: '30s', target: 200 },
      { duration: '15s', target: 0 },
    ],
  },
  spike: {
    executor: 'ramping-arrival-rate',
    startRate: 50,
    timeUnit: '1s',
    preAllocatedVUs: 100,
    maxVUs: 500,
    stages: [
      { duration: '10s', target: 50 },
      { duration: '10s', target: 1000 },
      { duration: '20s', target: 1000 },
      { duration: '10s', target: 50 },
    ],
  },
};

export const options = {
  scenarios: { primary: scenarios[SCENARIO] },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<200', 'p(99)<500'],
  },
};

function ctxBody() {
  // Vary the user id so bucketing exercises the hash path.
  const id = `u-${__VU}-${__ITER}`;
  return JSON.stringify({
    context: {
      user: {
        id,
        tier: __VU % 3 === 0 ? 'pro' : 'free',
        region: `rgn-${__VU % 5}`,
      },
    },
  });
}

const headers = { 'Content-Type': 'application/json' };

export default function () {
  // Alternate between the two endpoints so both get coverage.
  if (__ITER % 2 === 0) {
    const r = http.post(`${BASE_URL}/api/v1/eval/${FLAG_KEY}`, ctxBody(), { headers });
    evalSingleLatency.add(r.timings.duration);
    check(r, { 'eval/{key} 200': (res) => res.status === 200 });
  } else {
    const r = http.post(`${BASE_URL}/api/v1/eval`, ctxBody(), { headers });
    evalAllLatency.add(r.timings.duration);
    check(r, { 'eval (all) 200': (res) => res.status === 200 });
  }
}
