import '@rhds/elements/rh-code-block/rh-code-block.js';

const socket = new WebSocket('ws://localhost:8080/logstream');
const code = document
  .getElementById('logs')
  .querySelector('rh-code-block');
  console.log(code)
code.textContent = '';
// Handle incoming log messages
socket.addEventListener('message', function (event) {
  code.textContent += event.data + '\n';
  console.log(code.textContent)
});