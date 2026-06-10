"use client";

import { useState } from "react";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Input } from "@repo/ui/components/ui/input";
import { Button } from "@repo/ui/components/ui/button";
import { Textarea } from "@repo/ui/components/ui/textarea";
import { CountryDropdown } from "@repo/ui/components/ui/country-dropdown";
import {
  ShoppingBag,
  Wrench,
  LayoutGrid,
  ChevronRight,
  ChevronLeft,
  Check,
  Loader2,
  AlertCircle,
} from "lucide-react";
import { Card } from "@repo/ui/components/ui/card";
import { cn } from "@repo/ui/utils";
import { ImageUploader } from "@/components/ImageUploader";
import { authClient } from "@/lib/auth-client";

// ── Validation schema ─────────────────────────────────────────────────────────

const onboardingSchema = z.object({
  // Step 1
  logo: z.string().optional(),
  name: z
    .string()
    .min(2, "Business name must be at least 2 characters")
    .max(80, "Too long"),
  slug: z
    .string()
    .min(2, "Slug must be at least 2 characters")
    .max(40, "Too long")
    .regex(/^[a-z0-9-]+$/, "Only lowercase letters, numbers, and hyphens"),
  description: z.string().max(2000, "Max 2000 characters"),

  // Step 2
  type: z.enum(["product", "service", "both"]),

  // Step 3
  phone: z.string().min(7, "Enter a valid phone number"),
  website: z.string().url("Enter a valid URL").optional().or(z.literal("")),
  address: z.string().max(200).optional(),
  city: z.string().min(1, "City is required").max(80),
  country: z.string().min(1, "Country is required"),
});

type OnboardingValues = z.infer<typeof onboardingSchema>;

// ── Business type options ─────────────────────────────────────────────────────

const BUSINESS_TYPES = [
  {
    value: "product" as const,
    label: "Product seller",
    description: "You sell physical or digital products",
    icon: ShoppingBag,
  },
  {
    value: "service" as const,
    label: "Service business",
    description: "You sell services",
    icon: Wrench,
  },
  {
    value: "both" as const,
    label: "Both",
    description: "You sell products and services",
    icon: LayoutGrid,
  },
];

const STEPS = [
  { id: 1, label: "Identity", title: "Tell us about your business" },
  { id: 2, label: "Type", title: "What do you offer?" },
  { id: 3, label: "Contact", title: "How can customers reach you?" },
];

// ── Small UI helpers ──────────────────────────────────────────────────────────

function FieldError({ message }: { message?: string }) {
  if (!message) return null;
  return (
    <p
      role="alert"
      className="flex items-center gap-1.5 text-xs font-medium text-destructive motion-safe:animate-in motion-safe:fade-in-0 motion-safe:slide-in-from-top-1 duration-200"
    >
      <AlertCircle className="w-3 h-3 shrink-0" aria-hidden="true" />
      {message}
    </p>
  );
}

// ── Component ─────────────────────────────────────────────────────────────────

interface BusinessOnboardingFormProps {
  onSuccess?: (values: OnboardingValues) => void;
}

