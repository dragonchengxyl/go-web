'use client'

import { Sparkles, Zap } from 'lucide-react'
import { HexBlitzTile } from '@/lib/api-client'
import { cn } from '@/lib/utils'

const HEX_SIZE = 34
const BOARD_PADDING = 56

const COLOR_STYLES: Record<
  HexBlitzTile['color'],
  { gradient: string; glow: string }
> = {
  ember: {
    gradient: 'linear-gradient(180deg, #ffb36a 0%, #ff7a45 48%, #eb4e30 100%)',
    glow: '0 12px 26px rgba(255, 122, 69, 0.42)',
  },
  lagoon: {
    gradient: 'linear-gradient(180deg, #70e7ff 0%, #2fc2ff 45%, #178fdf 100%)',
    glow: '0 12px 26px rgba(47, 194, 255, 0.42)',
  },
  mint: {
    gradient: 'linear-gradient(180deg, #89ffd0 0%, #3fd9a5 45%, #1ca979 100%)',
    glow: '0 12px 26px rgba(63, 217, 165, 0.4)',
  },
  sun: {
    gradient: 'linear-gradient(180deg, #ffe88c 0%, #ffc847 45%, #ff9d1d 100%)',
    glow: '0 12px 26px rgba(255, 200, 71, 0.42)',
  },
  violet: {
    gradient: 'linear-gradient(180deg, #d8b4ff 0%, #a66cff 45%, #6b46ff 100%)',
    glow: '0 12px 26px rgba(166, 108, 255, 0.42)',
  },
}

function axialToPixel(q: number, r: number) {
  return {
    x: Math.sqrt(3) * HEX_SIZE * (q + r / 2),
    y: 1.5 * HEX_SIZE * r,
  }
}

function computeBoardBounds(tiles: HexBlitzTile[]) {
  if (tiles.length === 0) {
    return {
      minX: 0,
      minY: 0,
      width: HEX_SIZE * 8,
      height: HEX_SIZE * 8,
    }
  }

  const points = tiles.map((tile) => axialToPixel(tile.q, tile.r))
  const xs = points.map((point) => point.x)
  const ys = points.map((point) => point.y)
  const minX = Math.min(...xs)
  const maxX = Math.max(...xs)
  const minY = Math.min(...ys)
  const maxY = Math.max(...ys)

  return {
    minX,
    minY,
    width: maxX - minX + HEX_SIZE * 2 + BOARD_PADDING * 2,
    height: maxY - minY + HEX_SIZE * 2 + BOARD_PADDING * 2,
  }
}

export function HexBlitzBoard({
  tiles,
  interactive = false,
  highlightedTileIds,
  onTileClick,
  onHoverChange,
  className,
}: {
  tiles: HexBlitzTile[]
  interactive?: boolean
  highlightedTileIds?: Set<string>
  onTileClick?: (tileId: string) => void
  onHoverChange?: (tileId: string | null) => void
  className?: string
}) {
  const bounds = computeBoardBounds(tiles)

  return (
    <div
      className={cn(
        'relative mx-auto rounded-[32px] border border-white/10 bg-[radial-gradient(circle_at_center,rgba(255,255,255,0.04),transparent_56%),linear-gradient(180deg,rgba(255,255,255,0.02),rgba(0,0,0,0.18))]',
        className
      )}
      style={{
        width: `${bounds.width}px`,
        height: `${bounds.height}px`,
      }}
    >
      <div className="absolute inset-5 rounded-[26px] border border-white/8 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.03),transparent_45%)]" />
      <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,transparent_0%,rgba(0,0,0,0.2)_100%)]" />
      {tiles.map((tile) => {
        const pixel = axialToPixel(tile.q, tile.r)
        const left =
          pixel.x -
          bounds.minX +
          BOARD_PADDING -
          Math.sqrt(3) * HEX_SIZE / 2
        const top = pixel.y - bounds.minY + BOARD_PADDING - HEX_SIZE
        const style = COLOR_STYLES[tile.color]
        const highlighted = highlightedTileIds?.has(tile.id) ?? false

        return (
          <button
            key={tile.id}
            type="button"
            onMouseEnter={() => onHoverChange?.(tile.id)}
            onMouseLeave={() => onHoverChange?.(null)}
            onClick={() => onTileClick?.(tile.id)}
            className={cn(
              'absolute flex items-center justify-center text-slate-950 transition-all duration-150',
              interactive ? 'cursor-pointer hover:-translate-y-1' : 'cursor-default'
            )}
            style={{
              left,
              top,
              width: `${Math.sqrt(3) * HEX_SIZE}px`,
              height: `${HEX_SIZE * 2}px`,
              clipPath:
                'polygon(25% 6%, 75% 6%, 100% 50%, 75% 94%, 25% 94%, 0 50%)',
              background: style.gradient,
              boxShadow: highlighted
                ? `${style.glow}, 0 0 0 3px rgba(255,255,255,0.16)`
                : style.glow,
              transform: highlighted ? 'translateY(-4px) scale(1.04)' : 'translateY(0) scale(1)',
            }}
          >
            <div className="absolute inset-[2px] rounded-[22px] bg-[linear-gradient(180deg,rgba(255,255,255,0.22),transparent_55%)] opacity-70" />
            {tile.special === 'spark' && (
              <Sparkles className="relative h-5 w-5 text-white drop-shadow-[0_2px_8px_rgba(0,0,0,0.35)]" />
            )}
            {tile.special === 'burst' && (
              <Zap className="relative h-5 w-5 text-white drop-shadow-[0_2px_8px_rgba(0,0,0,0.35)]" />
            )}
          </button>
        )
      })}
    </div>
  )
}
