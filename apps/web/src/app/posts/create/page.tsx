'use client';

import { useState, useRef, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { X, ImagePlus, Loader2 } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { useOSSUpload } from '@/hooks/use-oss-upload';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

const DRAFT_KEY = 'post_draft';

interface ImageItem {
  url: string;
  preview: string;
  progress: number;
  uploading: boolean;
}

export default function CreatePostPage() {
  const router = useRouter();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [tags, setTags] = useState('');
  const [visibility, setVisibility] = useState('public');
  const [isAIGenerated, setIsAIGenerated] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [images, setImages] = useState<ImageItem[]>([]);
  const { upload } = useOSSUpload();

  // Draft restore on mount
  useEffect(() => {
    const raw = localStorage.getItem(DRAFT_KEY);
    if (!raw) return;
    try {
      const draft = JSON.parse(raw);
      if (draft.title || draft.content) {
        if (confirm('检测到未保存的草稿，是否恢复？')) {
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

  const handleFileChange = useCallback(async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || []);
    if (!files.length) return;
    const remaining = 9 - images.length;
    const toUpload = files.slice(0, remaining);
    if (fileInputRef.current) fileInputRef.current.value = '';

    const placeholders: ImageItem[] = toUpload.map(file => ({
      url: '',
      preview: URL.createObjectURL(file),
      progress: 0,
      uploading: true,
    }));

    setImages(prev => [...prev, ...placeholders]);
    const startIdx = images.length;

    await Promise.all(
      toUpload.map(async (file, i) => {
        const idx = startIdx + i;
        try {
          const ossHook = { upload };
          const url = await ossHook.upload(file, 'post');
          setImages(prev =>
            prev.map((item, j) =>
              j === idx ? { ...item, url, progress: 100, uploading: false } : item
            )
          );
        } catch (err: any) {
          setError(err.message || '图片上传失败');
          setImages(prev => prev.filter((_, j) => j !== idx));
        }
      })
    );
  }, [images.length, upload]);

  function removeImage(index: number) {
    setImages(prev => prev.filter((_, i) => i !== index));
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!content.trim()) {
      setError('内容不能为空');
      return;
    }
    if (images.some(img => img.uploading)) {
      setError('请等待图片上传完成');
      return;
    }
    setLoading(true);
    setError('');
    try {
      const post = await apiClient.createPost({
        title: title || undefined,
        content,
        media_urls: images.filter(img => img.url).map(img => img.url),
        tags: tags ? tags.split(',').map((t) => t.trim()).filter(Boolean) : [],
        visibility: visibility as 'public' | 'followers_only' | 'private',
        is_ai_generated: isAIGenerated || undefined,
      });
      localStorage.removeItem(DRAFT_KEY);
      router.push(`/posts/${post.id}?submitted=1`);
    } catch (err: any) {
      setError(err.message || '发布失败');
    } finally {
      setLoading(false);
    }
  };

  const anyUploading = images.some(img => img.uploading);

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-8">
      <h1 className="text-2xl font-bold mb-6">发布动态</h1>
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
              <div key={i} className="relative aspect-square rounded-lg overflow-hidden bg-muted">
                <img src={img.preview} alt="" className="w-full h-full object-cover" />
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
        </div>

        {/* AI Generated label */}
        <label className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400 cursor-pointer">
          <input
            type="checkbox"
            checked={isAIGenerated}
            onChange={e => setIsAIGenerated(e.target.checked)}
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
            {loading ? '发布中...' : '发布'}
          </Button>
          <Button type="button" variant="outline" onClick={() => router.back()}>
            取消
          </Button>
        </div>
      </form>
    </div>
  );
}
