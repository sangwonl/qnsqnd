import http from "k6/http";
import { check } from "k6";

const HOST = `${__ENV.HOST}`;
const TOPIC = `${__ENV.TOPIC}`;

export default function() {
    const params = { headers: { "Content-Type": "application/json" } };
    const payload = JSON.stringify({
        topic: TOPIC,
        payload: {
            hello: "world",
            timestamp: 1554525076
        }
    });
    const res = http.post(`http://${HOST}/publish`, payload, params);
    check(res, { "is status 200": (r) => r.status === 202 });
};
