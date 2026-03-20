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

export type PlayerRepeatMode = 'off' | 'one' | 'all'

export interface PlayerSourceContext {
  kind: string
  label?: string
  href?: string
  entityId?: string
}

export interface PlayerSession {
  id: string
  queue: PlayerTrack[]
  currentIndex: number
  repeatMode: PlayerRepeatMode
  shuffle: boolean
  sourceContext?: PlayerSourceContext
  openedFrom?: string
  createdAt: number
}

export interface PlayerSessionInput {
  queue: PlayerTrack[]
  currentIndex?: number
  repeatMode?: PlayerRepeatMode
  shuffle?: boolean
  sourceContext?: PlayerSourceContext
  openedFrom?: string
  autoplay?: boolean
}

export interface SetTrackOptions {
  queue?: PlayerTrack[]
  currentIndex?: number
  repeatMode?: PlayerRepeatMode
  shuffle?: boolean
  sourceContext?: PlayerSourceContext
  openedFrom?: string
  autoplay?: boolean
}

const PLAYER_HISTORY_KEY = 'audio_player_history'
const MAX_PLAYER_HISTORY = 12

interface PlayerState {
  session: PlayerSession | null
  currentTrack: PlayerTrack | null
  queue: PlayerTrack[]
  history: PlayerTrack[]
  currentIndex: number
  repeatMode: PlayerRepeatMode
  shuffle: boolean
  sourceContext: PlayerSourceContext | null
  openedFrom: string | null
  isPlaying: boolean
  volume: number
  currentTime: number
  duration: number
  playbackVersion: number
  setSession: (input: PlayerSessionInput) => void
  setTrack: (track: PlayerTrack, options?: SetTrackOptions) => void
  replaceQueue: (queue: PlayerTrack[], options?: Omit<PlayerSessionInput, 'queue'>) => void
  selectTrack: (trackId: string) => void
  hydrateHistory: (tracks: PlayerTrack[]) => void
  rememberTrack: (track: PlayerTrack) => void
  clearTrack: () => void
  play: () => void
  pause: () => void
  togglePlay: () => void
  playNext: () => void
  playPrevious: () => void
  handleTrackEnded: () => void
  setVolume: (volume: number) => void
  seek: (time: number) => void
  setCurrentTime: (time: number) => void
  setDuration: (duration: number) => void
  setRepeatMode: (mode: PlayerRepeatMode) => void
  toggleShuffle: () => void
  setShuffle: (enabled: boolean) => void
}

const DEFAULT_REPEAT_MODE: PlayerRepeatMode = 'off'

function readStoredHistory(): PlayerTrack[] {
  if (typeof window === 'undefined') return []
  try {
    const raw = window.localStorage.getItem(PLAYER_HISTORY_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return parsed.filter((item): item is PlayerTrack => {
      return (
        !!item &&
        typeof item === 'object' &&
        typeof item.id === 'string' &&
        typeof item.title === 'string' &&
        typeof item.artist === 'string' &&
        typeof item.audioUrl === 'string' &&
        item.kind === 'audio_work'
      )
    })
  } catch {
    return []
  }
}

function writeStoredHistory(history: PlayerTrack[]) {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(PLAYER_HISTORY_KEY, JSON.stringify(history))
  } catch {
    // ignore local persistence failures
  }
}

function updateHistory(history: PlayerTrack[], track: PlayerTrack) {
  const next = [track, ...history.filter((item) => item.id !== track.id)]
  return next.slice(0, MAX_PLAYER_HISTORY)
}

function clampIndex(index: number | undefined, queueLength: number) {
  if (queueLength <= 0) return -1
  if (typeof index !== 'number' || Number.isNaN(index)) return 0
  return Math.max(0, Math.min(queueLength - 1, index))
}

function trackAt(session: PlayerSession | null) {
  if (!session || session.currentIndex < 0 || session.currentIndex >= session.queue.length) {
    return null
  }
  return session.queue[session.currentIndex] ?? null
}

function buildSession(input: PlayerSessionInput, previous?: PlayerSession | null) {
  if (!input.queue.length) return null

  const currentIndex = clampIndex(input.currentIndex, input.queue.length)
  const createdAt = Date.now()

  return {
    id: `${input.queue[currentIndex].id}:${createdAt}`,
    queue: [...input.queue],
    currentIndex,
    repeatMode: input.repeatMode ?? previous?.repeatMode ?? DEFAULT_REPEAT_MODE,
    shuffle: input.shuffle ?? previous?.shuffle ?? false,
    sourceContext: input.sourceContext ?? previous?.sourceContext,
    openedFrom: input.openedFrom ?? previous?.openedFrom,
    createdAt,
  } satisfies PlayerSession
}

