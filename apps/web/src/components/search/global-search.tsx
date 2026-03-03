'use client'

import { useState, useEffect, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { useQuery } from '@tanstack/react-query'
import { Search, X } from 'lucide-react'
import { apiClient } from '@/lib/api-client'
import { Input } from '@/components/ui/input'

interface SearchSuggestion {
  type: 'game' | 'album'
  title: string
  url: string
}

export function GlobalSearch() {
  const router = useRouter()
  const [query, setQuery] = useState('')
  const [isOpen, setIsOpen] = useState(false)
  const searchRef = useRef<HTMLDivElement>(null)

  const { data: suggestions } = useQuery<SearchSuggestion[]>({
    queryKey: ['search-suggestions', query],
    queryFn: async () => {
      if (query.length < 2) return []
      const response = await apiClient.get('/search/suggestions', {
        params: { q: query },
      })
      return response.data.data
    },
    enabled: query.length >= 2,
  })

  const { data: popularSearches } = useQuery<string[]>({
    queryKey: ['popular-searches'],
    queryFn: async () => {
      const response = await apiClient.get('/search/popular')
      return response.data.data
    },
  })

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (searchRef.current && !searchRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handleSearch = (searchQuery: string) => {
    if (searchQuery.trim()) {
      router.push(`/search?q=${encodeURIComponent(searchQuery.trim())}`)
      setIsOpen(false)
      setQuery('')
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSearch(query)
    }
  }

  const handleSuggestionClick = (url: string) => {
    router.push(url)
    setIsOpen(false)
    setQuery('')
  }

  return (
    <div ref={searchRef} className="relative w-full max-w-xl">
      <div className="relative">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
        <Input
          type="text"
          placeholder="搜索游戏、音乐..."
          value={query}
          onChange={(e) => {
            setQuery(e.target.value)
            setIsOpen(true)
          }}
          onFocus={() => setIsOpen(true)}
          onKeyDown={handleKeyDown}
          className="pl-10 pr-10"
        />
        {query && (
          <button
            onClick={() => {
              setQuery('')
              setIsOpen(false)
            }}
            className="absolute right-3 top-1/2 transform -translate-y-1/2"
          >
            <X className="h-4 w-4 text-gray-400 hover:text-gray-600" />
          </button>
        )}
      </div>

      {isOpen && (
        <div className="absolute top-full mt-2 w-full bg-white border rounded-lg shadow-lg z-50 max-h-96 overflow-y-auto">
          {query.length >= 2 && suggestions && suggestions.length > 0 ? (
            <div className="py-2">
              <p className="px-4 py-2 text-xs font-medium text-gray-500 uppercase">搜索建议</p>
              {suggestions.map((suggestion, index) => (
                <button
                  key={index}
                  onClick={() => handleSuggestionClick(suggestion.url)}
                  className="w-full px-4 py-2 text-left hover:bg-gray-50 flex items-center gap-3"
                >
                  <Search className="h-4 w-4 text-gray-400" />
                  <div>
                    <p className="font-medium">{suggestion.title}</p>
                    <p className="text-xs text-gray-500">
                      {suggestion.type === 'game' ? '游戏' : '音乐'}
                    </p>
                  </div>
                </button>
              ))}
            </div>
          ) : query.length < 2 && popularSearches && popularSearches.length > 0 ? (
            <div className="py-2">
              <p className="px-4 py-2 text-xs font-medium text-gray-500 uppercase">热门搜索</p>
              {popularSearches.map((search, index) => (
                <button
                  key={index}
                  onClick={() => handleSearch(search)}
                  className="w-full px-4 py-2 text-left hover:bg-gray-50 flex items-center gap-3"
                >
                  <Search className="h-4 w-4 text-gray-400" />
                  <span>{search}</span>
                </button>
              ))}
            </div>
          ) : query.length >= 2 ? (
            <div className="py-8 text-center text-gray-500">
              <p>未找到相关结果</p>
            </div>
          ) : null}
        </div>
      )}
    </div>
  )
}
