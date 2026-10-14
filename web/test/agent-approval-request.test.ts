import { afterEach, expect, test } from "bun:test";
import { agentImageApproval, agentVideoApproval } from "../src/lib/canvas/agent-media-approval";
import { decideAgentApproval } from "../src/services/api/agent";
import { apiClient } from "../src/services/api/request";

const originalAdapter = apiClient.defaults.adapter;
afterEach(() => {
    apiClient.defaults.adapter = originalAdapter;
});

test.each(["image", "video"])("%s approval sends only editable media settings", async (mode) => {
    const detail = {
        toolName: "generate_media",
        arguments: {
            mode,
            prompt: "保持主体外观",
            size: "16:9",
            quality: "720p",
            channelId: "platform",
            channelModelKey: "test-model",
            referenceNodeIds: ["reference"],
            referenceTransientIds: ["transient"],
            sourceNodeId: "source",
            ...(mode === "video" ? { durationSeconds: 5, videoGenerateAudio: false, videoWatermark: false, videoArkPrivateAssetUpload: false } : {}),
        },
    };
    const initial = mode === "video" ? agentVideoApproval(detail) : agentImageApproval(detail);
    if (!initial) throw new Error("approval fixture must parse");
    // The video option editor retains the initial preview fields when changing size.
    const settings = { ...initial, size: "9:16" };
    let body: Record<string, unknown> | undefined;
    apiClient.defaults.adapter = async (config) => {
        body = JSON.parse(config.data);
        return { data: { code: 0, data: { accepted: true }, msg: "ok" }, status: 200, statusText: "OK", headers: {}, config };
    };

    await expect(decideAgentApproval("run", "approval", "approve", "", undefined, settings)).resolves.toEqual({ accepted: true });
    expect(body).toEqual({
        decision: "approve",
        mediaSettings: {
            logicalModelId: "",
            channelId: "platform",
            channelModelKey: "test-model",
            size: "9:16",
            quality: "720p",
            ...(mode === "video" ? { durationSeconds: 5, videoGenerateAudio: false, videoWatermark: false, videoArkPrivateAssetUpload: false } : {}),
        },
    });
});
