 
let allRequests = [];
window.currentRequest = null;
let selectedReplayAttempt = null;
let activeMethodFilter = 'ALL';
let activeSourceFilter = 'all';
let activeFormat = 'json';

window.addEventListener('DOMContentLoaded', () => {
  lucide.createIcons();
  loadRequests();
  initPanelResizer();
  connectSSE((newReq) => {
    allRequests.unshift(newReq);
    renderList();
    if (!window.currentRequest) {
      selectRequest(newReq.id);
    }
  });

   window.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      closeSimulateModal();
      closeSnippetModal();
      closeEditReplayModal();
      closeClearModal();
    }
  });
});

function initPanelResizer() {
  const resizer = document.getElementById('panel-resizer');
  const panel = document.getElementById('bottom-replay-panel');
  const container = document.getElementById('active-detail');
  if (!resizer || !panel) return;

  function applySavedHeight() {
    const parent = container || panel.parentElement;
    const parentH = parent ? parent.clientHeight : window.innerHeight - 150;
    const savedHeight = localStorage.getItem('hooklet_replay_panel_height');
    if (savedHeight) {
      const h = parseInt(savedHeight, 10);
      const minH = 90;
      const maxH = Math.max(minH, parentH - 140);
      if (!isNaN(h) && h >= minH && h <= maxH) {
        panel.style.height = `${h}px`;
      } else if (h > maxH) {
        panel.style.height = `${maxH}px`;
      }
    }
  }

   applySavedHeight();

  resizer.addEventListener('mousedown', (e) => {
    e.preventDefault();
    const startY = e.clientY;
    const startHeight = panel.offsetHeight;
    const parent = container || panel.parentElement;
    const parentH = parent ? parent.clientHeight : window.innerHeight - 150;

    document.body.style.cursor = 'row-resize';
    document.body.style.userSelect = 'none';

    function onMouseMove(moveEvent) {
      const delta = startY - moveEvent.clientY;
      const minH = 90;
       const maxH = Math.max(minH, parentH - 140);
      const newHeight = Math.max(minH, Math.min(maxH, startHeight + delta));
      panel.style.height = `${newHeight}px`;
    }

    function onMouseUp() {
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
      document.removeEventListener('mousemove', onMouseMove);
      document.removeEventListener('mouseup', onMouseUp);
      localStorage.setItem('hooklet_replay_panel_height', panel.offsetHeight);
    }

    document.addEventListener('mousemove', onMouseMove);
    document.addEventListener('mouseup', onMouseUp);
  });
}

async function loadRequests() {
  try {
    const res = await fetch('/api/requests');
    if (res.ok) {
      allRequests = await res.json() || [];
      renderList();
      if (allRequests.length > 0 && !window.currentRequest) {
        selectRequest(allRequests[0].id);
      }
    }
  } catch (err) {
    console.error('Failed to load requests:', err);
  }
}

