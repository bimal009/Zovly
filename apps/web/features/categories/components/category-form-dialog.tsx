"use client";

import * as React from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@repo/ui/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@repo/ui/components/ui/dialog";
import { Input } from "@repo/ui/components/ui/input";
import { Label } from "@repo/ui/components/ui/label";
import { Textarea } from "@repo/ui/components/ui/textarea";
import { type CreateCategoryInput } from "../types/categories";

const CategoryFormSchema = z.object({
  name: z.string().min(1, "Name is required"),
  slug: z.string(),
  description: z.string(),
});

type CategoryFormValues = z.infer<typeof CategoryFormSchema>;

const DEFAULT_VALUES: CategoryFormValues = {
  name: "",
  slug: "",
  description: "",
};

// "Summer Collection" → "summer-collection"
function slugify(value: string) {
  return value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

interface CategoryFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (data: CreateCategoryInput) => void;
  saving: boolean;
}

export function CategoryFormDialog({
  open,
  onOpenChange,
  onSave,
  saving,
}: CategoryFormDialogProps) {
  const {
    register,
    handleSubmit,
    reset,
    setValue,
    getValues,
    formState: { errors },
  } = useForm<CategoryFormValues>({
    resolver: zodResolver(CategoryFormSchema),
    defaultValues: DEFAULT_VALUES,
  });

  // track whether the user has manually edited the slug, so we stop auto-syncing
  const slugEdited = React.useRef(false);

  React.useEffect(() => {
    if (open) {
      reset(DEFAULT_VALUES);
      slugEdited.current = false;
    }
  }, [open, reset]);

  function handleNameChange(value: string) {
    if (!slugEdited.current) {
      setValue("slug", slugify(value), { shouldDirty: true });
    }
  }

  function onSubmit(values: CategoryFormValues) {
    onSave({
      name: values.name.trim(),
      slug: values.slug.trim() || undefined,
      description: values.description.trim() || undefined,
    });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Add Category</DialogTitle>
          <DialogDescription>
            Group your products to keep your catalog organized.
          </DialogDescription>
        </DialogHeader>

        <form
          id="category-form"
          onSubmit={handleSubmit(onSubmit)}
          className="flex flex-col gap-4 py-2"
        >
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="c-name">Name *</Label>
            <Input
              id="c-name"
              placeholder="e.g. Summer Collection"
              {...register("name", {
                onChange: (e) => handleNameChange(e.target.value),
              })}
            />
            {errors.name && (
              <p className="text-xs text-destructive">{errors.name.message}</p>
            )}
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="c-slug">Slug</Label>
            <Input
              id="c-slug"
              placeholder="e.g. summer-collection"
              {...register("slug", {
                onChange: () => {
                  slugEdited.current = getValues("slug").length > 0;
                },
              })}
            />
            <p className="text-xs text-muted-foreground">
              Used in URLs. Auto-filled from the name.
            </p>
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="c-desc">Description</Label>
            <Textarea
              id="c-desc"
              placeholder="Short description…"
              rows={3}
              {...register("description")}
            />
          </div>
        </form>

        <DialogFooter>
          <Button
            variant="outline"
            className="cursor-pointer"
            onClick={() => onOpenChange(false)}
            disabled={saving}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            form="category-form"
            className="cursor-pointer"
            disabled={saving}
          >
            {saving ? "Saving…" : "Add Category"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
