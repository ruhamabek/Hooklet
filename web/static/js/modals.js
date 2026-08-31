 
const PRESETS = {
  github_push: {
    method: 'POST',
    path: '/wh/github',
    headers: {
      "Content-Type": "application/json",
      "User-Agent": "GitHub-Hookshot/908b1a",
      "X-GitHub-Event": "push",
      "X-Hub-Signature-256": "sha256=d57c2927232047225965e378b309b556529"
    },
    payload: {
      "ref": "refs/heads/main",
      "before": "a1b2c3d4e5f6",
      "after": "c0ffee998877",
      "repository": { "id": 81828471, "name": "Hooklet", "full_name": "sapphire/Hooklet" },
      "pusher": { "name": "sapphire", "email": "dev@hooklet.local" },
      "commits": [
        { "id": "c0ffee", "message": "feat(engine): resolve buffered channel deadlock on rapid replays", "author": { "name": "sapphire" } }
      ]
    }
  },
  github_pr: {
    method: 'POST',
    path: '/wh/github',
    headers: {
      "Content-Type": "application/json",
      "X-GitHub-Event": "pull_request",
      "X-Hub-Signature-256": "sha256=fa77b391fcc699026da66928e08d6692289f81a17c"
    },
    payload: {
      "action": "opened",
      "number": 42,
      "pull_request": {
        "id": 994812,
        "title": "fix(engine): resolve buffered channel deadlock on rapid replays",
        "state": "open",
        "user": { "login": "alex-engineer" }
      },
      "repository": { "name": "Hooklet", "full_name": "sapphire/Hooklet" }
    }
  },
  stripe_payment: {
    method: 'POST',
    path: '/wh/stripe',
    headers: {
      "Content-Type": "application/json",
      "Stripe-Signature": "t=1690000000,v1=5257a869e7ecebeda32affa62cd492316e6"
    },
    payload: {
      "id": "evt_3Mv0002eZvKYlo2C",
      "object": "event",
      "type": "payment_intent.succeeded",
      "data": {
        "object": {
          "id": "pi_3Mv0002eZvKYlo2C",
          "amount": 14900,
          "currency": "usd",
          "customer": "cus_sapphire_enterprise",
          "status": "succeeded"
        }
      }
    }
  },
  shopify_order: {
    method: 'POST',
    path: '/wh/shopify',
    headers: {
      "Content-Type": "application/json",
      "X-Shopify-Topic": "orders/create",
      "X-Shopify-Hmac-Sha256": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY="
    },
    payload: {
      "id": 450789469,
      "email": "customer@example.com",
      "total_price": "249.50",
      "line_items": [
        { "id": 1, "title": "Self-hosted Webhook Pro License", "price": "249.50", "quantity": 1 }
      ]
    }
  },
  slack_action: {
    method: 'POST',
    path: '/wh/slack',
    headers: {
      "Content-Type": "application/json",
      "X-Slack-Signature": "v0=a2114d57b48eac39b9ad189dd831627a4148a31f"
    },
    payload: {
      "type": "block_actions",
      "user": { "id": "U012AB3CD", "username": "sapphire" },
      "actions": [
        { "action_id": "approve_deployment", "value": "release-v0.1" }
      ]
    }
  }
};

let currentSnippetLang = 'curl';

 function openSimulateModal() {
  selectPreset('github_push');
  setSimHeadersMode('highlight');
  setSimPayloadMode('highlight');
  document.getElementById('modal-simulate').classList.remove('hidden');
  lucide.createIcons();
}

function closeSimulateModal() {
  document.getElementById('modal-simulate').classList.add('hidden');
}

