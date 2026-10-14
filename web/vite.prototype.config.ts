// PROTOTYPE ONLY. Anonymous fixture responses; no requests reach a real backend.
import { mergeConfig, defineConfig } from "vite";
import base from "./vite.config";

export default mergeConfig(base, defineConfig({
    cacheDir: ".prototype-vite-cache",
    plugins: [{
        name: "mediacore-prototype-api-fixtures",
        configureServer(server) {
            server.middlewares.use((req, res, next) => {
                if (!req.url?.startsWith("/api/")) return next();
                res.setHeader("Content-Type", "application/json");
                if (req.method !== "GET") {
                    res.statusCode = 405;
                    return res.end(JSON.stringify({ code: 405, msg: "原型不执行写入或生成" }));
                }
                const path = req.url.split("?")[0];
                const fixtures: Record<string, unknown> = {
                    "/api/auth/session": { user: null },
                    "/api/public/appearance": { appearance: { brandName: "Mediacore", seoTitle: "Mediacore · 创作 UI 原型" } },
                    "/api/model-catalog": { source: "system", channels: [] },
                    "/api/announcements": { announcements: [] },
                    "/api/banner-announcements": { announcements: [] },
                };
                if (!(path in fixtures)) {
                    res.statusCode = 404;
                    return res.end(JSON.stringify({ code: 404, msg: "此接口不属于 UI 原型" }));
                }
                res.end(JSON.stringify({ code: 0, data: fixtures[path] }));
            });
        },
    }],
}));
