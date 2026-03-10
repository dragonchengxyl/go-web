import { cn } from '@/lib/utils';

export function Skeleton({ className }: { className?: string }) {
  return (
    <div
      className={cn(
        'rounded-md bg-gradient-to-r from-muted via-muted/40 to-muted bg-[length:200%_100%] animate-shimmer',
        className
      )}
    />
  );
}

export function PostCardSkeleton() {
  return (
    <div className="bg-card border rounded-xl p-4 space-y-3">
      {/* Header */}
      <div className="flex items-center gap-3">
        <Skeleton className="w-10 h-10 rounded-full flex-shrink-0" />
        <div className="flex-1 space-y-1.5">
          <Skeleton className="h-3.5 w-28" />
          <Skeleton className="h-3 w-16" />
        </div>
      </div>
      {/* Title */}
      <Skeleton className="h-5 w-3/4" />
      {/* Content lines */}
      <div className="space-y-2">
        <Skeleton className="h-3.5 w-full" />
        <Skeleton className="h-3.5 w-5/6" />
        <Skeleton className="h-3.5 w-4/6" />
      </div>
      {/* Image placeholder */}
      <Skeleton className="h-40 w-full rounded-lg" />
      {/* Actions */}
      <div className="flex gap-4 pt-1 border-t">
        <Skeleton className="h-4 w-12" />
        <Skeleton className="h-4 w-12" />
      </div>
    </div>
  );
}