function setSimHeadersMode(mode) {
  const preview = document.getElementById('sim-headers-preview');
  const textarea = document.getElementById('sim-headers');
  const btnHl = document.getElementById('sim-headers-mode-hl');
  const btnEdit = document.getElementById('sim-headers-mode-edit');
  if (!preview || !textarea) return;

  if (mode === 'highlight') {
    updateSimHeadersPreview();
    preview.classList.remove('hidden');
    textarea.classList.add('hidden');
    if (btnHl) btnHl.className = 'px-2 py-0.5 rounded bg-blue-600 text-white font-medium';
    if (btnEdit) btnEdit.className = 'px-2 py-0.5 rounded text-slate-400 hover:text-slate-200';
  } else {
    preview.classList.add('hidden');
    textarea.classList.remove('hidden');
    if (btnHl) btnHl.className = 'px-2 py-0.5 rounded text-slate-400 hover:text-slate-200';
    if (btnEdit) btnEdit.className = 'px-2 py-0.5 rounded bg-blue-600 text-white font-medium';
    textarea.focus();
  }
}

function updateSimHeadersPreview() {
  const textarea = document.getElementById('sim-headers');
  const preview = document.getElementById('sim-headers-preview');
  if (!textarea || !preview) return;
  const val = textarea.value.trim();
  const code = preview.querySelector('code');
  if (!code) return;
  try {
    const parsed = JSON.parse(val);
    code.textContent = JSON.stringify(parsed, null, 2);
  } catch (e) {
    code.textContent = val;
  }
  if (window.Prism) {
    Prism.highlightElement(code);
  }
}

function setSimPayloadMode(mode) {
  const preview = document.getElementById('sim-payload-preview');
  const textarea = document.getElementById('sim-payload');
  const btnHl = document.getElementById('sim-payload-mode-hl');
  const btnEdit = document.getElementById('sim-payload-mode-edit');
  if (!preview || !textarea) return;

  if (mode === 'highlight') {
    updateSimPayloadPreview();
    preview.classList.remove('hidden');
    textarea.classList.add('hidden');
    if (btnHl) btnHl.className = 'px-2 py-0.5 rounded bg-blue-600 text-white font-medium';
    if (btnEdit) btnEdit.className = 'px-2 py-0.5 rounded text-slate-400 hover:text-slate-200';
  } else {
    preview.classList.add('hidden');
    textarea.classList.remove('hidden');
    if (btnHl) btnHl.className = 'px-2 py-0.5 rounded text-slate-400 hover:text-slate-200';
    if (btnEdit) btnEdit.className = 'px-2 py-0.5 rounded bg-blue-600 text-white font-medium';
    textarea.focus();
  }
}

function updateSimPayloadPreview() {
  const textarea = document.getElementById('sim-payload');
  const preview = document.getElementById('sim-payload-preview');
  if (!textarea || !preview) return;
  const val = textarea.value.trim();
  const code = preview.querySelector('code');
  if (!code) return;
  try {
    const parsed = JSON.parse(val);
    code.textContent = JSON.stringify(parsed, null, 2);
  } catch (e) {
    code.textContent = val;
  }
  if (window.Prism) {
    Prism.highlightElement(code);
  }
}

function selectPreset(name) {
  const preset = PRESETS[name];
  if (!preset) return;

  Object.keys(PRESETS).forEach(k => {
    const card = document.getElementById('preset-' + k);
    if (card) {
      card.className = k === name 
        ? 'p-2.5 rounded-lg border border-blue-500 bg-blue-950/30 text-left transition flex flex-col gap-1'
        : 'p-2.5 rounded-lg border border-canvas-800 bg-canvas-950 text-left transition hover:border-canvas-700 flex flex-col gap-1';
    }
  });

  document.getElementById('sim-method').value = preset.method;
  document.getElementById('sim-path').value = preset.path;
  document.getElementById('sim-headers').value = JSON.stringify(preset.headers, null, 2);
  document.getElementById('sim-payload').value = JSON.stringify(preset.payload, null, 2);
  updateSimHeadersPreview();
  updateSimPayloadPreview();
}

function formatSimPayload() {
  try {
    const val = document.getElementById('sim-payload').value;
    document.getElementById('sim-payload').value = JSON.stringify(JSON.parse(val), null, 2);
    updateSimPayloadPreview();
  } catch (e) {}
}

