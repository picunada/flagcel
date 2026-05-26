// k6 load test for the /eval endpoints.
//
// Usage:
//   k6 run scripts/load/eval.js
//   k6 run -e BASE_URL=http://localhost:8080 -e FLAG_KEY=my-flag scripts/load/eval.js
//   k6 run -e SCENARIO=ramp scripts/load/eval.js
//   k6 run -e SETUP=true -e ADMIN_EMAIL=admin@localhost -e ADMIN_PASSWORD=flagcel-dev-password scripts/load/eval.js
//
// Required setup unless SETUP=true: a flag must already exist whose key matches
// FLAG_KEY (default "bench-flag"), and it must be linked to a context declaring:
// user.id, user.tier, and user.region. Set API_TOKEN when auth is enabled.

import http from 'k6/http';
import { check } from 'k6';
import { Trend } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const FLAG_KEY = __ENV.FLAG_KEY || 'bench-flag';
const SCENARIO = __ENV.SCENARIO || 'steady';
const SETUP = __ENV.SETUP === 'true';
const API_TOKEN = __ENV.API_TOKEN || '';
const ADMIN_EMAIL = __ENV.ADMIN_EMAIL || 'admin@localhost';
const ADMIN_PASSWORD = __ENV.ADMIN_PASSWORD || 'flagcel-dev-password';
const CONTEXT_NAME = __ENV.CONTEXT_NAME || `loadtest-${FLAG_KEY}`;
const RULE_EXPR = __ENV.RULE_EXPR || 'user.tier == "pro"';
const ROLLOUT_PERCENTAGE = Number(__ENV.ROLLOUT_PERCENTAGE || 100);

const evalSingleLatency = new Trend('eval_single_ms', true);
const evalAllLatency = new Trend('eval_all_ms', true);

const contextFields = [
  { path: 'user.id', type: 'string' },
  { path: 'user.tier', type: 'string' },
  { path: 'user.region', type: 'string' },
];

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
  return contextBody(id, __VU % 3 === 0 ? 'pro' : 'free', `rgn-${__VU % 5}`);
}

function contextBody(id, tier, region) {
  return JSON.stringify({
    context: {
      user: {
        id,
        tier,
        region,
      },
    },
  });
}

function jsonHeaders(extra = {}) {
  return Object.assign({ 'Content-Type': 'application/json' }, extra);
}

function evalHeaders(token) {
  const headers = jsonHeaders();
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  return headers;
}

function dataFrom(res) {
  return JSON.parse(res.body).data;
}

function cookieHeader(res) {
  const parts = [];
  for (const name in res.cookies) {
    if (res.cookies[name] && res.cookies[name][0]) {
      parts.push(`${name}=${res.cookies[name][0].value}`);
    }
  }
  return parts.join('; ');
}

function adminRequest(method, path, body, cookie) {
  return http.request(method, `${BASE_URL}${path}`, body ? JSON.stringify(body) : null, {
    headers: jsonHeaders({ Cookie: cookie }),
  });
}

function getOrCreateContext(cookie) {
  const create = adminRequest(
    'POST',
    '/api/v1/contexts',
    {
      name: CONTEXT_NAME,
      description: 'k6 evaluation load test context',
      fields: contextFields,
    },
    cookie,
  );
  if (create.status === 200) {
    return dataFrom(create);
  }
  if (create.status !== 409) {
    throw new Error(`create context failed: ${create.status} ${create.body}`);
  }

  const list = adminRequest('GET', '/api/v1/contexts', null, cookie);
  if (list.status !== 200) {
    throw new Error(`list contexts failed: ${list.status} ${list.body}`);
  }
  const existing = dataFrom(list).find((ctx) => ctx.name === CONTEXT_NAME);
  if (!existing) {
    throw new Error(`context ${CONTEXT_NAME} already exists but was not returned`);
  }
  const update = adminRequest(
    'PUT',
    `/api/v1/contexts/${encodeURIComponent(existing.id)}`,
    {
      name: CONTEXT_NAME,
      description: existing.description || 'k6 evaluation load test context',
      fields: contextFields,
    },
    cookie,
  );
  if (update.status !== 200) {
    throw new Error(`update context failed: ${update.status} ${update.body}`);
  }
  return dataFrom(update);
}

function seedFlag(cookie, contextID) {
  const res = adminRequest(
    'POST',
    '/api/v1/flags',
    {
      key: FLAG_KEY,
      enabled: true,
      default_value: false,
      context_id: contextID,
      rules: [
        {
          expression: RULE_EXPR,
          rollout: {
            percentage: ROLLOUT_PERCENTAGE,
            bucket_by: 'user.id',
          },
        },
      ],
    },
    cookie,
  );
  if (res.status !== 200) {
    throw new Error(`create flag failed: ${res.status} ${res.body}`);
  }
}

function createAPIKey(cookie) {
  const res = adminRequest(
    'POST',
    '/api/v1/api-keys',
    { name: `loadtest-${FLAG_KEY}-${Date.now()}` },
    cookie,
  );
  if (res.status !== 200) {
    throw new Error(`create api key failed: ${res.status} ${res.body}`);
  }
  return dataFrom(res).token;
}

function assertEvalToken(token) {
  if (!token) {
    throw new Error('eval API token is empty');
  }
  const res = http.post(
    `${BASE_URL}/api/v1/eval/${FLAG_KEY}`,
    contextBody('setup-user', 'pro', 'rgn-setup'),
    { headers: evalHeaders(token) },
  );
  if (res.status !== 200) {
    throw new Error(`eval token probe failed: ${res.status} ${res.body}`);
  }
}

export function setup() {
  if (!SETUP) {
    assertEvalToken(API_TOKEN);
    return { token: API_TOKEN };
  }

  const login = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    { headers: jsonHeaders() },
  );
  if (login.status !== 200) {
    throw new Error(`admin login failed: ${login.status} ${login.body}`);
  }

  const cookie = cookieHeader(login);
  const context = getOrCreateContext(cookie);
  seedFlag(cookie, context.id);
  const token = createAPIKey(cookie);
  assertEvalToken(token);
  return { token };
}

export default function (data) {
  if (!data || !data.token) {
    throw new Error('missing eval API token from setup');
  }
  const headers = evalHeaders(data.token);
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
