import { describe, expect, test } from "bun:test";
import { renderToStaticMarkup } from "react-dom/server";

import { LoadingContent } from "../src/components/canvas/canvas-node-status-content";
import { canvasThemes } from "../src/lib/canvas-theme";
import { CanvasNodeType, type CanvasNodeMetadata } from "../src/types/canvas";

function renderLoading(metadata: CanvasNodeMetadata) {
    return renderToStaticMarkup(<LoadingContent node={{ id: "video", type: CanvasNodeType.Video, title: "视频", position: { x: 0, y: 0 }, width: 400, height: 300, metadata: { status: "loading", ...metadata } }} theme={canvasThemes.dark} />);
}

describe("画布生成节点状态文案", () => {
    for (const stage of [undefined, "排队中"]) {
        test(`排队阶段为 ${stage ?? "空"} 时只显示一次排队中`, () => {
            const markup = renderLoading({ taskId: "task-1", taskStatus: "queued", taskStage: stage });
            expect(markup.match(/排队中/g)).toHaveLength(1);
            expect(markup).toContain("详情");
        });
    }

    test("保留与状态不同的阶段说明", () => {
        const markup = renderLoading({ taskId: "task-1", taskStatus: "queued", taskStage: "等待队列调度" });
        expect(markup).toContain("正在准备创作");
        expect(markup.match(/排队中/g)).toHaveLength(1);
    });

    test("生成阶段为空时只显示一次生成中", () => {
        const markup = renderLoading({ taskId: "task-1", taskStatus: "running" });
        expect(markup.match(/生成中/g)).toHaveLength(1);
    });

    test("尚未创建任务时保留准备说明", () => {
        expect(renderLoading({})).toContain("正在创建任务");
    });

    test("提交结果不确定时保留避免重复扣费说明", () => {
        const markup = renderLoading({ taskId: "task-1", taskStatus: "running", taskStage: "submission_unknown" });
        expect(markup).toContain("提交结果待确认");
        expect(markup).toContain("为避免重复扣费，未自动重试");
    });
});
