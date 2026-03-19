"use client";

import Link from "next/link";
import { FormEvent, ReactNode, useEffect, useRef, useState } from "react";
import {
  ArrowUp,
  ChevronDown,
  Loader2,
  Sparkles,
  ThumbsDown,
  ThumbsUp,
  X,
} from "lucide-react";
import {
  apiClient,
  AssistantCard,
  AssistantChatMessage,
  AssistantConversation,
  AssistantInsight,
} from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/contexts/auth-context";
import { useAssistantPageContext } from "@/contexts/assistant-page-context";

const STORAGE_KEY = "furry_assistant_messages_v1";
const CONVERSATION_KEY = "furry_assistant_current_conversation_v1";

const START_PROMPT = "能快速带我了解这个社区吗？";

const WELCOME_CARDS: AssistantCard[] = [
  {
    kind: "page",
    title: "发现页",
    summary: "适合第一次来先看热门内容和社区氛围。",
    href: "/explore",
    meta: "/explore",
  },
  {
    kind: "page",
    title: "圈子广场",
    summary: "按兴趣找同好，快速加入你喜欢的圈子。",
    href: "/groups",
    meta: "/groups",
  },
  {
    kind: "page",
    title: "活动广场",
    summary: "看近期线上线下活动，想参加可以直接报名。",
    href: "/events",
    meta: "/events",
  },
];

const WELCOME_MESSAGE: AssistantChatMessage = {
  role: "assistant",
  content:
    "我是霜牙，你的站内导览助手。你可以问我“先逛哪里”“推荐几个圈子”“最近有什么活动”“怎么发第一条动态”。",
  cards: WELCOME_CARDS,
};

const SOURCE_KIND_LABELS: Record<string, string> = {
  page: "页面",
  post: "帖子",
  user: "用户",
  tag: "标签",
  group: "圈子",
  event: "活动",
};

const INSIGHT_KIND_LABELS: Record<string, string> = {
  draft_polish: "正文润色",
  title_options: "标题备选",
  tag_suggestions: "标签建议",
  visibility_suggestion: "可见性",
  group_atmosphere: "圈子氛围",
  rules_summary: "规则摘要",
  join_suggestion: "加入建议",
  posting_ideas: "发帖方向",
  event_summary: "活动概览",
  fit_assessment: "适配判断",
  signup_notes: "参与提示",
  preparation_checklist: "准备清单",
};

function MascotAvatar({ compact = false }: { compact?: boolean }) {
  const size = compact ? "h-11 w-11" : "h-14 w-14";
  const ear = compact ? "h-4 w-4" : "h-5 w-5";
  return (
    <div className={`relative ${size}`}>
      <div
        className={`absolute left-1 top-0 ${ear} rotate-[-18deg] rounded-t-[14px] rounded-bl-[10px] bg-amber-200 shadow-sm`}
      />
      <div
        className={`absolute right-1 top-0 ${ear} rotate-[18deg] rounded-t-[14px] rounded-br-[10px] bg-amber-200 shadow-sm`}
      />
      <div className="absolute inset-x-0 top-1 bottom-0 rounded-[22px] bg-gradient-to-br from-amber-100 via-orange-100 to-orange-200 shadow-[0_10px_25px_rgba(251,146,60,0.35)]">
        <div className="absolute inset-x-2 top-3 h-5 rounded-full bg-slate-800/90">
          <div className="absolute left-2 top-1 h-1.5 w-1.5 rounded-full bg-cyan-300 shadow-[0_0_8px_rgba(34,211,238,0.9)]" />
          <div className="absolute right-2 top-1 h-1.5 w-1.5 rounded-full bg-cyan-300 shadow-[0_0_8px_rgba(34,211,238,0.9)]" />
        </div>
        <div className="absolute inset-x-0 bottom-3 flex items-center justify-center gap-1">
          <span className="h-2.5 w-2.5 rounded-full bg-rose-300/70" />
          <span className="h-1.5 w-2 rounded-full bg-slate-700" />
          <span className="h-2.5 w-2.5 rounded-full bg-rose-300/70" />
        </div>
      </div>
      {!compact && (
        <div className="absolute -right-1 -top-1 rounded-full bg-slate-950 px-1.5 py-0.5 text-[10px] font-semibold text-cyan-200 ring-1 ring-cyan-400/30">
          AI
        </div>
      )}
    </div>
  );
}

