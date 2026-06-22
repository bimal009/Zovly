"use client";

import * as React from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Check, Loader2 } from "lucide-react";
import { cn } from "@repo/ui/utils";
import { Button } from "@repo/ui/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@repo/ui/components/ui/dialog";
import { Input } from "@repo/ui/components/ui/input";
import { Label } from "@repo/ui/components/ui/label";
import { Textarea } from "@repo/ui/components/ui/textarea";
import type { CreateFaqInput, Faq, UpdateFaqInput } from "../api/faq";

const ANSWER_MAX = 500;

const FaqFormSchema = z.object({
  question: z.string().min(1, "Question is required"),
  answer: z.string().min(1, "Answer is required").max(ANSWER_MAX, `Answer must be ${ANSWER_MAX} characters or less`),
});

type FaqFormValues = z.infer<typeof FaqFormSchema>;
const DEFAULT_VALUES: FaqFormValues = { question: "", answer: "" };

const STEPS = [
  { label: "Saving FAQ", delay: 0 },
  { label: "Generating embeddings", delay: 500 },
  { label: "Vectorizing content", delay: 1100 },
];

type Phase = "form" | "processing" | "done";

interface FaqFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  editing: Faq | null;
  onSave: (data: CreateFaqInput | { id: string; input: UpdateFaqInput }) => void;
  saving: boolean;
  saveError: Error | null;
}