function renderList() {
  const listEl = document.getElementById('request-list');
  const filtered = getFilteredRequests();
  document.getElementById('webhook-count-pill').innerText = allRequests.length;

  if (filtered.length === 0) {
    listEl.innerHTML = `
      <div class="p-8 text-center text-xs text-slate-500 flex flex-col items-center">
        <i data-lucide="inbox" class="w-6 h-6 mb-2 text-slate-600"></i>
        <span>No matching webhooks</span>
      </div>
    `;
    lucide.createIcons();
    return;
  }

  listEl.innerHTML = filtered.map(req => {
    const isSelected = window.currentRequest && window.currentRequest.id === req.id;
    const timeStr = new Date(req.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    const sizeStr = formatBytes(req.content_length || (req.body ? atob(req.body).length : 0));
    const eventName = detectEventName(req);
    const replaysCount = req.replay_attempts ? req.replay_attempts.length : 0;
    const hasSignature = detectSignature(req);

    let methodBadgeClass = 'badge-post';
    if (req.method === 'GET') methodBadgeClass = 'badge-get';
    if (req.method === 'PUT') methodBadgeClass = 'badge-put';
    if (req.method === 'DELETE') methodBadgeClass = 'badge-delete';

    return `
      <div onclick="selectRequest('${req.id}')"
           class="group p-3.5 cursor-pointer transition flex flex-col gap-1.5 ${isSelected ? 'bg-canvas-850/90 border-l-2 border-blue-500' : 'hover:bg-canvas-850/40'}">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2 truncate">
            <span class="px-1.5 py-0.2 text-[10px] font-bold font-mono rounded ${methodBadgeClass}">${req.method}</span>
            <span class="text-xs font-mono font-medium text-slate-200 truncate">${req.path}</span>
          </div>
          <div class="flex items-center gap-1.5 flex-shrink-0">
            <span class="text-[11px] font-mono text-slate-500 group-hover:hidden">${timeStr}</span>
            <button onclick="deleteSingleRequest('${req.id}', event)" class="hidden group-hover:flex p-1 rounded hover:bg-rose-950 text-slate-500 hover:text-rose-400 transition" title="Delete this webhook">
              <i data-lucide="trash-2" class="w-3.5 h-3.5"></i>
            </button>
          </div>
        </div>

        <div class="flex items-center justify-between text-[11px] font-mono">
          <div class="flex items-center gap-1.5 text-slate-400 truncate max-w-[200px]">
            <i data-lucide="tag" class="w-3 h-3 text-slate-500"></i>
            <span class="truncate">${eventName}</span>
          </div>
          <span class="text-slate-500 text-[10px]">${sizeStr}</span>
        </div>

        <div class="flex items-center gap-1.5 mt-0.5 text-[10px] font-mono">
          ${replaysCount > 0 
            ? `<span class="px-1.5 py-0.2 rounded bg-canvas-800 text-blue-400 border border-blue-500/20">${replaysCount} replay${replaysCount > 1 ? 's' : ''}</span>`
            : `<span class="text-slate-600">Unplayed</span>`
          }
          ${hasSignature ? `<span class="px-1.5 py-0.2 rounded bg-canvas-800 text-amber-400 border border-amber-500/20">Signed</span>` : ''}
        </div>
      </div>
    `;
  }).join('');

  lucide.createIcons();
}

function getFilteredRequests() {
  const query = document.getElementById('search-filter').value.toLowerCase().trim();
  return allRequests.filter(req => {
    if (activeMethodFilter !== 'ALL' && req.method !== activeMethodFilter) return false;
    if (activeSourceFilter !== 'all' && !req.path.toLowerCase().includes(activeSourceFilter)) return false;
    if (query) {
      const inPath = req.path.toLowerCase().includes(query);
      const inMethod = req.method.toLowerCase().includes(query);
      const inEvent = detectEventName(req).toLowerCase().includes(query);
      return inPath || inMethod || inEvent;
    }
    return true;
  });
}

function detectEventName(req) {
  if (!req.headers) return 'webhook.received';
  if (req.headers['X-Github-Event']) return req.headers['X-Github-Event'][0];
  if (req.headers['X-Shopify-Topic']) return req.headers['X-Shopify-Topic'][0];

  try {
    const bodyStr = atob(req.body);
    const parsed = JSON.parse(bodyStr);
    if (parsed.type) return parsed.type;
    if (parsed.event) return parsed.event;
    if (parsed.action) return parsed.action;
  } catch (e) {}

  return req.path.replace(/^\/wh\//, '') || 'incoming.event';
}

function detectSignature(req) {
  if (!req.headers) return false;
  const keys = Object.keys(req.headers).map(k => k.toLowerCase());
  return keys.some(k => k.includes('signature') || k.includes('sig') || k.includes('hmac'));
}

async function selectRequest(id) {
  try {
    const res = await fetch('/api/requests/' + id);
    if (!res.ok) return;
    window.currentRequest = await res.json();
    renderList();
    renderActiveWorkspace();
  } catch (err) {
    console.error('Failed to select request:', err);
  }
}
window.selectRequest = selectRequest;

function renderActiveWorkspace() {
  const req = window.currentRequest;
  if (!req) {
    document.getElementById('empty-state').classList.remove('hidden');
    document.getElementById('active-detail').classList.add('hidden');
    return;
  }

  document.getElementById('empty-state').classList.add('hidden');
  document.getElementById('active-detail').classList.remove('hidden');

  document.getElementById('detail-method-badge').innerText = req.method;
  document.getElementById('detail-path').innerText = req.path;
  document.getElementById('detail-id-short').innerText = req.id.substring(0, 8) + '...';
  document.getElementById('detail-time').innerText = new Date(req.created_at).toLocaleString();
  document.getElementById('detail-ip').innerText = req.remote_addr || '127.0.0.1';

  const rawBytes = req.body ? atob(req.body) : '';
  const sizeStr = formatBytes(rawBytes.length);
  document.getElementById('detail-size').innerText = sizeStr;
  document.getElementById('tab-payload-size').innerText = sizeStr;

  const globalTarget = document.getElementById('global-target').value;
  document.getElementById('detail-target').value = globalTarget;

  renderPayloadTab(rawBytes);
  renderHeadersTab();
  renderQueryTab();
  renderSignatureTab();
  renderReplayHistory();

  lucide.createIcons();
}

function renderPayloadTab(rawBytes) {
  const el = document.getElementById('payload-content');
  if (!rawBytes) {
    el.innerHTML = '<span class="text-slate-500 italic">/* Empty body */</span>';
    return;
  }

  if (activeFormat === 'json') {
    try {
      const parsed = JSON.parse(rawBytes);
      el.innerHTML = syntaxHighlightJson(parsed);
    } catch (e) {
      el.innerText = rawBytes;
    }
  } else {
    el.innerText = rawBytes;
  }
}

function renderHeadersTab() {
  const tbody = document.getElementById('headers-tbody');
  const headers = window.currentRequest.headers || {};
  const entries = Object.entries(headers);
  document.getElementById('tab-headers-count').innerText = entries.length;

  if (entries.length === 0) {
    tbody.innerHTML = '<tr><td colspan="2" class="p-4 text-center text-slate-500 italic">No headers captured</td></tr>';
    return;
  }

  tbody.innerHTML = entries.map(([key, vals]) => `
    <tr class="hover:bg-canvas-850/50 transition">
      <td class="py-2 px-4 text-blue-400 font-medium">${escapeHtml(key)}</td>
      <td class="py-2 px-4 text-slate-300 break-all">${escapeHtml(vals.join(', '))}</td>
    </tr>
  `).join('');
}

function renderQueryTab() {
  document.getElementById('query-content').innerText = window.currentRequest.query_string || 'None (no query parameters)';
  document.getElementById('remote-addr-content').innerText = window.currentRequest.remote_addr || '127.0.0.1';
  document.getElementById('content-type-content').innerText = window.currentRequest.content_type || 'None';
}

function renderSignatureTab() {
  const el = document.getElementById('signature-content');
  const headers = window.currentRequest.headers || {};
  const sigHeaders = Object.entries(headers).filter(([k]) => {
    const lower = k.toLowerCase();
    return lower.includes('signature') || lower.includes('sig') || lower.includes('hmac');
  });

  if (sigHeaders.length === 0) {
    el.innerHTML = '<span class="text-slate-500 italic">No cryptographic signature headers found in this request.</span>';
    return;
  }

  el.innerHTML = sigHeaders.map(([k, vals]) => `
    <div class="mb-2">
      <span class="text-amber-400 font-bold block">${escapeHtml(k)}:</span>
      <span class="text-slate-300 break-all">${escapeHtml(vals.join(', '))}</span>
    </div>
  `).join('');
}

function renderReplayHistory() {
  const listEl = document.getElementById('replay-runs-list');
  const replays = window.currentRequest.replay_attempts || [];
  document.getElementById('replay-runs-count').innerText = `${replays.length} run${replays.length !== 1 ? 's' : ''}`;

  if (replays.length === 0) {
    listEl.innerHTML = '<div class="text-slate-500 text-xs italic text-center py-8">No replays triggered yet. Click "REPLAY REQUEST" above.</div>';
    document.getElementById('target-response-body').innerText = 'Select a replay run on the left to view target server response.';
    return;
  }

  listEl.innerHTML = replays.map((att, idx) => {
    const is2xx = att.status_code >= 200 && att.status_code < 300;
    const timeStr = new Date(att.created_at).toLocaleTimeString();
    const attemptNum = replays.length - idx;
    const isSelected = selectedReplayAttempt && selectedReplayAttempt.id === att.id;

    return `
      <div onclick="selectReplayAttempt('${att.id}')"
           class="p-2.5 rounded-lg border cursor-pointer transition flex flex-col gap-1 text-xs font-mono ${isSelected ? 'bg-canvas-800 border-blue-500' : 'bg-canvas-950/80 border-canvas-800 hover:border-canvas-750'}">
        <div class="flex items-center justify-between">
          <span class="px-1.5 py-0.2 rounded font-bold text-[10px] ${is2xx ? 'bg-emerald-950 text-emerald-400 border border-emerald-800' : 'bg-rose-950 text-rose-400 border border-rose-800'}">
            ${att.status_code || 'ERR'}
          </span>
          <span class="text-slate-500 text-[10px] flex items-center gap-1">
            <i data-lucide="clock" class="w-3 h-3"></i>
            <span>${att.latency_ms}ms</span>
          </span>
        </div>
        <div class="text-slate-300 text-[11px] truncate">${escapeHtml(att.target_url)}</div>
        <div class="flex items-center justify-between text-[10px] text-slate-500">
          <span>${timeStr}</span>
          <span>Attempt #${attemptNum}</span>
        </div>
      </div>
    `;
  }).join('');

  if (!selectedReplayAttempt && replays.length > 0) {
    selectReplayAttempt(replays[replays.length - 1].id);
  } else if (selectedReplayAttempt) {
    selectReplayAttempt(selectedReplayAttempt.id);
  }

  lucide.createIcons();
}

function selectReplayAttempt(id) {
  const replays = window.currentRequest.replay_attempts || [];
  selectedReplayAttempt = replays.find(a => a.id === id) || replays[replays.length - 1];
  if (!selectedReplayAttempt) return;

  const bodyEl = document.getElementById('target-response-body');
  if (selectedReplayAttempt.error) {
    bodyEl.innerHTML = `<span class="text-rose-400 font-bold">Network Connection Error:</span>\n${escapeHtml(selectedReplayAttempt.error)}`;
    return;
  }

  const raw = selectedReplayAttempt.response_body ? atob(selectedReplayAttempt.response_body) : '';
  try {
    const parsed = JSON.parse(raw);
    bodyEl.innerHTML = syntaxHighlightJson(parsed);
  } catch (e) {
    bodyEl.innerText = raw || '/* Empty 200 OK Response */';
  }
}

async function triggerReplay() {
  if (!window.currentRequest) return;
  const btn = document.getElementById('btn-replay');
  const originalHtml = btn.innerHTML;
  btn.disabled = true;
  btn.innerHTML = `<i data-lucide="loader-2" class="w-3.5 h-3.5 animate-spin"></i><span>DISPATCHING...</span>`;
  lucide.createIcons();

  const target = document.getElementById('detail-target').value || document.getElementById('global-target').value;
  try {
    await fetch(`/api/requests/${window.currentRequest.id}/replay?target=${encodeURIComponent(target)}`, {
      method: 'POST'
    });
    showToast('Replay dispatched to target!');
  } catch (err) {
    console.error('Replay failed:', err);
    showToast('Network error during replay');
  } finally {
    await selectRequest(window.currentRequest.id);
    btn.disabled = false;
    btn.innerHTML = originalHtml;
    lucide.createIcons();
  }
}

function switchTab(tab) {
  ['payload', 'headers', 'query', 'signature'].forEach(t => {
    document.getElementById('tab-' + t).classList.add('hidden');
    const btn = document.getElementById('tab-btn-' + t);
    btn.classList.remove('border-blue-500', 'text-blue-400');
    btn.classList.add('border-transparent', 'text-slate-400');
  });

  document.getElementById('tab-' + tab).classList.remove('hidden');
  const activeBtn = document.getElementById('tab-btn-' + tab);
  activeBtn.classList.remove('border-transparent', 'text-slate-400');
  activeBtn.classList.add('border-blue-500', 'text-blue-400');
}

function setFormat(fmt) {
  activeFormat = fmt;
  document.getElementById('fmt-json').className = fmt === 'json' ? 'px-2 py-0.5 rounded bg-blue-600 text-white font-medium' : 'px-2 py-0.5 rounded text-slate-400 hover:text-slate-200';
  document.getElementById('fmt-raw').className = fmt === 'raw' ? 'px-2 py-0.5 rounded bg-blue-600 text-white font-medium' : 'px-2 py-0.5 rounded text-slate-400 hover:text-slate-200';
  if (window.currentRequest) {
    renderPayloadTab(window.currentRequest.body ? atob(window.currentRequest.body) : '');
  }
}

function setMethodFilter(method) {
  activeMethodFilter = method;
  ['ALL', 'POST', 'GET', 'PUT'].forEach(m => {
    const btn = document.getElementById('filter-' + m.toLowerCase());
    if (m === method) {
      btn.className = 'px-2 py-0.5 rounded bg-blue-600 text-white font-medium';
    } else {
      btn.className = 'px-2 py-0.5 rounded bg-canvas-800 text-slate-400 hover:text-slate-200';
    }
  });
  renderList();
}

function setSourceFilter(source) {
  activeSourceFilter = source;
  ['all', 'github', 'stripe', 'shopify'].forEach(s => {
    const btn = document.getElementById('source-' + s);
    if (s === source) {
      btn.className = 'px-2 py-0.5 rounded bg-blue-600 text-white font-medium';
    } else {
      btn.className = 'px-2 py-0.5 rounded bg-canvas-800 text-slate-400 hover:text-slate-200';
    }
  });
  renderList();
}

function applyFilters() {
  renderList();
}

function copyIngressUrl() {
  const url = window.location.origin + '/wh/';
  navigator.clipboard.writeText(url);
  showToast('Ingress URL copied: ' + url);
}

function copyCurrentID() {
  if (!window.currentRequest) return;
  navigator.clipboard.writeText(window.currentRequest.id);
  showToast('Webhook ID copied');
}

function copyPayload() {
  const content = document.getElementById('payload-content').innerText;
  navigator.clipboard.writeText(content);
  showToast('Payload copied to clipboard');
}

function copyTargetResponse() {
  const content = document.getElementById('target-response-body').innerText;
  navigator.clipboard.writeText(content);
  showToast('Target response copied to clipboard');
}

function exportRequestsJSON() {
  const dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(allRequests, null, 2));
  const downloadAnchor = document.createElement('a');
  downloadAnchor.setAttribute("href", dataStr);
  downloadAnchor.setAttribute("download", "hooklet-webhooks.json");
  document.body.appendChild(downloadAnchor);
  downloadAnchor.click();
  downloadAnchor.remove();
  showToast('Exported ' + allRequests.length + ' webhooks as JSON');
}

function clearAllRequests() {
  const modal = document.getElementById('modal-clear-confirm');
  const countEl = document.getElementById('clear-modal-count');
  if (countEl) {
    countEl.innerText = `${allRequests.length} webhook${allRequests.length === 1 ? '' : 's'} will be deleted`;
  }
  if (modal) {
    modal.classList.remove('hidden');
    lucide.createIcons();
  }
}

function closeClearModal() {
  const modal = document.getElementById('modal-clear-confirm');
  if (modal) modal.classList.add('hidden');
}

async function executeClearRequests() {
  const btn = document.getElementById('btn-confirm-delete');
  const originalHtml = btn ? btn.innerHTML : '';
  if (btn) {
    btn.disabled = true;
    btn.innerHTML = `<i data-lucide="loader-2" class="w-3.5 h-3.5 animate-spin"></i><span>Clearing...</span>`;
    lucide.createIcons();
  }

  try {
    const res = await fetch('/api/requests', { method: 'DELETE' });
    if (res.ok) {
      allRequests = [];
      window.currentRequest = null;
      renderList();
      renderActiveWorkspace();
      closeClearModal();
      showToast('All captured webhooks and replays cleared');
    } else {
      alert('Failed to clear: ' + res.statusText);
    }
  } catch (e) {
    alert('Failed to clear: ' + e);
  } finally {
    if (btn) {
      btn.disabled = false;
      btn.innerHTML = originalHtml;
      lucide.createIcons();
    }
  }
}

async function deleteSingleRequest(id, event) {
  if (event) event.stopPropagation();

  const idx = allRequests.findIndex(r => r.id === id);
  if (idx === -1) return;
  const deletedReq = allRequests[idx];
  allRequests.splice(idx, 1);

  // If the deleted webhook was currently inspected, select the next available one
  if (window.currentRequest && window.currentRequest.id === id) {
    if (allRequests.length > 0) {
      const nextIdx = Math.min(idx, allRequests.length - 1);
      selectRequest(allRequests[nextIdx].id);
    } else {
      window.currentRequest = null;
      renderActiveWorkspace();
    }
  }
  renderList();

  try {
    const res = await fetch('/api/requests/' + id, { method: 'DELETE' });
    if (res.ok) {
      showToast(`Deleted ${deletedReq.method} ${deletedReq.path}`);
    } else {
      showToast('Failed to delete webhook');
    }
  } catch (err) {
    console.error('Delete request failed:', err);
    showToast('Network error deleting webhook');
  }
}

function deleteCurrentRequest() {
  if (!window.currentRequest) return;
  deleteSingleRequest(window.currentRequest.id);
}

function formatBytes(bytes) {
  if (!bytes || bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

function escapeHtml(str) {
  if (!str) return '';
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}
