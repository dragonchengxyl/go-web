'use client'

import { create } from 'zustand'
import { useEffect, useRef } from 'react'
import { AudioLines, Pause, Play, Waves } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { type AudioWork } from '@/lib/api-client'

export interface PlayerTrack {
  id: string
  title: string
  artist: string
  cover?: string
  audioUrl: string
  kind: 'audio_work'
}

interface PlayerState {
  currentTrack: PlayerTrack | null
  isPlaying: boolean
  volume: number
  currentTime: number
  duration: number
  setTrack: (track: PlayerTrack) => void
  clearTrack: () => void
  play: () => void
  pause: () => void
  togglePlay: () => void
  setVolume: (volume: number) => void
  seek: (time: number) => void
  setCurrentTime: (time: number) => void
  setDuration: (duration: number) => void
}

export const usePlayerStore = create<PlayerState>((set) => ({
  currentTrack: null,
  isPlaying: false,
  volume: 0.7,
  currentTime: 0,
  duration: 0,
  setTrack: (track) => set({ currentTrack: track, isPlaying: true, currentTime: 0 }),
  clearTrack: () => set({ currentTrack: null, isPlaying: false, currentTime: 0, duration: 0 }),
  play: () => set({ isPlaying: true }),
  pause: () => set({ isPlaying: false }),
  togglePlay: () => set((state) => ({ isPlaying: !state.isPlaying })),
  setVolume: (volume) => set({ volume }),
  seek: (time) => set({ currentTime: time }),
  setCurrentTime: (time) => set({ currentTime: time }),
  setDuration: (duration) => set({ duration }),
}))

export function audioWorkToPlayerTrack(work: AudioWork): PlayerTrack {
  return {
    id: work.id,
    title: work.title,
    artist: work.author_username ? `@${work.author_username}` : work.author_id,
    cover: work.cover_image_url,
    audioUrl: work.audio_url,
    kind: 'audio_work',
  }
}

export function MusicPlayer() {
  const audioRef = useRef<HTMLAudioElement>(null)
  const {
    currentTrack,
    isPlaying,
    volume,
    currentTime,
    duration,
    pause,
    togglePlay,
    setCurrentTime,
    setDuration,
    clearTrack,
  } = usePlayerStore()

  useEffect(() => {
    if (!audioRef.current) return

    if (isPlaying) {
      audioRef.current.play()
    } else {
      audioRef.current.pause()
    }
  }, [isPlaying])

  useEffect(() => {
    if (!audioRef.current) return
    audioRef.current.volume = volume
  }, [volume])

  useEffect(() => {
    if (!audioRef.current || !currentTrack) return
    audioRef.current.src = currentTrack.audioUrl
    audioRef.current.play()
  }, [currentTrack])

  const handleTimeUpdate = () => {
    if (audioRef.current) {
      setCurrentTime(audioRef.current.currentTime)
    }
  }

  const handleLoadedMetadata = () => {
    if (audioRef.current) {
      setDuration(audioRef.current.duration)
    }
  }

  const handleSeek = (e: React.ChangeEvent<HTMLInputElement>) => {
    const time = parseFloat(e.target.value)
    if (audioRef.current) {
      audioRef.current.currentTime = time
      setCurrentTime(time)
    }
  }

  const formatTime = (seconds: number) => {
    if (isNaN(seconds)) return '0:00'
    const mins = Math.floor(seconds / 60)
    const secs = Math.floor(seconds % 60)
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }

  if (!currentTrack) return null

  return (
    <>
      <audio
        ref={audioRef}
        onTimeUpdate={handleTimeUpdate}
        onLoadedMetadata={handleLoadedMetadata}
        onEnded={() => pause()}
      />

      <div className="fixed bottom-0 left-0 right-0 z-50 border-t border-white/10 bg-slate-950/92 text-white shadow-2xl backdrop-blur-xl">
        <div className="container mx-auto px-4 py-3">
          <div className="flex flex-col gap-3 md:flex-row md:items-center md:gap-4">
            <div
              className="flex h-14 w-14 items-center justify-center rounded-xl border border-white/10 bg-gradient-to-br from-emerald-500/30 to-cyan-500/20"
              style={
                currentTrack.cover
                  ? {
                      backgroundImage: `url(${currentTrack.cover})`,
                      backgroundSize: 'cover',
                      backgroundPosition: 'center',
                    }
                  : undefined
              }
            >
              {!currentTrack.cover ? <AudioLines className="h-5 w-5 text-white/80" /> : null}
            </div>

            <div className="min-w-0 flex-1">
              <p className="font-medium truncate">{currentTrack.title}</p>
              <div className="mt-1 flex items-center gap-2 text-sm text-slate-300">
                <span className="truncate">{currentTrack.artist}</span>
                <span className="text-slate-600">·</span>
                <span className="inline-flex items-center gap-1 text-xs uppercase tracking-[0.18em] text-emerald-300">
                  <Waves className="h-3 w-3" />
                  Audio Work
                </span>
              </div>
            </div>

            <div className="flex items-center gap-3 md:w-[460px]">
              <Button
                variant="secondary"
                size="sm"
                onClick={togglePlay}
                className="w-20 bg-white/10 text-white hover:bg-white/20"
              >
                {isPlaying ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4" />}
              </Button>

              <div className="flex flex-1 items-center gap-2">
                <span className="w-10 text-right text-xs text-slate-400">
                  {formatTime(currentTime)}
                </span>
                <input
                  type="range"
                  min="0"
                  max={duration || 0}
                  value={currentTime}
                  onChange={handleSeek}
                  className="flex-1 accent-emerald-400"
                />
                <span className="w-10 text-xs text-slate-400">
                  {formatTime(duration)}
                </span>
              </div>

              <Button
                variant="ghost"
                size="sm"
                onClick={clearTrack}
                className="text-slate-300 hover:bg-white/10 hover:text-white"
              >
                关闭
              </Button>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}
