"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Sparkles, Tag, X, ImagePlus, Loader2, Users } from "lucide-react";
import { apiClient, MediaAnalysisItem } from "@/lib/api-client";
import { useAuth } from "@/contexts/auth-context";
import { useAssistantPageContext } from "@/contexts/assistant-page-context";
import { useOSSUpload } from "@/hooks/use-oss-upload";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

const DRAFT_KEY = "post_draft";

interface ImageItem {
  url: string;
  preview: string;
  progress: number;
  uploading: boolean;
}

function truncateForAssistant(value: string, limit: number) {
  const trimmed = value.trim();
  if (!trimmed || trimmed.length <= limit) return trimmed;
  return `${trimmed.slice(0, limit).trim()}...`;
}

export default function CreatePostPage() {
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const { setPageContext, clearPageContext } = useAssistantPageContext();
  const [groupId, setGroupId] = useState<string | null>(null);
  const [groupName, setGroupName] = useState("");
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [tags, setTags] = useState("");
  const [visibility, setVisibility] = useState("public");
  const [isAIGenerated, setIsAIGenerated] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [images, setImages] = useState<ImageItem[]>([]);
  const [mediaAnalysis, setMediaAnalysis] = useState<MediaAnalysisItem[]>([]);
  const [analyzingImages, setAnalyzingImages] = useState(false);
  const { upload } = useOSSUpload();
  const lastAnalysisKeyRef = useRef("");

  // Draft restore on mount
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const gid = params.get("group_id");
    if (gid) {
      setGroupId(gid);
      apiClient
        .getGroup(gid)
        .then((group) => setGroupName(group.name))
        .catch(() => {
          setGroupId(null);
          setGroupName("");
        });
    }

    const raw = localStorage.getItem(DRAFT_KEY);
    if (!raw) return;
    try {
      const draft = JSON.parse(raw);
      if (draft.title || draft.content) {
        if (confirm("检测到未保存的草稿，是否恢复？")) {
          if (draft.title) setTitle(draft.title);
          if (draft.content) setContent(draft.content);
          if (draft.tags) setTags(draft.tags);
        }
      }
    } catch {
      // ignore
    }
    localStorage.removeItem(DRAFT_KEY);
  }, []);

  // Draft auto-save with 2s debounce
  useEffect(() => {
    if (!content && !title) return;
    const timer = setTimeout(() => {
      localStorage.setItem(DRAFT_KEY, JSON.stringify({ title, content, tags }));
    }, 2000);
    return () => clearTimeout(timer);
  }, [title, content, tags]);

  useEffect(() => {
    const fields: Record<string, string> = {};
    if (title.trim()) {
      fields.draft_title = truncateForAssistant(title, 60);
    }
    if (content.trim()) {
      fields.draft_content = truncateForAssistant(content, 240);
    }
    if (tags.trim()) {
      fields.draft_tags = truncateForAssistant(tags, 80);
    }
    if (groupName.trim()) {
      fields.group_name = groupName.trim();
    }
    if (visibility) {
      fields.visibility =
        visibility === "followers_only"
          ? "仅关注者可见"
          : visibility === "private"
            ? "私密"
            : "公开";
    }
    if (isAIGenerated) {
      fields.ai_generated = "已勾选";
    }
    if (mediaAnalysis.length > 0) {
      const mergedTags = Array.from(
        new Set(mediaAnalysis.flatMap((item) => item.tags ?? [])),
      );
      if (mergedTags.length > 0) {
        fields.image_tags = mergedTags.join("、");
      }
      const firstAlt = mediaAnalysis
        .map((item) => item.alt_text)
        .find((item) => item?.trim());
      if (firstAlt) {
        fields.image_alt_notes = truncateForAssistant(firstAlt, 120);
      }
    }

    setPageContext({
      path: "/posts/create",
      kind: "post_create",
      title: groupName ? `发布到圈子：${groupName}` : "发布动态",
      summary: groupName
        ? "用户当前正在一个圈子里撰写动态，适合提供更贴合圈子氛围的发帖建议。"
        : "用户当前正在撰写一条新动态，适合提供标题、正文、标签和可见性建议。",
      prompt_hints: groupName
        ? [
            "帮我把这条动态改得更适合发在这个圈子",
            "根据我现在的草稿，推荐 3 个更贴切的标签",
            "帮我润色一下这段内容，但保留原本语气",
          ]
        : [
            "帮我润色一下这条动态",
            "根据我现在的草稿，推荐 3 个标签",
            "这条动态更适合公开还是仅关注者可见？",
          ],
      fields,
    });
  }, [
    clearPageContext,
    content,
    groupName,
    isAIGenerated,
    setPageContext,
    tags,
    title,
    visibility,
    mediaAnalysis,
  ]);

  useEffect(() => () => clearPageContext(), [clearPageContext]);

  useEffect(() => {
    const urls = images.filter((item) => item.url && !item.uploading).map((item) => item.url);
    if (urls.length === 0) {
      setMediaAnalysis([]);
      lastAnalysisKeyRef.current = "";
      return;
    }

    const key = urls.join("|");
    if (key === lastAnalysisKeyRef.current) return;

    let cancelled = false;
    setAnalyzingImages(true);
    apiClient
      .analyzeAssistantMedia(urls, "post_create")
      .then((result) => {
        if (cancelled) return;
        setMediaAnalysis(result.items ?? []);
        lastAnalysisKeyRef.current = key;
      })
      .catch(() => {
        if (!cancelled) {
          setMediaAnalysis([]);
        }
      })
      .finally(() => {
        if (!cancelled) {
          setAnalyzingImages(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [images]);

  const handleFileChange = useCallback(
    async (e: React.ChangeEvent<HTMLInputElement>) => {
      const files = Array.from(e.target.files || []);
      if (!files.length) return;
      const remaining = 9 - images.length;
      const toUpload = files.slice(0, remaining);
      if (fileInputRef.current) fileInputRef.current.value = "";

      const placeholders: ImageItem[] = toUpload.map((file) => ({
        url: "",
        preview: URL.createObjectURL(file),
        progress: 0,
        uploading: true,
      }));

      setImages((prev) => [...prev, ...placeholders]);
      const startIdx = images.length;

      await Promise.all(
        toUpload.map(async (file, i) => {
          const idx = startIdx + i;
          try {
            const ossHook = { upload };
            const url = await ossHook.upload(file, "post");
            setImages((prev) =>
              prev.map((item, j) =>
                j === idx
                  ? { ...item, url, progress: 100, uploading: false }
                  : item,
              ),
            );
          } catch (err: any) {
            setError(err.message || "图片上传失败");
            setImages((prev) => prev.filter((_, j) => j !== idx));
          }
        }),
      );
    },
    [images.length, upload],
  );

  function removeImage(index: number) {
    setImages((prev) => prev.filter((_, i) => i !== index));
  }

  function applySuggestedTags() {
    const suggested = mediaAnalysis.flatMap((item) => item.tags ?? []);
    if (suggested.length === 0) return;
    const merged = Array.from(
      new Set(
        [
          ...tags
            .split(",")
            .map((item) => item.trim())
            .filter(Boolean),
          ...suggested.map((item) => item.replace(/^#/, "").trim()).filter(Boolean),
        ],
      ),
    );
    setTags(merged.join(", "));
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!content.trim()) {
      setError("内容不能为空");
      return;
    }
    if (images.some((img) => img.uploading)) {
      setError("请等待图片上传完成");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const post = await apiClient.createPost({
        title: title || undefined,
        content,
        media_urls: images.filter((img) => img.url).map((img) => img.url),
        tags: tags
          ? tags
              .split(",")
              .map((t) => t.trim())
              .filter(Boolean)
          : [],
        visibility: visibility as "public" | "followers_only" | "private",
        group_id: groupId || undefined,
        is_ai_generated: isAIGenerated || undefined,
      });
      localStorage.removeItem(DRAFT_KEY);
      router.push(`/posts/${post.id}?submitted=1`);
    } catch (err: any) {
      setError(err.message || "发布失败");
    } finally {
      setLoading(false);
    }
  };

  const anyUploading = images.some((img) => img.uploading);
  const emailVerified = !!user?.email_verified_at;

  if (!authLoading && user && !emailVerified) {
    return (
      <div className="max-w-2xl mx-auto pt-20 px-4 pb-8">
        <div className="rounded-2xl border bg-card p-8 text-center">
          <h1 className="text-2xl font-bold mb-2">先验证邮箱</h1>
          <p className="text-muted-foreground mb-6">
            为了保护社区内容安全，发布动态前需要先完成邮箱验证。
          </p>
          <div className="flex justify-center gap-3">
            <Button onClick={() => router.push("/settings")}>去设置验证</Button>
            <Button variant="outline" onClick={() => router.push("/feed")}>
              返回动态
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-8">
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <h1 className="text-2xl font-bold">发布动态</h1>
        <Link
          href={{
            pathname: "/agent",
            query: {
              scenario: "post_agent",
              source_path: "/posts/create",
              draft_title: truncateForAssistant(title, 60),
              draft_content: truncateForAssistant(content, 240),
              draft_tags: truncateForAssistant(tags, 80),
              group_name: groupName,
              visibility:
                visibility === "followers_only"
                  ? "仅关注者可见"
                  : visibility === "private"
                    ? "私密"
                    : "公开",
            },
          }}
          className="inline-flex items-center justify-center rounded-xl border border-orange-200 bg-orange-50 px-4 py-2 text-sm font-medium text-orange-700 transition-colors hover:border-orange-300 hover:bg-orange-100"
        >
          <Sparkles className="mr-2 h-4 w-4" />
          交给 Agent 处理
        </Link>
      </div>
      {groupId && (
        <div className="mb-4 rounded-2xl border bg-card p-4 flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center">
            <Users className="h-5 w-5 text-primary" />
          </div>
          <div>
            <p className="text-sm font-medium">发布到圈子</p>
            <p className="text-sm text-muted-foreground">
              {groupName || "正在读取圈子信息..."}
            </p>
          </div>
        </div>
      )}
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <Label htmlFor="title">标题（可选）</Label>
          <Input
            id="title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="长帖标题..."
            className="mt-1"
          />
        </div>
        <div>
          <Label htmlFor="content">内容 *</Label>
          <Textarea
            id="content"
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="分享你的furry日常、创作、想法..."
            rows={8}
            className="mt-1"
            required
          />
        </div>

        {/* Image upload */}
        <div>
          <Label>图片（最多 9 张）</Label>
          <div className="mt-2 grid grid-cols-3 gap-2">
            {images.map((img, i) => (
              <div
                key={i}
                className="relative aspect-square rounded-lg overflow-hidden bg-muted"
              >
                <img
                  src={img.preview}
                  alt=""
                  className="w-full h-full object-cover"
                />
                {img.uploading && (
                  <div className="absolute inset-0 bg-black/40 flex flex-col items-center justify-center gap-1">
                    <Loader2 className="h-5 w-5 text-white animate-spin" />
                    <div className="w-3/4 h-1.5 bg-white/30 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-white rounded-full transition-all duration-200"
                        style={{ width: `${img.progress}%` }}
                      />
                    </div>
                  </div>
                )}
                {!img.uploading && (
                  <button
                    type="button"
                    onClick={() => removeImage(i)}
                    className="absolute top-1 right-1 bg-black/60 rounded-full p-0.5 hover:bg-black/80 transition-colors"
                  >
                    <X className="h-3 w-3 text-white" />
                  </button>
                )}
              </div>
            ))}
            {images.length < 9 && (
              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                disabled={anyUploading}
                className="aspect-square rounded-lg border-2 border-dashed border-muted-foreground/30 flex flex-col items-center justify-center gap-1 hover:border-primary/50 hover:bg-muted/50 transition-colors text-muted-foreground disabled:opacity-50"
              >
                <ImagePlus className="h-5 w-5" />
                <span className="text-xs">添加图片</span>
              </button>
            )}
          </div>
          <input
            ref={fileInputRef}
            type="file"
            accept="image/jpeg,image/png,image/gif,image/webp"
            multiple
            className="hidden"
            onChange={handleFileChange}
          />
          {(analyzingImages || mediaAnalysis.length > 0) && (
            <div className="mt-4 rounded-2xl border border-cyan-200/70 bg-cyan-50/60 p-4">
              <div className="flex items-center justify-between gap-3">
                <div>
                  <p className="text-sm font-semibold text-slate-900">图片理解建议</p>
                  <p className="text-xs text-slate-500">
                    自动生成 alt text、标签和摘要，方便你补充图片信息。
                  </p>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={applySuggestedTags}
                  disabled={mediaAnalysis.length === 0}
                >
                  <Tag className="mr-1 h-4 w-4" />
                  应用推荐标签
                </Button>
              </div>

              {analyzingImages ? (
                <div className="mt-3 flex items-center gap-2 text-sm text-slate-500">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  正在分析图片内容...
                </div>
              ) : (
                <div className="mt-3 grid gap-3">
                  {mediaAnalysis.map((item, index) => (
                    <div
                      key={`${item.media_url}-${index}`}
                      className="rounded-2xl border border-white/80 bg-white px-4 py-3"
                    >
                      <div className="flex items-center gap-2 text-xs text-slate-500">
                        <Sparkles className="h-3.5 w-3.5 text-cyan-600" />
                        图片 {index + 1}
                        {item.fallback && (
                          <span className="rounded-full bg-amber-100 px-2 py-0.5 text-amber-700">
                            fallback
                          </span>
                        )}
                      </div>
                      <p className="mt-2 text-sm font-medium text-slate-900">
                        Alt Text：{item.alt_text}
                      </p>
                      {item.image_summary && (
                        <p className="mt-2 text-xs leading-5 text-slate-600">
                          {item.image_summary}
                        </p>
                      )}
                      {item.tags?.length ? (
                        <div className="mt-2 flex flex-wrap gap-2">
                          {item.tags.map((tag) => (
                            <span
                              key={`${item.media_url}-${tag}`}
                              className="rounded-full bg-cyan-100 px-2.5 py-1 text-xs font-medium text-cyan-800"
                            >
                              {tag}
                            </span>
                          ))}
                        </div>
                      ) : null}
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>

        {/* AI Generated label */}
        <label className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400 cursor-pointer">
          <input
            type="checkbox"
            checked={isAIGenerated}
            onChange={(e) => setIsAIGenerated(e.target.checked)}
            className="rounded"
          />
          <span>此内容包含 AI 生成内容（请如实标注）</span>
        </label>

        <div>
          <Label htmlFor="tags">标签（逗号分隔）</Label>
          <Input
            id="tags"
            value={tags}
            onChange={(e) => setTags(e.target.value)}
            placeholder="furry, 兽设, 创作..."
            className="mt-1"
          />
        </div>
        <div>
          <Label htmlFor="visibility">可见性</Label>
          <select
            id="visibility"
            value={visibility}
            onChange={(e) => setVisibility(e.target.value)}
            className="mt-1 w-full rounded-md border bg-background px-3 py-2 text-sm"
          >
            <option value="public">公开</option>
            <option value="followers_only">仅关注者可见</option>
            <option value="private">私密</option>
          </select>
        </div>
        {error && <p className="text-destructive text-sm">{error}</p>}
        <div className="flex gap-3 pt-2">
          <Button type="submit" disabled={loading || anyUploading}>
            {loading ? "发布中..." : "发布"}
          </Button>
          <Button type="button" variant="outline" onClick={() => router.back()}>
            取消
          </Button>
        </div>
      </form>
    </div>
  );
}
