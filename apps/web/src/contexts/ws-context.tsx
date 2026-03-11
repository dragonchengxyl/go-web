'use client'

import React, { createContext, useCallback, useContext, useEffect, useReducer, useRef } from 'react'

const WS_URL = process.env.NEXT_PUBLIC_WS_URL || ''

// Build the WebSocket base URL at runtime so it follows the page's host/port
// (works behind the Next.js dev proxy without hardcoding localhost:8080).
function getWsBase(): string {
  if (WS_URL) return WS_URL
  if (typeof window === 'undefined') return 'ws://localhost:8080'
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${window.location.host}`
}
const BACKOFF = [1000, 2000, 4000, 8000, 16000, 30000]

type WSStatus = 'connecting' | 'connected' | 'disconnected'

interface WSContextValue {
  status: WSStatus
  subscribe: (type: string, handler: (payload: unknown) => void) => () => void
}

interface WSMessage {
  type: string
  conversation_id?: string
  payload: unknown
}

type State = { status: WSStatus }
type Action = { type: 'CONNECTING' } | { type: 'CONNECTED' } | { type: 'DISCONNECTED' }

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case 'CONNECTING': return { status: 'connecting' }
    case 'CONNECTED': return { status: 'connected' }
    case 'DISCONNECTED': return { status: 'disconnected' }
    default: return state
  }
}

const WSContext = createContext<WSContextValue>({
  status: 'disconnected',
  subscribe: () => () => {},
})

export function useWS() {
  return useContext(WSContext)
}

export function WSProvider({ children }: { children: React.ReactNode }) {
  const [state, dispatch] = useReducer(reducer, { status: 'disconnected' })
  const wsRef = useRef<WebSocket | null>(null)
  const attemptRef = useRef(0)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const handlersRef = useRef<Map<string, Set<(payload: unknown) => void>>>(new Map())
  const mountedRef = useRef(true)

  const getToken = () => {
    if (typeof window === 'undefined') return null
    return localStorage.getItem('access_token')
  }

  const isTokenValid = (token: string): boolean => {
    try {
      const parts = token.split('.')
      if (parts.length !== 3) return false
      const payload = JSON.parse(atob(parts[1]))
      return payload.exp * 1000 > Date.now()
    } catch {
      return false
    }
  }

  const clearTimer = () => {
    if (timerRef.current) {
      clearTimeout(timerRef.current)
      timerRef.current = null
    }
  }

  const connect = useCallback(() => {
    if (!mountedRef.current) return
    if (wsRef.current) {
      wsRef.current.onclose = null
      wsRef.current.close()
      wsRef.current = null
    }

    const token = getToken()
    if (!token) return
    if (!isTokenValid(token)) return

    dispatch({ type: 'CONNECTING' })
    const ws = new WebSocket(`${getWsBase()}/ws/chat?token=${token}`)
    wsRef.current = ws

    ws.onopen = () => {
      if (!mountedRef.current) { ws.close(); return }
      dispatch({ type: 'CONNECTED' })
      attemptRef.current = 0
    }

    ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data)
        const handlers = handlersRef.current.get(msg.type)
        handlers?.forEach(h => h(msg.payload))
      } catch {
        // ignore
      }
    }

    ws.onclose = () => {
      if (!mountedRef.current) return
      dispatch({ type: 'DISCONNECTED' })
      wsRef.current = null

      if (document.visibilityState === 'hidden') return

      clearTimer()
      const delay = BACKOFF[Math.min(attemptRef.current, BACKOFF.length - 1)]
      attemptRef.current++
      timerRef.current = setTimeout(connect, delay)
    }

    ws.onerror = () => {
      ws.close()
    }
  }, [])

  useEffect(() => {
    mountedRef.current = true

    const handleVisibility = () => {
      if (document.visibilityState === 'visible' && !wsRef.current) {
        clearTimer()
        connect()
      }
    }
    document.addEventListener('visibilitychange', handleVisibility)

    connect()

    return () => {
      mountedRef.current = false
      clearTimer()
      document.removeEventListener('visibilitychange', handleVisibility)
      if (wsRef.current) {
        wsRef.current.onclose = null
        wsRef.current.close()
        wsRef.current = null
      }
    }
  }, [connect])

  const subscribe = useCallback((type: string, handler: (payload: unknown) => void) => {
    if (!handlersRef.current.has(type)) {
      handlersRef.current.set(type, new Set())
    }
    handlersRef.current.get(type)!.add(handler)

    return () => {
      handlersRef.current.get(type)?.delete(handler)
    }
  }, [])

  return (
    <WSContext.Provider value={{ status: state.status, subscribe }}>
      {children}
    </WSContext.Provider>
  )
}
