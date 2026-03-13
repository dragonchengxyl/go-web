"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import {
  ArrowLeft,
  Calendar,
  MapPin,
  Globe,
  Users,
  CheckCircle2,
  Clock,
  Bookmark,
} from "lucide-react";
import { apiClient, Event, EventAttendee } from "@/lib/api-client";
import { Button } from "@/components/ui/button";

export default function EventDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [event, setEvent] = useState<Event | null>(null);
  const [attendees, setAttendees] = useState<EventAttendee[]>([]);
  const [loading, setLoading] = useState(true);
  const [attending, setAttending] = useState(false);
  const [hasAttended, setHasAttended] = useState(false);
  const [error, setError] = useState("");
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [bookmarked, setBookmarked] = useState(false);
  const [bookmarkLoading, setBookmarkLoading] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem("access_token");
    if (token) {
      apiClient.setToken(token);
      setIsLoggedIn(true);
      if (id) {
        apiClient
          .checkBookmark("event", id)
          .then((res) => setBookmarked(res.bookmarked))
          .catch(() => {});
      }
    }
    if (!id) return;
    Promise.all([apiClient.getEvent(id), apiClient.listEventAttendees(id)])
      .then(([ev, atRes]) => {
        setEvent(ev);
        const list = atRes.attendees ?? [];
        setAttendees(list);
        if (token) {
          apiClient
            .getMe()
            .then((me) => {
              setHasAttended(list.some((a) => a.user_id === me?.id));
            })
            .catch(() => {});
        }
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [id]);

  const handleAttend = async () => {
    if (!id) return;
    setAttending(true);
    setError("");
    try {
      await apiClient.attendEvent(id);
      const [updated, atRes] = await Promise.all([
        apiClient.getEvent(id),
        apiClient.listEventAttendees(id),
      ]);
      setEvent(updated);
      setAttendees(atRes.attendees ?? []);
      setHasAttended(true);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "报名失败");
    } finally {
      setAttending(false);
    }
  };

  const handleBookmark = async () => {
    if (!id) return;
    setBookmarkLoading(true);
    setError("");
    try {
      if (bookmarked) {
        await apiClient.unbookmarkEvent(id);
        setBookmarked(false);
      } else {
        await apiClient.bookmarkEvent(id);
        setBookmarked(true);
      }
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "收藏失败");
    } finally {
      setBookmarkLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto pt-20 px-4">
        <div className="h-8 w-32 bg-muted animate-pulse rounded mb-6" />
        <div className="h-48 bg-muted animate-pulse rounded-2xl mb-4" />
        <div className="h-24 bg-muted animate-pulse rounded-2xl" />
      </div>
    );
  }

  if (!event) {
    return (
      <div className="max-w-2xl mx-auto pt-20 px-4 text-center py-16 text-muted-foreground">
        <p className="text-xl font-medium mb-2">活动不存在</p>
        <p className="text-sm mb-6">该活动可能已被删除</p>
        <Link href="/events">
          <Button variant="outline">返回活动列表</Button>
        </Link>
      </div>
    );
  }

  const start = new Date(event.start_time);
  const end = new Date(event.end_time);
  const isPublished = event.status === "published";
  const isFull =
    event.max_capacity > 0 && event.attendee_count >= event.max_capacity;

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-12">
      {/* Back */}
      <Link href="/events">
        <Button variant="ghost" size="sm" className="mb-6 -ml-2">
          <ArrowLeft className="h-4 w-4 mr-1" />
          返回活动
        </Button>
      </Link>

      {/* Header card */}
      <div className="bg-card border rounded-2xl overflow-hidden mb-5">
        {/* Banner */}
        <div className="h-32 bg-gradient-to-br from-brand-purple/40 via-brand-teal/30 to-brand-coral/20 relative">
          <div className="absolute top-3 right-3">
            <span
              className={`text-xs px-2.5 py-1 rounded-full font-medium ${
                isPublished
                  ? "bg-green-500/20 text-green-500 border border-green-500/30"
                  : event.status === "cancelled"
                    ? "bg-red-500/20 text-red-500 border border-red-500/30"
                    : "bg-muted text-muted-foreground border border-border"
              }`}
            >
              {isPublished
                ? "报名中"
                : event.status === "cancelled"
                  ? "已取消"
                  : event.status === "completed"
                    ? "已结束"
                    : "草稿"}
            </span>
          </div>
        </div>

        <div className="p-6">
          <div className="flex items-start justify-between gap-3 mb-3">
            <h1 className="text-2xl font-bold">{event.title}</h1>
            {isLoggedIn && (
              <Button
                variant="outline"
                size="sm"
                onClick={handleBookmark}
                disabled={bookmarkLoading}
                className={
                  bookmarked ? "border-brand-purple text-brand-purple" : ""
                }
              >
                <Bookmark
                  className={`h-4 w-4 mr-1 ${bookmarked ? "fill-current" : ""}`}
                />
                {bookmarked ? "已收藏" : "收藏"}
              </Button>
            )}
          </div>

          {event.tags?.length > 0 && (
            <div className="flex flex-wrap gap-1.5 mb-4">
              {event.tags.map((t) => (
                <Link key={t} href={`/tags/${encodeURIComponent(t)}`}>
                  <span className="text-xs px-2 py-0.5 rounded-full bg-primary/10 text-primary hover:bg-primary/20 transition-colors cursor-pointer">
                    #{t}
                  </span>
                </Link>
              ))}
            </div>
          )}

          {/* Info */}
          <div className="space-y-2.5 text-sm text-muted-foreground">
            <div className="flex items-center gap-2.5">
              <Calendar className="h-4 w-4 text-primary flex-shrink-0" />
              <span>
                {start.toLocaleDateString("zh-CN", {
                  year: "numeric",
                  month: "long",
                  day: "numeric",
                })}{" "}
                {start.toLocaleTimeString("zh-CN", {
                  hour: "2-digit",
                  minute: "2-digit",
                })}
                {" — "}
                {end.toLocaleTimeString("zh-CN", {
                  hour: "2-digit",
                  minute: "2-digit",
                })}
              </span>
            </div>
            {event.is_online ? (
              <div className="flex items-center gap-2.5">
                <Globe className="h-4 w-4 text-primary flex-shrink-0" />
                <span>线上活动</span>
              </div>
            ) : event.location ? (
              <div className="flex items-center gap-2.5">
                <MapPin className="h-4 w-4 text-primary flex-shrink-0" />
                <span>{event.location}</span>
              </div>
            ) : null}
            <div className="flex items-center gap-2.5">
              <Users className="h-4 w-4 text-primary flex-shrink-0" />
              <span>
                {event.attendee_count} 人参与
                {event.max_capacity > 0 && ` / 上限 ${event.max_capacity} 人`}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Description */}
      {event.description && (
        <div className="bg-card border rounded-2xl p-6 mb-5">
          <h2 className="font-semibold mb-3">活动详情</h2>
          <p className="text-sm text-muted-foreground whitespace-pre-wrap leading-relaxed">
            {event.description}
          </p>
        </div>
      )}

      {/* Attend button */}
      {isPublished && (
        <div className="mb-6">
          {error && <p className="text-destructive text-sm mb-2">{error}</p>}
          {hasAttended ? (
            <div className="flex items-center gap-2 p-4 rounded-xl bg-green-500/10 border border-green-500/20 text-green-600 dark:text-green-400">
              <CheckCircle2 className="h-5 w-5 flex-shrink-0" />
              <p className="text-sm font-medium">你已报名参加此活动</p>
            </div>
          ) : isFull ? (
            <Button disabled className="w-full" size="lg">
              <Users className="h-4 w-4 mr-2" />
              名额已满
            </Button>
          ) : isLoggedIn ? (
            <Button
              onClick={handleAttend}
              disabled={attending}
              size="lg"
              className="w-full bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110"
            >
              <Calendar className="h-4 w-4 mr-2" />
              {attending ? "提交中…" : "我要参加"}
            </Button>
          ) : (
            <Link href="/login">
              <Button
                size="lg"
                className="w-full bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110"
              >
                登录后报名
              </Button>
            </Link>
          )}
        </div>
      )}

      {/* Attendees */}
      {attendees.length > 0 && (
        <div className="bg-card border rounded-2xl p-6">
          <h2 className="font-semibold mb-4">参与者（{attendees.length}）</h2>
          <div className="flex flex-wrap gap-2">
            {attendees.map((a) => (
              <Link key={a.user_id} href={`/users/${a.user_id}`}>
                <div
                  className="w-10 h-10 rounded-full bg-gradient-to-br from-brand-purple to-brand-teal flex items-center justify-center text-xs text-white font-bold hover:brightness-110 transition-all"
                  title={a.user_id}
                >
                  {a.user_id.slice(0, 2).toUpperCase()}
                </div>
              </Link>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
