"use client";

import Link from "next/link";
import { FormEvent, ReactNode, useEffect, useRef, useState } from "react";
import { ArrowUp, Loader2, Sparkles, X } from "lucide-react";
import {
  apiClient,
  AssistantCard,
  AssistantChatMessage,
  AssistantConversation,
} from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/contexts/auth-context";

const STORAGE_KEY = "furry_assistant_messages_v1";
const CONVERSATION_KEY = "furry_assistant_current_conversation_v1";

const QUICK_PROMPTS = [
  "我第一次来，先逛哪里？",
  "推荐几个有意思的圈子",
  "最近有什么活动值得看？",
  "怎么发布我的第一条动态？",
  "推荐几个值得关注的用户",
];

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
    <div className="mt-3 grid gap-2">
      {cards.map((card) => (
        <Link
          key={`${card.kind}-${card.href}-${card.title}`}
          href={card.href}
          className="rounded-2xl border border-amber-200/70 bg-white/80 p-3 transition-colors hover:border-orange-300 hover:bg-orange-50 dark:border-orange-900/70 dark:bg-slate-900/70 dark:hover:bg-slate-900"
        >
          <div className="mb-1 flex items-center gap-2">
            {card.ref && (
              <span className="rounded-full bg-slate-900 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-white dark:bg-orange-500 dark:text-slate-950">
                {card.ref}
              </span>
            )}
            <span className="rounded-full bg-orange-100 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.18em] text-orange-700 dark:bg-orange-950/70 dark:text-orange-300">
              {card.kind}
            </span>
            {card.meta && (
              <span className="text-[11px] text-muted-foreground">
                {card.meta}
              </span>
            )}
          </div>
          <p className="text-sm font-semibold text-slate-900 dark:text-slate-100">
            {card.title}
          </p>
          <p className="mt-1 text-xs leading-5 text-slate-600 dark:text-slate-400">
            {card.summary}
          </p>
          {card.reason && (
            <p className="mt-2 text-[11px] leading-5 text-slate-500 dark:text-slate-400">
              推荐理由：{card.reason}
            </p>
          )}
          {card.source && (
            <p className="mt-1 text-[11px] leading-5 text-slate-500 dark:text-slate-400">
              来源：{card.source}
            </p>
          )}
        </Link>
      ))}
    </div>
  );
}

