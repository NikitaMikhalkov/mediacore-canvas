import { App, Button, Form, Input, Skeleton } from "antd";
import { Switch } from "@/pages/admin/ui/controls";
import { AlertTriangle, BadgeCheck, FileText, RefreshCw, RotateCcw, Save, ShieldCheck, Sparkles, UserPlus, UsersRound } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { useBlocker } from "react-router";

import { useAppearanceStore } from "@/stores/use-appearance-store";
import { getAdminRegistrationSetting, updateAdminRegistrationSetting, type RegistrationSetting } from "@/services/api/wallet";
import { AdminStatusBadge, SettingsSectionCard } from "./admin-ui";

const DEFAULT_AGREEMENT_TEMPLATE = `一、服务说明与接受
欢迎使用本平台提供的 AI 影视与内容创作服务。在注册或使用本平台各项功能前，请务必仔细阅读并理解本协议。当您勾选同意或点击注册、登录按钮，即表示您已自愿接受本协议所有条款的约束。

二、账号注册与安全
1. 用户在注册时应提供真实、准确、合法的个人信息，并自行妥善保管账号密码及相关凭据。
2. 任何通过该账号进行的行为均视为用户本人的行为，因保管不当造成的损失由用户自行承担。

三、内容合规与用户承诺
1. 用户使用本平台生成、上传或发布的文本、图片、视频、音频等所有内容，必须遵守国家相关法律法规及公序良俗。
2. 严禁利用本平台从事任何违反国家法律、侵害他人合法权益（包括但不限于肖像权、名誉权、知识产权、个人隐私等）或危害网络安全的活动。

四、知识产权与免责声明
1. 用户利用本平台创作产生的合法作品归用户所有或依法享有相应权利。
2. AI 模型生成的内容受算法模型与提示词影响，平台对其完整性、准确性不作绝对保证，用户应自行审慎评估和使用。

五、服务变更与终止
平台有权根据业务发展、技术升级或合规要求适时调整服务内容及本协议条款。修改后的条款一经发布即发生效力。`;

