import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import vm from 'node:vm';

test('passwordless profile offers verified-contact reauthentication', () => {
  const source = readFileSync(new URL('../src/pages/settings/profile-pane.tsx', import.meta.url), 'utf8');
  assert.match(source, /requestProfileReauthCode/);
  assert.match(source, /confirmProfileReauth/);
  assert.match(source, /phoneVerifiedAt/);
  assert.match(source, /emailVerifiedAt/);
  assert.match(source, /profileQuery\.refetch/);
});

// Callback-level race tests (not a browser/React DOM test): the real ProfileOTPReauth
// async handlers are extracted with Babel and run with hand-driven hooks.
const require = createRequire(import.meta.url);
const { transformSync } = require('@babel/core');
const presetTypeScript = require.resolve('@babel/preset-typescript');
const paneSource = readFileSync(new URL('../src/pages/settings/profile-pane.tsx', import.meta.url), 'utf8');
const paneStart = paneSource.indexOf('function ProfileOTPReauth(');
const paneEnd = paneSource.indexOf('export function ProfilePane()', paneStart);
assert.ok(paneStart >= 0 && paneEnd > paneStart);
const component = paneSource.slice(paneStart, paneEnd);
const returnStart = component.indexOf('    return <section');
assert.ok(returnStart > 0);
const compiled = transformSync(component.slice(0, returnStart) + '    return { send, confirm };\n}\n', {
  filename: 'profile-pane.tsx', configFile: false, babelrc: false,
  presets: [[presetTypeScript, { isTSX: true, allExtensions: true }]],
}).code;

class FakeApiError extends Error { constructor(m, retryAfterMs) { super(m); this.retryAfterMs = retryAfterMs; } }

function harness({ requestProfileReauthCode, confirmProfileReauth }) {
  const state = [], deps = [], refs = [];
  const toasts = [], verified = [];
  let cursor = 0, pendingEffects = [];
  const context = vm.createContext({
    App: { useApp: () => ({ message: { success: m => toasts.push(['success', m]), error: m => toasts.push(['error', m]) } }) },
    Date, ApiError: FakeApiError, requestProfileReauthCode, confirmProfileReauth,
    errorMessage: (error, fallback) => error?.message || fallback,
    window: { setInterval: () => 1, clearInterval() {} },
    useState(initial) {
      const index = cursor++;
      if (!(index in state)) state[index] = initial;
      return [state[index], value => { state[index] = value; }];
    },
    useRef(initial) {
      const index = cursor++;
      if (!(index in refs)) refs[index] = { current: initial };
      return refs[index];
    },
    useEffect(effect, next) {
      const index = cursor++;
      if (!deps[index] || next.some((value, i) => !Object.is(value, deps[index][i]))) {
        deps[index] = [...next];
        pendingEffects.push(effect);
      }
    },
  });
  vm.runInContext(compiled, context);
  return {
    // state slots: 0 ticket, 1 code, 2 busy, 3 retryAt
    state, toasts, verified,
    render(user) {
      cursor = 0; pendingEffects = [];
      const callbacks = context.ProfileOTPReauth({ user, onVerified: () => verified.push(user.id) });
      for (const effect of pendingEffects) effect();
      return callbacks;
    },
  };
}

const oldUser = { id: 'u1', phone: '+8613800138000', phoneVerifiedAt: '2026-01-01T00:00:00Z' };
const newUser = { ...oldUser, phone: '+8613900139000', phoneVerifiedAt: '2026-01-02T00:00:00Z' };
const deferred = () => { let resolve, reject; const promise = new Promise((a, b) => { resolve = a; reject = b; }); return { promise, resolve, reject }; };

test('a late send response must not restore a ticket or cooldown after the verified contact changes', async () => {
  const pending = deferred();
  const h = harness({ requestProfileReauthCode: () => pending.promise });
  const sending = h.render(oldUser).send('sms');
  h.render(newUser);
  pending.resolve({ ticket: 'old-contact-ticket', expiresIn: 600, retryAfter: 60 });
  await sending;
  assert.equal(h.state[0], null, 'late response restored an old-contact ticket');
  assert.equal(h.state[3], 0, 'late response put the new contact into resend cooldown');
  assert.equal(h.state[2], false, 'stale request left the new context busy');
  assert.deepEqual(h.toasts, [], 'late response leaked a toast into the new context');
});

test('a late send failure must not set cooldown or error toast in the new context', async () => {
  const pending = deferred();
  const h = harness({ requestProfileReauthCode: () => pending.promise });
  const sending = h.render(oldUser).send('sms');
  h.render(newUser);
  pending.reject(new FakeApiError('old failure', 30000));
  await sending;
  assert.equal(h.state[3], 0);
  assert.equal(h.state[2], false);
  assert.deepEqual(h.toasts, []);
});

test('a send that stays in the same context still applies its response', async () => {
  const h = harness({ requestProfileReauthCode: async () => ({ ticket: 't', expiresIn: 600, retryAfter: 60 }) });
  await h.render(oldUser).send('sms');
  assert.equal(h.state[0].value, 't');
  assert.ok(h.state[3] > 0);
});

test('a late confirm result must not clear the new context or report success', async () => {
  for (const outcome of ['resolve', 'reject']) {
    const pending = deferred();
    const h = harness({
      requestProfileReauthCode: async () => ({ ticket: 't', expiresIn: 600, retryAfter: 60 }),
      confirmProfileReauth: () => pending.promise,
    });
    await h.render(oldUser).send('sms');
    h.state[1] = '123456';
    const confirming = h.render(oldUser).confirm();
    h.render(newUser);
    // A fresh ticket for the new contact arrives while the stale confirm is pending.
    h.state[0] = { value: 'new-ticket', method: 'sms', expiresAt: Date.now() + 600000 };
    h.state[1] = '654321';
    h.toasts.length = 0;
    pending[outcome](outcome === 'reject' ? new FakeApiError('old rejected') : undefined);
    await confirming;
    assert.equal(h.state[0]?.value, 'new-ticket', `late ${outcome} cleared the new ticket`);
    assert.equal(h.state[1], '654321');
    assert.equal(h.state[2], false);
    assert.deepEqual(h.verified, [], 'late confirm must not call onVerified');
    assert.deepEqual(h.toasts, []);
  }
});
