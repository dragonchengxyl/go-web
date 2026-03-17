"use client";

import { usePathname } from "next/navigation";
import { Header } from "@/components/layout/header";
import { MusicPlayer } from "@/components/music-player";
import { ModerationToast } from "@/components/moderation-toast";
import { VerifyEmailBanner } from "@/components/verify-email-banner";
import { FurryAssistant } from "@/components/assistant/furry-assistant";

export function AppChrome({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const isAdminRoute = pathname.startsWith("/admin");

  if (isAdminRoute) {
    return <>{children}</>;
  }

  return (
    <>
      <Header />
      <VerifyEmailBanner />
      {children}
      <MusicPlayer />
      <ModerationToast />
      <FurryAssistant />
    </>
  );
}
