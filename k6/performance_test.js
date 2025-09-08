import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';

// A small array of sample questions to ask the LLM.
const questions = new SharedArray('questions', function () {
    return [
        'What is Go and why is it popular for backend development?',
        'Explain Docker in simple terms.',
        'Write a short story about a robot who discovers music.',
        'How does Redis work?',
    ];
});

export const options = {
    // Simulate 10 virtual users over a 30-second duration.
    vus: 10,
    duration: '30s',
};

export default function () {
    const baseUrl = 'http://localhost:8080/api';

    // 1. Test GET /api/chats endpoint
    const chatsRes = http.get(`${baseUrl}/chats`);
    check(chatsRes, {
        'GET /chats status is 200': (r) => r.status === 200,
    });

    sleep(1); // Wait for 1 second between requests

    // 2. Test POST /api/chats/messages to create a new chat and get a response.
    // This test won't check the SSE stream itself, but will verify the initial connection.
    const payload = JSON.stringify({
        content: questions[Math.floor(Math.random() * questions.length)],
        model: 'llama3:latest', // Make sure this model is available in Ollama
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    const streamRes = http.post(`${baseUrl}/chats/messages`, payload, params);

    check(streamRes, {
        'POST /chats/messages connection is successful (status 200)': (r) => r.status === 200,
        'POST /chats/messages response is SSE': (r) => r.headers['Content-Type'] === 'text/event-stream',
    });

    sleep(2); // Wait for 2 seconds before the next iteration.
}