export default function AccessSettingsPanel() {
    const { message, modal } = App.useApp();
    const brandName = useAppearanceStore((state) => state.appearance.brandName) || "平台";
    const [registration, setRegistration] = useState<RegistrationSetting | null>(null);
    const [loading, setLoading] = useState(true);
    const [refreshing, setRefreshing] = useState(false);
    const [savingRegistration, setSavingRegistration] = useState(false);
    const [savingAgreement, setSavingAgreement] = useState(false);
    const [agreementTitle, setAgreementTitle] = useState("");
    const [agreementContent, setAgreementContent] = useState("");
    const [agreementDirty, setAgreementDirty] = useState(false);
    const [loadError, setLoadError] = useState("");
    const [saveError, setSaveError] = useState("");
    const requestVersionRef = useRef(0);
    const navigationConfirmOpenRef = useRef(false);
    const navigationTriggerRef = useRef<HTMLElement | null>(null);

    const load = useCallback(
        async (initial = false, announce = false) => {
            const requestVersion = ++requestVersionRef.current;
            if (initial) setLoading(true);
            else setRefreshing(true);
            setLoadError("");
            try {
                const registrationData = await getAdminRegistrationSetting();
                if (requestVersion !== requestVersionRef.current) return;
                setRegistration(registrationData.setting);
                setAgreementTitle(registrationData.setting.agreementTitle || "");
                setAgreementContent(registrationData.setting.agreementContent || "");
                setAgreementDirty(false);
                setSaveError("");
                if (announce) message.success("已重新读取当前登录与注册配置");
            } catch (error) {
                if (requestVersion !== requestVersionRef.current) return;
                const errorMessage = error instanceof Error ? error.message : "读取登录与注册配置失败";
                setLoadError(errorMessage);
                if (!initial) message.error(errorMessage);
            } finally {
                if (requestVersion === requestVersionRef.current) {
                    setLoading(false);
                    setRefreshing(false);
                }
            }
        },
        [message],
    );

    useEffect(() => {
        void load(true);
        return () => {
            requestVersionRef.current += 1;
        };
    }, [load]);

    const unsaved = agreementDirty && !savingAgreement;
    const blocker = useBlocker(unsaved);

    // 条款编辑区下方的字数与段落提示：与注册页的分段规则保持一致（按空行分段），
    // 让管理员在保存前就能看到条款会被拆成几段。
    const agreementParagraphsCount = agreementContent.trim() ? agreementContent.trim().split(/\n\s*\n/).length : 0;

    useEffect(() => {
        const beforeUnload = (event: BeforeUnloadEvent) => {
            if (!unsaved) return;
            event.preventDefault();
        };
        window.addEventListener("beforeunload", beforeUnload);
        return () => window.removeEventListener("beforeunload", beforeUnload);
    }, [unsaved]);

    useEffect(() => {
        if (blocker.state !== "blocked" || navigationConfirmOpenRef.current) return;
        navigationConfirmOpenRef.current = true;
        navigationTriggerRef.current = document.activeElement instanceof HTMLElement && document.activeElement !== document.body ? document.activeElement : null;
        modal.confirm({
            title: "放弃未保存的登录与注册调整？",
            content: "当前有尚未保存的表单调整，离开后这些内容会丢失。用户注册状态不受影响。",
            okText: "放弃并离开",
            cancelText: "继续编辑",
            okButtonProps: { danger: true },
            onOk: () => {
                navigationConfirmOpenRef.current = false;
                navigationTriggerRef.current = null;
                blocker.proceed();
            },
            onCancel: () => {
                navigationConfirmOpenRef.current = false;
                blocker.reset();
                window.requestAnimationFrame(() => {
                    const fallback = document.querySelector<HTMLButtonElement>(".admin-access-command-actions button");
                    const target = navigationTriggerRef.current?.isConnected ? navigationTriggerRef.current : fallback;
                    target?.focus();
                    navigationTriggerRef.current = null;
                });
            },
        });
    }, [blocker, modal]);

    const saveAgreement = async () => {
        if (!registration || savingAgreement) return;
        setSavingAgreement(true);
        setSaveError("");
        try {
            const data = await updateAdminRegistrationSetting({
                enabled: registration.enabled,
                agreementTitle: agreementTitle.trim(),
                agreementContent: agreementContent.trim(),
            });
            setRegistration(data.setting);
            setAgreementTitle(data.setting.agreementTitle || "");
            setAgreementContent(data.setting.agreementContent || "");
            setAgreementDirty(false);
            message.success("服务协议已保存，注册页将立即生效");
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : "保存服务协议失败";
            setSaveError(`${errorMessage}。未自动重试，请重新读取当前配置后再决定是否保存。`);
            message.error(errorMessage);
        } finally {
            setSavingAgreement(false);
        }
    };

    const resetAgreementDraft = () => {
        if (!registration || savingAgreement) return;
        setAgreementTitle(registration.agreementTitle || "");
        setAgreementContent(registration.agreementContent || "");
        setAgreementDirty(false);
        message.info("已撤销服务协议的未保存调整");
    };

    const requestRefresh = () => {
        if (!agreementDirty) {
            void load(false, true);
            return;
        }
        modal.confirm({
            title: "放弃调整并重新读取？",
            content: "重新读取会丢弃当前表单中的未保存内容（含服务协议），并以服务端配置为准。",
            okText: "放弃并刷新",
            cancelText: "继续编辑",
            okButtonProps: { danger: true },
            onOk: () => load(false, true),
        });
    };

    const toggleRegistration = async (enabled: boolean) => {
        setSavingRegistration(true);
        try {
            const data = await updateAdminRegistrationSetting(enabled);
            setRegistration(data.setting);
            message.success(enabled ? "用户注册已开启" : "用户注册已关闭");
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : "更新注册设置失败";
            message.error(errorMessage);
            throw error;
        } finally {
            setSavingRegistration(false);
        }
    };

    const requestRegistrationChange = (enabled: boolean) => {
        if (!registration || enabled === registration.enabled || savingRegistration) return;
        void toggleRegistration(enabled).catch(() => undefined);
    };

    if (loading && !registration) {
        return (
            <div className="admin-settings-stack admin-access-settings" aria-label="正在读取登录与注册配置" role="status">
                <div className="admin-access-loading-card">
                    <Skeleton active paragraph={{ rows: 6 }} />
                </div>
            </div>
        );
    }

    if (!registration) {
        return (
            <div className="admin-settings-stack admin-access-settings">
                <div className="admin-access-load-error" role="alert">
                    <span className="admin-access-load-error-icon">
                        <AlertTriangle className="size-5" aria-hidden="true" />
                    </span>
                    <div>
                        <h2>无法读取登录与注册配置</h2>
                        <p>{loadError || "当前没有可显示的配置，请稍后重试。"}</p>
                    </div>
                    <Button icon={<RefreshCw className="size-4" />} loading={refreshing} onClick={() => void load(false, true)}>
                        重新读取
                    </Button>
                </div>
            </div>
        );
    }

    return (
        <div className="admin-settings-stack admin-access-settings">
            <div className="admin-access-command-bar">
                <div className="admin-access-command-copy" aria-live="polite">
                    <span className="admin-access-command-icon">
                        <ShieldCheck className="size-4" aria-hidden="true" />
                    </span>
                    <div>
                        <strong>登录与注册设置</strong>
                        <p>{`新用户注册${registration.enabled ? "已开放" : "已关闭"}`}</p>
                    </div>
                </div>
                <div className="admin-access-command-actions">
                    <Button icon={<RefreshCw className="size-4" />} loading={refreshing} disabled={savingAgreement || savingRegistration} onClick={requestRefresh}>
                        刷新状态
                    </Button>
                </div>
            </div>

            {loadError || saveError ? (
                <div className="admin-access-inline-alert" role="alert">
                    <AlertTriangle className="size-4 shrink-0" aria-hidden="true" />
                    <span>{saveError || `${loadError}。页面仍显示上一次成功读取的配置。`}</span>
                </div>
            ) : null}

            <div id="admin-access-registration" className="admin-settings-anchor">
                <SettingsSectionCard
                    className="admin-access-section admin-access-registration-section"
                    icon={<UserPlus className="size-4" aria-hidden="true" />}
                    title="1. 是否允许创建新账号"
                    description="先决定是否开放注册。关闭后已有账号仍可登录，已有数据不会删除。"
                    status={<AdminStatusBadge label={registration.enabled ? "已开放" : "已关闭"} tone={registration.enabled ? "success" : "neutral"} />}
                >
                    <div className="admin-access-registration-policy">
                        <span className="admin-access-policy-icon">
                            <UsersRound className="size-5" aria-hidden="true" />
                        </span>
                        <div className="admin-access-policy-copy">
                            <div className="flex flex-wrap items-center gap-2">
                                <strong>允许创建新账号</strong>
                                <AdminStatusBadge label="切换即保存" tone="info" />
                            </div>
                            <p>关闭后，新用户注册会被拒绝；已有账号仍可继续登录。</p>
                            <span>{formatSettingTime(registration.updatedAt, "当前来自部署环境默认值")}</span>
                        </div>
                        <Switch checked={registration.enabled} loading={savingRegistration} disabled={loading || refreshing} onChange={requestRegistrationChange} aria-label="允许创建新账号，切换后立即生效" />
                    </div>
                </SettingsSectionCard>
            </div>

            <div id="admin-access-agreement" className="admin-settings-anchor">
                <SettingsSectionCard
                    className="admin-access-section admin-access-agreement-section"
                    icon={<FileText className="size-4" aria-hidden="true" />}
                    title="2. 服务协议名称与条款"
                    description="注册页勾选框与弹窗展示的内容。协议名称留空时自动跟随品牌名生成，条款留空时注册页给出待补充提示。"
                    status={<AdminStatusBadge label={agreementDirty ? "待保存" : agreementContent.trim() ? "已自定义" : "待补充"} tone={agreementDirty ? "warning" : agreementContent.trim() ? "success" : "neutral"} />}
                    footer={
                        <>
                            <div className="admin-access-footer-note">
                                <BadgeCheck className="size-4" aria-hidden="true" />
                                <span>保存后注册页立即生效，无需重启服务</span>
                            </div>
                            <div className="flex flex-wrap items-center gap-2">
                                {agreementDirty ? (
                                    <Button icon={<RotateCcw className="size-4" />} disabled={savingAgreement} onClick={resetAgreementDraft}>
                                        撤销
                                    </Button>
                                ) : null}
                                <Button type="primary" icon={<Save className="size-4" />} loading={savingAgreement} disabled={!agreementDirty || loading || refreshing} onClick={() => void saveAgreement()}>
                                    保存服务协议
                                </Button>
                            </div>
                        </>
                    }
                >
                    <Form layout="vertical" requiredMark={false} disabled={loading || refreshing || savingAgreement}>
                        <Form.Item
                            label="协议名称"
                            extra={`留空时自动跟随品牌名生成，即《${brandName}服务协议》。此处只需填名称本身，书名号由注册页补充。`}
                            className="admin-access-agreement-title-field"
                        >
                            <Input
                                value={agreementTitle}
                                maxLength={60}
                                placeholder={`${brandName}服务协议`}
                                onChange={(event) => {
                                    setAgreementTitle(event.target.value);
                                    setAgreementDirty(true);
                                }}
                            />
                        </Form.Item>
                        <Form.Item
                            label="协议条款内容"
                            extra="支持多段落文本，按空行分段展示；留空时注册页弹窗显示“服务协议内容待补充”。"
                            className="admin-access-agreement-content-field"
                        >
                            <Input.TextArea
                                value={agreementContent}
                                maxLength={20000}
                                autoSize={{ minRows: 8, maxRows: 20 }}
                                placeholder="输入服务协议条款正文"
                                onChange={(event) => {
                                    setAgreementContent(event.target.value);
                                    setAgreementDirty(true);
                                }}
                            />
                        </Form.Item>
                        <div className="admin-access-agreement-toolbar">
                            <div className="admin-access-agreement-template-note">
                                <Sparkles className="size-3.5 shrink-0" aria-hidden="true" />
                                <span>没有准备条款时，可先填入通用模板再按需修改。</span>
                            </div>
                            <div className="admin-access-agreement-template-actions">
                                <span className="admin-access-agreement-metric">
                                    {agreementContent.trim() ? `${agreementParagraphsCount} 段 · ${agreementContent.length} 字` : "尚未填写条款"}
                                </span>
                                <Button
                                    size="small"
                                    disabled={savingAgreement || loading || refreshing}
                                    onClick={() => {
                                        setAgreementContent(DEFAULT_AGREEMENT_TEMPLATE);
                                        setAgreementDirty(true);
                                    }}
                                >
                                    填入通用模板
                                </Button>
                                {agreementContent.trim() ? (
                                    <Button
                                        size="small"
                                        type="text"
                                        danger
                                        disabled={savingAgreement || loading || refreshing}
                                        onClick={() => {
                                            setAgreementContent("");
                                            setAgreementDirty(true);
                                        }}
                                    >
                                        清空
                                    </Button>
                                ) : null}
                            </div>
                        </div>
                    </Form>
                </SettingsSectionCard>
            </div>
        </div>
    );
}

function formatSettingTime(value: string | undefined, fallback: string) {
    if (!value) return fallback;
    const date = new Date(value);
    if (Number.isNaN(date.getTime()) || date.getFullYear() < 2000) return fallback;
    return `更新于 ${date.toLocaleString("zh-CN", { hour12: false })}`;
}