function buildStateFromSession(
  session: PlayerSession | null,
  state: Pick<PlayerState, 'isPlaying' | 'volume' | 'currentTime' | 'duration' | 'playbackVersion'>
) {
  const currentTrack = trackAt(session)
  return {
    session,
    currentTrack,
    queue: session?.queue ?? [],
    currentIndex: session?.currentIndex ?? -1,
    repeatMode: session?.repeatMode ?? DEFAULT_REPEAT_MODE,
    shuffle: session?.shuffle ?? false,
    sourceContext: session?.sourceContext ?? null,
    openedFrom: session?.openedFrom ?? null,
    isPlaying: currentTrack ? state.isPlaying : false,
    volume: state.volume,
    currentTime: currentTrack ? state.currentTime : 0,
    duration: currentTrack ? state.duration : 0,
    playbackVersion: state.playbackVersion,
  }
}

function chooseNextIndex(session: PlayerSession, direction: 'next' | 'previous') {
  const queueLength = session.queue.length
  if (queueLength <= 1) {
    if (direction === 'next' && session.repeatMode === 'all') return 0
    return session.currentIndex
  }

  if (session.shuffle) {
    const choices = session.queue
      .map((_, index) => index)
      .filter((index) => index !== session.currentIndex)
    return choices[Math.floor(Math.random() * choices.length)] ?? session.currentIndex
  }

  if (direction === 'next') {
    if (session.currentIndex < queueLength - 1) {
      return session.currentIndex + 1
    }
    return session.repeatMode === 'all' ? 0 : session.currentIndex
  }

  if (session.currentIndex > 0) {
    return session.currentIndex - 1
  }
  return session.repeatMode === 'all' ? queueLength - 1 : session.currentIndex
}

function advanceSession(
  state: PlayerState,
  direction: 'next' | 'previous',
  options?: { resetToTrackStart?: boolean }
) {
  const session = state.session
  if (!session) {
    return state
  }

  if (direction === 'previous' && options?.resetToTrackStart && state.currentTime > 3) {
    return {
      ...state,
      currentTime: 0,
      playbackVersion: state.playbackVersion + 1,
      isPlaying: true,
    }
  }

  const nextIndex = chooseNextIndex(session, direction)
  if (nextIndex === session.currentIndex) {
    return state
  }

  const nextSession = { ...session, currentIndex: nextIndex }
  return {
    ...buildStateFromSession(nextSession, {
      isPlaying: true,
      volume: state.volume,
      currentTime: 0,
      duration: 0,
      playbackVersion: state.playbackVersion + 1,
    }),
  }
}