function ReferenceList({ cards }: { cards?: AssistantCard[] }) {
  if (!cards || cards.length === 0) return null;

  return (
    <div className="mt-3 rounded-2xl border border-orange-200/70 bg-orange-50/70 p-3 dark:border-orange-900/50 dark:bg-orange-950/20">
      <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-orange-700 dark:text-orange-300">
        参考来源
      </p>
      <div className="mt-2 space-y-2">
        {cards.map((card, index) => (
          <Link
            key={`ref-${card.kind}-${card.href}-${index}`}
            id={card.ref ? `ref-${card.ref}` : undefined}
            href={card.href}
            className="block rounded-xl px-2 py-2 transition-colors hover:bg-white/80 dark:hover:bg-slate-900/50"
          >
            <div className="flex items-center gap-2 text-xs text-slate-600 dark:text-slate-300">
              <span className="inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-orange-200/80 px-1.5 font-semibold text-orange-900 dark:bg-orange-900/60 dark:text-orange-100">
                {card.ref || index + 1}
              </span>
              <span className="font-medium text-slate-900 dark:text-slate-100">
                {card.title}
              </span>
              <span className="text-muted-foreground">{card.meta}</span>
            </div>
            {(card.reason || card.source) && (
              <p className="mt-1 pl-7 text-[11px] leading-5 text-slate-500 dark:text-slate-400">
                {card.reason ? `推荐理由：${card.reason}` : ""}
                {card.reason && card.source ? " · " : ""}
                {card.source ? `来源：${card.source}` : ""}
              </p>
            )}
          </Link>
        ))}
      </div>
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
    <div className="space-y-3">
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
              className="space-y-1.5 pl-4 text-sm leading-6 text-inherit"
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
          <p key={`p-${blockIndex}`} className="text-sm leading-6 text-inherit">
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
  const scrollerRef = useRef<HTMLDivElement>(null);
  const abortRef = useRef<AbortController | null>(null);

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
        } finally {
          setHistoryLoading(false);
        }
      })();
      return;
    }

    setConversations([]);
    setConversationId(null);
    setHistoryLoading(false);
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
    if (!scrollerRef.current) return;
    scrollerRef.current.scrollTop = scrollerRef.current.scrollHeight;
  }, [messages, loading, open]);

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
                cards: meta.cards,
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
    if (isLoggedIn) {
      setConversationId(null);
      localStorage.removeItem(CONVERSATION_KEY);
      return;
    }
    localStorage.removeItem(STORAGE_KEY);
  }

  const userMessageCount = messages.filter((msg) => msg.role === "user").length;

  return (
    <div className="pointer-events-none fixed bottom-5 right-4 z-[60] sm:bottom-6 sm:right-6">
      {open && (
        <div className="pointer-events-auto mb-4 w-[calc(100vw-2rem)] max-w-[390px] overflow-hidden rounded-[28px] border border-orange-200/70 bg-[radial-gradient(circle_at_top,#fff7ed,white_48%,#fff_100%)] shadow-[0_28px_80px_rgba(15,23,42,0.22)] dark:border-orange-900/50 dark:bg-[radial-gradient(circle_at_top,#1f1724,#0f172a_52%,#0b1120_100%)]">
          <div className="border-b border-orange-200/70 bg-white/70 px-4 py-4 backdrop-blur dark:border-orange-900/50 dark:bg-slate-950/50">
            <div className="flex items-start gap-3">
              <MascotAvatar />
              <div className="min-w-0 flex-1">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="text-sm font-semibold text-slate-900 dark:text-slate-100">
                      霜牙
                    </p>
                    <p className="mt-0.5 text-xs leading-5 text-slate-600 dark:text-slate-400">
                      Furry 站内导览助手，能帮你找页面、帖子、圈子和活动。
                    </p>
                  </div>
                  <button
                    type="button"
                    onClick={() => setOpen(false)}
                    className="rounded-full p-1 text-slate-500 transition-colors hover:bg-orange-100 hover:text-slate-900 dark:hover:bg-slate-800 dark:hover:text-slate-100"
                    aria-label="关闭助手"
                  >
                    <X className="h-4 w-4" />
                  </button>
                </div>
                <div className="mt-3 flex items-center gap-2">
                  <span className="inline-flex items-center gap-1 rounded-full bg-emerald-100 px-2.5 py-1 text-[11px] font-medium text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-300">
                    <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
                    {fallbackMode
                      ? "站内检索模式"
                      : `${providerLabel} + 站内检索`}
                  </span>
                  <button
                    type="button"
                    onClick={clearConversation}
                    className="text-[11px] text-muted-foreground transition-colors hover:text-foreground"
                  >
                    {isLoggedIn ? "新建对话" : "清空对话"}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div
            ref={scrollerRef}
            className="max-h-[min(65vh,560px)] space-y-4 overflow-y-auto px-4 py-4"
          >
            {isLoggedIn && conversations.length > 0 && (
              <div className="space-y-2">
                <p className="text-[11px] font-medium uppercase tracking-[0.18em] text-muted-foreground">
                  最近对话
                </p>
                <div className="flex gap-2 overflow-x-auto pb-1">
                  {conversations.slice(0, 8).map((conversation) => (
                    <button
                      key={conversation.id}
                      type="button"
                      onClick={() => void openConversation(conversation.id)}
                      className={`min-w-[140px] max-w-[180px] rounded-2xl border px-3 py-2 text-left transition-colors ${
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
                  message.role === "user" ? "ml-12 flex justify-end" : "mr-8"
                }
              >
                <div
                  className={
                    message.role === "user"
                      ? "max-w-[85%] rounded-[20px] rounded-br-md bg-slate-900 px-4 py-3 text-sm leading-6 text-white shadow-sm dark:bg-orange-500"
                      : "max-w-full rounded-[22px] rounded-bl-md border border-orange-200/80 bg-white/90 px-4 py-3 text-sm leading-6 text-slate-800 shadow-sm dark:border-slate-700 dark:bg-slate-900/90 dark:text-slate-100"
                  }
                >
                  {message.role === "assistant" ? (
                    message.content ? (
                      <AssistantMarkdown text={message.content} />
                    ) : loading && index === messages.length - 1 ? (
                      <p className="whitespace-pre-wrap break-words">
                        正在整理站内信息...
                      </p>
                    ) : null
                  ) : (
                    <p className="whitespace-pre-wrap break-words">
                      {message.content}
                    </p>
                  )}
                  {message.role === "assistant" && (
                    <ReferenceList cards={message.cards} />
                  )}
                  {message.role === "assistant" && (
                    <CardList cards={message.cards} />
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
              <div className="flex flex-wrap gap-2">
                {QUICK_PROMPTS.map((prompt) => (
                  <button
                    key={prompt}
                    type="button"
                    onClick={() => void askAssistant(prompt)}
                    className="rounded-full border border-orange-200 bg-orange-50 px-3 py-1.5 text-xs text-orange-700 transition-colors hover:border-orange-300 hover:bg-orange-100 dark:border-orange-900/60 dark:bg-orange-950/30 dark:text-orange-200"
                  >
                    {prompt}
                  </button>
                ))}
              </div>
            )}
          </div>

          <div className="border-t border-orange-200/70 bg-white/70 px-4 py-4 backdrop-blur dark:border-orange-900/50 dark:bg-slate-950/50">
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
                className="min-h-[92px] w-full resize-none rounded-[22px] border border-orange-200 bg-white px-4 py-3 text-sm leading-6 text-slate-900 outline-none transition-colors placeholder:text-slate-400 focus:border-orange-400 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:placeholder:text-slate-500"
              />

              <div className="flex items-center justify-between gap-3">
                <p className="text-[11px] leading-5 text-muted-foreground">
                  回答会结合站内公开内容做推荐。
                </p>
                {loading ? (
                  <Button
                    type="button"
                    variant="outline"
                    onClick={handleStop}
                    className="rounded-full"
                  >
                    停止
                  </Button>
                ) : (
                  <Button
                    type="submit"
                    disabled={!input.trim()}
                    className="rounded-full bg-slate-900 text-white hover:bg-slate-800 dark:bg-orange-500 dark:text-slate-950 dark:hover:bg-orange-400"
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
        className="pointer-events-auto group flex items-center gap-3 rounded-full border border-orange-200 bg-white/95 px-3 py-2 shadow-[0_16px_45px_rgba(15,23,42,0.2)] backdrop-blur transition-all hover:-translate-y-0.5 hover:shadow-[0_22px_55px_rgba(15,23,42,0.24)] dark:border-orange-900/60 dark:bg-slate-950/92"
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