function CardList({ cards }: { cards?: AssistantCard[] }) {
  if (!cards || cards.length === 0) return null;

  return (
    <div className="grid gap-2.5">
      {cards.map((card) => (
        <Link
          key={`${card.kind}-${card.href}-${card.title}`}
          id={card.ref ? `ref-${card.ref}` : undefined}
          href={card.href}
          className="group rounded-[22px] border border-slate-200/85 bg-[linear-gradient(180deg,rgba(255,255,255,0.96),rgba(255,247,237,0.86))] p-3.5 transition-all hover:-translate-y-0.5 hover:border-orange-300 hover:shadow-[0_16px_30px_rgba(251,146,60,0.16)] dark:border-slate-800 dark:bg-[linear-gradient(180deg,rgba(15,23,42,0.94),rgba(15,23,42,0.72))] dark:hover:border-orange-500/50"
        >
          <div className="mb-2 flex items-center gap-2">
            {card.ref && (
              <span className="rounded-full bg-slate-950 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-white dark:bg-orange-500 dark:text-slate-950">
                {card.ref}
              </span>
            )}
            <span className="rounded-full bg-orange-100 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.18em] text-orange-700 dark:bg-orange-950/70 dark:text-orange-300">
              {SOURCE_KIND_LABELS[card.kind] || card.kind}
            </span>
            {card.meta && (
              <span className="truncate text-[11px] text-muted-foreground">
                {card.meta}
              </span>
            )}
          </div>
          <p className="text-sm font-semibold text-slate-900 transition-colors group-hover:text-orange-700 dark:text-slate-100 dark:group-hover:text-orange-200">
            {card.title}
          </p>
          <p className="mt-1.5 text-xs leading-5 text-slate-600 dark:text-slate-400">
            {card.summary}
          </p>
          {(card.reason || card.source) && (
            <div className="mt-3 space-y-1 border-t border-orange-100/80 pt-3 text-[11px] leading-5 text-slate-500 dark:border-slate-800 dark:text-slate-400">
              {card.reason ? <p>推荐理由：{card.reason}</p> : null}
              {card.source ? <p>来源：{card.source}</p> : null}
            </div>
          )}
        </Link>
      ))}
    </div>
  );
}

function InsightList({ insights }: { insights?: AssistantInsight[] }) {
  if (!insights || insights.length === 0) return null;

  return (
    <div className="grid gap-2.5">
      {insights.map((insight, index) => (
        <div
          key={`${insight.kind}-${index}`}
          className="rounded-[22px] border border-cyan-200/80 bg-[linear-gradient(180deg,rgba(236,254,255,0.88),rgba(255,255,255,0.94))] p-3.5 dark:border-cyan-900/40 dark:bg-[linear-gradient(180deg,rgba(8,47,73,0.22),rgba(15,23,42,0.54))]"
        >
          <div className="mb-2 flex items-center gap-2">
            <span className="rounded-full bg-cyan-100 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.16em] text-cyan-800 dark:bg-cyan-950/60 dark:text-cyan-200">
              {INSIGHT_KIND_LABELS[insight.kind] || insight.kind}
            </span>
            <span className="text-sm font-semibold text-slate-900 dark:text-slate-100">
              {insight.title}
            </span>
          </div>
          {insight.summary && (
            <p className="text-xs leading-5 text-slate-700 dark:text-slate-300">
              {insight.summary}
            </p>
          )}
          {insight.bullets?.length ? (
            <ul className="mt-2 space-y-1.5 pl-4 text-xs leading-5 text-slate-600 dark:text-slate-300">
              {insight.bullets.map((bullet, bulletIndex) => (
                <li key={`${insight.kind}-${bulletIndex}`} className="list-disc">
                  {bullet}
                </li>
              ))}
            </ul>
          ) : null}
        </div>
      ))}
    </div>
  );
}

function AttachmentSection({
  title,
  summary,
  badge,
  expanded,
  onToggle,
  children,
}: {
  title: string;
  summary: string;
  badge: string;
  expanded: boolean;
  onToggle: () => void;
  children: ReactNode;
}) {
  return (
    <div className="mt-4 overflow-hidden rounded-[24px] border border-slate-200/85 bg-white/75 shadow-[0_10px_24px_rgba(15,23,42,0.05)] backdrop-blur-sm dark:border-slate-800 dark:bg-slate-950/40">
      <button
        type="button"
        onClick={onToggle}
        className="flex w-full items-center justify-between gap-3 px-3.5 py-3 text-left transition-colors hover:bg-slate-50/70 dark:hover:bg-slate-900/50"
      >
        <div className="min-w-0">
          <p className="text-xs font-semibold uppercase tracking-[0.16em] text-slate-800 dark:text-slate-100">
            {title}
          </p>
          <p className="mt-1 text-[11px] leading-5 text-slate-500 dark:text-slate-400">
            {summary}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <span className="rounded-full bg-slate-100 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.14em] text-slate-700 dark:bg-slate-800 dark:text-slate-200">
            {badge}
          </span>
          <ChevronDown
            className={`h-4 w-4 shrink-0 text-slate-400 transition-transform dark:text-slate-500 ${
              expanded ? "rotate-180" : ""
            }`}
          />
        </div>
      </button>
      {expanded ? (
        <div className="border-t border-slate-200/80 px-3.5 py-3 dark:border-slate-800">
          {children}
        </div>
      ) : null}
    </div>
  );
}