export function FaqFormDialog({
  open,
  onOpenChange,
  editing,
  onSave,
  saving,
  saveError,
}: FaqFormDialogProps) {
  const [phase, setPhase] = React.useState<Phase>("form");
  const [activeStep, setActiveStep] = React.useState(-1);
  const [completedSteps, setCompletedSteps] = React.useState<Set<number>>(new Set());

  const { register, handleSubmit, reset, watch, formState: { errors } } = useForm<FaqFormValues>({
    resolver: zodResolver(FaqFormSchema),
    defaultValues: DEFAULT_VALUES,
  });

  const answerLength = watch("answer")?.length ?? 0;

  // Reset when dialog opens
  React.useEffect(() => {
    if (open) {
      reset(editing ? { question: editing.question, answer: editing.answer } : DEFAULT_VALUES);
      setPhase("form");
      setActiveStep(-1);
      setCompletedSteps(new Set());
    }
  }, [editing, open, reset]);

  // Start step animations when saving begins
  React.useEffect(() => {
    if (!saving || phase !== "form") return;
    setPhase("processing");

    const timers: ReturnType<typeof setTimeout>[] = [];
    STEPS.forEach((step, idx) => {
      timers.push(setTimeout(() => setActiveStep(idx), step.delay));
    });
    return () => timers.forEach(clearTimeout);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [saving]);

  // Advance completedSteps when activeStep moves to the next one
  React.useEffect(() => {
    if (activeStep > 0) {
      setCompletedSteps((prev) => new Set([...prev, activeStep - 1]));
    }
  }, [activeStep]);

  // Detect save finishing (saving flips true→false while processing)
  const prevSavingRef = React.useRef(saving);
  React.useEffect(() => {
    const wasSaving = prevSavingRef.current;
    prevSavingRef.current = saving;

    if (!wasSaving || saving || phase !== "processing") return;

    if (saveError) {
      // API failed — go back to the form so the user can retry
      setPhase("form");
    } else {
      // Success — snap all steps to done and show the banner
      setCompletedSteps(new Set([0, 1, 2]));
      setActiveStep(-1);
      setPhase("done");
    }
  }, [saving, saveError, phase]);

  // Auto-close after "done" banner has been visible briefly
  React.useEffect(() => {
    if (phase !== "done") return;
    const t = setTimeout(() => onOpenChange(false), 900);
    return () => clearTimeout(t);
  }, [phase, onOpenChange]);

  function onSubmit(values: FaqFormValues) {
    if (editing) {
      onSave({ id: editing.id, input: values });
    } else {
      onSave(values);
    }
  }

  // Block Escape / overlay click while animating
  function handleOpenChange(v: boolean) {
    if (!v && phase !== "form") return;
    onOpenChange(v);
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg overflow-hidden">
        <DialogHeader>
          <DialogTitle>{editing ? "Edit FAQ" : "Add FAQ"}</DialogTitle>
        </DialogHeader>

        {phase === "form" ? (
          <>
            <form
              id="faq-form"
              onSubmit={handleSubmit(onSubmit)}
              className="flex flex-col gap-4 py-2"
            >
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="faq-question">Question *</Label>
                <Input
                  id="faq-question"
                  placeholder="e.g. What are your business hours?"
                  {...register("question")}
                />
                {errors.question && (
                  <p className="text-xs text-destructive">{errors.question.message}</p>
                )}
              </div>

              <div className="flex flex-col gap-1.5">
                <Label htmlFor="faq-answer">Answer *</Label>
                <Textarea
                  id="faq-answer"
                  placeholder="Provide a clear and helpful answer…"
                  rows={4}
                  {...register("answer")}
                />
                <div className="flex items-center justify-between">
                  {errors.answer ? (
                    <p className="text-xs text-destructive">{errors.answer.message}</p>
                  ) : (
                    <span />
                  )}
                  <span className={cn(
                    "text-xs tabular-nums",
                    answerLength > ANSWER_MAX ? "text-destructive" : answerLength > ANSWER_MAX * 0.85 ? "text-warning" : "text-muted-foreground",
                  )}>
                    {answerLength}/{ANSWER_MAX}
                  </span>
                </div>
              </div>
            </form>

            <DialogFooter>
              <Button
                variant="outline"
                className="cursor-pointer"
                onClick={() => onOpenChange(false)}
              >
                Cancel
              </Button>
              <Button type="submit" form="faq-form" className="cursor-pointer">
                {editing ? "Save Changes" : "Add FAQ"}
              </Button>
            </DialogFooter>
          </>
        ) : (
          <div className="flex flex-col items-center gap-8 py-8">
            <div className="flex flex-col gap-5 w-full max-w-xs">
              {STEPS.map((step, idx) => {
                const done = completedSteps.has(idx) || phase === "done";
                const active = activeStep === idx && !done;

                return (
                  <div key={step.label} className="flex items-center gap-3">
                    <div
                      className={cn(
                        "flex h-8 w-8 shrink-0 items-center justify-center rounded-full border-2 transition-all duration-500",
                        done
                          ? "border-primary bg-primary"
                          : active
                            ? "border-primary bg-transparent"
                            : "border-muted-foreground/30 bg-transparent",
                      )}
                    >
                      {done ? (
                        <Check className="h-4 w-4 text-primary-foreground" />
                      ) : active ? (
                        <Loader2 className="h-4 w-4 animate-spin text-primary" />
                      ) : (
                        <span className="h-2 w-2 rounded-full bg-muted-foreground/30" />
                      )}
                    </div>

                    <div className="flex flex-col">
                      <span
                        className={cn(
                          "text-sm font-medium transition-colors duration-300",
                          done || active ? "text-foreground" : "text-muted-foreground",
                        )}
                      >
                        {step.label}
                      </span>
                      {active && (
                        <span className="text-xs text-muted-foreground animate-in fade-in duration-300">
                          In progress…
                        </span>
                      )}
                      {done && (
                        <span className="text-xs text-primary animate-in fade-in duration-300">
                          Complete
                        </span>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>

            {phase === "done" && (
              <div className="flex items-center gap-2 rounded-full bg-primary/10 px-4 py-2 text-sm font-medium text-primary animate-in fade-in zoom-in-95 duration-300">
                <Check className="h-4 w-4" />
                FAQ added to knowledge base
              </div>
            )}

            {phase === "processing" && (
              <p className="text-xs text-muted-foreground animate-in fade-in duration-500">
                Hang tight — this only takes a moment
              </p>
            )}
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
