import { test } from "node:test";
import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";

// Static source checks (not browser tests): Linux.do login/bind/reauth and its historical display are removed from the web app.
const read = (path) => readFileSync(new URL(`../${path}`, import.meta.url), "utf8");
const section = (source, from, to) => {
    const start = source.indexOf(from);
    const end = source.indexOf(to, start);
    assert.ok(start >= 0 && end > start, `section ${from} not found`);
    return source.slice(start, end);
};

test("no Linux.do entry, setting or OAuth plumbing remains in the live web code", () => {
    for (const path of [
        "src/pages/auth/login.tsx",
        "src/pages/auth/register.tsx",
        "src/pages/auth/auth-scene.tsx",
        "src/pages/settings/profile-pane.tsx",
        "src/pages/admin/components/access-settings-panel.tsx",
        "src/services/api/auth.ts",
        "src/services/api/profile.ts",
        "src/services/api/wallet.ts",
        "src/services/api/login-logs.ts",
        "src/stores/use-user-store.ts",
        "src/components/layout/workspace-sidebar-footer.tsx",
        "src/pages/admin/users/users-columns.tsx",
        "src/pages/admin/login-logs/login-log-display.ts",
        "vite.config.ts",
    ])
        assert.doesNotMatch(read(path), /linux\.?do/i, `${path} still references Linux.do`);
});

test("Linux.do badge component, icon asset and identity fields are gone", () => {
    assert.equal(existsSync(new URL("../src/components/layout/identity-provider-badge.tsx", import.meta.url)), false);
    assert.equal(existsSync(new URL("../public/icons/linuxdo.svg", import.meta.url)), false);
    for (const path of ["src/stores/use-user-store.ts", "src/services/api/auth.ts", "src/components/layout/workspace-sidebar-footer.tsx"]) assert.doesNotMatch(read(path), /identityProvider|identityUsername|IdentityProviderBadge/, path);
});

test("admin access panel keeps registration and agreement settings", () => {
    const panel = read("src/pages/admin/components/access-settings-panel.tsx");
    assert.match(panel, /1\. 是否允许创建新账号/);
    assert.match(panel, /2\. 服务协议名称与条款/);
    assert.match(panel, /getAdminRegistrationSetting/);
});

test("profile email change uses the current password only and never the OTP reauth marker", () => {
    const pane = read("src/pages/settings/profile-pane.tsx");
    const emailFlow = section(pane, "const sendEmailCode", "const resetEmailTicket");
    assert.doesNotMatch(emailFlow, /reauth/i);
    assert.match(emailFlow, /!profile\?\.hasPassword|!profile\.hasPassword/);
    // The email form is only rendered for password accounts; no-password accounts are told to set a password first.
    const emailCard = section(pane, 'id="profile-email-title"', "avatarDialogOpen");
    assert.doesNotMatch(emailCard, /identityReady|reauth/i);
    assert.match(emailCard, /profile\.hasPassword/);
    // OTP reauth still gates setting the password.
    assert.match(section(pane, "const savePassword", "const sendEmailCode"), /profileReauthReady/);
});
