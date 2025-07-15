const MAX_MESSAGES = 20;
const ul = document.getElementById("messages");

// Add a single message, trim to MAX_MESSAGES, then auto-scroll
function addMessage(m) {
  // Create list item
  const li = document.createElement("li");
  li.className = "p-2 bg-gray-50 rounded";
  const time = new Date(m.timestamp).toLocaleTimeString();
  li.textContent = `[${time}] ${m.username}: ${m.message}`;

  // Append and remove oldest if over limit
  ul.appendChild(li);
  if (ul.children.length > MAX_MESSAGES) {
    ul.removeChild(ul.firstElementChild);
  }

  // Scroll to bottom
  ul.scrollTop = ul.scrollHeight;
}

// Initial load of last 20 messages
async function loadInitial() {
  const res = await fetch("/api/messages");
  const msgs = await res.json();
  // Keep only the last MAX_MESSAGES
  const initial = msgs.slice(-MAX_MESSAGES);
  initial.forEach(addMessage);
}

// Example SSE setup (if implemented backend endpoint)
function subscribeLive() {
  const es = new EventSource("/api/stream");
  es.onmessage = e => {
    const msg = JSON.parse(e.data);
    addMessage(msg);
  };
}

// Kick it off
loadInitial();
// If you have SSE:
subscribeLive();
