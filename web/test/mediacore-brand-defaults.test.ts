import { expect, test } from "bun:test";

import { resolvePublicAppearance } from "../src/services/appearance-bootstrap";
import { brandStudioLabel, normalizePublicAppearance } from "../src/stores/use-appearance-store";

test("a failed first appearance request presents the bundled Mediacore identity", async () => {
    const appearance = await resolvePublicAppearance(async () => {
        throw new Error("offline");
    });

    expect(appearance.brandName).toBe("Mediacore");
    expect(appearance.brandSlug).toBe("mediacore");
    expect(brandStudioLabel(appearance)).toBe("MEDIACORE STUDIO");
    expect(appearance.seoTitle).toBe("Mediacore");
    expect(appearance.seoDescription).toBe("Mediacore，面向 AI 影视与短剧创作的工作台。");
    expect(appearance.footerCopyright).toContain("Mediacore. All rights reserved.");
    expect(appearance.canvas?.agentName).toBe("Mediacore");
    expect(appearance.logoUrl).toBe("/logo.svg");
    expect(appearance.authVideoUrl).toBe("");
    expect(appearance.authVideoPosterUrl).toBe("/brand/mediacore-auth-poster.svg");
});

test("incomplete or unsafe appearance responses keep a safe Mediacore fallback", () => {
    const appearance = normalizePublicAppearance({
        brandName: " ",
        brandSlug: "not a slug",
        logoUrl: "javascript:alert(1)",
        darkLogoUrl: "//untrusted.example/logo.svg",
        authVideoUrl: "http://untrusted.example/brand.mp4",
        authVideoPosterUrl: "data:image/svg+xml,<svg/>",
        redeemPurchaseUrl: "javascript:alert(1)",
    });

    expect(appearance.brandName).toBe("Mediacore");
    expect(brandStudioLabel(appearance)).toBe("MEDIACORE STUDIO");
    expect(appearance.logoUrl).toBe("/logo.svg");
    expect(appearance.darkLogoUrl).toBe("/logo.svg");
    expect(appearance.authVideoUrl).toBe("");
    expect(appearance.authVideoPosterUrl).toBe("/brand/mediacore-auth-poster.svg");
    expect(appearance.redeemPurchaseUrl).toBe("");
});

test("explicit custom branding and secure asset locations remain configurable", () => {
    const appearance = normalizePublicAppearance({
        brandName: "My Studio",
        brandSlug: "my-studio",
        logoConfigured: true,
        logoUrl: "https://assets.example.com/logo.svg",
        authVideoConfigured: true,
        authVideoUrl: "/api/public/appearance/assets/video?v=custom",
    });

    expect(appearance.brandName).toBe("My Studio");
    expect(brandStudioLabel(appearance)).toBe("MY STUDIO");
    expect(appearance.logoUrl).toBe("https://assets.example.com/logo.svg");
    expect(appearance.authVideoUrl).toBe("/api/public/appearance/assets/video?v=custom");
    expect(appearance.authVideoPosterUrl).toBe("");
});
