// Throwaway shared UI-study control. Never rendered by production builds.
import { useEffect } from "react";
import { ChevronLeft, ChevronRight, FlaskConical } from "lucide-react";
import { useSearchParams } from "react-router";

export const prototypeVariants = { A: "雾白工作室", B: "灵感桌面", C: "暗场导演台" };
export type PrototypeVariant = keyof typeof prototypeVariants;
export function PrototypeSwitcher({ variant, state }: { variant: PrototypeVariant; state: Record<string, unknown> }) {
    const [, setSearch] = useSearchParams();
    const cycle = (offset: number) => {
        const keys = Object.keys(prototypeVariants) as PrototypeVariant[];
        const next = keys[(keys.indexOf(variant) + offset + keys.length) % keys.length];
        setSearch((current) => { current.set("variant", next); return current; }, { replace: true });
    };
    useEffect(() => { console.info("Mediacore UI prototype", { variant, ...state }); }, [variant, state]);
    useEffect(() => {
        const onKey = (event: KeyboardEvent) => {
            if ((event.target as HTMLElement).closest("input, textarea, select, [contenteditable=true], [role=dialog]")) return;
            if (event.key === "ArrowLeft" || event.key === "ArrowRight") { event.preventDefault(); cycle(event.key === "ArrowLeft" ? -1 : 1); }
        };
        window.addEventListener("keydown", onKey);
        return () => window.removeEventListener("keydown", onKey);
    });
    if (!import.meta.env.DEV) return null;
    return <aside className="mc-prototype-switcher" aria-label="原型方案切换">
        <span className="mc-prototype-flag"><FlaskConical size={14} /> 设计原型</span>
        <button aria-label="上一个设计方案" onClick={() => cycle(-1)}><ChevronLeft size={18} /></button>
        <strong>{variant} · {prototypeVariants[variant]}</strong>
        <button aria-label="下一个设计方案" onClick={() => cycle(1)}><ChevronRight size={18} /></button>
        <details><summary>演示状态</summary><pre>{JSON.stringify({ variant, ...state, data: "示例数据，操作仅在内存中生效，不调用 AI" }, null, 2)}</pre></details>
    </aside>;
}
