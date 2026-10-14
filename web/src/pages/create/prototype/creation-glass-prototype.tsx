/**
 * THROWAWAY: three structurally different creation homes on /?variant=A|B|C.
 * Question: which frosted-glass hierarchy makes starting and continuing work clearest?
 * Existing route/auth/data hooks remain in the host. All actions below are memory-only stubs.
 * Verdict pending user comparison. Never merge this prototype directly into production.
 */
import { useEffect, useRef, useState, type ReactNode } from "react";
import { useSearchParams } from "react-router";
import { ArrowDownToLine, ArrowRight, ArrowUpRight, Bell, BookOpen, Check, ChevronDown, Clapperboard, Clock3, Coins, Film, FolderOpen, Image as ImageIcon, Layers3, LayoutDashboard, Lightbulb, Maximize2, MessageSquare, Mic, Moon, MoreHorizontal, Paperclip, Plus, SlidersHorizontal, Sparkles, Sun, WandSparkles, X } from "lucide-react";
import { Modal } from "antd";
import { PrototypeSwitcher, type PrototypeVariant } from "@/components/ui/prototype-switcher";
import { curatedInspirations, type CreationInspiration, inspirationSourceOf } from "@/lib/inspirations/catalog";
import "./creation-glass-prototype.css";

type Mode = "image" | "video" | "text" | "agent";
type Surface = { nav: ReactNode; header: ReactNode; intro: ReactNode; composer: ReactNode; shortcuts: ReactNode; gallery: ReactNode; stage: ReactNode; history: ReactNode };
const modes = [{ id: "image", label: "图片", icon: ImageIcon }, { id: "video", label: "视频", icon: Film }, { id: "text", label: "文本", icon: MessageSquare }, { id: "agent", label: "Agent", icon: Sparkles }] as const;
const modelNames: Record<Exclude<Mode, "agent">, string[]> = {
    image: ["GPT_Image2", "Gemini_Image", "Grok_Imagine"],
    video: ["Seedance2.0_Mini", "Seedance2.0_Fast", "Grok_Imagine"],
    text: ["GPT6.1_Sol", "Gemini", "Claude"],
};
const navGroups = [
    [{ label: "创作", icon: WandSparkles, route: "/", purpose: "从提示词开始，快速生成文本、图片和视频。" }, { label: "灵感", icon: Lightbulb, route: "/inspirations", purpose: "浏览案例，复用提示词；选用灵感不会自动发起生成。" }, { label: "短剧制作", icon: Clapperboard, route: "/projects", purpose: "以项目组织故事、角色、场景与分镜，沿用原有制作流程。" }, { label: "自由画布", icon: LayoutDashboard, route: "/canvas", purpose: "在无限画布中组织节点、连接素材，并与 Agent 协作。" }],
    [{ label: "图片生成", icon: ImageIcon, route: "/toolbox/image", purpose: "图片专属工作台，集中管理参考图、模型与生成参数。" }, { label: "视频生成", icon: Film, route: "/toolbox/video", purpose: "视频专属工作台，处理参考素材、时长与任务结果。" }, { label: "技能", icon: BookOpen, route: "/skills", purpose: "加入可复用的创作技能，通过 @ 引用或交给 Agent 使用。" }],
    [{ label: "资产", icon: FolderOpen, route: "/assets", purpose: "管理上传素材与生成结果，供不同创作页面引用。" }, { label: "创作历史", icon: Clock3, route: "/tasks", purpose: "查看跨页面的生成任务和结果；与本页历史对话区分。" }],
];
const featured = [curatedInspirations[0], curatedInspirations[1], curatedInspirations[3], curatedInspirations.find((item) => item.title === "月面独行") || curatedInspirations[2]];

