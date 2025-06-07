// scripts/config-flow.js
import http from 'k6/http';
import { check, sleep, fail } from 'k6';

export const options = {
    thresholds: {
        http_req_failed:   ['rate<0.01'],   // <1 % ошибок
        http_req_duration: ['p(95)<200'],   // p95 < 200 мс
    },

    scenarios: {
        default: {
            executor: 'ramping-vus',
            gracefulStop: '30s',          // сколько даём VU завершиться
            stages: [
                // ramp-up  ➜ plateau  ➜ ramp-up …
                { duration: '15s', target: 200 },
                { duration: '30s', target: 200 },

                { duration: '15s', target: 400 },
                { duration: '30s', target: 400 },

                { duration: '15s', target: 800 },
                { duration: '30s', target: 800 },

                { duration: '15s', target: 1000 },
                { duration: '1m', target: 1000 },

                // аккуратное снижение — чтобы не ронять сервер резким обрывом
                { duration: '30s', target: 0 },
            ],
        },
    },
};

const BASE = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
    /* — ваш сценарий без изменений — */
    const res = http.get(`${BASE}/config/components?category=cpu`);
    check(res, { 'GET /components 200': r => r.status === 200 });

    const list = res.json();
    if (!Array.isArray(list) || list.length === 0) {
        fail(`Нет CPU в ответе. status=${res.status}, body=${res.body?.substring(0,200)}`);
    }
    const cpu = list[Math.floor(Math.random() * list.length)].name;

    const res2 = http.post(
        `${BASE}/config/compatible`,
        JSON.stringify({ category: 'motherboard', bases: [{ category: 'cpu', name: cpu }] }),
        { headers: { 'Content-Type': 'application/json' } },
    );
    check(res2, { 'POST /compatible 200': r => r.status === 200 });

    sleep(1);                // чтобы один VU делал ≈1 итерацию/сек.
}
