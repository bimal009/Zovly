"use client";

import React, { useCallback, useEffect, useRef, useState } from "react";
import {
  X,
  Upload,
  ImagePlus,
  Camera,
  Loader2,
  AlertCircle,
  RotateCcw,
} from "lucide-react";
import { cn } from "@repo/ui/utils";
import { fetchIKAuth } from "@/features/images/api/image";

// ── Types ──────────────────────────────────────────────────────────────────────

type UploadStatus = "uploading" | "done" | "error";

interface UploadEntry {
  id: string;
  file: File;
  preview: string;
  status: UploadStatus;
  url?: string;
  error?: string;
}

export interface ImageUploaderProps {
  /** 1 = compact logo/avatar mode. 0 = unlimited gallery. >1 = capped gallery. */
  maxImages?: number;
  maxFileSizeBytes?: number;
  /** ImageKit folder path, e.g. "/logos" */
  folder?: string;
  accept?: string[];
  /** Called with the CDN URL each time an image finishes uploading */
  onUpload?: (url: string) => void;
  /** Called with the full list of done CDN URLs whenever it changes */
  onUploadsChange?: (urls: string[]) => void;
  className?: string;
  /** Label shown above the uploader */
  label?: string;
}

// ── Helpers ───────────────────────────────────────────────────────────────────

const formatSize = (bytes: number) => {
  if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
};

