import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const loginDuration = new Trend('login_duration');
const gameListDuration = new Trend('game_list_duration');
const gameDetailDuration = new Trend('game_detail_duration');
const requestCounter = new Counter('requests');

// Test configuration
export const options = {
  stages: [
    { duration: '30s', target: 50 },   // Ramp up to 50 users
    { duration: '1m', target: 100 },   // Ramp up to 100 users
    { duration: '2m', target: 100 },   // Stay at 100 users
    { duration: '30s', target: 200 },  // Spike to 200 users
    { duration: '1m', target: 200 },   // Stay at 200 users
    { duration: '30s', target: 0 },    // Ramp down to 0 users
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500', 'p(99)<1000'],
    'http_req_failed': ['rate<0.01'],
    'errors': ['rate<0.1'],
    'login_duration': ['p(99)<200'],
    'game_list_duration': ['p(99)<50'],
    'game_detail_duration': ['p(99)<100'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Test data
const testUsers = [
  { username: 'testuser1', email: 'test1@example.com', password: 'Test1234' },
  { username: 'testuser2', email: 'test2@example.com', password: 'Test1234' },
  { username: 'testuser3', email: 'test3@example.com', password: 'Test1234' },
];

// Helper function to get random user
function getRandomUser() {
  return testUsers[Math.floor(Math.random() * testUsers.length)];
}

// Helper function to login
function login(user) {
  const loginStart = Date.now();
  const res = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({
    username: user.username,
    password: user.password,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  loginDuration.add(Date.now() - loginStart);
  requestCounter.add(1);

  const success = check(res, {
    'login status is 200': (r) => r.status === 200,
    'login has access token': (r) => r.json('data.access_token') !== undefined,
  });

  errorRate.add(!success);

  if (success) {
    return res.json('data.access_token');
  }
  return null;
}

// Test scenarios
export default function () {
  const user = getRandomUser();

  // Scenario 1: Health check (10% of requests)
  if (Math.random() < 0.1) {
    const res = http.get(`${BASE_URL}/health`);
    requestCounter.add(1);
    check(res, {
      'health check status is 200': (r) => r.status === 200,
    });
    sleep(1);
    return;
  }

  // Scenario 2: Browse games without login (30% of requests)
  if (Math.random() < 0.3) {
    const listStart = Date.now();
    const res = http.get(`${BASE_URL}/api/v1/games?page=1&page_size=20`);
    gameListDuration.add(Date.now() - listStart);
    requestCounter.add(1);

    const success = check(res, {
      'game list status is 200': (r) => r.status === 200,
      'game list has data': (r) => r.json('data') !== undefined,
    });
    errorRate.add(!success);

    sleep(1);
    return;
  }

  // Scenario 3: Login and browse (60% of requests)
  const token = login(user);
  if (!token) {
    sleep(1);
    return;
  }

  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };

  // Get game list
  const listStart = Date.now();
  const listRes = http.get(`${BASE_URL}/api/v1/games?page=1&page_size=20`, { headers });
  gameListDuration.add(Date.now() - listStart);
  requestCounter.add(1);

  const listSuccess = check(listRes, {
    'authenticated game list status is 200': (r) => r.status === 200,
  });
  errorRate.add(!listSuccess);

  if (listSuccess && listRes.json('data.games') && listRes.json('data.games').length > 0) {
    const games = listRes.json('data.games');
    const randomGame = games[Math.floor(Math.random() * games.length)];

    // Get game detail
    const detailStart = Date.now();
    const detailRes = http.get(`${BASE_URL}/api/v1/games/${randomGame.id}`, { headers });
    gameDetailDuration.add(Date.now() - detailStart);
    requestCounter.add(1);

    const detailSuccess = check(detailRes, {
      'game detail status is 200': (r) => r.status === 200,
      'game detail has data': (r) => r.json('data') !== undefined,
    });
    errorRate.add(!detailSuccess);
  }

  sleep(1);
}

// Setup function (runs once at the beginning)
export function setup() {
  console.log('Starting load test...');
  console.log(`Base URL: ${BASE_URL}`);
  console.log(`Test users: ${testUsers.length}`);
  return { startTime: Date.now() };
}

// Teardown function (runs once at the end)
export function teardown(data) {
  const duration = (Date.now() - data.startTime) / 1000;
  console.log(`Load test completed in ${duration.toFixed(2)} seconds`);
}