async function dispatchSimulatedWebhook() {
  const path = document.getElementById('sim-path').value.trim();
  const method = document.getElementById('sim-method').value.trim();
  let headers = {};
  try {
    headers = JSON.parse(document.getElementById('sim-headers').value);
  } catch (e) {
    alert('Invalid JSON in Headers');
    return;
  }
  const body = document.getElementById('sim-payload').value;

  try {
    const res = await fetch(path, {
      method: method,
      headers: headers,
      body: body
    });
    if (res.ok) {
      closeSimulateModal();
      showToast('Webhook dispatched into live stream!');
    } else {
      alert('Dispatch failed: ' + res.statusText);
    }
  } catch (err) {
    alert('Error dispatching simulated webhook: ' + err);
  }
}

 function openSnippetModal() {
  if (!window.currentRequest) return;
  setSnippetLang('curl');
  document.getElementById('modal-snippet').classList.remove('hidden');
  lucide.createIcons();
}

function closeSnippetModal() {
  document.getElementById('modal-snippet').classList.add('hidden');
}

function setSnippetLang(lang) {
  currentSnippetLang = lang;
  ['curl', 'js', 'python', 'go'].forEach(l => {
    const tab = document.getElementById('snip-tab-' + l);
    if (!tab) return;
    tab.className = l === lang ? 'px-3 py-1.5 rounded-t font-medium bg-blue-600 text-white' : 'px-3 py-1.5 rounded-t font-medium text-slate-400 hover:text-white';
  });

  const req = window.currentRequest;
  if (!req) return;
  const fullUrl = window.location.origin + req.path;
  const rawBody = req.body ? atob(req.body) : '';
  const headers = req.headers || {};
  const codeBox = document.getElementById('snippet-code');
  if (!codeBox) return;

  let snippet = '';
  let prismLang = 'bash';

  if (lang === 'curl') {
    prismLang = 'bash';
    let headersStr = Object.entries(headers).map(([k, v]) => `  -H "${k}: ${v.join(', ')}" \\`).join('\n');
    let dataPart = rawBody ? `  -d '${rawBody.replace(/'/g, "'\\''")}'` : '';
    snippet = `curl -X ${req.method} "${fullUrl}" \\\n${headersStr}\n${dataPart}`.trim();
  } else if (lang === 'js') {
    prismLang = 'javascript';
    let formattedBody = rawBody;
    try {
      formattedBody = JSON.stringify(JSON.parse(rawBody), null, 2);
    } catch(e) {}
    snippet = `// JavaScript (Fetch API)\nconst url = "${fullUrl}";\nconst headers = ${JSON.stringify(headers, null, 2)};\nconst payload = ${formattedBody || '""'};\n\nfetch(url, {\n  method: "${req.method}",\n  headers: headers,\n  body: typeof payload === "object" ? JSON.stringify(payload) : payload\n})\n  .then(response => response.json())\n  .then(data => console.log("Response:", data))\n  .catch(error => console.error("Error:", error));`;
  } else if (lang === 'python') {
    prismLang = 'python';
    let formattedBody = rawBody;
    try {
      formattedBody = JSON.stringify(JSON.parse(rawBody), null, 2);
    } catch(e) {}
    snippet = `# Python 3 (Requests)\nimport requests\nimport json\n\nurl = "${fullUrl}"\nheaders = ${JSON.stringify(headers, null, 2)}\npayload = ${formattedBody || '""'}\n\nresponse = requests.${req.method.toLowerCase()}(\n    url,\n    headers=headers,\n    json=payload if isinstance(payload, dict) else None,\n    data=payload if not isinstance(payload, dict) else None\n)\n\nprint(f"Status Code: {response.status_code}")\nprint(response.text)`;
  } else if (lang === 'go') {
    prismLang = 'go';
    snippet = `// Go (net/http)\npackage main\n\nimport (\n\t"bytes"\n\t"fmt"\n\t"io"\n\t"net/http"\n)\n\nfunc main() {\n\turl := "${fullUrl}"\n\tpayload := []byte(\`${rawBody}\`)\n\n\treq, err := http.NewRequest("${req.method}", url, bytes.NewReader(payload))\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\n`;
    for (const [k, vals] of Object.entries(headers)) {
      snippet += `\treq.Header.Set("${k}", "${vals.join(', ')}")\n`;
    }
    snippet += `\n\tresp, err := http.DefaultClient.Do(req)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\tdefer resp.Body.Close()\n\n\tbody, _ := io.ReadAll(resp.Body)\n\tfmt.Printf("Status: %d\\nBody: %s\\n", resp.StatusCode, string(body))\n}`;
  }

  codeBox.className = `language-${prismLang}`;
  codeBox.textContent = snippet;
  if (window.Prism) {
    Prism.highlightElement(codeBox);
  }
}

