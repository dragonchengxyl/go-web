'use client';

import { Header } from '@/components/layout/header';
import { Footer } from '@/components/layout/footer';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Heart, Code, Gamepad2, Music } from 'lucide-react';

export default function AboutPage() {
  return (
    <div className="min-h-screen">
      <Header />
      <main className="pt-16">
        {/* Hero Section */}
        <section className="bg-gradient-to-br from-primary/10 to-secondary/10 py-20">
          <div className="container mx-auto px-4 text-center">
            <h1 className="text-5xl font-bold mb-6">关于我们</h1>
            <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
              我们是一群热爱游戏的独立开发者，致力于创造独特而有趣的游戏体验
            </p>
          </div>
        </section>

        {/* Mission */}
        <section className="container mx-auto px-4 py-16">
          <div className="max-w-3xl mx-auto text-center mb-16">
            <h2 className="text-3xl font-bold mb-6">我们的使命</h2>
            <p className="text-lg text-muted-foreground">
              通过创新的游戏设计和精美的艺术风格，为玩家带来难忘的游戏体验。
              我们相信独立游戏的力量，相信小团队也能创造出令人惊叹的作品。
            </p>
          </div>

          {/* Values */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <Card>
              <CardContent className="pt-6 text-center">
                <Heart className="h-12 w-12 text-primary mx-auto mb-4" />
                <h3 className="font-bold text-lg mb-2">用心创作</h3>
                <p className="text-sm text-muted-foreground">
                  每一个细节都经过精心打磨
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="pt-6 text-center">
                <Code className="h-12 w-12 text-primary mx-auto mb-4" />
                <h3 className="font-bold text-lg mb-2">技术创新</h3>
                <p className="text-sm text-muted-foreground">
                  探索新技术，突破传统界限
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="pt-6 text-center">
                <Gamepad2 className="h-12 w-12 text-primary mx-auto mb-4" />
                <h3 className="font-bold text-lg mb-2">玩家至上</h3>
                <p className="text-sm text-muted-foreground">
                  倾听玩家声音，持续改进
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="pt-6 text-center">
                <Music className="h-12 w-12 text-primary mx-auto mb-4" />
                <h3 className="font-bold text-lg mb-2">艺术追求</h3>
                <p className="text-sm text-muted-foreground">
                  游戏不仅是娱乐，更是艺术
                </p>
              </CardContent>
            </Card>
          </div>
        </section>

        {/* Team */}
        <section className="bg-muted/30 py-16">
          <div className="container mx-auto px-4">
            <div className="max-w-3xl mx-auto text-center">
              <h2 className="text-3xl font-bold mb-6">我们的团队</h2>
              <p className="text-lg text-muted-foreground mb-8">
                一个由程序员、设计师、音乐家和游戏爱好者组成的多元化团队
              </p>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
                {[1, 2, 3, 4].map((i) => (
                  <div key={i} className="text-center">
                    <div className="w-24 h-24 bg-primary rounded-full mx-auto mb-3" />
                    <p className="font-medium">团队成员 {i}</p>
                    <p className="text-sm text-muted-foreground">职位</p>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </section>

        {/* Contact */}
        <section className="container mx-auto px-4 py-16">
          <div className="max-w-2xl mx-auto text-center">
            <h2 className="text-3xl font-bold mb-6">联系我们</h2>
            <p className="text-lg text-muted-foreground mb-8">
              有任何问题或建议？欢迎随时与我们联系
            </p>
            <div className="space-y-4">
              <p className="text-muted-foreground">
                邮箱: contact@studio.example.com
              </p>
              <p className="text-muted-foreground">
                Discord: studio-community
              </p>
              <p className="text-muted-foreground">
                Twitter: @IndieStudio
              </p>
            </div>
          </div>
        </section>
      </main>
      <Footer />
    </div>
  );
}
