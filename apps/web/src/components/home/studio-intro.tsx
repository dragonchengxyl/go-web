export function StudioIntro() {
  return (
    <section className="py-20">
      <div className="container mx-auto px-4">
        <div className="max-w-3xl mx-auto text-center">
          <h2 className="text-4xl font-bold mb-6">关于工作室</h2>
          <p className="text-lg text-muted-foreground mb-8">
            我们是一支充满热情的独立游戏开发团队，致力于创造独特而富有创意的游戏体验。
            每一款作品都倾注了我们的热情与匠心，希望能为玩家带来难忘的游戏时光。
          </p>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-8 mt-12">
            <div>
              <div className="text-4xl font-bold text-primary mb-2">5+</div>
              <div className="text-sm text-muted-foreground">游戏作品</div>
            </div>
            <div>
              <div className="text-4xl font-bold text-primary mb-2">10K+</div>
              <div className="text-sm text-muted-foreground">玩家</div>
            </div>
            <div>
              <div className="text-4xl font-bold text-primary mb-2">20+</div>
              <div className="text-sm text-muted-foreground">音乐专辑</div>
            </div>
            <div>
              <div className="text-4xl font-bold text-primary mb-2">100%</div>
              <div className="text-sm text-muted-foreground">用心制作</div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
