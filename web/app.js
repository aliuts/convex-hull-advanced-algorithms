const $ = (selector) => document.querySelector(selector);
const canvas = $('#plane');
const ctx = canvas.getContext('2d');

const LABELS = { graham_scan: 'Graham Scan', jarvis_march: 'Jarvis March', chan: "Chan's Algorithm" };
const PAD = 12;    // pixels between the 0–1000 data range and the canvas edge
const PICK = 10;   // how close (in pixels) a right-click must be to remove a point

let points = [];
let hull = [];

// Data coordinates run 0–1000 on both axes, with y pointing up.
function toScreen(p) {
  const size = canvas.clientWidth - 2 * PAD;
  return { x: PAD + (p.x / 1000) * size, y: PAD + (1 - p.y / 1000) * size };
}

function toData(event) {
  const rect = canvas.getBoundingClientRect();
  const size = rect.width - 2 * PAD;
  const clamp = (v) => Math.min(1000, Math.max(0, v));
  return {
    x: clamp(((event.clientX - rect.left - PAD) / size) * 1000),
    y: clamp((1 - (event.clientY - rect.top - PAD) / size) * 1000),
  };
}

function draw() {
  const dpr = window.devicePixelRatio || 1;
  canvas.width = canvas.clientWidth * dpr;
  canvas.height = canvas.clientHeight * dpr;
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

  const style = getComputedStyle(document.documentElement);
  if (hull.length > 1) {
    ctx.beginPath();
    hull.map(toScreen).forEach((p, i) => (i ? ctx.lineTo(p.x, p.y) : ctx.moveTo(p.x, p.y)));
    ctx.closePath();
    ctx.strokeStyle = style.getPropertyValue('--hull');
    ctx.lineWidth = 2;
    ctx.stroke();
  }
  ctx.fillStyle = style.getPropertyValue('--point');
  for (const p of points.map(toScreen)) {
    ctx.beginPath();
    ctx.arc(p.x, p.y, 3, 0, 2 * Math.PI);
    ctx.fill();
  }
  $('#point-count').textContent = `${points.length} point${points.length === 1 ? '' : 's'}`;
}

// Any change to the points makes the old hull and timings stale.
function setPoints(next) {
  points = next;
  hull = [];
  $('#results').innerHTML = '';
  draw();
}

async function api(path, body) {
  const res = await fetch(path, body && {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Request failed');
  return data;
}

// Runs an action, showing any error and disabling the button while it works.
async function withButton(button, action) {
  $('#error').hidden = true;
  button.disabled = true;
  try {
    await action();
  } catch (err) {
    $('#error').textContent = err.message;
    $('#error').hidden = false;
  } finally {
    button.disabled = false;
  }
}

function formatTime(ns) {
  return ns < 1e6 ? `${(ns / 1e3).toFixed(1)} µs` : `${(ns / 1e6).toFixed(2)} ms`;
}

canvas.addEventListener('click', (e) => setPoints([...points, toData(e)]));

canvas.addEventListener('contextmenu', (e) => {
  e.preventDefault();
  const rect = canvas.getBoundingClientRect();
  const click = { x: e.clientX - rect.left, y: e.clientY - rect.top };
  const i = points.findIndex((p) => {
    const s = toScreen(p);
    return Math.hypot(s.x - click.x, s.y - click.y) < PICK;
  });
  if (i >= 0) setPoints(points.filter((_, j) => j !== i));
});

$('#generate').addEventListener('click', (e) => withButton(e.target, async () => {
  const data = await api('/api/dataset', {
    kind: $('#kind').value,
    count: Number($('#count').value),
    seed: Number($('#seed').value),
  });
  setPoints(data.points);
}));

$('#clear').addEventListener('click', () => setPoints([]));

$('#run').addEventListener('click', (e) => withButton(e.target, async () => {
  if (!points.length) throw new Error('Add or generate some points first.');
  const data = await api('/api/run', { points, algorithm: $('#algorithm').value });
  $('#results').innerHTML = data.results.map((r) => `
    <tr>
      <td>${LABELS[r.result.algorithm] || r.result.algorithm}</td>
      <td>${formatTime(r.durationNs)}</td>
      <td>${r.hullVertices}</td>
    </tr>`).join('');
  hull = data.results[0].result.hull || [];
  draw();
}));

window.addEventListener('resize', draw);

// Fill the algorithm picker from the server, so it always matches what's registered.
api('/api/algorithms').then(({ algorithms }) => {
  for (const name of algorithms) $('#algorithm').add(new Option(LABELS[name] || name, name));
});

draw();
