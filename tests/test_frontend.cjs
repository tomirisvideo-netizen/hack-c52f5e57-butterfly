const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const { test } = require('node:test');

// Calendar dates must remain the same west of UTC as well as east of it.
process.env.TZ = 'America/Los_Angeles';
const source = fs.readFileSync(path.join(__dirname, '../frontend/app.js'), 'utf8');

function workspace(fetch, { storageUnavailable = false } = {}) {
  const controls = [{ disabled: false }, { disabled: false }];
  const downloads = [];
  const elements = new Map(['#app', '#toast', '#settings-dialog', '#settings-form-error', '#meeting-dialog', '#meeting-form-error'].map((selector) => [selector, {
    innerHTML: '', textContent: '', className: '', closed: false,
    querySelectorAll: () => controls,
    close() { this.closed = true; },
  }]));
  const listeners = new Map();
  class FormData {
    constructor(form) { this.values = new Map(Object.entries(form?.fields || {})); }
    get(name) { return this.values.get(name); }
    delete(name) { this.values.delete(name); }
  }
  class BrowserURL extends URL {
    static createObjectURL() { return 'blob:download'; }
    static revokeObjectURL() {}
  }
  const context = vm.createContext({
    console, Date, URL: BrowserURL, AbortController, FormData, File: class File {}, fetch,
    setTimeout: () => 1, clearTimeout: () => {},
    localStorage: {
      getItem: () => '',
      setItem: () => { if (storageUnavailable) throw new Error('Storage disabled'); },
    },
    document: {
      hidden: false,
      querySelector: (selector) => elements.get(selector) || null,
      querySelectorAll: () => [],
      addEventListener: (name, listener) => listeners.set(name, listener),
      createElement: () => ({ click() { downloads.push({ href: this.href, name: this.download }); }, remove() {} }),
      body: { appendChild() {} },
    },
    window: { addEventListener: () => {} },
  });
  // Exercise the browser script without bootstrapping an HTTP session.
  const script = source.replace(/  connect\(\);\s*\}\)\(\);\s*$/, '  globalThis.app = { state, formatDate, flyDemoPage, loadMeeting, saveDraft, actionCard, connect, createDemo, exportMeeting };\n})();');
  assert.notEqual(script, source, 'frontend bootstrap should be available to the test harness');
  vm.runInContext(script, context, { filename: 'frontend/app.js' });
  return {
    ...context.app, controls, elements, downloads,
    click: (action, data = {}) => listeners.get('click')({ target: { closest: () => ({ dataset: { action, ...data } }) } }),
    submit: (fields) => listeners.get('submit')({ preventDefault() {}, target: { id: 'settings-form', fields } }),
    submitMeeting: (fields) => listeners.get('submit')({ preventDefault() {}, target: { id: 'new-meeting-form', fields, querySelector: () => ({ disabled: false }) } }),
  };
}

function meeting(id = 'one') {
  return { id, title: id, date: '2026-10-07', status: 'needs_review', summary: 'Summary', actions: [], transcript: [], events: [] };
}

function ready(app) {
  app.state.loading = false;
  app.state.selected = meeting();
  app.state.draft = meeting();
  app.state.dirty = true;
}

function response(body) { return { ok: true, status: 200, json: async () => body }; }

test('meeting calendar dates do not move to the previous day in western time zones', () => {
  const app = workspace();
  assert.equal(app.formatDate('2026-10-07', { day: 'numeric', month: 'numeric', year: 'numeric' }), '07.10.2026');
  assert.equal(app.formatDate('2026-10-07T00:30:00Z', { day: 'numeric', month: 'numeric', year: 'numeric' }), '06.10.2026');
});

test('returning home cancels a pending meeting load and removes the loading screen', async () => {
  let resolve;
  const app = workspace(() => new Promise((done) => { resolve = done; }));
  app.state.loading = false;
  const loading = app.loadMeeting('one');
  assert.equal(app.state.loadingMeeting, true);
  await app.click('home');
  assert.equal(app.state.loadingMeeting, false);
  assert.match(app.elements.get('#app').innerHTML, /overview-page/);
  resolve(response(meeting()));
  await loading;
  assert.equal(app.state.selected, null);
});

test('saving locks editing and navigation and rejects duplicate writes until completion', async () => {
  let resolve;
  const calls = [];
  const app = workspace((url, options) => { calls.push({ url, options }); return new Promise((done) => { resolve = done; }); });
  ready(app);
  const saving = app.saveDraft();
  assert.equal(app.state.busy, true);
  assert.ok(app.controls.every((control) => control.disabled));
  assert.equal(await app.saveDraft(), false);
  await app.click('home');
  assert.equal(app.state.selected.id, 'one');
  assert.equal(calls.length, 1);
  resolve(response(meeting()));
  assert.equal(await saving, true);
  assert.equal(app.state.busy, false);
  assert.equal(app.state.dirty, false);
});

test('a failed save unlocks controls and retains the draft for retry', async () => {
  const app = workspace(async () => ({ ok: false, status: 503, json: async () => ({ detail: 'Try again' }) }));
  ready(app);
  app.state.draft.summary = 'Keep my draft';
  assert.equal(await app.saveDraft(), false);
  assert.equal(app.state.busy, false);
  assert.equal(app.state.dirty, true);
  assert.equal(app.state.draft.summary, 'Keep my draft');
  assert.equal(app.elements.get('#toast').textContent, 'Try again');
});

