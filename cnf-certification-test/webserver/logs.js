import '@rhds/elements/rh-code-block/rh-code-block.js';

const socket = new WebSocket('ws://localhost:8084/logstream');
const code = document
  .getElementById('logs')
  .querySelector('rh-code-block');
code.textContent = '';
// Handle incoming log messages
socket.addEventListener('message', function (event) {
  code.innerHTML += event.data + '\n';
});