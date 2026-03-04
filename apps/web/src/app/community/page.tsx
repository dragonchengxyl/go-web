'use client';

import { Header } from '@/components/layout/header';
import { Footer } from '@/components/layout/footer';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { MessageSquare, Users, Trophy, Calendar } from 'lucide-react';

export default function CommunityPage() {
  return (
    <div className="min-h-screen">
      <Header />
      <main className="pt-16">
        {/* Hero Section */}
        <section className="bg-gradient-to-br from-primary/10 to-secondary/10 py-20">
          <div className="container mx-auto px-4 text-center">
            <h1 className="text-5xl font-bold mb-6">游戏社区</h1>
            <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
              与全球玩家交流，分享游戏心得，参与社区活动
            </p>
          </div>
        </section>

        {/* Features */}
        <section className="container mx-auto px-4 py-12">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <Card>
              <CardHeader>
                <MessageSquare className="h-12 w-12 text-primary mb-4" />
                <CardTitle>论坛讨论</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-muted-foreground">
                  在论坛中与其他玩家交流游戏心得和攻略
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <Users className="h-12 w-12 text-primary mb-4" />
                <CardTitle>玩家组队</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-muted-foreground">
                  寻找志同道合的队友，一起畅玩游戏
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <Trophy className="h-12 w-12 text-primary mb-4" />
                <CardTitle>竞赛活动</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-muted-foreground">
                  参与官方举办的各类竞赛，赢取丰厚奖励
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <Calendar className="h-12 w-12 text-primary mb-4" />
                <CardTitle>社区活动</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-muted-foreground">
                  定期举办线上线下活动，增进玩家友谊
                </p>
              </CardContent>
            </Card>
          </div>
        </section>

        {/* Coming Soon */}
        <section className="container mx-auto px-4 py-12">
          <Card className="text-center py-12">
            <CardContent>
              <h2 className="text-3xl font-bold mb-4">社区功能即将上线</h2>
              <p className="text-muted-foreground mb-6">
                我们正在努力打造一个活跃、友好的游戏社区，敬请期待！
              </p>
              <Button size="lg" disabled>
                敬请期待
              </Button>
            </CardContent>
          </Card>
        </section>
      </main>
      <Footer />
    </div>
  );
}
