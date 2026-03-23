'use client';

import { Bot, Crown, Layers3, Radio, Sparkles, Users2 } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';

const implementationSteps = [
  'Phase 0：Games Hub、详情页与试玩页壳子已经接入。',
  'Phase 1：服务端规则引擎、牌型比较和状态机。',
  'Phase 2：真人房、WebSocket、人机演示模式与断线重连。',
  'Phase 3：战报、最近对局、后台多游戏观测。',
];

const roomModes = [
  {
    title: '真人房模式',
    icon: Users2,
    eyebrow: 'Planned PVP Room',
    description:
      '登录用户创建房间、邀请两位玩家加入，服务端负责发牌、叫分、出牌裁决和结算落库。',
    badge: 'Phase 2',
    cta: '房间服务开发中',
  },
  {
    title: '快速 AI 演示',
    icon: Bot,
    eyebrow: 'Demo Fallback',
    description:
      '单人即可开始一局，系统自动补齐两名机器人，确保演示、录屏和面试场景不依赖临时凑人。',
    badge: 'Phase 2',
    cta: '机器人模式开发中',
  },
];

export function DouDizhuPlayStage() {
  return (
    <div className="space-y-6">
      <div className="grid gap-6 xl:grid-cols-[1.08fr_0.92fr]">
        <Card className="border-white/10 bg-white/[0.04] text-white">
          <CardContent className="p-6 md:p-7">
            <div className="mb-5 flex items-center gap-3">
              <Sparkles className="h-5 w-5 text-amber-300" />
              <div>
                <p className="text-xs uppercase tracking-[0.28em] text-slate-500">
                  Phase 0 Shell
                </p>
                <h2 className="mt-2 text-2xl font-black tracking-tight text-white">
                  斗地主已接入站内试玩入口
                </h2>
              </div>
            </div>

            <div className="space-y-4 text-sm leading-7 text-slate-300">
              <p>
                当前阶段先把斗地主正式接进现有网站的信息架构：你现在已经可以在
                Games Hub、详情页和试玩页里看到这款游戏，不再是文档里的孤立企划。
              </p>
              <p>
                下一阶段会把这块接成真正的三人斗地主服务端链路，包括规则引擎、房间状态机、
                WebSocket 实时同步、结果落库，以及单人可演示的人机模式。
              </p>
            </div>

            <div className="mt-6 grid gap-3 sm:grid-cols-3">
              <div className="rounded-2xl border border-white/10 bg-black/20 px-4 py-4">
                <div className="mb-2 flex items-center gap-2 text-sm text-slate-400">
                  <Layers3 className="h-4 w-4" />
                  当前阶段
                </div>
                <div className="text-lg font-semibold text-white">入口接入</div>
              </div>
              <div className="rounded-2xl border border-white/10 bg-black/20 px-4 py-4">
                <div className="mb-2 flex items-center gap-2 text-sm text-slate-400">
                  <Radio className="h-4 w-4" />
                  实时目标
                </div>
                <div className="text-lg font-semibold text-white">房间 + WS</div>
              </div>
              <div className="rounded-2xl border border-white/10 bg-black/20 px-4 py-4">
                <div className="mb-2 flex items-center gap-2 text-sm text-slate-400">
                  <Crown className="h-4 w-4" />
                  演示兜底
                </div>
                <div className="text-lg font-semibold text-white">1 人 + 2 机器人</div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="border-white/10 bg-white/[0.04] text-white">
          <CardContent className="p-6 md:p-7">
            <div className="mb-5 flex items-center gap-3">
              <Crown className="h-5 w-5 text-amber-300" />
              <div>
                <h2 className="text-2xl font-black tracking-tight text-white">
                  为什么先把人机模式列进计划
                </h2>
                <p className="mt-1 text-sm text-slate-400">
                  这是演示兜底能力，不是先做重 AI 项目。
                </p>
              </div>
            </div>

            <div className="space-y-3">
              {[
                '现场演示不一定总能凑齐 3 个真人玩家。',
                '录屏和自测需要稳定跑通一整局，而不是等人联调。',
                '机器人会走同一套服务端规则引擎，只做合法且节奏自然的基础策略。',
              ].map((item) => (
                <div
                  key={item}
                  className="rounded-2xl border border-white/10 bg-black/20 px-4 py-3 text-sm leading-6 text-slate-300"
                >
                  {item}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        {roomModes.map((mode) => {
          const Icon = mode.icon;

          return (
            <Card key={mode.title} className="border-white/10 bg-white/[0.04] text-white">
              <CardContent className="p-6">
                <div className="mb-5 flex items-start justify-between gap-4">
                  <div className="flex items-center gap-3">
                    <div className="rounded-2xl border border-white/10 bg-black/20 p-3">
                      <Icon className="h-5 w-5 text-amber-300" />
                    </div>
                    <div>
                      <p className="text-xs uppercase tracking-[0.28em] text-slate-500">
                        {mode.eyebrow}
                      </p>
                      <h3 className="mt-2 text-xl font-semibold text-white">{mode.title}</h3>
                    </div>
                  </div>
                  <Badge className="border-white/15 bg-white/8 text-white">{mode.badge}</Badge>
                </div>

                <p className="text-sm leading-7 text-slate-300">{mode.description}</p>

                <div className="mt-6">
                  <Button
                    disabled
                    className="border-white/10 bg-white/10 text-white hover:bg-white/10"
                  >
                    {mode.cta}
                  </Button>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      <Card className="border-white/10 bg-white/[0.04] text-white">
        <CardContent className="p-6 md:p-7">
          <div className="mb-5 flex items-center gap-3">
            <Sparkles className="h-5 w-5 text-sky-300" />
            <div>
              <h2 className="text-2xl font-black tracking-tight text-white">阶段路线</h2>
              <p className="mt-1 text-sm text-slate-400">
                按网站接入优先，不先为了抽象去重写现有 Hex Blitz 链路。
              </p>
            </div>
          </div>

          <div className="grid gap-3 lg:grid-cols-2">
            {implementationSteps.map((step) => (
              <div
                key={step}
                className="rounded-2xl border border-white/10 bg-black/20 px-4 py-3 text-sm leading-6 text-slate-300"
              >
                {step}
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
