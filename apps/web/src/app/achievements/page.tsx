'use client'

import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

interface Achievement {
  id: number
  slug: string
  name: string
  description: string
  rarity: 'common' | 'rare' | 'epic' | 'legendary'
  points: number
  is_secret: boolean
}

const rarityColor: Record<string, string> = {
  common:    'bg-gray-200 text-gray-800',
  rare:      'bg-blue-100 text-blue-800',
  epic:      'bg-purple-100 text-purple-800',
  legendary: 'bg-yellow-100 text-yellow-800',
}

const rarityLabel: Record<string, string> = {
  common:    '普通',
  rare:      '稀有',
  epic:      '史诗',
  legendary: '传说',
}

const rarityIcon: Record<string, string> = {
  common:    '⚪',
  rare:      '🔵',
  epic:      '🟣',
  legendary: '🌟',
}

export default function AchievementsPage() {
  const { data: achievements, isLoading } = useQuery<Achievement[]>({
    queryKey: ['achievements'],
    queryFn: () => apiClient.get('/achievements'),
  })

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-8 text-center">加载中...</div>
    )
  }

  const grouped = (achievements ?? []).reduce<Record<string, Achievement[]>>((acc, a) => {
    acc[a.rarity] = acc[a.rarity] ?? []
    acc[a.rarity].push(a)
    return acc
  }, {})

  const order = ['legendary', 'epic', 'rare', 'common']

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">成就系统</h1>
        <p className="text-gray-600">完成特定任务解锁成就，赢得积分奖励</p>
      </div>

      <div className="space-y-8">
        {order.map((rarity) => {
          const list = grouped[rarity]
          if (!list?.length) return null
          return (
            <section key={rarity}>
              <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
                <span>{rarityIcon[rarity]}</span>
                <span>{rarityLabel[rarity]}成就</span>
                <Badge className={rarityColor[rarity]}>{list.length}</Badge>
              </h2>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {list.map((a) => (
                  <Card key={a.id} className="hover:shadow-md transition-shadow">
                    <CardHeader className="pb-2">
                      <div className="flex items-start justify-between">
                        <CardTitle className="text-base">{a.name}</CardTitle>
                        <Badge className={rarityColor[a.rarity]}>
                          {rarityLabel[a.rarity]}
                        </Badge>
                      </div>
                    </CardHeader>
                    <CardContent>
                      <p className="text-sm text-gray-600 mb-3">{a.description}</p>
                      <div className="flex items-center gap-1 text-sm font-medium text-yellow-600">
                        <span>⭐</span>
                        <span>+{a.points} 积分</span>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </section>
          )
        })}

        {(!achievements || achievements.length === 0) && (
          <div className="text-center py-20 text-gray-500">暂无成就数据</div>
        )}
      </div>
    </div>
  )
}