function renderInlineMarkdown(text: string): ReactNode[] {
  const nodes: ReactNode[] = [];
  let remaining = text;
  let key = 0;

  while (remaining.length > 0) {
    const citationMatch = remaining.match(/\[(R\d+)\]/);
    const linkMatch = remaining.match(/\[([^\]]+)\]\(([^)]+)\)/);
    const boldMatch = remaining.match(/\*\*([^*]+)\*\*/);
    const codeMatch = remaining.match(/`([^`]+)`/);

    const matches = [citationMatch, linkMatch, boldMatch, codeMatch]
      .filter((item): item is RegExpMatchArray => !!item)
      .map((item) => ({
        match: item,
        index: item.index ?? 0,
      }))
      .sort((a, b) => a.index - b.index);

    if (matches.length === 0) {
      nodes.push(<span key={`text-${key++}`}>{remaining}</span>);
      break;
    }

    const { match, index } = matches[0];
    if (index > 0) {
      nodes.push(
        <span key={`text-${key++}`}>{remaining.slice(0, index)}</span>,
      );
    }

    if (match[0] === citationMatch?.[0]) {
      const ref = citationMatch?.[1] ?? "";
      nodes.push(
        <a
          key={`cite-${key++}`}
          href={`#ref-${ref}`}
          className="mx-0.5 inline-flex rounded-full bg-orange-100 px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-orange-800 no-underline dark:bg-orange-950/60 dark:text-orange-200"
        >
          {ref}
        </a>,
      );
    } else if (match[0] === linkMatch?.[0]) {
      const href = linkMatch?.[2] ?? "#";
      nodes.push(
        <a
          key={`link-${key++}`}
          href={href}
          target={href.startsWith("http") ? "_blank" : undefined}
          rel={href.startsWith("http") ? "noopener noreferrer" : undefined}
          className="font-medium text-orange-600 underline decoration-orange-300 underline-offset-4 dark:text-orange-300"
        >
          {linkMatch?.[1]}
        </a>,
      );
    } else if (match[0] === boldMatch?.[0]) {
      nodes.push(
        <strong
          key={`strong-${key++}`}
          className="font-semibold text-foreground"
        >
          {boldMatch?.[1]}
        </strong>,
      );
    } else if (match[0] === codeMatch?.[0]) {
      nodes.push(
        <code
          key={`code-${key++}`}
          className="rounded-md bg-slate-900/90 px-1.5 py-0.5 text-[12px] text-orange-100 dark:bg-slate-800"
        >
          {codeMatch?.[1]}
        </code>,
      );
    }

    remaining = remaining.slice(index + match[0].length);
  }

  return nodes;
}

function AssistantMarkdown({ text }: { text: string }) {
  const blocks = text.split(/\n{2,}/).filter((item) => item.trim());

  return (
    <div className="space-y-4 text-[15px] leading-7">
      {blocks.map((block, blockIndex) => {
        const lines = block.split("\n").filter((item) => item.trim());
        const isList = lines.every(
          (line) =>
            line.trim().startsWith("- ") || line.trim().startsWith("* "),
        );

        if (isList) {
          return (
            <ul
              key={`list-${blockIndex}`}
              className="space-y-2 pl-5 text-[15px] leading-7 text-inherit"
            >
              {lines.map((line, lineIndex) => (
                <li
                  key={`item-${blockIndex}-${lineIndex}`}
                  className="list-disc"
                >
                  {renderInlineMarkdown(line.trim().slice(2))}
                </li>
              ))}
            </ul>
          );
        }

        return (
          <p key={`p-${blockIndex}`} className="text-[15px] leading-7 text-inherit">
            {lines.map((line, lineIndex) => (
              <span key={`line-${blockIndex}-${lineIndex}`}>
                {renderInlineMarkdown(line)}
                {lineIndex < lines.length - 1 && <br />}
              </span>
            ))}
          </p>
        );
      })}
    </div>
  );
}

