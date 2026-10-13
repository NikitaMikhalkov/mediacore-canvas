import { Drawer, type DrawerProps } from "antd";
import type { CSSProperties } from "react";
import { appGlassSurfaceStyle } from "@/lib/app-theme";

// 用途：产品侧栏壳。消费 token 由内容自己负责。默认 destroyOnHidden，可选 flush 去内边距。
export type AppDrawerProps = DrawerProps & {
    flush?: boolean;
};

export function AppDrawer({ flush = false, destroyOnHidden = true, styles, className, rootClassName, ...props }: AppDrawerProps) {
    const mergeResolvedStyles = (resolved: unknown) => {
        const base = resolved && typeof resolved === "object" ? (resolved as Record<string, unknown>) : {};
        const content = base.content && typeof base.content === "object" ? (base.content as CSSProperties) : {};
        const body = base.body && typeof base.body === "object" ? (base.body as CSSProperties) : {};
        return {
            ...base,
            content: {
                ...appGlassSurfaceStyle,
                ...content,
            },
            body: flush ? { ...body, padding: 0 } : body,
        };
    };
    const mergedStyles =
        typeof styles === "function"
            ? (info: { props: DrawerProps }) => mergeResolvedStyles(styles(info))
            : mergeResolvedStyles(styles);

    return <Drawer destroyOnHidden={destroyOnHidden} className={className} rootClassName={rootClassName} {...props} styles={mergedStyles as DrawerProps["styles"]} />;
}
