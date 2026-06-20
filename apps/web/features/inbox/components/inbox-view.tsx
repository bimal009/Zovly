"use client";

import { useState } from "react";
import { useConversations, useMessages } from "../client/inbox";
import { Conversation, Message, Platform } from "../types/inbox";
import { IconSearch, IconInbox, IconRobot, IconUser } from "@tabler/icons-react";

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
      className={`${dim} rounded-full bg-gradient-to-br from-violet-500 to-blue-500 flex items-center justify-center font-semibold text-white flex-shrink-0`}
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
  const name = conv.contact_name ?? conv.contact_username ?? conv.contact_id;

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

// ─── message bubble ───────────────────────────────────────────────────────────

function MessageBubble({ msg }: { msg: Message }) {
  const isInbound = msg.direction === "in";

  return (
    <div className={`flex ${isInbound ? "justify-start" : "justify-end"} mb-2`}>
      <div className={`max-w-[70%] flex flex-col ${isInbound ? "items-start" : "items-end"} gap-1`}>
        {msg.sent_by && (
          <div className="flex items-center gap-1 px-1">
            {msg.sent_by === "ai" ? (
              <>
                <IconRobot size={11} className="text-violet-500" />
                <span className="text-[10px] text-violet-500 font-medium">AI</span>
              </>
            ) : (
              <>
                <IconUser size={11} className="text-blue-500" />
                <span className="text-[10px] text-blue-500 font-medium">You</span>
              </>
            )}
          </div>
        )}

        <div
          className={`px-3.5 py-2.5 rounded-2xl text-sm leading-relaxed ${
            isInbound
              ? "bg-muted text-foreground rounded-tl-sm"
              : "bg-primary text-primary-foreground rounded-tr-sm"
          }`}
        >
          {msg.content ?? <span className="italic text-xs opacity-70">[media]</span>}
        </div>

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
  const name = conversation.contact_name ?? conversation.contact_username ?? conversation.contact_id;

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
          {conversation.contact_username && (
            <p className="text-xs text-muted-foreground">@{conversation.contact_username}</p>
          )}
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
