'use client';

import { Bell } from 'lucide-react';

export default function NotificationsPage() {
  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-8">
      <h1 className="text-2xl font-bold mb-6 flex items-center gap-2">
        <Bell className="h-6 w-6" />
        通知
      </h1>
      <div className="text-center py-16 text-muted-foreground">
        <Bell className="h-12 w-12 mx-auto mb-4 opacity-30" />
        <p>暂无通知</p>
      </div>
    </div>
  );
}