export const usePlayerStore = create<PlayerState>((set) => ({
  session: null,
  currentTrack: null,
  queue: [],
  history: [],
  currentIndex: -1,
  repeatMode: DEFAULT_REPEAT_MODE,
  shuffle: false,
  sourceContext: null,
  openedFrom: null,
  isPlaying: false,
  volume: 0.7,
  currentTime: 0,
  duration: 0,
  playbackVersion: 0,

  setSession: (input) =>
    set((state) => {
      const session = buildSession(input, state.session)
      return {
        ...buildStateFromSession(session, {
          isPlaying: input.autoplay ?? true,
          volume: state.volume,
          currentTime: 0,
          duration: 0,
          playbackVersion: state.playbackVersion + 1,
        }),
      }
    }),

  setTrack: (track, options) =>
    set((state) => {
      const queue = options?.queue?.length ? options.queue : [track]
      const currentIndex =
        typeof options?.currentIndex === 'number'
          ? options.currentIndex
          : Math.max(
              0,
              queue.findIndex((item) => item.id === track.id)
            )
      const session = buildSession(
        {
          queue,
          currentIndex,
          repeatMode: options?.repeatMode,
          shuffle: options?.shuffle,
          sourceContext: options?.sourceContext,
          openedFrom: options?.openedFrom,
        },
        state.session
      )

      return {
        ...buildStateFromSession(session, {
          isPlaying: options?.autoplay ?? true,
          volume: state.volume,
          currentTime: 0,
          duration: 0,
          playbackVersion: state.playbackVersion + 1,
        }),
      }
    }),

  replaceQueue: (queue, options) =>
    set((state) => {
      const fallbackIndex = queue.findIndex((item) => item.id === state.currentTrack?.id)
      const session = buildSession(
        {
          queue,
          currentIndex:
            typeof options?.currentIndex === 'number'
              ? options.currentIndex
              : fallbackIndex >= 0
                ? fallbackIndex
                : 0,
          repeatMode: options?.repeatMode,
          shuffle: options?.shuffle,
          sourceContext: options?.sourceContext,
          openedFrom: options?.openedFrom,
        },
        state.session
      )

      return {
        ...buildStateFromSession(session, {
          isPlaying: session ? state.isPlaying : false,
          volume: state.volume,
          currentTime: 0,
          duration: 0,
          playbackVersion: state.playbackVersion + (session ? 1 : 0),
        }),
      }
    }),

  selectTrack: (trackId) =>
    set((state) => {
      if (!state.session) return state

      const nextIndex = state.session.queue.findIndex((track) => track.id === trackId)
      if (nextIndex < 0 || nextIndex === state.session.currentIndex) {
        return state
      }

      const session = { ...state.session, currentIndex: nextIndex }
      return {
        ...buildStateFromSession(session, {
          isPlaying: true,
          volume: state.volume,
          currentTime: 0,
          duration: 0,
          playbackVersion: state.playbackVersion + 1,
        }),
      }
    }),

  hydrateHistory: (tracks) => set({ history: tracks.slice(0, MAX_PLAYER_HISTORY) }),

  rememberTrack: (track) =>
    set((state) => {
      const nextHistory = updateHistory(state.history, track)
      writeStoredHistory(nextHistory)
      return { history: nextHistory }
    }),

  clearTrack: () =>
    set((state) => ({
      ...buildStateFromSession(null, {
        isPlaying: false,
        volume: state.volume,
        currentTime: 0,
        duration: 0,
        playbackVersion: state.playbackVersion + 1,
      }),
    })),

  play: () => set({ isPlaying: true }),
  pause: () => set({ isPlaying: false }),
  togglePlay: () => set((state) => ({ isPlaying: !state.isPlaying })),
  playNext: () => set((state) => advanceSession(state, 'next')),
  playPrevious: () => set((state) => advanceSession(state, 'previous', { resetToTrackStart: true })),

  handleTrackEnded: () =>
    set((state) => {
      if (!state.session) return state

      if (state.repeatMode === 'one') {
        return {
          ...state,
          currentTime: 0,
          isPlaying: true,
          playbackVersion: state.playbackVersion + 1,
        }
      }

      const nextIndex = chooseNextIndex(state.session, 'next')
      if (nextIndex !== state.session.currentIndex) {
        const session = { ...state.session, currentIndex: nextIndex }
        return {
          ...buildStateFromSession(session, {
            isPlaying: true,
            volume: state.volume,
            currentTime: 0,
            duration: 0,
            playbackVersion: state.playbackVersion + 1,
          }),
        }
      }

      if (state.repeatMode === 'all' && state.session.queue.length > 0) {
        const session = { ...state.session, currentIndex: 0 }
        return {
          ...buildStateFromSession(session, {
            isPlaying: true,
            volume: state.volume,
            currentTime: 0,
            duration: 0,
            playbackVersion: state.playbackVersion + 1,
          }),
        }
      }

      return {
        ...state,
        isPlaying: false,
        currentTime: state.duration,
      }
    }),

  setVolume: (volume) => set({ volume }),
  seek: (time) => set({ currentTime: time }),
  setCurrentTime: (time) => set({ currentTime: time }),
  setDuration: (duration) => set({ duration }),

  setRepeatMode: (mode) =>
    set((state) => {
      if (!state.session) {
        return { repeatMode: mode }
      }
      const session = { ...state.session, repeatMode: mode }
      return {
        ...state,
        session,
        repeatMode: mode,
      }
    }),

  toggleShuffle: () =>
    set((state) => {
      const nextShuffle = !state.shuffle
      if (!state.session) {
        return { shuffle: nextShuffle }
      }
      const session = { ...state.session, shuffle: nextShuffle }
      return {
        ...state,
        session,
        shuffle: nextShuffle,
      }
    }),

  setShuffle: (enabled) =>
    set((state) => {
      if (!state.session) {
        return { shuffle: enabled }
      }
      const session = { ...state.session, shuffle: enabled }
      return {
        ...state,
        session,
        shuffle: enabled,
      }
    }),
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

export function audioWorksToPlayerTracks(works: AudioWork[]) {
  return works.map(audioWorkToPlayerTrack)
}

export function MusicPlayer() {
  const audioRef = useRef<HTMLAudioElement>(null)
  const hydratedHistoryRef = useRef(false)
  const {
    currentTrack,
    isPlaying,
    volume,
    currentTime,
    duration,
    togglePlay,
    setCurrentTime,
    setDuration,
    hydrateHistory,
    rememberTrack,
    clearTrack,
    handleTrackEnded,
    playbackVersion,
  } = usePlayerStore()

  useEffect(() => {
    if (hydratedHistoryRef.current) return
    hydratedHistoryRef.current = true
    hydrateHistory(readStoredHistory())
  }, [hydrateHistory])

  useEffect(() => {
    if (!audioRef.current) return

    if (isPlaying) {
      void audioRef.current.play().catch(() => undefined)
    } else {
      audioRef.current.pause()
    }
  }, [isPlaying, playbackVersion])

  useEffect(() => {
    if (!audioRef.current) return
    audioRef.current.volume = volume
  }, [volume])

  useEffect(() => {
    if (!audioRef.current || !currentTrack) return
    audioRef.current.src = currentTrack.audioUrl
    audioRef.current.currentTime = 0
    if (isPlaying) {
      void audioRef.current.play().catch(() => undefined)
    }
  }, [currentTrack, isPlaying, playbackVersion])

  useEffect(() => {
    if (!currentTrack) return
    rememberTrack(currentTrack)
  }, [currentTrack, rememberTrack])

  useEffect(() => {
    if (!audioRef.current || !Number.isFinite(currentTime)) return
    if (Math.abs(audioRef.current.currentTime - currentTime) > 0.5) {
      audioRef.current.currentTime = currentTime
    }
  }, [currentTime])

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
        onEnded={handleTrackEnded}
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
