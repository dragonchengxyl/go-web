'use client'

import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

interface LeaderboardEntry {
  rank: number
  user_id: string
  username: string
  score: number
}

const medalColor = (rank: number) => {
  if (rank === 1) return 'text-yellow-500'
  if (rank === 2) return 'text-gray-400'
  if (rank === 3) return 'text-amber-600'
  return 'text-gray-500'
}

const medalIcon = (rank: number) => {
  if (rank === 1) return '🥇'
  if (rank === 2) return '🥈'
  if (rank === 3) return '🥉'
  return `#${rank}`
}

function LeaderboardTable({ entries }: { entries: LeaderboardEntry[] }) {
  if (entries.length === 0) {
    return <div className="text-center py-12 text-gray-500">暂无数据</div>
  }
  return (
    <div className="space-y-2">
      {entries.map((entry) => (
        <div
          key={entry.user_id}
          className={`flex items-center gap-4 p-4 rounded-lg border ${
            entry.rank <= 3 ? 'bg-yellow-50 border-yellow-200' : 'bg-white'
          }`}
        >
          <span className={`text-2xl font-bold w-10 text-center ${medalColor(entry.rank)}`}>
            {medalIcon(entry.rank)}
          </span>
          <div className="flex-1">
            <p className="font-semibold">
              {entry.username || entry.user_id.slice(0, 8) + '...'}
            </p>
          </div>
          <div className="text-right">
            <p className="font-bold text-lg">{Math.round(entry.score).toLocaleString()}</p>
            <p className="text-xs text-gray-500">积分</p>
          </div>
        </div>
      ))}
    </div>
  )
}

export default function LeaderboardPage() {
  const [tab, setTab] = useState('total')

  const { data: totalBoard, isLoading: loadingTotal } = useQuery<LeaderboardEntry[]>({
    queryKey: ['leaderboard', 'total'],
    queryFn: () => apiClient.get('/leaderboard?limit=20'),
    enabled: tab === 'total',
  })

  const { data: weeklyBoard, isLoading: loadingWeekly } = useQuery<LeaderboardEntry[]>({
    queryKey: ['leaderboard', 'weekly'],
    queryFn: () => apiClient.get('/leaderboard/weekly?limit=20'),
    enabled: tab === 'weekly',
  })

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">积分排行榜</h1>
        <p className="text-gray-600">完成任务、发表评论、购买内容获取积分，登上排行榜</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>排行榜</CardTitle>
        </CardHeader>
        <CardContent>
          <Tabs value={tab} onValueChange={setTab}>
            <TabsList className="mb-6">
              <TabsTrigger value="total">总榜</TabsTrigger>
              <TabsTrigger value="weekly">本周榜</TabsTrigger>
            </TabsList>

            <TabsContent value="total">
              {loadingTotal
                ? <div className="text-center py-8">加载中...</div>
                : <LeaderboardTable entries={totalBoard ?? []} />
              }
            </TabsContent>

            <TabsContent value="weekly">
              {loadingWeekly
                ? <div className="text-center py-8">加载中...</div>
                : <LeaderboardTable entries={weeklyBoard ?? []} />
              }
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>

      {/* Points info */}
      <Card className="mt-6">
        <CardHeader>
          <CardTitle className="text-base">积分获取说明</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
            {[
              { label: '首次注册', points: '+10' },
              { label: '发布评论', points: '+1' },
              { label: '评论获赞', points: '+3' },
              { label: '购买游戏', points: '+50' },
              { label: '解锁成就', points: '+20起' },
              { label: '首次下载', points: '+10' },
            ].map((item) => (
              <div key={item.label} className="flex justify-between items-center p-3 bg-gray-50 rounded">
                <span className="text-gray-600">{item.label}</span>
                <span className="font-bold text-green-600">{item.points}</span>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
