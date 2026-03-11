import Link from 'next/link';

export default function PrivacyPage() {
  return (
    <div className="max-w-3xl mx-auto pt-24 px-4 pb-16">
      <div className="mb-8">
        <Link href="/" className="text-sm text-muted-foreground hover:text-foreground transition-colors">
          ← 返回首页
        </Link>
      </div>

      <h1 className="text-3xl font-bold mb-2">隐私政策</h1>
      <p className="text-muted-foreground text-sm mb-10">最后更新：2026 年 3 月</p>

      <div className="prose prose-neutral dark:prose-invert max-w-none space-y-8 text-sm leading-7">
        <section>
          <h2 className="text-lg font-semibold mb-3">1. 我们收集的信息</h2>
          <p>我们收集以下类型的信息：</p>
          <ul className="list-disc pl-6 space-y-1 mt-2">
            <li><strong>账号信息</strong>：用户名、邮箱地址、加密后的密码</li>
            <li><strong>个人资料</strong>：您主动填写的兽名、物种、简介、头像等</li>
            <li><strong>内容数据</strong>：您发布的帖子、评论、消息等内容</li>
            <li><strong>使用数据</strong>：访问日志、IP 地址、浏览器信息（用于安全和服务改进）</li>
          </ul>
        </section>

        <section>
          <h2 className="text-lg font-semibold mb-3">2. 信息使用方式</h2>
          <p>我们使用收集的信息用于：</p>
          <ul className="list-disc pl-6 space-y-1 mt-2">
            <li>提供、维护和改进平台服务</li>
            <li>账号认证和安全防护</li>
            <li>发送与服务相关的通知</li>
            <li>内容审核与社区安全</li>
            <li>分析平台使用情况以优化用户体验</li>
          </ul>
        </section>

        <section>
          <h2 className="text-lg font-semibold mb-3">3. 信息共享</h2>
          <p>
            我们不会出售您的个人信息。仅在以下情形下共享信息：（a）经您明确同意；
            （b）履行法律义务；（c）保护平台或用户安全；（d）与受保密协议约束的服务提供商合作。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-semibold mb-3">4. 数据存储与安全</h2>
          <p>
            您的数据存储于安全的云服务器（PostgreSQL 数据库，加密传输）。
            密码使用 Argon2id 算法进行哈希处理，不以明文存储。
            我们定期进行安全评估，但无法保证 100% 安全。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-semibold mb-3">5. Cookie 与本地存储</h2>
          <p>
            本平台使用 Cookie 和浏览器本地存储（localStorage）来保存登录状态和偏好设置。
            您可以通过浏览器设置管理 Cookie，但这可能影响部分功能的正常使用。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-semibold mb-3">6. 您的权利</h2>
          <p>您有权：</p>
          <ul className="list-disc pl-6 space-y-1 mt-2">
            <li>查看和修改您的个人信息（通过个人资料页）</li>
            <li>删除您发布的内容</li>
            <li>申请注销账号及删除账号相关数据</li>
            <li>对数据处理提出异议</li>
          </ul>
        </section>

        <section>
          <h2 className="text-lg font-semibold mb-3">7. 未成年人保护</h2>
          <p>
            本平台不面向 16 周岁以下人士。如发现我们无意收集了未成年人信息，
            请联系我们，我们将立即删除。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-semibold mb-3">8. 政策更新</h2>
          <p>
            我们可能不定期更新本隐私政策。重大变更将通过平台通知或邮件告知。
            继续使用平台即视为您接受更新后的政策。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-semibold mb-3">9. 联系我们</h2>
          <p>
            如对本隐私政策有任何疑问，请通过平台内的意见反馈渠道联系我们。
          </p>
        </section>
      </div>
    </div>
  );
}
