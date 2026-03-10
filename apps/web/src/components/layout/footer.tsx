import Link from 'next/link';
import { Github, Twitter, Mail } from 'lucide-react';

export function Footer() {
  return (
    <footer className="bg-muted/50 border-t">
      <div className="container mx-auto px-4 py-12">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
          {/* About */}
          <div>
            <h3 className="font-bold text-lg mb-4">关于社区</h3>
            <p className="text-sm text-muted-foreground">
              一个温暖而包容的 Furry 同好社区，分享你的故事与创作
            </p>
          </div>

          {/* Links */}
          <div>
            <h3 className="font-bold text-lg mb-4">快速链接</h3>
            <ul className="space-y-2">
              <li>
                <Link href="/feed" className="text-sm text-muted-foreground hover:text-primary">
                  动态
                </Link>
              </li>
              <li>
                <Link href="/explore" className="text-sm text-muted-foreground hover:text-primary">
                  发现
                </Link>
              </li>
              <li>
                <Link href="/notifications" className="text-sm text-muted-foreground hover:text-primary">
                  通知
                </Link>
              </li>
              <li>
                <Link href="/sponsor" className="text-sm text-muted-foreground hover:text-primary">
                  赞助支持
                </Link>
              </li>
            </ul>
          </div>

          {/* Account */}
          <div>
            <h3 className="font-bold text-lg mb-4">账号</h3>
            <ul className="space-y-2">
              <li>
                <Link href="/settings" className="text-sm text-muted-foreground hover:text-primary">
                  设置
                </Link>
              </li>
              <li>
                <Link href="/privacy" className="text-sm text-muted-foreground hover:text-primary">
                  隐私政策
                </Link>
              </li>
              <li>
                <Link href="/terms" className="text-sm text-muted-foreground hover:text-primary">
                  服务条款
                </Link>
              </li>
            </ul>
          </div>

          {/* Social */}
          <div>
            <h3 className="font-bold text-lg mb-4">关注我们</h3>
            <div className="flex space-x-4">
              <a
                href="https://github.com"
                target="_blank"
                rel="noopener noreferrer"
                className="text-muted-foreground hover:text-primary"
              >
                <Github className="h-5 w-5" />
              </a>
              <a
                href="https://twitter.com"
                target="_blank"
                rel="noopener noreferrer"
                className="text-muted-foreground hover:text-primary"
              >
                <Twitter className="h-5 w-5" />
              </a>
              <a
                href="mailto:contact@furry.community"
                className="text-muted-foreground hover:text-primary"
              >
                <Mail className="h-5 w-5" />
              </a>
            </div>
          </div>
        </div>

        <div className="mt-8 pt-8 border-t text-center text-sm text-muted-foreground">
          <p>&copy; 2026 Furry 同好社区. All rights reserved.</p>
        </div>
      </div>
    </footer>
  );
}
