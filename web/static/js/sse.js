 
function connectSSE(onNewRequest) {
  const source = new EventSource('/api/events');

  source.onmessage = (event) => {
    try {
      const req = JSON.parse(event.data);
      onNewRequest(req);
      showToast(`Captured ${req.method} ${req.path}`);
    } catch (err) {
      console.error('Error parsing SSE event:', err);
    }
  };

  source.onerror = () => {
    const el = document.getElementById('live-indicator');
    if (el) el.innerText = 'RECONNECTING...';
  };

  source.onopen = () => {
    const el = document.getElementById('live-indicator');
    if (el) el.innerText = 'LIVE CAPTURE';
  };
}

function showToast(msg) {
  const toast = document.getElementById('toast');
  const msgEl = document.getElementById('toast-message');
  if (!toast || !msgEl) return;
  msgEl.innerText = msg;
  toast.classList.remove('translate-y-16', 'opacity-0');
  setTimeout(() => {
    toast.classList.add('translate-y-16', 'opacity-0');
  }, 3000);
}
