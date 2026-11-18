import { test } from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";

const read = (path) => readFileSync(new URL(path, import.meta.url), "utf8");
test("phone-first login keeps password as a separate entrance", () => {
    const source = read("../src/pages/auth/login.tsx");
    assert.match(source, /setMethod\(available\[0\] \?\? "password"\)/);
    assert.match(source, /methods\.length > 1/);
    assert.match(source, /密码登录/);
    assert.match(source, /手机号 \/ 邮箱 \/ 用户名/);
});
test("verified registration permits optional credentials and keeps agreement", () => {
    const source = read("../src/pages/auth/register.tsx");
    assert.match(source, /useState<VerificationMethod>\("sms"\)/);
    assert.match(source, /用户名（可选）/);
    assert.match(source, /密码（可选）/);
    assert.match(source, /required=\{settings\?\.firstUser\}/);
    assert.match(source, /acceptedTerms: agreementAccepted/);
});
test("registration requires password when verification login is unavailable", () => {
    const source = read("../src/pages/auth/register.tsx");
    assert.match(source, /const passwordRequired =/);
    assert.match(source, /required=\{passwordRequired\}/);
    assert.match(source, /当前联系方式未开启验证码登录/);
});
test("login delivery acknowledgement does not reveal account existence", () => {
    const source = read("../src/components/auth/verification-fields.tsx");
    assert.match(source, /已受理.*注册/);
});