export function VariantA(p: Surface) {
    return <div className="mc-layout-a">{p.nav}<div className="mc-page">{p.header}<main className="mc-main-a"><div className="mc-intro-row">{p.intro}{p.history}</div>{p.composer}{p.shortcuts}{p.gallery}</main></div></div>;
}
export function VariantB(p: Surface) {
    return <div className="mc-layout-b">{p.nav}<div className="mc-page">{p.header}<main className="mc-main-b"><div className="mc-b-heading">{p.intro}{p.history}</div><div className="mc-b-desk"><section className="mc-b-board">{p.stage}{p.gallery}</section><section className="mc-b-compose">{p.composer}{p.shortcuts}</section></div></main></div></div>;
}
export function VariantC(p: Surface) {
    return <div className="mc-layout-c">{p.nav}<div className="mc-page">{p.header}<main className="mc-main-c"><div className="mc-c-heading">{p.intro}{p.history}</div><div className="mc-c-desk"><section className="mc-c-control">{p.composer}{p.shortcuts}</section><section className="mc-c-monitor">{p.stage}{p.gallery}</section></div></main></div></div>;
}

export default function CreationGlassPrototype() {
    const [search] = useSearchParams();
    const variant: PrototypeVariant = search.get("variant") === "B" ? "B" : search.get("variant") === "C" ? "C" : "A";
    const [themeOverride, setThemeOverride] = useState<boolean | null>(null);
    const dark = themeOverride ?? variant === "C";
    const [mode, setMode] = useState<Mode>("image");
    const [prompt, setPrompt] = useState("");
    const [models, setModels] = useState({ image: modelNames.image[0], video: modelNames.video[0], text: modelNames.text[0] });
    const [ratio, setRatio] = useState("16:9");
    const [quality, setQuality] = useState("自动");
    const [count, setCount] = useState("1");
    const [seconds, setSeconds] = useState("6");
    const [thinking, setThinking] = useState(false);
    const [streaming, setStreaming] = useState(true);
    const [panel, setPanel] = useState<"model" | "settings" | null>(null);
    const [references, setReferences] = useState<Array<{ name: string; image?: string }>>([]);
    const [skills, setSkills] = useState<string[]>([]);
    const [dialog, setDialog] = useState<"references" | "skills" | "history" | "account" | "result" | null>(null);
    const [notice, setNotice] = useState<{ title: string; body: string } | null>(null);
    const [activeArt, setActiveArt] = useState(1);
    const [filter, setFilter] = useState("全部");
    const [history, setHistory] = useState<Array<{ prompt: string; mode: Mode }>>([]);
    const textarea = useRef<HTMLTextAreaElement>(null);
    const fileInput = useRef<HTMLInputElement>(null);
    const objectURLs = useRef<string[]>([]);
    const model = mode === "agent" ? "画布 Agent" : models[mode];
    const price = mode === "video" && model === "Seedance2.0_Mini" ? 0 : mode === "image" ? Number(count) : 1;
    const chosenArt = featured[activeArt];
    const notify = (title: string, body: string) => setNotice({ title, body });
    const selectMode = (next: Mode) => { setMode(next); setPanel(null); };
    const focusPrompt = () => requestAnimationFrame(() => textarea.current?.focus());
    const useIdea = (item: CreationInspiration) => { selectMode(item.mode); setPrompt(item.prompt); focusPrompt(); };
    const demoGenerate = () => { if (!prompt.trim()) return; setHistory((items) => [{ prompt, mode }, ...items]); setDialog("result"); };
    useEffect(() => () => { objectURLs.current.forEach(URL.revokeObjectURL); }, []);
    useEffect(() => { setPanel(null); setThemeOverride(null); }, [variant]);
    const hint = mode === "image" ? "描述你想看见的画面、人物、光线与风格…" : mode === "video" ? "描述镜头内容、运动、光线与节奏…" : "从一个故事、一个角色，或一句对白开始…";

    const nav = <aside className="mc-nav mc-glass">
        <button className="mc-wordmark" onClick={() => { setPrompt(""); selectMode("image"); }} aria-label="Mediacore 创作首页"><span className="mc-brand-symbol"><i /><i /><i /></span><span>Mediacore<small>映核 · 创作空间</small></span></button>
        <div className="mc-nav-groups">{navGroups.map((group, index) => <div className="mc-nav-group" key={index}>{index > 0 && <span className="mc-nav-label">{index === 1 ? "创作工具" : "我的空间"}</span>}{group.map(({ label, icon: Icon, route, purpose }) => <button key={label} className={route === "/" ? "is-active" : ""} aria-current={route === "/" ? "page" : undefined} title={label} onClick={() => route === "/" ? (setPrompt(""), setDialog(null)) : notify(label, `${purpose}\n原有入口：${route}。本次只预览创作首页，其他页面待确定风格后沿用原功能改版。`)}><Icon /><span>{label}</span>{route === "/" && <i className="mc-nav-dot" />}</button>)}</div>)}</div>
        <div className="mc-nav-bottom"><button className="mc-storage" onClick={() => notify("账号容量", "沿用原有容量统计，包含素材与 Agent 附件。这里展示的是示例容量。") }><span><Layers3 /> <span>云端空间<small>4 MB / 20 GB</small></span></span><i><b /></i></button><button className="mc-profile" onClick={() => setDialog("account")}><span className="mc-avatar">M</span><span>我的工作空间<small>个人创作账户</small></span><MoreHorizontal /></button></div>
    </aside>;
    const header = <header className="mc-header"><div className="mc-breadcrumb"><span>工作空间</span><span>/</span><strong>创作</strong></div><div className="mc-header-actions"><button className="mc-credit" onClick={() => notify("积分中心", "余额、充值、兑换码与账单沿用原生功能。原型显示 120 示例积分，不扣费。") }><Coins /><span>120 <small>积分</small></span><Plus size={12} /></button><span className="mc-divider" /><button className="mc-icon" aria-label="系统公告" onClick={() => notify("系统公告", "公告入口保留；右上角不放回更新日志。") }><Bell /></button><button className="mc-icon" aria-label="切换明暗主题" onClick={() => setThemeOverride(!dark)}>{dark ? <Sun /> : <Moon />}</button><button className="mc-header-avatar" aria-label="账户菜单" onClick={() => setDialog("account")}>M</button></div></header>;
    const intro = <div className="mc-intro"><div className="mc-eyebrow"><i /> YOUR NEXT FRAME</div><h1>{variant === "B" ? <>灵感在这里，<br />下一帧由你。</> : variant === "C" ? <>你的想法，<em>即将开场。</em></> : <>让想象，<em>慢慢成形。</em></>}</h1><p>{variant === "B" ? "看见一个好画面，然后创造属于你的版本。" : variant === "C" ? "从画面、镜头到故事，把注意力留给创作。" : "从一句描述出发，创造画面、镜头与故事。"}</p></div>;
    const historyButton = <button className="mc-history-button" onClick={() => setDialog("history")}><Clock3 />历史对话<ChevronDown size={13} /></button>;
    const modeTabs = <div className="mc-mode-tabs" aria-label="创作模式">{modes.map(({ id, label, icon: Icon }) => <button key={id} aria-pressed={mode === id} onClick={() => selectMode(id)}><Icon /><span>{label}</span>{id === "agent" && <i />}</button>)}</div>;
    const composer = <section className="mc-composer mc-glass" aria-label="创作输入区">
        <div className="mc-composer-heading">{modeTabs}<span className="mc-composer-note">{mode === "agent" ? "画布协作" : "开始新的创作"}</span></div>
        {mode === "agent" ? <div className="mc-agent"><div className="mc-agent-mark"><Sparkles /></div><span className="mc-eyebrow">CANVAS AGENT</span><h2>从想法，到一整张画布。</h2><p>让 Agent 结合素材与技能，帮你规划、生成和整理。<br />生成与修改保留原有的审批步骤。</p><button className="mc-primary" onClick={() => notify("新建画布，与 Agent 创作", "正式功能会先创建并保存画布，再进入同一个画布 Agent。\n本次仅演示入口，不创建真实画布、不发起 AI 请求。")}>新建画布，与 Agent 创作<ArrowUpRight /></button><button className="mc-text-button" onClick={() => notify("在已有画布中继续", "沿用 /canvas?agent=1 的画布选择入口，继续使用已有素材与上下文。") }><FolderOpen />在已有画布中继续</button></div> : <>
            <textarea ref={textarea} aria-label="创作提示词" placeholder={hint} value={prompt} onChange={(event) => setPrompt(event.target.value)} onKeyDown={(event) => { if (event.key === "Enter" && !event.shiftKey && !event.nativeEvent.isComposing) { event.preventDefault(); demoGenerate(); } }} />
            <div className="mc-reference-row">{references.map((ref, index) => <span className="mc-reference" key={`${ref.name}-${index}`}>{ref.image ? <img src={ref.image} alt="参考素材" /> : <Paperclip />}<span>{ref.name}</span><button aria-label={`移除参考 ${ref.name}`} onClick={() => setReferences((items) => items.filter((_, i) => index !== i))}><X size={12} /></button></span>)}{skills.map((skill) => <span className="mc-reference mc-skill-chip" key={skill}><BookOpen /><span>{skill}</span><button aria-label="移除技能" onClick={() => setSkills([])}><X size={12} /></button></span>)}<button className="mc-attachment" onClick={() => setDialog("references")}><Plus />添加参考</button><button className="mc-attachment mc-skill-add" onClick={() => setDialog("skills")}>@ 技能</button><span className="mc-input-hint">Shift + Enter 换行</span></div>
            <div className="mc-composer-footer"><div className="mc-composer-tools"><button className="mc-icon mc-voice" aria-label="实时对话" onClick={() => notify("实时对话", "原有入口保留。本原型不请求麦克风，也不启用尚未接通的音频生成或转写。") }><Mic /></button><div className="mc-popover-host"><button className="mc-model-button" aria-expanded={panel === "model"} onClick={() => setPanel(panel === "model" ? null : "model")}><span className="mc-model-mark">{mode === "video" ? <Film /> : mode === "text" ? <MessageSquare /> : <ImageIcon />}</span>{model}<ChevronDown /></button>{panel === "model" && <div className="mc-popover mc-model-menu"><span className="mc-small-label">4aion · {mode === "image" ? "图片" : mode === "text" ? "文本" : "视频"}模型示例</span>{modelNames[mode].map((name) => <button key={name} onClick={() => { setModels((current) => ({ ...current, [mode]: name })); setPanel(null); }}><span><strong>{name}</strong><small>{name === "Seedance2.0_Mini" ? 0 : 1} 积分 / 次 · 示例</small></span>{model === name && <Check size={15} />}</button>)}</div>}</div><div className="mc-popover-host"><button className="mc-param-button" aria-label="生成设置" aria-expanded={panel === "settings"} onClick={() => setPanel(panel === "settings" ? null : "settings")}><SlidersHorizontal /><span>{mode === "text" ? thinking ? "深度思考" : "生成设置" : `${ratio} · ${mode === "video" ? "720P" : quality}`}</span><ChevronDown /></button>{panel === "settings" && <div className="mc-popover mc-settings-menu"><span className="mc-small-label">生成设置 · 布局演示</span>{mode === "text" ? <><label>流式输出<input type="checkbox" checked={streaming} onChange={(event) => setStreaming(event.target.checked)} /></label><label>深度思考<input type="checkbox" checked={thinking} onChange={(event) => setThinking(event.target.checked)} /></label></> : <><label>画面比例<select value={ratio} onChange={(event) => setRatio(event.target.value)}>{["16:9", "9:16", "1:1", "4:3", "3:4"].map((item) => <option key={item}>{item}</option>)}</select></label>{mode === "image" && <><label>图片质量<select value={quality} onChange={(event) => setQuality(event.target.value)}>{["自动", "标准", "高清"].map((item) => <option key={item}>{item}</option>)}</select></label><label>生成张数<select value={count} onChange={(event) => setCount(event.target.value)}>{["1", "2", "3", "4"].map((item) => <option key={item}>{item}</option>)}</select></label></>}{mode === "video" && <label>视频时长<select value={seconds} onChange={(event) => setSeconds(event.target.value)}>{["4", "6", "8", "10"].map((item) => <option key={item}>{item}</option>)}</select></label>}</>}<small>正式版仍按所选模型支持的参数显示。</small></div>}</div>{mode === "video" && <button className="mc-param-button mc-duration" onClick={() => setPanel("settings")}><Clock3 />{seconds}s</button>}</div><button className="mc-generate mc-primary" disabled={!prompt.trim()} onClick={demoGenerate}><span>开始创作<small>{price} 积分</small></span><ArrowRight /></button></div>
        </>}
    </section>;
    const shortcuts = <div className="mc-shortcuts"><button onClick={() => useIdea(featured[0])}><Clapperboard /><span>生成第一个镜头</span><ArrowUpRight /></button><button onClick={() => { selectMode("image"); setDialog("references"); }}><ImageIcon /><span>从参考图开始</span><ArrowUpRight /></button><button onClick={() => { selectMode("text"); setPrompt("帮我续写一个短剧故事，先聊聊剧情走向："); focusPrompt(); }}><MessageSquare /><span>续写故事</span><ArrowUpRight /></button><button onClick={() => setDialog("skills")}><BookOpen /><span>引用创作技能</span><ArrowUpRight /></button></div>;
    const filteredArt = featured.filter((item) => filter === "全部" || (filter === "图片" ? item.mode === "image" : item.mode === "video"));
    const gallery = <section className="mc-gallery"><div className="mc-section-heading"><div><span className="mc-eyebrow">A LITTLE INSPIRATION</span><h2>换个角度，寻找灵感</h2></div><button className="mc-text-button" onClick={() => notify("全部灵感", "沿用灵感库 /inspirations，可浏览、搜索、收藏并复用提示词。")}>全部灵感<ArrowUpRight /></button></div><div className="mc-gallery-filters">{["全部", "图片", "视频"].map((item) => <button key={item} aria-pressed={filter === item} onClick={() => setFilter(item)}>{item}</button>)}<span>选用灵感，填入提示词</span></div><div className="mc-gallery-grid">{filteredArt.map((item) => <button key={item.title} className="mc-inspiration-card" onClick={() => { setActiveArt(featured.indexOf(item)); useIdea(item); }}><div className="mc-art"><img src={item.image} alt={item.title} /><span className="mc-art-type">{item.mode === "video" ? <Film /> : <ImageIcon />}{item.mode === "video" ? "视频" : "图片"}</span><span className="mc-art-arrow"><ArrowUpRight /></span></div><div className="mc-art-description"><strong>{item.title.split(" · ")[0]}</strong><span>{item.description}</span></div></button>)}</div><div className="mc-sources">案例封面复用开源项目素材，仅作灵感示意。{inspirationSourceOf(chosenArt).label}</div></section>;
    const stage = <section className="mc-stage"><div className="mc-stage-toolbar"><span><i />{variant === "C" ? "灵感预览" : "今日灵感"}</span><button className="mc-icon" aria-label="放大灵感" onClick={() => notify(chosenArt.title, `${chosenArt.description}\n这是现有灵感封面，并非本次生成结果。点击“使用这个创意”会把原提示词填入创作区。`)}><Maximize2 /></button></div><div className="mc-stage-picture"><img src={chosenArt.image} alt={chosenArt.title} /><div className="mc-stage-caption"><span>STORY STARTS HERE</span><h2>{chosenArt.title.split(" · ")[0]}</h2><p>{chosenArt.description}</p><button onClick={() => useIdea(chosenArt)}>使用这个创意<ArrowUpRight /></button></div><span className="mc-stage-counter">0{activeArt + 1} <i>/ 04</i></span></div><div className="mc-stage-thumbs">{featured.map((art, index) => <button key={art.title} aria-label={`预览灵感 ${art.title}`} aria-pressed={index === activeArt} onClick={() => setActiveArt(index)}><img src={art.image} alt="" /><span>0{index + 1}</span></button>)}<div>一个画面，<br /><strong>也可以是故事的开始。</strong></div></div></section>;
    const surface: Surface = { nav, header, intro, composer, shortcuts, gallery, stage, history: historyButton };
    return <div className={`mc-prototype mc-variant-${variant} ${dark ? "mc-dark" : "mc-light"}`}>
        <div className="mc-ambient" aria-hidden><i /><i /><i /></div>
        {variant === "B" ? <VariantB {...surface} /> : variant === "C" ? <VariantC {...surface} /> : <VariantA {...surface} />}
        <PrototypeSwitcher variant={variant} state={{ mode, model, ratio, quality, count, seconds, thinking, streaming, prompt, references: references.map((r) => r.name), skills, historyCount: history.length, dark, panel, dialog }} />
        <Modal open={Boolean(dialog)} title={dialog === "references" ? "添加参考内容" : dialog === "skills" ? "引用创作技能" : dialog === "history" ? "历史对话" : dialog === "account" ? "我的账户" : "创作结果 · 交互示意"} onCancel={() => setDialog(null)} footer={null} width={dialog === "result" ? 640 : 560} destroyOnHidden>
            <div className="mc-demo-modal">
                {dialog === "references" && <><p>从资产中选取，或上传本地文件。原型仅保存在当前页面内存中。</p><div className="mc-demo-assets">{featured.map((item) => <button key={item.title} onClick={() => { setReferences((items) => [...items, { name: item.title, image: item.image }]); setDialog(null); }}><img src={item.image} alt="" /><span>{item.title}</span></button>)}</div><input hidden ref={fileInput} type="file" accept={mode === "image" ? "image/*" : "image/*,video/*,audio/*,.txt,.pdf"} onChange={(event) => { const file = event.target.files?.[0]; if (!file) return; const url = URL.createObjectURL(file); objectURLs.current.push(url); setReferences((items) => [...items, { name: file.name, image: file.type.startsWith("image/") ? url : undefined }]); setDialog(null); }} /><button className="mc-demo-action" onClick={() => fileInput.current?.click()}><Paperclip />上传本地参考</button></>}
                {dialog === "skills" && <><p>技能是可复用的创作指引，与模型选择各有用途。</p><button className="mc-demo-skill" onClick={() => { setSkills(["Mediacore 短剧制作工作流"]); setDialog(null); focusPrompt(); }}><BookOpen /><span><strong>Mediacore 短剧制作工作流</strong><small>故事构思、角色、分镜与镜头规划</small></span><Plus /></button></>}
                {dialog === "history" && <><p>这里是本页的历史对话；所有生成任务仍从“创作历史”进入。</p>{history.length ? history.map((item, index) => <button className="mc-demo-history" key={index} onClick={() => { setPrompt(item.prompt); selectMode(item.mode); setDialog(null); }}><MessageSquare /><span>{item.prompt}</span><ArrowUpRight /></button>) : <div className="mc-demo-empty"><Clock3 /><h3>从第一段想法开始</h3><p>本原型暂无对话。试着填入灵感并点击“开始创作”。</p></div>}</>}
                {dialog === "account" && <><p>示例账户 · 120 积分</p>{["充值与兑换", "账户与设置", "管理员后台", "退出登录"].map((label) => <button className="mc-demo-history" key={label} onClick={() => { setDialog(null); notify(label, "保留现有账户功能；设计原型不修改真实账户或登录状态。"); }}><span>{label}</span><ArrowUpRight /></button>)}</>}
                {dialog === "result" && <><p>这是静态结果占位，用于体验生成后的操作位置。没有发起 AI 请求，也没有扣除积分。</p><blockquote>{prompt}</blockquote>{mode === "text" ? <div className="mc-demo-empty"><MessageSquare /><h3>文本结果显示在这里</h3><p>沿用原有流式输出、继续对话与画布接续能力。</p></div> : <div className="mc-demo-result"><img src={chosenArt.image} alt="静态示例封面，并非生成结果" /><span>静态示意 · 非生成结果</span></div>}<div className="mc-demo-result-actions"><button onClick={() => { setDialog(null); focusPrompt(); }}><WandSparkles />调整后再创作</button><button onClick={() => notify("保存结果", "正式版沿用原有资源下载与资产保存。本原型没有生成新资源。") }><ArrowDownToLine />保存</button><button onClick={() => notify("画布中继续", "沿用现有创作会话与资源交接，把结果带入自由画布；不重写保存和同步逻辑。") }><LayoutDashboard />画布中继续</button></div></>}
            </div>
        </Modal>
        <Modal open={Boolean(notice)} title={notice?.title} onCancel={() => setNotice(null)} footer={null}><p style={{ whiteSpace: "pre-line", lineHeight: 1.9, paddingBlock: 12 }}>{notice?.body}</p></Modal>
    </div>;
}