test('Fly uses the configured backend, including its path prefix', () => {
  const app = workspace();
  app.state.api = 'https://butterfly.example/workspace';
  assert.match(app.flyDemoPage(), /src="https:\/\/butterfly\.example\/workspace\/fly-demo"/);
});

test('deadline editing uses a calendar control matching the backend date contract', () => {
  const app = workspace();
  ready(app);
  assert.match(app.actionCard({ id: 'task', text: 'Review', deadline: '2026-10-08' }, 0), /<input[^>]*type="date"[^>]*data-field="deadline"/);
});

test('API settings can connect for this session when browser storage is disabled', async () => {
  const calls = [];
  const app = workspace(async (url) => { calls.push(url); return response(url.endsWith('/api/meetings') ? [] : {}); }, { storageUnavailable: true });
  await app.submit({ api: 'https://butterfly.example///' });
  assert.equal(app.state.api, 'https://butterfly.example');
  assert.equal(app.state.loadingMeeting, false);
  assert.deepEqual(calls, ['https://butterfly.example/api/session', 'https://butterfly.example/api/meetings', 'https://butterfly.example/health']);
  assert.equal(app.elements.get('#settings-dialog').closed, true);
});

test('changing the API saves a dirty meeting to its original server first', async () => {
  const calls = [];
  const app = workspace(async (url, options) => {
    calls.push({ url, method: options.method });
    return response(options.method === 'PATCH' ? meeting() : url.endsWith('/api/meetings') ? [] : {});
  });
  ready(app);
  app.state.api = 'https://old.example';
  await app.submit({ api: 'https://new.example' });
  assert.deepEqual(calls[0], { url: 'https://old.example/api/meetings/one', method: 'PATCH' });
  assert.equal(calls[1].url, 'https://new.example/api/session');
  assert.equal(app.state.selected, null);
  assert.equal(app.state.dirty, false);
});

test('opening a new demo preserves edits to the previously selected meeting', async () => {
  const calls = [];
  const app = workspace(async (url, options) => {
    calls.push({ url, method: options.method });
    return response(options.method === 'PATCH' ? meeting() : meeting('demo'));
  });
  ready(app);
  await app.createDemo();
  assert.deepEqual(calls[0], { url: '/api/meetings/one', method: 'PATCH' });
  assert.deepEqual(calls[1], { url: '/api/demo', method: 'POST' });
  assert.equal(app.state.selected.id, 'demo');
});

test('creating a meeting saves the previous draft and prevents duplicate submissions', { timeout: 1000 }, async () => {
  const calls = [];
  let resolveCreation;
  const app = workspace((url, options) => {
    calls.push({ url, method: options.method });
    if(url === '/api/meetings') return new Promise((resolve) => { resolveCreation = resolve; });
    return Promise.resolve(response(options.method === 'PATCH' ? meeting() : meeting('created')));
  });
  ready(app);
  const creating = app.submitMeeting({ title: 'New meeting', transcript: 'Discussion' });
  // Allow the autosave to finish and the creation request to begin.
  while(!resolveCreation) await new Promise((resolve) => setImmediate(resolve));
  assert.deepEqual(calls[0], { url: '/api/meetings/one', method: 'PATCH' });
  assert.equal(app.state.busy, true);
  await app.submitMeeting({ title: 'New meeting', transcript: 'Discussion' });
  assert.equal(calls.filter((call) => call.url === '/api/meetings').length, 1);
  resolveCreation(response(meeting('created')));
  await creating;
  assert.equal(app.state.busy, false);
  assert.equal(app.state.selected.id, 'created');
});

test('an obsolete connection error cannot replace a successful new API connection', async () => {
  let rejectOld;
  const app = workspace((url) => {
    if(url === 'https://old.example/api/session') return new Promise((resolve, reject) => { rejectOld = reject; });
    return Promise.resolve(response(url.endsWith('/api/meetings') ? [meeting('new')] : {}));
  });
  app.state.api = 'https://old.example';
  const old = app.connect();
  app.state.api = 'https://new.example';
  await app.connect();
  rejectOld(new Error('Old host unavailable'));
  await old;
  assert.equal(app.state.error, '');
  assert.equal(app.state.meetings[0].id, 'new');
});

test('a poll started before saving cannot restore an older version of the draft', async () => {
  let resolvePoll;
  const app = workspace((url, options) => {
    if(options.method === 'PATCH') return Promise.resolve(response({ ...meeting(), summary: 'Saved edit' }));
    return new Promise((resolve) => { resolvePoll = resolve; });
  });
  ready(app);
  const polling = app.loadMeeting('one', true);
  app.state.draft.summary = 'Saved edit';
  await app.saveDraft();
  resolvePoll(response({ ...meeting(), summary: 'Old summary' }));
  await polling;
  assert.equal(app.state.selected.summary, 'Saved edit');
  assert.equal(app.state.draft.summary, 'Saved edit');
});

test('an export still downloads the requested meeting after navigating home', async () => {
  let resolveExport;
  const app = workspace(() => new Promise((resolve) => { resolveExport = resolve; }));
  ready(app);
  app.state.dirty = false;
  const exporting = app.exportMeeting('json');
  await app.click('home');
  resolveExport({ ok: true, status: 200, blob: async () => ({}) });
  await exporting;
  assert.deepEqual(app.downloads, [{ href: 'blob:download', name: 'one.json' }]);
});
