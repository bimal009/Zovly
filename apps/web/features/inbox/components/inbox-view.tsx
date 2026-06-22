"use client";

import { useState } from "react";
import { useConversations, useMessages } from "../client/inbox";
import { Conversation, Message, Platform } from "../types/inbox";
import {
  IconSearch,
  IconInbox,
  IconRobot,
  IconUser,
  IconFileText,
  IconDownload,
  IconVideo,
} from "@tabler/icons-react";

// ─── platform assets ──────────────────────────────────────────────────────────

function FacebookLogo({ size = 16 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none">
      <rect width="24" height="24" rx="6" fill="#1877F2" />
      <path
        d="M16.5 12H14V10.5C14 9.67 14.67 9.5 15 9.5H16.5V7H14.5C12.57 7 11.5 8.43 11.5 10V12H10V14.5H11.5V21H14V14.5H16L16.5 12Z"
        fill="white"
      />
    </svg>
  );
}

function InstagramLogo({ size = 16 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none">
      <defs>
        <linearGradient id="ig-grad" x1="0%" y1="100%" x2="100%" y2="0%">
          <stop offset="0%" stopColor="#FFDC80" />
          <stop offset="25%" stopColor="#FCAF45" />
          <stop offset="50%" stopColor="#F77737" />
          <stop offset="75%" stopColor="#C13584" />
          <stop offset="100%" stopColor="#833AB4" />
        </linearGradient>
      </defs>
      <rect width="24" height="24" rx="6" fill="url(#ig-grad)" />
      <circle cx="12" cy="12" r="3.5" stroke="white" strokeWidth="1.8" />
      <circle cx="17" cy="7" r="1.2" fill="white" />
      <rect
        x="4"
        y="4"
        width="16"
        height="16"
        rx="5"
        stroke="white"
        strokeWidth="1.8"
        fill="none"
      />
    </svg>
  );
}

function PlatformBadge({ platform }: { platform: Platform }) {
  return platform === "facebook" ? <FacebookLogo size={18} /> : <InstagramLogo size={18} />;
}

// ─── helpers ──────────────────────────────────────────────────────────────────

