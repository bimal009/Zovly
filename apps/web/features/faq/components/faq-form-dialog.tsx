"use client";

import * as React from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Loader2 } from "lucide-react";
import { Button } from "@repo/ui/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@repo/ui/components/ui/dialog";
import { Label } from "@repo/ui/components/ui/label";
import { Textarea } from "@repo/ui/components/ui/textarea";
import { cn } from "@repo/ui/utils";
import { CreateFaqInput, createFaqSchema, Faq, UpdateFaqInput } from "@repo/types";

const QUESTION_MAX = 150; // must match schema
const ANSWER_MAX = 300;   // must match schema

type FaqFormValues = z.infer<typeof createFaqSchema>;
const DEFAULT_VALUES: FaqFormValues = {
  question: "",
  answer: "",
  isActive: true,
};

interface FaqFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  editing: Faq | null;
  onSave: (data: CreateFaqInput | { id: string; input: UpdateFaqInput }) => void;
  saving: boolean;
}

export function FaqFormDialog({
  open,
  onOpenChange,
  editing,
  onSave,
  saving,
}: FaqFormDialogProps) {
  const {
    register,
    handleSubmit,
    reset,
    watch,
    formState: { errors },
  } = useForm<FaqFormValues>({
    resolver: zodResolver(createFaqSchema),
    defaultValues: DEFAULT_VALUES,
  });

  const questionLength = watch("question")?.length ?? 0;
  const answerLength   = watch("answer")?.length   ?? 0;

  React.useEffect(() => {
    if (open) {
      reset(
        editing
          ? { question: editing.question, answer: editing.answer, isActive: editing.isActive ?? true }
          : DEFAULT_VALUES
      );
    }
  }, [editing, open, reset]);

  function onSubmit(values: FaqFormValues) {
    if (editing) {
      onSave({ id: editing.id, input: values });
    } else {
      onSave(values);
    }
  }

  function handleOpenChange(v: boolean) {
    if (!v && saving) return;
    onOpenChange(v);
  }

  // Shared helper so both counters use identical colour thresholds
  function counterClass(length: number, max: number) {
    return cn(
      "text-xs tabular-nums",
      length > max
        ? "text-destructive"
        : length > max * 0.85
          ? "text-warning"
          : "text-muted-foreground"
    );
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg overflow-hidden">
        <DialogHeader>
          <DialogTitle>{editing ? "Edit FAQ" : "Add FAQ"}</DialogTitle>
        </DialogHeader>

        <form
          id="faq-form"
          onSubmit={handleSubmit(onSubmit)}
          className="flex flex-col gap-4 py-2 min-w-0"
        >
          {/* ── Question ── */}
          <div className="flex flex-col gap-1.5 min-w-0">
            <Label htmlFor="faq-question">Question *</Label>
            <Textarea
              id="faq-question"
              placeholder="e.g. What are your business hours?"
              rows={2}
              disabled={saving}
              className="w-full break-words"
              {...register("question")}
            />
            <div className="flex items-center justify-between">
              {errors.question ? (
                <p className="text-xs text-destructive">{errors.question.message}</p>
              ) : (
                <span />
              )}
              <span className={counterClass(questionLength, QUESTION_MAX)}>
                {questionLength}/{QUESTION_MAX}
              </span>
            </div>
          </div>

          {/* ── Answer ── */}
          <div className="flex flex-col gap-1.5 min-w-0">
            <Label htmlFor="faq-answer">Answer *</Label>
            <Textarea
              id="faq-answer"
              placeholder="Provide a clear and helpful answer…"
              rows={4}
              disabled={saving}
              className="w-full break-words"
              {...register("answer")}
            />
            <div className="flex items-center justify-between">
              {errors.answer ? (
                <p className="text-xs text-destructive">{errors.answer.message}</p>
              ) : (
                <span />
              )}
              <span className={counterClass(answerLength, ANSWER_MAX)}>
                {answerLength}/{ANSWER_MAX}
              </span>
            </div>
          </div>
        </form>

        <DialogFooter>
          <Button
            variant="outline"
            className="cursor-pointer"
            disabled={saving}
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            form="faq-form"
            className="cursor-pointer"
            disabled={saving}
          >
            {saving && <Loader2 className="mr-1.5 h-4 w-4 animate-spin" />}
            {editing ? "Save Changes" : "Add FAQ"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}