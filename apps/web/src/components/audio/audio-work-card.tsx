import Link from 'next/link';
import { AudioLines, Heart, MessageCircle, PlayCircle, UserRound, Waves } from 'lucide-react';
import { type AudioWork } from '@/lib/api-client';
import { cn } from '@/lib/utils';

const GRADIENTS = [
  'from-emerald-500 to-cyan-500',
  'from-sky-500 to-indigo-500',
  'from-fuchsia-500 to-rose-500',
  'from-amber-500 to-orange-500',
  'from-violet-500 to-cyan-500',
];

function hashGradient(str: string): string {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = (hash * 31 + str.charCodeAt(i)) | 0;
  }
  return GRADIENTS[Math.abs(hash) % GRADIENTS.length];
}

function formatDuration(seconds: number) {
  if (!seconds || Number.isNaN(seconds)) return '0:00';
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins}:${secs.toString().padStart(2, '0')}`;
}

export function AudioWorkCard({
  work,
  compact = false,
}: {
  work: AudioWork;
  compact?: boolean;
}) {
  const gradient = hashGradient(work.id);
  const bars = work.waveform_preview?.length ? work.waveform_preview : [0.16, 0.3, 0.42, 0.58, 0.74, 0.48, 0.34, 0.2];

  return (
    <Link href={`/audio/works/${work.id}`} className="group block">
      <div className="overflow-hidden rounded-2xl border border-border/60 bg-card shadow-sm transition-all duration-200 hover:-translate-y-0.5 hover:border-primary/30 hover:shadow-lg hover:shadow-primary/5">
        <div className={cn('relative overflow-hidden', compact ? 'aspect-[16/10]' : 'aspect-[16/11]')}>
          {work.cover_image_url ? (
            <div
              className="h-full w-full transition-transform duration-300 group-hover:scale-105"
              style={{
                backgroundImage: `url(${work.cover_image_url})`,
                backgroundSize: 'cover',
                backgroundPosition: 'center',
              }}
              aria-label={work.title}
            />
          ) : (
            <div className={cn('h-full w-full bg-gradient-to-br', gradient)}>
              <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.18),transparent_42%)]" />
              <div className="absolute inset-x-0 bottom-0 px-4 pb-4">
                <div className="rounded-2xl border border-white/10 bg-black/20 p-3 backdrop-blur-sm">
                  <div className="mb-2 flex items-center gap-2 text-[11px] uppercase tracking-[0.18em] text-white/70">
                    <AudioLines className="h-3.5 w-3.5" />
                    Audio Work
                  </div>
                  <div className="flex h-16 items-end gap-1">
                    {bars.slice(0, compact ? 18 : 28).map((value, index) => (
                      <div
                        key={`${work.id}-wave-${index}`}
                        className="min-w-0 flex-1 rounded-full bg-white/80"
                        style={{ height: `${Math.max(10, Math.min(100, value * 100))}%` }}
                      />
                    ))}
                  </div>
                </div>
              </div>
            </div>
          )}

          <div className="absolute inset-0 bg-black/0 transition-colors duration-200 group-hover:bg-black/20" />
          <div className="absolute left-3 top-3 inline-flex items-center gap-1 rounded-full bg-black/45 px-2.5 py-1 text-[11px] font-medium text-white backdrop-blur-sm">
            <PlayCircle className="h-3.5 w-3.5" />
            {formatDuration(work.duration_sec)}
          </div>
        </div>

        <div className="space-y-3 px-4 py-3.5">
          <div>
            <p className="line-clamp-1 text-sm font-semibold">{work.title}</p>
            {work.description ? (
              <p className="mt-1 line-clamp-2 text-xs leading-relaxed text-muted-foreground">{work.description}</p>
            ) : null}
          </div>

          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <span className="inline-flex items-center gap-1 truncate">
              <UserRound className="h-3.5 w-3.5" />
              {work.author_username ? `@${work.author_username}` : work.author_id}
            </span>
            <span className="text-border">·</span>
            <span className="inline-flex items-center gap-1">
              <Waves className="h-3.5 w-3.5" />
              {work.waveform_preview?.length ? '已生成波形' : '音频作品'}
            </span>
          </div>

          <div className="flex items-center gap-3 text-xs text-muted-foreground">
            <span className="inline-flex items-center gap-1">
              <Heart className="h-3.5 w-3.5" />
              {work.like_count}
            </span>
            <span className="inline-flex items-center gap-1">
              <MessageCircle className="h-3.5 w-3.5" />
              {work.comment_count}
            </span>
          </div>

          {work.tags && work.tags.length > 0 ? (
            <div className="flex flex-wrap gap-1.5">
              {work.tags.slice(0, 4).map((tag) => (
                <span key={`${work.id}-${tag}`} className="rounded-full border border-border/70 bg-muted/30 px-2 py-0.5 text-[11px] text-muted-foreground">
                  {tag}
                </span>
              ))}
            </div>
          ) : null}
        </div>
      </div>
    </Link>
  );
}