function formatRelativeTime(iso: string | null): string {
  if (!iso) return "";
  const diff = Date.now() - new Date(iso).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return "now";
  if (mins < 60) return `${mins}m`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h`;
  const days = Math.floor(hrs / 24);
  if (days < 7) return `${days}d`;
  return new Date(iso).toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString(undefined, { hour: "2-digit", minute: "2-digit" });
}

// Instagram/Facebook often withhold profile name + username (e.g. apps without
// Advanced Access), so those fields arrive as empty strings — `??` won't catch
// those. Treat blank/whitespace as missing and degrade to a platform label
// rather than showing a raw numeric contact id or an empty name.
function contactDisplayName(c: Conversation): string {
  const name = c.contact_name?.trim();
  if (name) return name;
  const username = c.contact_username?.trim();
  if (username) return username;
  return c.platform === "instagram" ? "Instagram user" : "Facebook user";
}

function Avatar({
  name,
  avatarUrl,
  size = "md",
}: {
  name: string | null;
  avatarUrl: string | null;
  size?: "sm" | "md";
}) {
  const initials = name
    ? name
        .split(" ")
        .map((w) => w[0])
        .join("")
        .slice(0, 2)
        .toUpperCase()
    : "?";
  const dim = size === "sm" ? "w-8 h-8 text-xs" : "w-10 h-10 text-sm";

  if (avatarUrl) {
    return (
      <img
        src={avatarUrl}
        alt={name ?? "contact"}
        className={`${dim} rounded-full object-cover flex-shrink-0`}
      />
    );
  }

  return (
    <div
      className={`${dim} bg-primary text-primary-foreground flex flex-shrink-0 items-center justify-center rounded-full font-semibold`}
    >
      {initials}
    </div>
  );
}

// ─── conversation list item ────────────────────────────────────────────────────

function ConversationItem({
  conv,
  selected,
  onClick,
}: {
  conv: Conversation;
  selected: boolean;
  onClick: () => void;
}) {
  const name = contactDisplayName(conv);

  return (
    <button
      onClick={onClick}
      className={`w-full flex items-start gap-3 px-4 py-3.5 text-left cursor-pointer transition-colors duration-150 border-b border-border/50 hover:bg-muted/60 ${
        selected ? "bg-muted border-l-2 border-l-primary" : ""
      }`}
    >
      <div className="relative flex-shrink-0">
        <Avatar name={conv.contact_name} avatarUrl={conv.contact_avatar_url} />
        <span className="absolute -bottom-0.5 -right-0.5">
          <PlatformBadge platform={conv.platform} />
        </span>
      </div>

      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between gap-1">
          <span className="font-medium text-sm text-foreground truncate">{name}</span>
          <span className="text-[11px] text-muted-foreground flex-shrink-0 whitespace-nowrap">
            {formatRelativeTime(conv.last_message_at)}
          </span>
        </div>
        <p className="text-xs text-muted-foreground truncate mt-0.5 capitalize">{conv.platform}</p>
      </div>
    </button>
  );
}

// ─── media content ──────────────────────────────────────────────────────────────

function fileNameFromUrl(url: string): string {
  try {
    const pathname = new URL(url).pathname;
    const last = pathname.split("/").filter(Boolean).pop();
    return last ? decodeURIComponent(last) : "Attachment";
  } catch {
    return "Attachment";
  }
}

function MediaContent({ msg }: { msg: Message }) {
  const url = msg.media_url;
  if (!url) return null;

  switch (msg.media_type) {
    case "image":
      return (
        <a
          href={url}
          target="_blank"
          rel="noopener noreferrer"
          className="block cursor-pointer overflow-hidden rounded-xl"
          aria-label="Open image in new tab"
        >
          <img
            src={url}
            alt={msg.content ?? "Image attachment"}
            loading="lazy"
            className="max-h-72 w-full max-w-xs rounded-xl object-cover transition-opacity duration-200 hover:opacity-90"
          />
        </a>
      );

    case "video":
      return (
        <video
          src={url}
          controls
          preload="metadata"
          className="max-h-72 w-full max-w-xs rounded-xl bg-black"
        >
          <a href={url} target="_blank" rel="noopener noreferrer">
            <IconVideo size={14} /> View video
          </a>
        </video>
      );

    case "audio":
      return (
        <audio src={url} controls preload="metadata" className="h-10 w-60 max-w-full">
          <a href={url} target="_blank" rel="noopener noreferrer">
            Play voice message
          </a>
        </audio>
      );

    case "document":
    default:
      return (
        <a
          href={url}
          target="_blank"
          rel="noopener noreferrer"
          className="flex max-w-xs cursor-pointer items-center gap-2.5 rounded-xl border border-border bg-background/60 px-3 py-2.5 transition-colors duration-200 hover:bg-muted"
          aria-label={`Download ${fileNameFromUrl(url)}`}
        >
          <span className="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-muted">
            <IconFileText size={18} className="text-muted-foreground" />
          </span>
          <span className="min-w-0 flex-1">
            <span className="block truncate text-xs font-medium text-foreground">
              {fileNameFromUrl(url)}
            </span>
            <span className="block text-[10px] text-muted-foreground capitalize">
              {msg.media_type ?? "file"}
            </span>
          </span>
          <IconDownload size={15} className="flex-shrink-0 text-muted-foreground" />
        </a>
      );
  }
}

// ─── message bubble ───────────────────────────────────────────────────────────

function MessageBubble({ msg }: { msg: Message }) {
  const isInbound = msg.direction === "in";
  const hasMedia = !!msg.media_url;
  const hasText = !!msg.content;
  // Voice/audio renders as a bare player; other media keep a bubble background.
  const bareMedia = hasMedia && !hasText && (msg.media_type === "image" || msg.media_type === "audio" || msg.media_type === "video");

  return (
    <div className={`flex ${isInbound ? "justify-start" : "justify-end"} mb-2`}>
      <div className={`max-w-[70%] flex flex-col ${isInbound ? "items-start" : "items-end"} gap-1`}>
        {msg.sent_by && (
          <div className="flex items-center gap-1 px-1">
            {msg.sent_by === "ai" ? (
              <>
                <IconRobot size={11} className="text-primary" />
                <span className="text-primary text-[10px] font-medium">AI</span>
              </>
            ) : (
              <>
                <IconUser size={11} className="text-muted-foreground" />
                <span className="text-muted-foreground text-[10px] font-medium">
                  You
                </span>
              </>
            )}
          </div>
        )}

        {bareMedia ? (
          <MediaContent msg={msg} />
        ) : (
          <div
            className={`flex flex-col gap-2 px-3.5 py-2.5 rounded-2xl text-sm leading-relaxed ${
              isInbound
                ? "bg-muted text-foreground rounded-tl-sm"
                : "bg-primary text-primary-foreground rounded-tr-sm"
            }`}
          >
            {hasMedia && <MediaContent msg={msg} />}
            {hasText ? (
              <span className="whitespace-pre-wrap break-words">{msg.content}</span>
            ) : !hasMedia ? (
              <span className="italic text-xs opacity-70">[media]</span>
            ) : null}
          </div>
        )}

        <div className="flex items-center gap-1.5 px-1">
          <span className="text-[10px] text-muted-foreground">{formatTime(msg.sent_at)}</span>
          {msg.status === "failed" && (
            <span className="text-[10px] text-destructive font-medium">Failed</span>
          )}
        </div>
      </div>
    </div>
  );
}

// ─── thread panel ─────────────────────────────────────────────────────────────

function ThreadPanel({ conversation }: { conversation: Conversation }) {
  const { data, isLoading } = useMessages(conversation.id);
  const messages = data?.data ?? [];
  const name = contactDisplayName(conversation);
  const username = conversation.contact_username?.trim();

  return (
    <div className="flex flex-col h-full">
      {/* Thread header */}
      <div className="flex items-center gap-3 px-5 py-3.5 border-b border-border bg-background/80 backdrop-blur-sm flex-shrink-0">
        <div className="relative flex-shrink-0">
          <Avatar name={conversation.contact_name} avatarUrl={conversation.contact_avatar_url} size="sm" />
          <span className="absolute -bottom-0.5 -right-0.5">
            <PlatformBadge platform={conversation.platform} />
          </span>
        </div>
        <div>
          <p className="font-semibold text-sm text-foreground">{name}</p>
          {username && <p className="text-xs text-muted-foreground">@{username}</p>}
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto px-5 py-4 space-y-1">
        {isLoading ? (
          <div className="flex flex-col gap-3 pt-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className={`flex ${i % 2 === 0 ? "justify-end" : "justify-start"}`}>
                <div className="h-9 w-48 rounded-2xl bg-muted animate-pulse" />
              </div>
            ))}
          </div>
        ) : messages.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center py-16">
            <IconInbox size={36} className="text-muted-foreground/40 mb-3" />
            <p className="text-sm text-muted-foreground">No messages yet</p>
          </div>
        ) : (
          messages.map((msg) => <MessageBubble key={msg.id} msg={msg} />)
        )}
      </div>
    </div>
  );
}

// ─── main inbox view ──────────────────────────────────────────────────────────

export function InboxView() {
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const { data, isLoading } = useConversations();

  const conversations = data?.data ?? [];
  const filtered = conversations.filter((c) => {
    const q = search.toLowerCase();
    return (
      !q ||
      c.contact_name?.toLowerCase().includes(q) ||
      c.contact_username?.toLowerCase().includes(q) ||
      c.platform.includes(q)
    );
  });

  const selectedConv = conversations.find((c) => c.id === selectedId) ?? null;

  return (
    <div className="flex h-[calc(100vh-var(--header-height,48px))] overflow-hidden border-t border-border">
      {/* ── Left: conversation list ── */}
      <div className="w-80 flex-shrink-0 flex flex-col border-r border-border bg-background">
        {/* Search */}
        <div className="px-3 py-3 border-b border-border">
          <div className="flex items-center gap-2 bg-muted rounded-lg px-3 py-2">
            <IconSearch size={15} className="text-muted-foreground flex-shrink-0" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search conversations..."
              className="flex-1 bg-transparent text-sm outline-none placeholder:text-muted-foreground text-foreground min-w-0"
            />
          </div>
        </div>

        {/* List */}
        <div className="flex-1 overflow-y-auto">
          {isLoading ? (
            <div className="flex flex-col gap-0">
              {[1, 2, 3, 4].map((i) => (
                <div key={i} className="flex items-center gap-3 px-4 py-3.5 border-b border-border/50">
                  <div className="w-10 h-10 rounded-full bg-muted animate-pulse flex-shrink-0" />
                  <div className="flex-1 space-y-1.5">
                    <div className="h-3.5 w-28 bg-muted animate-pulse rounded" />
                    <div className="h-3 w-20 bg-muted animate-pulse rounded" />
                  </div>
                </div>
              ))}
            </div>
          ) : filtered.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center px-4">
              <IconInbox size={32} className="text-muted-foreground/40 mb-2" />
              <p className="text-sm text-muted-foreground">
                {search ? "No matching conversations" : "No conversations yet"}
              </p>
            </div>
          ) : (
            filtered.map((conv) => (
              <ConversationItem
                key={conv.id}
                conv={conv}
                selected={conv.id === selectedId}
                onClick={() => setSelectedId(conv.id)}
              />
            ))
          )}
        </div>

        {/* Footer count */}
        {!isLoading && conversations.length > 0 && (
          <div className="px-4 py-2 border-t border-border">
            <p className="text-xs text-muted-foreground">
              {conversations.length} conversation{conversations.length !== 1 ? "s" : ""}
            </p>
          </div>
        )}
      </div>

      {/* ── Right: thread ── */}
      <div className="flex-1 overflow-hidden bg-background/50">
        {selectedConv ? (
          <ThreadPanel conversation={selectedConv} />
        ) : (
          <div className="flex flex-col items-center justify-center h-full text-center px-8">
            <div className="w-16 h-16 rounded-2xl bg-muted flex items-center justify-center mb-4">
              <IconInbox size={28} className="text-muted-foreground/60" />
            </div>
            <h3 className="font-semibold text-foreground mb-1">Select a conversation</h3>
            <p className="text-sm text-muted-foreground max-w-xs">
              Choose a conversation from the list to view messages from Facebook or Instagram.
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