function copySnippetCode() {
  const code = document.getElementById('snippet-code').textContent;
  navigator.clipboard.writeText(code);
  showToast('Code snippet copied to clipboard');
}

 function openEditReplayModal() {
  if (!window.currentRequest) return;
  document.getElementById('edit-modal-path-badge').innerText = window.currentRequest.path;
  document.getElementById('edit-method').value = window.currentRequest.method;
  document.getElementById('edit-target-url').value = document.getElementById('detail-target').value;

  const container = document.getElementById('edit-headers-container');
  container.innerHTML = '';
  const headers = window.currentRequest.headers || {};
  Object.entries(headers).forEach(([k, v]) => {
    addEditHeaderRow(k, v.join(', '));
  });

  const raw = window.currentRequest.body ? atob(window.currentRequest.body) : '';
  try {
    document.getElementById('edit-payload-body').value = JSON.stringify(JSON.parse(raw), null, 2);
  } catch (e) {
    document.getElementById('edit-payload-body').value = raw;
  }

  document.getElementById('modal-edit-replay').classList.remove('hidden');
  lucide.createIcons();
}

function closeEditReplayModal() {
  document.getElementById('modal-edit-replay').classList.add('hidden');
}

function addEditHeaderRow(key = '', val = '') {
  const container = document.getElementById('edit-headers-container');
  const row = document.createElement('div');
  row.className = 'flex items-center gap-2 mb-1';
  row.innerHTML = `
    <input type="text" placeholder="Header Key" value="${escapeHtml(key)}" class="flex-1 bg-canvas-950 border border-canvas-800 rounded px-2 py-1 text-xs text-blue-400 header-key font-mono" />
    <input type="text" placeholder="Header Value" value="${escapeHtml(val)}" class="flex-1 bg-canvas-950 border border-canvas-800 rounded px-2 py-1 text-xs text-slate-200 header-val font-mono" />
    <button onclick="this.parentElement.remove()" class="p-1 text-slate-500 hover:text-rose-400 transition" title="Delete header">
      <i data-lucide="trash-2" class="w-3.5 h-3.5"></i>
    </button>
  `;
  container.appendChild(row);
  lucide.createIcons();
}

function formatEditBody() {
  try {
    const val = document.getElementById('edit-payload-body').value;
    document.getElementById('edit-payload-body').value = JSON.stringify(JSON.parse(val), null, 2);
  } catch (e) {}
}

async function submitCustomReplay() {
  const target = document.getElementById('edit-target-url').value;
  const method = document.getElementById('edit-method').value;
  const body = document.getElementById('edit-payload-body').value;

  const headers = {};
  document.querySelectorAll('#edit-headers-container .flex').forEach(row => {
    const k = row.querySelector('.header-key').value.trim();
    const v = row.querySelector('.header-val').value.trim();
    if (k) {
      headers[k] = [v];
    }
  });

  const payload = {
    target: target,
    method: method,
    headers: headers,
    body: body
  };

  try {
   await fetch(`/api/requests/${window.currentRequest.id}/replay`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    closeEditReplayModal();
    showToast('Custom Replay executed!');
    if (window.selectRequest) {
      window.selectRequest(window.currentRequest.id);
    }
  } catch (err) {
    alert('Custom replay failed: ' + err);
  }
}
