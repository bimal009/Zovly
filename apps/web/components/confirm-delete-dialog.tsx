"use client";

import * as React from "react";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@repo/ui/components/ui/alert-dialog";

interface ConfirmDeleteDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title?: string;
  description?: React.ReactNode;
  onConfirm: () => void;
  loading?: boolean;
  confirmText?: string;
}

const wrapStyle: React.CSSProperties = {
  overflowWrap: "anywhere",
  wordBreak: "break-word",
  whiteSpace: "normal",
};

export function ConfirmDeleteDialog({
  open,
  onOpenChange,
  title = "Are you sure?",
  description = "This action cannot be undone.",
  onConfirm,
  loading = false,
  confirmText = "Delete",
}: ConfirmDeleteDialogProps) {
  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent
        className="min-w-0 max-w-lg"
        style={{ overflow: "hidden" }}
      >
        <AlertDialogHeader className="min-w-0">
          <AlertDialogTitle style={wrapStyle}>{title}</AlertDialogTitle>
          <AlertDialogDescription style={wrapStyle}>
            {description}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel className="cursor-pointer" disabled={loading}>
            Cancel
          </AlertDialogCancel>
          <AlertDialogAction
            className="cursor-pointer bg-destructive text-destructive-foreground hover:bg-destructive/90"
            onClick={onConfirm}
            disabled={loading}
          >
            {loading ? `${confirmText}…` : confirmText}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}