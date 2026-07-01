"use client";

import * as React from "react";
import { MessageCircleQuestion, Pencil, Plus, Trash2 } from "lucide-react";
import { Badge } from "@repo/ui/components/ui/badge";
import { Button } from "@repo/ui/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@repo/ui/components/ui/card";
import { ConfirmDeleteDialog } from "@/components/confirm-delete-dialog";
import {
  useCreateFaq,
  useDeleteFaq,
  useGetFaqs,
  useUpdateFaq,
} from "../client/faq";
import { FaqFormDialog } from "./faq-form-dialog";
import { toast } from "@repo/ui/components/ui/sonner";
import { CreateFaqInput, Faq, UpdateFaqInput } from "@repo/types";

export function FaqList(businessId:{businessId:string}) {
  const { data, isLoading } = useGetFaqs(businessId.businessId);
  const createMutation = useCreateFaq(businessId.businessId);
  const updateMutation = useUpdateFaq(businessId.businessId);
  const deleteMutation = useDeleteFaq(businessId.businessId);

  const faqs = data?.data ?? [];
  const saving = createMutation.isPending || updateMutation.isPending;

  const [formOpen, setFormOpen] = React.useState(false);
  const [editing, setEditing] = React.useState<Faq | null>(null);
  const [deleteTarget, setDeleteTarget] = React.useState<Faq | null>(null);

  function openCreate() {
    setEditing(null);
    setFormOpen(true);
  }

  function openEdit(faq: Faq) {
    setEditing(faq);
    setFormOpen(true);
  }

  function handleSave(
    data: CreateFaqInput | { id: string; input: UpdateFaqInput }
  ) {
    if ("id" in data) {
      updateMutation.mutate(data, {
        onSuccess: (res) => {
          toast.success(res?.message ?? "FAQ updated successfully");
          setFormOpen(false);
        },
        onError: (error) => {
          toast.error(error?.message ?? "Failed to update FAQ");
        },
      });
    } else {
      createMutation.mutate(data, {
        onSuccess: (res) => {
          toast.success(res?.message ?? "FAQ added successfully");
          setFormOpen(false);
        },
        onError: (error) => {
          toast.error(error?.message ?? "Failed to add FAQ");
        },
      });
    }
  }

  function handleConfirmDelete() {
    if (!deleteTarget) return;
    deleteMutation.mutate(deleteTarget.id, {
      onSuccess: (res) => {
        toast.success(res?.message ?? "FAQ deleted successfully");
        setDeleteTarget(null);
      },
      onError: (error) => {
        toast.error(error?.message ?? "Failed to delete FAQ");
      },
    });
  }

  return (
    <>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between pb-4">
          <div className="flex items-center gap-2">
            <MessageCircleQuestion className="h-5 w-5 text-muted-foreground" />
            <CardTitle className="text-base">
              Frequently Asked Questions
            </CardTitle>
            {!isLoading && (
              <Badge variant="secondary" className="ml-1">
                {faqs.length}
              </Badge>
            )}
          </div>
          <Button onClick={openCreate} size="sm" className="cursor-pointer">
            <Plus className="mr-1.5 h-4 w-4" />
            Add FAQ
          </Button>
        </CardHeader>

        <CardContent>
          {isLoading ? (
            <div className="flex flex-col gap-3">
              {[1, 2, 3].map((i) => (
                <div
                  key={i}
                  className="h-20 animate-pulse rounded-lg bg-muted"
                />
              ))}
            </div>
          ) : faqs.length === 0 ? (
            <div className="flex flex-col items-center gap-2 py-10 text-center text-muted-foreground">
              <MessageCircleQuestion className="h-8 w-8 opacity-40" />
              <p className="text-sm">No FAQs yet.</p>
              <p className="text-xs">
                Add common questions your customers ask.
              </p>
            </div>
          ) : (
            <div className="flex flex-col divide-y">
              {faqs.map((faq, idx) => (
                <div
                  key={faq.id}
                  className="group flex items-start gap-4 py-4 first:pt-0 last:pb-0"
                >
                  <span className="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-medium text-muted-foreground">
                    {idx + 1}
                  </span>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium leading-snug">
                      {faq.question}
                    </p>
                    <p className="mt-1 text-sm text-muted-foreground leading-relaxed line-clamp-3">
                      {faq.answer}
                    </p>
                  </div>
                  <div className="flex shrink-0 items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7 cursor-pointer"
                      onClick={() => openEdit(faq)}
                    >
                      <Pencil className="h-3.5 w-3.5" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7 cursor-pointer text-destructive hover:text-destructive"
                      onClick={() => setDeleteTarget(faq)}
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <FaqFormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        editing={editing}
        onSave={handleSave}
        saving={saving}
      />

      <ConfirmDeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete FAQ?"
        description={
          <>
            This will permanently delete the question{" "}
            <span className="font-semibold">"{deleteTarget?.question}"</span>.
          </>
        }
        onConfirm={handleConfirmDelete}
        loading={deleteMutation.isPending}
      />
    </>
  );
}