export function FurryAssistant() {
  const { isLoggedIn } = useAuth();
  const { pageContext } = useAssistantPageContext();
  const [open, setOpen] = useState(false);
  const [messages, setMessages] = useState<AssistantChatMessage[]>([
    WELCOME_MESSAGE,
  ]);
  const [conversations, setConversations] = useState<AssistantConversation[]>(
    [],
  );
  const [conversationId, setConversationId] = useState<string | null>(null);
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [error, setError] = useState("");
  const [providerLabel, setProviderLabel] = useState("AI");
  const [fallbackMode, setFallbackMode] = useState(false);
  const [intentLabel, setIntentLabel] = useState("综合导览");
  const [sourceCounts, setSourceCounts] = useState<Record<string, number>>({});
  const [feedbackLoadingId, setFeedbackLoadingId] = useState<string | null>(null);
  const [showConversationRail, setShowConversationRail] = useState(false);
  const [expandedSections, setExpandedSections] = useState<Record<string, boolean>>(
    {},
  );
  const scrollerRef = useRef<HTMLDivElement>(null);
  const abortRef = useRef<AbortController | null>(null);
  const lastMessageCountRef = useRef(0);

  async function loadConversationList() {
    if (!isLoggedIn) return;
    try {
      const data = await apiClient.getAssistantConversations(1, 12);
      setConversations(data.conversations ?? []);
    } catch {
      setConversations([]);
    }
  }

  async function openConversation(id: string) {
    if (!isLoggedIn) return;
    abortRef.current?.abort();
    abortRef.current = null;
    setLoading(false);
    setHistoryLoading(true);
    setError("");
    setFallbackMode(false);
    setProviderLabel("AI");
    setIntentLabel("综合导览");
    setSourceCounts({});
    setExpandedSections({});
    try {
      const data = await apiClient.getAssistantConversation(id);
      setConversationId(data.conversation.id);
      setMessages(data.messages?.length ? data.messages : [WELCOME_MESSAGE]);
      localStorage.setItem(CONVERSATION_KEY, data.conversation.id);
    } catch (err) {
      const message = err instanceof Error ? err.message : "读取对话失败";
      setError(message);
    } finally {
      setHistoryLoading(false);
    }
  }

  useEffect(() => {
    if (isLoggedIn) {
      localStorage.removeItem(STORAGE_KEY);
      setHistoryLoading(true);
      void (async () => {
        try {
          const data = await apiClient.getAssistantConversations(1, 12);
          setConversations(data.conversations ?? []);

          const savedConversationId = localStorage.getItem(CONVERSATION_KEY);
          if (savedConversationId) {
            try {
              const detail =
                await apiClient.getAssistantConversation(savedConversationId);
              setConversationId(detail.conversation.id);
              setMessages(
                detail.messages?.length ? detail.messages : [WELCOME_MESSAGE],
              );
              return;
            } catch {
              localStorage.removeItem(CONVERSATION_KEY);
            }
          }
          setConversationId(null);
          setMessages([WELCOME_MESSAGE]);
          setFallbackMode(false);
          setProviderLabel("AI");
          setIntentLabel("综合导览");
          setSourceCounts({});
          setExpandedSections({});
        } finally {
          setHistoryLoading(false);
        }
      })();
      return;
    }

    setConversations([]);
    setConversationId(null);
    setHistoryLoading(false);
    setFallbackMode(false);
    setProviderLabel("AI");
    setIntentLabel("综合导览");
    setSourceCounts({});
    setExpandedSections({});
    localStorage.removeItem(CONVERSATION_KEY);

    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      setMessages([WELCOME_MESSAGE]);
      return;
    }
    try {
      const parsed = JSON.parse(raw) as AssistantChatMessage[];
      if (Array.isArray(parsed) && parsed.length > 0) {
        setMessages(parsed);
      } else {
        setMessages([WELCOME_MESSAGE]);
      }
    } catch {
      setMessages([WELCOME_MESSAGE]);
    }
  }, [isLoggedIn]);

  useEffect(() => {
    if (isLoggedIn) return;
    const next = messages.slice(-20);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
  }, [isLoggedIn, messages]);

  useEffect(() => {
    if (!open || !scrollerRef.current) return;
    const container = scrollerRef.current;
    const behavior: ScrollBehavior =
      loading || messages.length === lastMessageCountRef.current ? "auto" : "smooth";
    lastMessageCountRef.current = messages.length;
    requestAnimationFrame(() => {
      container.scrollTo({
        top: container.scrollHeight,
        behavior,
      });
    });
  }, [historyLoading, loading, messages, open]);

  useEffect(() => {
    return () => abortRef.current?.abort();
  }, []);

  async function askAssistant(question: string) {
    const trimmed = question.trim();
    if (!trimmed || loading) return;

    const userMessage: AssistantChatMessage = {
      role: "user",
      content: trimmed,
    };
    const pendingAssistant: AssistantChatMessage = {
      role: "assistant",
      content: "",
      cards: [],
    };
    const nextMessages = [...messages, userMessage, pendingAssistant];

    setMessages(nextMessages);
    setInput("");
    setError("");
    setLoading(true);
    setIntentLabel("综合导览");
    setSourceCounts({});
    setShowConversationRail(false);

    const controller = new AbortController();
    abortRef.current = controller;

    try {
      await apiClient.streamAssistantChat(
        nextMessages,
        {
          signal: controller.signal,
          onMeta: (meta) => {
            setProviderLabel(
              meta.provider === "deepseek" ? "DeepSeek" : meta.provider || "AI",
            );
            setFallbackMode(meta.fallback);
            setIntentLabel(meta.intent_label || "综合导览");
            setSourceCounts(meta.source_counts ?? {});
            if (meta.conversation_id) {
              setConversationId(meta.conversation_id);
              localStorage.setItem(CONVERSATION_KEY, meta.conversation_id);
            }
            setMessages((prev) => {
              const copy = [...prev];
              const last = copy[copy.length - 1];
              if (!last || last.role !== "assistant") return copy;
              copy[copy.length - 1] = {
                ...last,
                id: meta.response_id || last.id,
                cards: meta.cards,
                insights: meta.insights,
                provider: meta.provider,
                fallback: meta.fallback,
                intent: meta.intent,
                source_counts: meta.source_counts,
                feedback_value: undefined,
              };
              return copy;
            });
          },
          onToken: (token) => {
            setMessages((prev) => {
              const copy = [...prev];
              const last = copy[copy.length - 1];
              if (!last || last.role !== "assistant") return copy;
              copy[copy.length - 1] = {
                ...last,
                content: `${last.content}${token}`,
              };
              return copy;
            });
          },
          onError: (message) => {
            setError(message);
          },
        },
        conversationId ?? undefined,
        pageContext ?? undefined,
      );
    } catch (err) {
      if (controller.signal.aborted) {
        setMessages((prev) => {
          const copy = [...prev];
          const last = copy[copy.length - 1];
          if (!last || last.role !== "assistant") return copy;
          if (last.content.trim()) return copy;
          copy[copy.length - 1] = {
            ...last,
            content: "这一段我先停住了。你可以继续追问，或者换个问法。",
          };
          return copy;
        });
      } else {
        const message =
          err instanceof Error ? err.message : "AI 助手暂时不可用";
        setError(message);
        setMessages((prev) => {
          const copy = [...prev];
          const last = copy[copy.length - 1];
          if (!last || last.role !== "assistant") return copy;
          if (last.content.trim()) return copy;
          copy[copy.length - 1] = {
            ...last,
            content:
              "我刚才没有顺利生成回复，但下面这些站内入口仍然值得你先看看。",
          };
          return copy;
        });
      }
    } finally {
      abortRef.current = null;
      setLoading(false);
      if (isLoggedIn) {
        void loadConversationList();
      }
    }
  }

  function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    void askAssistant(input);
  }

  function handleStop() {
    abortRef.current?.abort();
    abortRef.current = null;
    setLoading(false);
  }

  function clearConversation() {
    abortRef.current?.abort();
    abortRef.current = null;
    setMessages([WELCOME_MESSAGE]);
    setError("");
    setLoading(false);
    setFallbackMode(false);
    setProviderLabel("AI");
    setIntentLabel("综合导览");
    setSourceCounts({});
    setExpandedSections({});
    setShowConversationRail(false);
    if (isLoggedIn) {
      setConversationId(null);
      localStorage.removeItem(CONVERSATION_KEY);
      return;
    }
    localStorage.removeItem(STORAGE_KEY);
  }

  function findQueryForAssistantMessage(index: number) {
    for (let i = index - 1; i >= 0; i--) {
      if (messages[i]?.role === "user") {
        return messages[i]?.content || "";
      }
    }
    return "";
  }

  async function handleFeedback(
    message: AssistantChatMessage,
    index: number,
    value: "helpful" | "unhelpful",
  ) {
    if (!message.id || feedbackLoadingId === message.id) return;

    setFeedbackLoadingId(message.id);
    setError("");
    try {
      await apiClient.submitAssistantFeedback({
        response_id: message.id,
        conversation_id: conversationId ?? undefined,
        value,
        query: findQueryForAssistantMessage(index),
        reply_excerpt: message.content,
        provider: message.provider,
        intent: message.intent,
        fallback: message.fallback,
        page_path: pageContext?.path,
        source_counts: message.source_counts,
        cards: message.cards,
      });
      setMessages((prev) =>
        prev.map((item, itemIndex) =>
          itemIndex === index
            ? {
                ...item,
                feedback_value: value,
              }
            : item,
        ),
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "反馈提交失败");
    } finally {
      setFeedbackLoadingId(null);
    }
  }

  function toggleSection(sectionKey: string) {
    setExpandedSections((prev) => ({
      ...prev,
      [sectionKey]: !prev[sectionKey],
    }));
  }

  function handlePanelWheel(event: React.WheelEvent<HTMLDivElement>) {
    const container = scrollerRef.current;
    const target = event.target as HTMLElement | null;
    if (!container || !target || target.closest("textarea")) {
      return;
    }
    if (container.scrollHeight <= container.clientHeight) {
      return;
    }
    container.scrollTop += event.deltaY;
    event.preventDefault();
  }

  const userMessageCount = messages.filter((msg) => msg.role === "user").length;
  const sourceSummary = Object.entries(sourceCounts)
    .sort(([kindA], [kindB]) => kindA.localeCompare(kindB))
    .map(([kind, count]) => `${count} 个${SOURCE_KIND_LABELS[kind] || kind}`)
    .join(" · ");

  return (
    <div className="pointer-events-none fixed bottom-5 right-4 z-[60] sm:bottom-6 sm:right-6">
      {open && (
        <div
          className="pointer-events-auto relative mb-4 flex h-[min(84vh,820px)] w-[min(100vw-1rem,34rem)] min-h-0 flex-col overflow-hidden rounded-[32px] border border-orange-200/70 bg-[linear-gradient(180deg,rgba(255,248,240,0.96),rgba(255,255,255,0.92))] shadow-[0_32px_90px_rgba(15,23,42,0.24)] backdrop-blur dark:border-orange-900/50 dark:bg-[linear-gradient(180deg,rgba(17,24,39,0.98),rgba(2,6,23,0.94))] sm:w-[min(100vw-2rem,36rem)]"
          onWheelCapture={handlePanelWheel}
        >
          <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_right,rgba(251,191,36,0.18),transparent_36%),radial-gradient(circle_at_top_left,rgba(34,211,238,0.12),transparent_28%)]" />
          <div className="relative border-b border-orange-200/70 bg-white/72 px-4 py-4 backdrop-blur-xl dark:border-orange-900/50 dark:bg-slate-950/60">
            <div className="flex items-start gap-3">
              <MascotAvatar />
              <div className="min-w-0 flex-1">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="text-sm font-semibold tracking-[0.02em] text-slate-900 dark:text-slate-100">
                      霜牙
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    {isLoggedIn && conversations.length > 0 ? (
                      <button
                        type="button"
                        onClick={() => setShowConversationRail((prev) => !prev)}
                        className="rounded-full border border-slate-200 bg-white/80 px-3 py-1.5 text-[11px] font-medium text-slate-600 transition-colors hover:border-orange-300 hover:text-slate-900 dark:border-slate-700 dark:bg-slate-900/70 dark:text-slate-300 dark:hover:border-orange-500/50 dark:hover:text-slate-100"
                      >
                        {showConversationRail ? "收起历史" : "历史对话"}
                      </button>
                    ) : null}
                    <button
                      type="button"
                      onClick={() => setOpen(false)}
                      className="rounded-full p-1 text-slate-500 transition-colors hover:bg-orange-100 hover:text-slate-900 dark:hover:bg-slate-800 dark:hover:text-slate-100"
                      aria-label="关闭助手"
                    >
                      <X className="h-4 w-4" />
                    </button>
                  </div>
                </div>
                <div className="mt-4 flex flex-wrap items-center gap-2">
                  <span className="inline-flex items-center gap-1 rounded-full bg-emerald-100 px-2.5 py-1 text-[11px] font-medium text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-300">
                    <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
                    {fallbackMode
                      ? "站内检索模式"
                      : `${providerLabel} + 站内检索`}
                  </span>
                  <span className="rounded-full bg-slate-100 px-2.5 py-1 text-[11px] text-slate-600 dark:bg-slate-800 dark:text-slate-300">
                    意图：{intentLabel}
                  </span>
                  <button
                    type="button"
                    onClick={clearConversation}
                    className="rounded-full bg-white/75 px-2.5 py-1 text-[11px] text-muted-foreground transition-colors hover:text-foreground dark:bg-slate-900/70"
                  >
                    {isLoggedIn ? "新建对话" : "清空对话"}
                  </button>
                </div>
                <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px] text-slate-500 dark:text-slate-400">
                  {pageContext && (
                    <span className="rounded-full bg-slate-100 px-2.5 py-1 dark:bg-slate-800">
                      场景：{pageContext.title}
                    </span>
                  )}
                  {sourceSummary ? (
                    <span className="rounded-full bg-slate-100 px-2.5 py-1 dark:bg-slate-800">
                      已结合：{sourceSummary}
                    </span>
                  ) : null}
                </div>
              </div>
            </div>
          </div>

          <div
            ref={scrollerRef}
            className="relative min-h-0 flex-1 space-y-5 overflow-y-auto px-4 pb-6 pt-4 overscroll-contain sm:px-5"
          >
            {isLoggedIn && conversations.length > 0 && showConversationRail && (
              <div className="rounded-[24px] border border-slate-200/80 bg-white/72 p-3.5 shadow-[0_14px_32px_rgba(15,23,42,0.06)] backdrop-blur-sm dark:border-slate-800 dark:bg-slate-950/40">
                <p className="text-[11px] font-medium uppercase tracking-[0.18em] text-muted-foreground">
                  最近对话
                </p>
                <div className="mt-3 flex gap-2 overflow-x-auto pb-1">
                  {conversations.slice(0, 8).map((conversation) => (
                    <button
                      key={conversation.id}
                      type="button"
                      onClick={() => void openConversation(conversation.id)}
                      className={`min-w-[160px] max-w-[220px] rounded-[20px] border px-3 py-2.5 text-left transition-colors ${
                        conversationId === conversation.id
                          ? "border-orange-400 bg-orange-50 dark:border-orange-500 dark:bg-orange-950/40"
                          : "border-border bg-background/70 hover:border-orange-300 hover:bg-orange-50/70 dark:hover:bg-slate-900"
                      }`}
                    >
                      <p className="truncate text-xs font-semibold text-foreground">
                        {conversation.title}
                      </p>
                      <p className="mt-1 truncate text-[11px] text-muted-foreground">
                        {conversation.last_message_preview || "打开继续聊天"}
                      </p>
                    </button>
                  ))}
                </div>
              </div>
            )}

            {historyLoading && (
              <div className="mr-8">
                <div className="inline-flex items-center gap-2 rounded-full border border-orange-200/80 bg-white/85 px-3 py-2 text-xs text-slate-600 shadow-sm dark:border-slate-700 dark:bg-slate-900/90 dark:text-slate-300">
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                  正在读取历史会话
                </div>
              </div>
            )}

            {messages.map((message, index) => (
              <div
                key={`${message.role}-${index}`}
                className={
                  message.role === "user"
                    ? "ml-14 flex justify-end"
                    : "mr-3 sm:mr-10"
                }
              >
                <div
                  className={
                    message.role === "user"
                      ? "max-w-[80%] rounded-[24px] rounded-br-[10px] bg-[linear-gradient(135deg,#0f172a,#334155)] px-4 py-3 text-sm leading-6 text-white shadow-[0_16px_32px_rgba(15,23,42,0.24)] dark:bg-[linear-gradient(135deg,#f97316,#fb923c)] dark:text-slate-950"
                      : "max-w-[94%] rounded-[28px] rounded-bl-[12px] border border-orange-200/80 bg-[linear-gradient(180deg,rgba(255,255,255,0.96),rgba(255,247,237,0.88))] px-4 py-4 text-slate-800 shadow-[0_20px_40px_rgba(15,23,42,0.08)] backdrop-blur-sm dark:border-slate-700 dark:bg-[linear-gradient(180deg,rgba(15,23,42,0.95),rgba(15,23,42,0.84))] dark:text-slate-100"
                  }
                >
                  {message.role === "assistant" && (
                    <div className="mb-3 flex flex-wrap items-center gap-2 border-b border-orange-100/80 pb-3 text-[10px] uppercase tracking-[0.16em] text-slate-500 dark:border-slate-800 dark:text-slate-400">
                      <span className="rounded-full bg-slate-950 px-2.5 py-1 font-semibold text-white dark:bg-orange-500 dark:text-slate-950">
                        霜牙
                      </span>
                      {message.provider ? (
                        <span className="rounded-full bg-white/75 px-2.5 py-1 dark:bg-slate-800/70">
                          {message.provider === "deepseek"
                            ? "DeepSeek"
                            : message.provider}
                        </span>
                      ) : null}
                      {message.fallback ? (
                        <span className="rounded-full bg-amber-100 px-2.5 py-1 text-amber-700 dark:bg-amber-950/40 dark:text-amber-300">
                          检索整理
                        </span>
                      ) : null}
                    </div>
                  )}
                  {message.role === "assistant" ? (
                    message.content ? (
                      <AssistantMarkdown text={message.content} />
                    ) : loading && index === messages.length - 1 ? (
                      <p className="whitespace-pre-wrap break-words text-[15px] leading-7">
                        正在整理站内信息...
                      </p>
                    ) : null
                  ) : (
                    <p className="whitespace-pre-wrap break-words">
                      {message.content}
                    </p>
                  )}
                  {message.role === "assistant" && message.insights?.length ? (
                    <AttachmentSection
                      title="Copilot 建议"
                      summary="把标题、标签、规则摘要或准备清单收在这里，避免正文被信息块打断。"
                      badge={`${message.insights.length} 项`}
                      expanded={
                        expandedSections[`insights-${message.id || index}`] ??
                        userMessageCount === 0
                      }
                      onToggle={() =>
                        toggleSection(`insights-${message.id || index}`)
                      }
                    >
                      <InsightList insights={message.insights} />
                    </AttachmentSection>
                  ) : null}
                  {message.role === "assistant" && message.cards?.length ? (
                    <AttachmentSection
                      title="相关入口"
                      summary="需要时再展开看来源和可点击入口，阅读正文时不再堆满整屏。"
                      badge={`${message.cards.length} 个`}
                      expanded={
                        expandedSections[`cards-${message.id || index}`] ??
                        userMessageCount === 0
                      }
                      onToggle={() => toggleSection(`cards-${message.id || index}`)}
                    >
                      <CardList cards={message.cards} />
                    </AttachmentSection>
                  ) : null}
                  {message.role === "assistant" && message.id && message.content && (
                    <div className="mt-4 flex flex-wrap items-center gap-2 border-t border-orange-100 pt-3 text-[11px] text-slate-500 dark:border-slate-800 dark:text-slate-400">
                      <span>这条回复有帮助吗？</span>
                      <button
                        type="button"
                        onClick={() => void handleFeedback(message, index, "helpful")}
                        disabled={feedbackLoadingId === message.id}
                        className={`inline-flex items-center gap-1 rounded-full px-2 py-1 transition-colors ${
                          message.feedback_value === "helpful"
                            ? "bg-emerald-100 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300"
                            : "hover:bg-emerald-50 hover:text-emerald-700 dark:hover:bg-emerald-950/20"
                        }`}
                      >
                        <ThumbsUp className="h-3.5 w-3.5" />
                        有帮助
                      </button>
                      <button
                        type="button"
                        onClick={() => void handleFeedback(message, index, "unhelpful")}
                        disabled={feedbackLoadingId === message.id}
                        className={`inline-flex items-center gap-1 rounded-full px-2 py-1 transition-colors ${
                          message.feedback_value === "unhelpful"
                            ? "bg-rose-100 text-rose-700 dark:bg-rose-950/40 dark:text-rose-300"
                            : "hover:bg-rose-50 hover:text-rose-700 dark:hover:bg-rose-950/20"
                        }`}
                      >
                        <ThumbsDown className="h-3.5 w-3.5" />
                        没帮助
                      </button>
                    </div>
                  )}
                </div>
              </div>
            ))}

            {loading && (
              <div className="mr-8">
                <div className="inline-flex items-center gap-2 rounded-full border border-orange-200/80 bg-white/85 px-3 py-2 text-xs text-slate-600 shadow-sm dark:border-slate-700 dark:bg-slate-900/90 dark:text-slate-300">
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                  正在生成回复
                </div>
              </div>
            )}

            {!loading && userMessageCount === 0 && (
              <div className="space-y-3 rounded-[26px] border border-slate-200/80 bg-white/74 p-3.5 shadow-[0_16px_32px_rgba(15,23,42,0.06)] backdrop-blur-sm dark:border-slate-800 dark:bg-slate-950/40">
                <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-slate-500 dark:text-slate-400">
                  快速开始
                </p>
                <button
                  type="button"
                  onClick={() => void askAssistant(START_PROMPT)}
                  className="rounded-[22px] border border-orange-200/80 bg-[linear-gradient(180deg,rgba(255,247,237,0.95),rgba(255,255,255,0.92))] px-3.5 py-3 text-left text-sm leading-6 text-slate-700 transition-all hover:-translate-y-0.5 hover:border-orange-300 hover:bg-orange-50 dark:border-orange-900/40 dark:bg-[linear-gradient(180deg,rgba(124,45,18,0.22),rgba(15,23,42,0.52))] dark:text-slate-200 dark:hover:bg-orange-950/30"
                >
                  {START_PROMPT}
                </button>
              </div>
            )}
          </div>

          <div className="relative border-t border-orange-200/70 bg-white/74 px-4 py-4 backdrop-blur-xl dark:border-orange-900/50 dark:bg-slate-950/60">
            {error && (
              <p className="mb-3 rounded-2xl bg-rose-50 px-3 py-2 text-xs leading-5 text-rose-700 dark:bg-rose-950/30 dark:text-rose-300">
                {error}
              </p>
            )}

            <form onSubmit={handleSubmit} className="space-y-3">
              <textarea
                value={input}
                onChange={(e) => setInput(e.target.value)}
                placeholder="问我这个站有什么、去哪逛、推荐什么内容..."
                rows={3}
                className="min-h-[96px] w-full resize-none rounded-[24px] border border-orange-200 bg-white/94 px-4 py-3.5 text-sm leading-6 text-slate-900 outline-none transition-colors placeholder:text-slate-400 focus:border-orange-400 dark:border-slate-700 dark:bg-slate-900/92 dark:text-slate-100 dark:placeholder:text-slate-500"
              />

              <div className="flex items-center justify-between gap-3">
                <div />
                {loading ? (
                  <Button
                    type="button"
                    variant="outline"
                    onClick={handleStop}
                    className="rounded-full border-slate-300 bg-white/80 dark:border-slate-700 dark:bg-slate-900/70"
                  >
                    停止
                  </Button>
                ) : (
                  <Button
                    type="submit"
                    disabled={!input.trim()}
                    className="rounded-full bg-slate-900 text-white shadow-[0_12px_28px_rgba(15,23,42,0.24)] hover:bg-slate-800 dark:bg-orange-500 dark:text-slate-950 dark:hover:bg-orange-400"
                  >
                    <ArrowUp className="mr-1 h-4 w-4" />
                    发送
                  </Button>
                )}
              </div>
            </form>
          </div>
        </div>
      )}

      <button
        type="button"
        onClick={() => setOpen((prev) => !prev)}
        className="pointer-events-auto group flex items-center gap-3 rounded-full border border-orange-200 bg-[linear-gradient(135deg,rgba(255,255,255,0.98),rgba(255,247,237,0.92))] px-3 py-2 shadow-[0_18px_48px_rgba(15,23,42,0.22)] backdrop-blur transition-all hover:-translate-y-0.5 hover:shadow-[0_24px_60px_rgba(15,23,42,0.26)] dark:border-orange-900/60 dark:bg-[linear-gradient(135deg,rgba(15,23,42,0.96),rgba(17,24,39,0.92))]"
        aria-label="打开 AI 助手"
      >
        <MascotAvatar compact />
        <div className="pr-1 text-left">
          <div className="flex items-center gap-1.5">
            <p className="text-sm font-semibold text-slate-900 dark:text-slate-100">
              霜牙
            </p>
            <Sparkles className="h-3.5 w-3.5 text-orange-500" />
          </div>
          <p className="text-xs text-slate-500 dark:text-slate-400">
            点我，帮你逛站内内容
          </p>
        </div>
      </button>
    </div>
  );
}