export function BusinessOnboardingForm({
  onSuccess,
}: BusinessOnboardingFormProps) {
  const [step, setStep] = useState(1);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const {
    register,
    control,
    handleSubmit,
    setValue,
    watch,
    trigger,
    formState: { errors },
  } = useForm<OnboardingValues>({
    resolver: zodResolver(onboardingSchema),
    defaultValues: {
      logo: "",
      name: "",
      slug: "",
      description: "",
      type: "service",
      phone: "",
      website: "",
      address: "",
      city: "",
      country: "NPL",
    },
    mode: "onChange",
  });

  const handleNameChange = (value: string) => {
    const slug = value
      .toLowerCase()
      .trim()
      .replace(/[^a-z0-9\s-]/g, "")
      .replace(/\s+/g, "-")
      .replace(/-+/g, "-");
    setValue("slug", slug, { shouldValidate: true });
  };

  const validateStep = async (currentStep: number) => {
    const fieldsPerStep: Record<number, (keyof OnboardingValues)[]> = {
      1: ["name", "slug"],
      2: ["type"],
      3: ["phone", "city", "country", "website", "address"],
    };
    return trigger(fieldsPerStep[currentStep]);
  };

  const handleNext = async () => {
    const valid = await validateStep(step);
    if (valid) setStep((s) => s + 1);
  };

  const handleBack = () => setStep((s) => s - 1);

  const onSubmit = async (values: OnboardingValues) => {
    setIsSubmitting(true);
    try {
      await new Promise((r) => setTimeout(r, 1200));
      onSuccess?.(values);
    } finally {
      setIsSubmitting(false);
    }
  };

  const typeValue = watch("type");
  const descriptionValue = watch("description");
  const slugValue = watch("slug");
  const currentStep = STEPS.find((s) => s.id === step) ?? STEPS[0]!;

  return (
    <div className="relative min-h-screen bg-background flex items-center justify-center p-4 overflow-hidden">
      {/* Ambient backdrop */}

      <div className="relative w-full max-w-lg">
        {/* Header */}
        <div className="mb-7 text-center sm:text-left">
          <h1 className="text-[28px] leading-tight font-semibold tracking-tight text-foreground">
            Set up your business
          </h1>
          <p className="text-sm text-muted-foreground mt-1.5">
            Takes about a minute &mdash; you can change everything later.
          </p>
        </div>

        {/* Step progress */}
        <nav aria-label="Onboarding progress" className="mb-8">
          <ol className="flex items-center gap-3">
            {STEPS.map((s, i) => {
              const isComplete = step > s.id;
              const isActive = step === s.id;
              return (
                <li
                  key={s.id}
                  className="flex items-center gap-3 flex-1 last:flex-none"
                >
                  <div className="flex items-center gap-2">
                    <div
                      aria-current={isActive ? "step" : undefined}
                      className={cn(
                        "w-7 h-7 rounded-full flex items-center justify-center text-xs font-semibold shrink-0",
                        "transition-all duration-300",
                        isComplete &&
                          "bg-primary text-primary-foreground shadow-sm shadow-primary/30",
                        isActive &&
                          "bg-primary text-primary-foreground ring-4 ring-primary/15 shadow-sm shadow-primary/30",
                        !isComplete &&
                          !isActive &&
                          "bg-muted text-muted-foreground border border-border",
                      )}
                    >
                      {isComplete ? (
                        <Check
                          className="w-3.5 h-3.5 motion-safe:animate-in motion-safe:zoom-in-50 duration-200"
                          aria-hidden="true"
                        />
                      ) : (
                        s.id
                      )}
                    </div>
                    <span
                      className={cn(
                        "text-xs font-medium transition-colors duration-300 hidden sm:inline",
                        step >= s.id
                          ? "text-foreground"
                          : "text-muted-foreground",
                      )}
                    >
                      {s.label}
                    </span>
                  </div>
                  {i < STEPS.length - 1 && (
                    <div className="flex-1 h-1 rounded-full bg-border/70 overflow-hidden">
                      <div
                        className={cn(
                          "h-full rounded-full bg-primary transition-all duration-500 ease-out",
                          isComplete ? "w-full" : "w-0",
                        )}
                      />
                    </div>
                  )}
                </li>
              );
            })}
          </ol>
        </nav>

        {/* Form card */}
        <Card className="rounded-2xl border-border/70 p-6 sm:p-8 shadow-[0_1px_2px_rgb(0_0_0/0.04),0_12px_32px_-12px_rgb(0_0_0/0.12)]">
          <form onSubmit={(e) => e.preventDefault()} className="space-y-6">
            {/* Step title */}
            <div
              key={`title-${step}`}
              className="motion-safe:animate-in motion-safe:fade-in-0 duration-300"
            >
              <p className="text-[11px] font-semibold tracking-wide text-primary uppercase">
                Step {step} of {STEPS.length}
              </p>
              <h2 className="text-lg font-semibold tracking-tight mt-0.5">
                {currentStep.title}
              </h2>
            </div>

            <div
              key={step}
              className="motion-safe:animate-in motion-safe:fade-in-0 motion-safe:slide-in-from-bottom-2 duration-300"
            >
              {/* ── Step 1: Identity ── */}
              {step === 1 && (
                <div className="space-y-5">
                  {/* Logo */}
                  <div className="flex items-center gap-4 rounded-xl border border-border/70 bg-muted/30 p-4">
                    <Controller
                      control={control}
                      name="logo"
                      render={({ field }) => (
                        <ImageUploader
                          maxImages={1}
                          folder="/logos"
                          onUpload={(url) => field.onChange(url)}
                          accept={["image/jpeg", "image/png", "image/webp"]}
                          maxFileSizeBytes={2 * 1024 * 1024}
                        />
                      )}
                    />
                    <div className="space-y-0.5">
                      <p className="text-sm font-medium">Business logo</p>
                      <p className="text-xs text-muted-foreground leading-relaxed">
                        Optional &middot; PNG, JPG, WebP &middot; max 2 MB
                        <br />
                        Square images look best
                      </p>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <label
                      htmlFor="name"
                      className="text-sm font-medium leading-none"
                    >
                      Business name
                    </label>
                    <Input
                      id="name"
                      placeholder="Acme Studio"
                      autoFocus
                      className="h-10"
                      {...register("name", {
                        onChange: (e) => handleNameChange(e.target.value),
                      })}
                    />
                    <FieldError message={errors.name?.message} />
                  </div>

                  <div className="space-y-2">
                    <label
                      htmlFor="slug"
                      className="text-sm font-medium leading-none"
                    >
                      URL slug
                    </label>
                    <div
                      className={cn(
                        "flex items-center rounded-md border border-input bg-background overflow-hidden",
                        "transition-[border-color,box-shadow] duration-200",
                        "focus-within:border-ring focus-within:ring-2 focus-within:ring-ring/25",
                      )}
                    >
                      <span className="px-3 text-sm text-muted-foreground border-r bg-muted/60 select-none h-10 flex items-center font-medium">
                        zovly.com/
                      </span>
                      <Input
                        id="slug"
                        className="border-0 bg-transparent focus-visible:ring-0 rounded-none h-10 shadow-none"
                        placeholder="acme-studio"
                        {...register("slug")}
                      />
                      {slugValue && !errors.slug && (
                        <Check
                          className="w-4 h-4 text-primary mr-3 shrink-0 motion-safe:animate-in motion-safe:zoom-in-50 duration-200"
                          aria-hidden="true"
                        />
                      )}
                    </div>
                    {errors.slug ? (
                      <FieldError message={errors.slug.message} />
                    ) : (
                      <p className="text-xs text-muted-foreground">
                        This will be your public storefront address
                      </p>
                    )}
                  </div>

                  <div className="space-y-2">
                    <div className="flex items-baseline justify-between">
                      <label
                        htmlFor="description"
                        className="text-sm font-medium leading-none"
                      >
                        Description
                      </label>
                      <span className="text-xs text-muted-foreground tabular-nums">
                        {descriptionValue?.length ?? 0}/2000
                      </span>
                    </div>
                    <Textarea
                      id="description"
                      placeholder="What does your business do? Customers will see this on your page."
                      className="resize-none leading-relaxed"
                      rows={5}
                      {...register("description")}
                    />
                    <FieldError message={errors.description?.message} />
                  </div>
                </div>
              )}

              {/* ── Step 2: Business type ── */}
              {step === 2 && (
                <div className="space-y-4">
                  <p className="text-sm text-muted-foreground -mt-2">
                    This activates the right modules for you.
                  </p>
                  <div
                    role="radiogroup"
                    aria-label="Business type"
                    className="grid gap-3"
                  >
                    {BUSINESS_TYPES.map((t) => {
                      const Icon = t.icon;
                      const selected = typeValue === t.value;
                      return (
                        <button
                          key={t.value}
                          type="button"
                          role="radio"
                          aria-checked={selected}
                          onClick={() =>
                            setValue("type", t.value, { shouldValidate: true })
                          }
                          className={cn(
                            "w-full cursor-pointer text-left rounded-xl border p-4",
                            "transition-all duration-200",
                            "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40 focus-visible:ring-offset-1",
                            selected
                              ? "border-primary/50 bg-primary/[0.04] ring-1 ring-primary/25 shadow-sm"
                              : "border-border bg-transparent hover:border-primary/30 hover:bg-muted/40",
                          )}
                        >
                          <div className="flex items-center gap-3.5">
                            <div
                              className={cn(
                                "w-10 h-10 rounded-lg flex items-center justify-center shrink-0",
                                "transition-colors duration-200",
                                selected
                                  ? "bg-primary text-primary-foreground shadow-sm shadow-primary/30"
                                  : "bg-muted text-muted-foreground",
                              )}
                            >
                              <Icon
                                className="w-[18px] h-[18px]"
                                aria-hidden="true"
                              />
                            </div>
                            <div className="flex-1 min-w-0">
                              <p className="text-sm font-semibold">{t.label}</p>
                              <p className="text-xs text-muted-foreground mt-0.5">
                                {t.description}
                              </p>
                            </div>
                            <div
                              className={cn(
                                "w-5 h-5 rounded-full border-2 flex items-center justify-center shrink-0",
                                "transition-colors duration-200",
                                selected
                                  ? "border-primary bg-primary"
                                  : "border-muted-foreground/30",
                              )}
                              aria-hidden="true"
                            >
                              {selected && (
                                <Check className="w-3 h-3 text-primary-foreground motion-safe:animate-in motion-safe:zoom-in-50 duration-150" />
                              )}
                            </div>
                          </div>
                        </button>
                      );
                    })}
                  </div>
                  <FieldError message={errors.type?.message} />
                </div>
              )}

              {/* ── Step 3: Contact & location ── */}
              {step === 3 && (
                <div className="space-y-5">
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <label
                        htmlFor="phone"
                        className="text-sm font-medium leading-none"
                      >
                        Phone
                      </label>
                      <Input
                        id="phone"
                        placeholder="+977 98XXXXXXXX"
                        autoFocus
                        className="h-10"
                        {...register("phone")}
                      />
                      <FieldError message={errors.phone?.message} />
                    </div>

                    <div className="space-y-2">
                      <div className="flex items-baseline justify-between">
                        <label
                          htmlFor="website"
                          className="text-sm font-medium leading-none"
                        >
                          Website
                        </label>
                        <span className="text-xs text-muted-foreground">
                          optional
                        </span>
                      </div>
                      <Input
                        id="website"
                        placeholder="https://yoursite.com"
                        className="h-10"
                        {...register("website")}
                      />
                      <FieldError message={errors.website?.message} />
                    </div>
                  </div>

                  <div className="space-y-2">
                    <div className="flex items-baseline justify-between">
                      <label
                        htmlFor="address"
                        className="text-sm font-medium leading-none"
                      >
                        Address
                      </label>
                      <span className="text-xs text-muted-foreground">
                        optional
                      </span>
                    </div>
                    <Input
                      id="address"
                      placeholder="Thamel, Kathmandu 44600"
                      className="h-10"
                      {...register("address")}
                    />
                    <FieldError message={errors.address?.message} />
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <label
                        htmlFor="city"
                        className="text-sm font-medium leading-none"
                      >
                        City
                      </label>
                      <Input
                        id="city"
                        placeholder="Kathmandu"
                        className="h-10"
                        {...register("city")}
                      />
                      <FieldError message={errors.city?.message} />
                    </div>

                    <div className="space-y-2">
                      <label
                        htmlFor="country"
                        className="text-sm font-medium leading-none"
                      >
                        Country
                      </label>
                      <Controller
                        control={control}
                        name="country"
                        render={({ field }) => (
                          <CountryDropdown
                            defaultValue={field.value}
                            onChange={(country) =>
                              field.onChange(country.alpha3)
                            }
                            placeholder="Select country"
                          />
                        )}
                      />
                      <FieldError message={errors.country?.message} />
                    </div>
                  </div>
                </div>
              )}
            </div>

            {/* ── Navigation ── */}
            <div className="flex items-center justify-between border-t border-border/60 pt-5">
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={handleBack}
                disabled={step === 1}
                className={cn(
                  "gap-1.5 text-muted-foreground hover:text-foreground",
                  step === 1 && "invisible",
                )}
              >
                <ChevronLeft className="w-4 h-4" aria-hidden="true" />
                Back
              </Button>

              {step < STEPS.length ? (
                <Button
                  type="button"
                  variant="default"
                  onClick={handleNext}
                  className="gap-1.5 h-10 px-5 shadow-sm shadow-primary/25"
                >
                  Continue
                  <ChevronRight className="w-4 h-4" aria-hidden="true" />
                </Button>
              ) : (
                <Button
                  type="button"
                  variant="default"
                  disabled={isSubmitting}
                  onClick={handleSubmit(onSubmit)}
                  className="gap-1.5 h-10 px-5 min-w-40 shadow-sm shadow-primary/25"
                >
                  {isSubmitting ? (
                    <>
                      <Loader2
                        className="w-4 h-4 animate-spin"
                        aria-hidden="true"
                      />
                      Creating...
                    </>
                  ) : (
                    <>
                      Create business
                      <Check className="w-4 h-4" aria-hidden="true" />
                    </>
                  )}
                </Button>
              )}
            </div>
          </form>
        </Card>

        <p className="text-center text-xs text-muted-foreground mt-5">
          You can update all of this later in Settings
        </p>
      </div>
    </div>
  );
}
