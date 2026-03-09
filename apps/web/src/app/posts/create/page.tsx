'use client';

import { useState, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { X, ImagePlus, Loader2 } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

export default function CreatePostPage() {
  const router = useRouter();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [tags, setTags] = useState('');
  const [visibility, setVisibility] = useState('public');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [images, setImages] = useState<{ url: string; preview: string }[]>([]);
  const [uploading, setUploading] = useState(false);

  async function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const files = Array.from(e.target.files || []);
    if (!files.length) return;
    const remaining = 9 - images.length;
    const toUpload = files.slice(0, remaining);
    setUploading(true);
    try {
      const results = await Promise.all(
        toUpload.map(async (file) => {
          const preview = URL.createObjectURL(file);
          const { url } = await apiClient.uploadFile('/upload/image', file);
          return { url, preview };
        })
      );
      setImages(prev => [...prev, ...results]);
    } catch (e: any) {
      setError(e.message || '图片上传失败');
    } finally {
      setUploading(false);
      if (fileInputRef.current) fileInputRef.current.value = '';
    }
  }

  function removeImage(index: number) {
    setImages(prev => prev.filter((_, i) => i !== index));
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!content.trim()) {
      setError('内容不能为空');
      return;
    }
    setLoading(true);
    setError('');
    try {
      const post = await apiClient.createPost({
        title: title || undefined,
        content,
        media_urls: images.map(img => img.url),
        tags: tags ? tags.split(',').map((t) => t.trim()).filter(Boolean) : [],
        visibility: visibility as 'public' | 'followers_only' | 'private',
      });
      router.push(`/posts/${post.id}`);
    } catch (err: any) {
      setError(err.message || '发布失败');
    } finally {
      setLoading(false);
    }
  };

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
                <button
                  type="button"
                  onClick={() => removeImage(i)}
                  className="absolute top-1 right-1 bg-black/60 rounded-full p-0.5 hover:bg-black/80 transition-colors"
                >
                  <X className="h-3 w-3 text-white" />
                </button>
              </div>
            ))}
            {images.length < 9 && (
              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                disabled={uploading}
                className="aspect-square rounded-lg border-2 border-dashed border-muted-foreground/30 flex flex-col items-center justify-center gap-1 hover:border-primary/50 hover:bg-muted/50 transition-colors text-muted-foreground"
              >
                {uploading ? (
                  <Loader2 className="h-5 w-5 animate-spin" />
                ) : (
                  <>
                    <ImagePlus className="h-5 w-5" />
                    <span className="text-xs">添加图片</span>
                  </>
                )}
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
          <Button type="submit" disabled={loading || uploading}>
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