const genId = () => `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;

async function uploadToImageKit(file: File, folder: string): Promise<string> {
  const auth = await fetchIKAuth();
  const publicKey = process.env.NEXT_PUBLIC_IMAGEKIT_PUBLIC_KEY ?? "";

  const body = new FormData();
  body.append("file", file);
  body.append("fileName", file.name);
  body.append("publicKey", publicKey);
  body.append("signature", auth.signature);
  body.append("expire", String(auth.expire));
  body.append("token", auth.token);
  body.append("folder", folder);

  const res = await fetch("https://upload.imagekit.io/api/v1/files/upload", {
    method: "POST",
    body,
  });

  if (!res.ok) {
    const msg = await res.text().catch(() => res.statusText);
    throw new Error(msg || "Upload failed");
  }

  const json = await res.json();
  return json.url as string;
}

// ── Component ─────────────────────────────────────────────────────────────────

export const ImageUploader = React.forwardRef<
  HTMLDivElement,
  ImageUploaderProps
>(
  (
    {
      maxImages = 0,
      maxFileSizeBytes = 5 * 1024 * 1024,
      folder = "/uploads",
      accept = ["image/jpeg", "image/png", "image/webp"],
      onUpload,
      onUploadsChange,
      className,
      label,
    },
    ref,
  ) => {
    const [entries, setEntries] = useState<UploadEntry[]>([]);
    const [isDragging, setIsDragging] = useState(false);
    const [validationError, setValidationError] = useState<string | null>(null);
    const inputRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
      const urls = entries.filter((e) => e.status === "done").map((e) => e.url!);
      onUploadsChange?.(urls);
    }, [entries, onUploadsChange]);

    const isLogoMode = maxImages === 1;
    const isAtLimit = maxImages > 0 && entries.length >= maxImages;
    const acceptStr = accept.join(",");
    const acceptLabel = accept
      .map((a) => a.split("/")[1]?.toUpperCase())
      .filter(Boolean)
      .join(", ");

    const startUpload = useCallback(
      (id: string, file: File) => {
        uploadToImageKit(file, folder)
          .then((url) => {
            setEntries((prev) =>
              prev.map((e) =>
                e.id === id ? { ...e, status: "done" as const, url } : e,
              ),
            );
            onUpload?.(url);
          })
          .catch((err: Error) => {
            setEntries((prev) =>
              prev.map((e) =>
                e.id === id
                  ? { ...e, status: "error" as const, error: err.message }
                  : e,
              ),
            );
          });
      },
      [folder, onUpload],
    );

    const handleFiles = useCallback(
      (files: FileList | null) => {
        if (!files || files.length === 0) return;
        setValidationError(null);

        for (const file of Array.from(files)) {
          if (!accept.includes(file.type)) {
            setValidationError(`Unsupported format. Use ${acceptLabel}`);
            return;
          }
          if (maxFileSizeBytes > 0 && file.size > maxFileSizeBytes) {
            setValidationError(
              `File too large. Max ${formatSize(maxFileSizeBytes)}`,
            );
            return;
          }
          if (maxImages > 0 && entries.length >= maxImages) {
            setValidationError(
              `Maximum ${maxImages} ${maxImages === 1 ? "image" : "images"} allowed`,
            );
            return;
          }

          const id = genId();
          const preview = URL.createObjectURL(file);
          const entry: UploadEntry = { id, file, preview, status: "uploading" };

          setEntries((prev) => (isLogoMode ? [entry] : [...prev, entry]));
          startUpload(id, file);
        }

        if (inputRef.current) inputRef.current.value = "";
      },
      [
        accept,
        acceptLabel,
        maxFileSizeBytes,
        maxImages,
        entries.length,
        isLogoMode,
        startUpload,
      ],
    );

    const retryUpload = useCallback(
      (entry: UploadEntry) => {
        setEntries((prev) =>
          prev.map((e) =>
            e.id === entry.id
              ? { ...e, status: "uploading", error: undefined }
              : e,
          ),
        );
        startUpload(entry.id, entry.file);
      },
      [startUpload],
    );

    const removeEntry = useCallback(
      (id: string) => {
        setEntries((prev) => {
          const next = prev.filter((e) => e.id !== id);
          onUploadsChange?.(
            next.filter((e) => e.status === "done").map((e) => e.url!),
          );
          return next;
        });
        setValidationError(null);
      },
      [onUploadsChange],
    );

    // ── Logo / single-image mode ───────────────────────────────────────────────
    if (isLogoMode) {
      const entry = entries[0];

      return (
        <div
          ref={ref}
          className={cn("inline-flex flex-col gap-1.5", className)}
        >
          {label && (
            <span className="text-sm font-medium text-foreground">{label}</span>
          )}

          <div className="relative">
            <button
              type="button"
              onClick={() =>
                !entry || entry.status !== "uploading"
                  ? inputRef.current?.click()
                  : undefined
              }
              disabled={entry?.status === "uploading"}
              aria-label="Upload logo"
              className={cn(
                "group relative w-20 h-20 rounded-2xl overflow-hidden",
                "flex items-center justify-center",
                "transition-all duration-200",
                "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
                entry
                  ? "border border-border shadow-sm"
                  : cn(
                      "border-2 border-dashed border-border bg-muted/40 cursor-pointer",
                      "hover:border-primary/50 hover:bg-primary/[0.04] hover:shadow-sm",
                    ),
                entry?.status === "uploading" && "cursor-default",
              )}
            >
              {entry ? (
                <>
                  <img
                    src={entry.preview}
                    alt="Logo preview"
                    className="w-full h-full object-cover"
                  />

                  {entry.status === "uploading" && (
                    <div className="absolute inset-0 bg-background/60 backdrop-blur-[2px] flex items-center justify-center">
                      <Loader2 className="w-5 h-5 animate-spin text-foreground" />
                    </div>
                  )}

                  {entry.status === "done" && (
                    <div className="absolute inset-0 flex flex-col items-center justify-center gap-0.5 bg-black/0 hover:bg-black/50 transition-colors duration-200 cursor-pointer opacity-0 hover:opacity-100">
                      <Camera
                        className="w-4 h-4 text-white"
                        aria-hidden="true"
                      />
                      <span className="text-[10px] font-medium text-white">
                        Change
                      </span>
                    </div>
                  )}

                  {entry.status === "error" && (
                    <div className="absolute inset-0 bg-destructive/20 backdrop-blur-[2px] flex items-center justify-center">
                      <AlertCircle className="w-4 h-4 text-destructive" />
                    </div>
                  )}
                </>
              ) : (
                <div className="flex flex-col items-center gap-1.5 text-muted-foreground group-hover:text-primary transition-colors duration-200 select-none">
                  <ImagePlus className="w-5 h-5" aria-hidden="true" />
                  <span className="text-[10px] font-semibold tracking-wide">
                    LOGO
                  </span>
                </div>
              )}
            </button>

            {/* Remove / retry controls outside the main button */}
            {entry && entry.status !== "uploading" && (
              <div className="absolute -top-2 -right-2 flex gap-1 motion-safe:animate-in motion-safe:zoom-in-75 duration-150">
                {entry.status === "error" && (
                  <button
                    type="button"
                    onClick={() => retryUpload(entry)}
                    aria-label="Retry upload"
                    className="w-6 h-6 rounded-full bg-background border border-border shadow-md flex items-center justify-center hover:bg-muted transition-colors duration-150 cursor-pointer"
                  >
                    <RotateCcw className="w-3 h-3 text-muted-foreground" />
                  </button>
                )}
                <button
                  type="button"
                  onClick={() => removeEntry(entry.id)}
                  aria-label="Remove image"
                  className="w-6 h-6 rounded-full bg-background border border-border shadow-md flex items-center justify-center hover:bg-destructive hover:border-destructive hover:text-white transition-colors duration-150 cursor-pointer"
                >
                  <X className="w-3 h-3" />
                </button>
              </div>
            )}
          </div>

          {entry?.status === "error" && (
            <p className="text-[11px] font-medium text-destructive leading-tight max-w-20">
              {entry.error ?? "Upload failed"}
            </p>
          )}
          {validationError && (
            <p className="text-[11px] font-medium text-destructive leading-tight max-w-20">
              {validationError}
            </p>
          )}

          <input
            ref={inputRef}
            type="file"
            accept={acceptStr}
            onChange={(e) => handleFiles(e.target.files)}
            className="hidden"
            aria-hidden="true"
          />
        </div>
      );
    }

    // ── Gallery / multi-image mode ─────────────────────────────────────────────
    return (
      <div ref={ref} className={cn("space-y-3", className)}>
        {label && (
          <span className="text-sm font-medium text-foreground">{label}</span>
        )}

        {/* Drop zone — hidden when limit reached */}
        {!isAtLimit && (
          <div
            role="button"
            tabIndex={0}
            aria-label="Upload images"
            onKeyDown={(e) => e.key === "Enter" && inputRef.current?.click()}
            onClick={() => inputRef.current?.click()}
            onDragOver={(e) => {
              e.preventDefault();
              setIsDragging(true);
            }}
            onDragLeave={(e) => {
              e.preventDefault();
              setIsDragging(false);
            }}
            onDrop={(e) => {
              e.preventDefault();
              setIsDragging(false);
              handleFiles(e.dataTransfer.files);
            }}
            className={cn(
              "rounded-xl border-2 border-dashed cursor-pointer",
              "transition-all duration-200",
              "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
              isDragging
                ? "border-primary bg-primary/[0.05] shadow-[inset_0_0_0_1px] shadow-primary/20"
                : "border-border bg-muted/20 hover:border-primary/40 hover:bg-muted/40",
            )}
          >
            <div className="flex flex-col items-center justify-center py-8 gap-3.5">
              <div
                className={cn(
                  "w-11 h-11 rounded-xl border flex items-center justify-center",
                  "bg-gradient-to-b from-background to-muted/60 shadow-sm",
                  "transition-all duration-200",
                  isDragging
                    ? "border-primary/40 scale-110 shadow-md shadow-primary/10"
                    : "border-border",
                )}
              >
                <Upload
                  className={cn(
                    "w-[18px] h-[18px] transition-colors duration-200",
                    isDragging ? "text-primary" : "text-muted-foreground",
                  )}
                  aria-hidden="true"
                />
              </div>

              <div className="text-center space-y-1">
                <p className="text-sm font-medium text-foreground">
                  {isDragging ? (
                    "Drop to upload"
                  ) : (
                    <>
                      <span className="text-primary font-semibold">
                        Click to upload
                      </span>
                      <span className="text-muted-foreground font-normal">
                        {" "}
                        or drag and drop
                      </span>
                    </>
                  )}
                </p>
                <p className="text-xs text-muted-foreground">
                  {acceptLabel}
                  {maxFileSizeBytes > 0 &&
                    ` · up to ${formatSize(maxFileSizeBytes)}`}
                  {maxImages > 0 && ` · max ${maxImages}`}
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Limit reached notice */}
        {isAtLimit && (
          <p className="text-xs text-muted-foreground">
            Maximum of {maxImages} images reached &mdash; remove one to add
            another
          </p>
        )}

        {/* Validation error */}
        {validationError && (
          <div
            role="alert"
            className="flex items-center gap-2 px-3 py-2.5 rounded-lg bg-destructive/[0.08] border border-destructive/20 text-xs font-medium text-destructive motion-safe:animate-in motion-safe:fade-in-0 motion-safe:slide-in-from-top-1 duration-200"
          >
            <AlertCircle className="w-3.5 h-3.5 shrink-0" aria-hidden="true" />
            {validationError}
          </div>
        )}

        {/* Image grid */}
        {entries.length > 0 && (
          <div
            className={cn(
              "grid gap-2.5",
              entries.length === 1
                ? "grid-cols-2 max-w-[180px]"
                : "grid-cols-3 sm:grid-cols-4",
            )}
          >
            {entries.map((entry) => (
              <div
                key={entry.id}
                className={cn(
                  "group relative aspect-square rounded-xl overflow-hidden",
                  "border border-border bg-muted/30 shadow-sm",
                  "transition-shadow duration-200 hover:shadow-md",
                  "motion-safe:animate-in motion-safe:fade-in-0 motion-safe:zoom-in-95 duration-300",
                )}
              >
                <img
                  src={entry.preview}
                  alt={entry.file.name}
                  className="w-full h-full object-cover"
                />

                {/* Uploading overlay */}
                {entry.status === "uploading" && (
                  <div className="absolute inset-0 bg-background/60 backdrop-blur-[2px] flex flex-col items-center justify-center gap-1.5">
                    <Loader2 className="w-5 h-5 animate-spin text-foreground" />
                    <span className="text-[10px] font-medium text-muted-foreground">
                      Uploading…
                    </span>
                  </div>
                )}

                {/* Error overlay */}
                {entry.status === "error" && (
                  <div className="absolute inset-0 bg-destructive/15 backdrop-blur-[2px] flex flex-col items-center justify-center gap-1.5">
                    <AlertCircle className="w-4 h-4 text-destructive" />
                    <button
                      type="button"
                      onClick={() => retryUpload(entry)}
                      className="text-[10px] font-semibold text-destructive underline-offset-2 underline cursor-pointer hover:no-underline"
                    >
                      Retry
                    </button>
                  </div>
                )}

                {/* Subtle bottom scrim on hover for filename */}
                {entry.status === "done" && (
                  <div className="absolute inset-x-0 bottom-0 px-1.5 pb-1 pt-4 bg-gradient-to-t from-black/55 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-200 pointer-events-none">
                    <p className="text-[10px] text-white truncate font-medium">
                      {entry.file.name}
                    </p>
                  </div>
                )}

                {/* Controls — shown on hover when not uploading */}
                {entry.status !== "uploading" && (
                  <button
                    type="button"
                    onClick={() => removeEntry(entry.id)}
                    aria-label={`Remove ${entry.file.name}`}
                    className={cn(
                      "absolute top-1.5 right-1.5 w-6 h-6 rounded-full",
                      "bg-background/95 border border-border shadow-md",
                      "flex items-center justify-center cursor-pointer",
                      "opacity-0 group-hover:opacity-100 focus-visible:opacity-100 transition-opacity duration-150",
                      "hover:bg-destructive hover:border-destructive hover:text-white",
                    )}
                  >
                    <X className="w-3 h-3" aria-hidden="true" />
                  </button>
                )}
              </div>
            ))}
          </div>
        )}

        <input
          ref={inputRef}
          type="file"
          multiple
          accept={acceptStr}
          onChange={(e) => handleFiles(e.target.files)}
          className="hidden"
          aria-hidden="true"
        />
      </div>
    );
  },
);

ImageUploader.displayName = "ImageUploader";
