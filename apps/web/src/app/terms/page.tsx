import Link from 'next/link';

export default function TermsPage() {
  return (
    <div className="max-w-3xl mx-auto pt-24 px-4 pb-16">
      <div className="mb-8">
        <Link href="/" className="text-sm text-muted-foreground hover:text-foreground transition-colors">
          ← 返回首页
        </Link>
      </div>

      <h1 className="text-3xl font-bold mb-2">服务条款</h1>
      <p className="text-muted-foreground text-sm mb-10">最后更新：2026 年 3 月</p>

      <div className="prose prose-neutral dark:prose-invert max-w-none space-y-8 text-sm leading-7">
        <section>
          <h2 className="text-lg font-semibold mb-3">1. 接受条款</h2>
          <p>
            欢迎使用 Furry 同好社区（以下简称「本平台」）。在注册账号或使用本平台的任何服务之前，
            请仔细阅读本服务条款。通过注册或使用本平台，即表示您同意受本条款约束。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-semibold mb-3">2. 用户资格</h2>
          <p>
            使用本平台，您必须：（a）年满 16 周岁；（b）具有完全民事行为能力；
            （c）提供真实、准确、完整的注册信息，并在信息变更时及时更新。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-semibold mb-3">3. 账号安全</h2>
          <p>
            您有责任维护账号和密码的保密性。您同意对账号下发生的所有活动负责。
            如发现账号遭到未授权使用，请立即通知我们。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-semibold mb-3">4. 内容规范</h2>
          <p>您承诺不发布以下内容：</p>
          <ul className="list-disc pl-6 space-y-1 mt-2">
            <li>违反中华人民共和国法律法规的内容</li>
            <li>侵犯他人知识产权、隐私权或其他合法权益的内容</li>
            <li>骚扰、威胁、诽谤或歧视性内容</li>
            <li>未经授权的商业广告或垃圾信息</li>
            <li>未加适当提示的成人内容</li>
          </ul>
        </section>

        <section>
          <h2 className="text-lg font-semibold mb-3">5. 知识产权</h2>
          <p>
            您对自己发布的原创内容保留知识产权。通过发布内容，您授予本平台非独家、
            全球范围内的免费许可，用于在平台内展示、传播和推广该内容。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-semibold mb-3">6. 服务变更与终止</h2>
          <p>
            我们保留随时修改、暂停或终止部分或全部服务的权利，并将提前通知。
            对于违反本条款的用户，我们有权限制或终止其账号，恕不另行通知。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-semibold mb-3">7. 免责声明</h2>
          <p>
            本平台以「现状」提供服务。我们不对平台的持续可用性、数据准确性或用户内容承担责任。
            在法律允许的最大范围内，本平台不承担任何间接损失。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-semibold mb-3">8. 联系方式</h2>
          <p>
            如对本条款有任何疑问，请通过平台内的意见反馈渠道联系我们。
          </p>
        </section>
      </div>
    </div>
  );
}
