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
} from "lucide-react";
import { Card } from "@repo/ui/components/ui/card";
import { cn } from "@repo/ui/utils";

// ── Validation schema ─────────────────────────────────────────────────────────

const onboardingSchema = z.object({
  // Step 1
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
  { id: 1, label: "Identity" },
  { id: 2, label: "Type" },
  { id: 3, label: "Contact" },
];

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
  const currentStepLabel = STEPS.find((s) => s.id === step)?.label ?? "";

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <div className="w-full max-w-lg">
        {/* Header */}
        <div className="mb-6">
          <h1 className="text-2xl font-semibold tracking-tight">
            Set up your business
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            Step {step} of {STEPS.length} &mdash;{" "}
            <span className="text-foreground font-medium">
              {currentStepLabel}
            </span>
          </p>
        </div>

        {/* Step progress */}
        <div className="flex items-center gap-2 mb-8">
          {STEPS.map((s, i) => (
            <div key={s.id} className="flex items-center gap-2 flex-1">
              <div
                className={cn(
                  "w-6 h-6 rounded-full flex items-center justify-center text-xs font-semibold shrink-0 transition-colors duration-200",
                  step > s.id
                    ? "bg-primary text-primary-foreground"
                    : step === s.id
                      ? "bg-primary text-primary-foreground"
                      : "bg-muted text-muted-foreground",
                )}
              >
                {step > s.id ? <Check className="w-3 h-3" /> : s.id}
              </div>
              <span
                className={cn(
                  "text-xs font-medium transition-colors duration-200",
                  step >= s.id ? "text-foreground" : "text-muted-foreground",
                )}
              >
                {s.label}
              </span>
              {i < STEPS.length - 1 && (
                <div
                  className={cn(
                    "flex-1 h-px transition-colors duration-300",
                    step > s.id ? "bg-primary" : "bg-border",
                  )}
                />
              )}
            </div>
          ))}
        </div>

        {/* Form card */}
        <Card className="p-6">
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
            {/* ── Step 1: Identity ── */}
            {step === 1 && (
              <div className="space-y-4">
                <div className="space-y-1.5">
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
                    {...register("name", {
                      onChange: (e) => handleNameChange(e.target.value),
                    })}
                  />
                  {errors.name && (
                    <p className="text-xs text-destructive">
                      {errors.name.message}
                    </p>
                  )}
                </div>

                <div className="space-y-1.5">
                  <label
                    htmlFor="slug"
                    className="text-sm font-medium leading-none"
                  >
                    URL slug
                  </label>
                  <div className="flex items-center rounded-md border bg-muted/40 focus-within:ring-1 focus-within:ring-ring overflow-hidden">
                    <span className="px-3 text-sm text-muted-foreground border-r bg-muted select-none h-9 flex items-center">
                      zovly.com/
                    </span>
                    <Input
                      id="slug"
                      className="border-0 bg-transparent focus-visible:ring-0 rounded-none h-9"
                      placeholder="acme-studio"
                      {...register("slug")}
                    />
                  </div>
                  {errors.slug ? (
                    <p className="text-xs text-destructive">
                      {errors.slug.message}
                    </p>
                  ) : (
                    <p className="text-xs text-muted-foreground">
                      Lowercase letters, numbers, and hyphens only
                    </p>
                  )}
                </div>

                <div className="space-y-1.5">
                  <div className="flex items-baseline justify-between">
                    <label
                      htmlFor="description"
                      className="text-sm font-medium leading-none"
                    >
                      Description
                    </label>
                  </div>
                  <Textarea
                    id="description"
                    placeholder="What does your business do?"
                    className="resize-none"
                    rows={6}
                    {...register("description")}
                  />
                  <div className="flex items-center justify-between">
                    {errors.description ? (
                      <p className="text-xs text-destructive">
                        {errors.description.message}
                      </p>
                    ) : (
                      <span />
                    )}
                    <span className="text-xs text-muted-foreground">
                      {descriptionValue?.length ?? 0}/2000
                    </span>
                  </div>
                </div>
              </div>
            )}

            {/* ── Step 2: Business type ── */}
            {step === 2 && (
              <div className="space-y-4">
                <div>
                  <p className="text-sm font-medium">
                    What does your business do?
                  </p>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    This activates the right modules for you.
                  </p>
                </div>
                <div className="grid gap-2.5">
                  {BUSINESS_TYPES.map((t) => {
                    const Icon = t.icon;
                    const selected = typeValue === t.value;
                    return (
                      <button
                        key={t.value}
                        type="button"
                        onClick={() =>
                          setValue("type", t.value, { shouldValidate: true })
                        }
                        className={cn(
                          "w-full cursor-pointer text-left rounded-lg border p-4 transition-all duration-200",
                          "hover:border-primary/40 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring",
                          selected
                            ? "border-primary bg-primary/5"
                            : "border-border bg-transparent",
                        )}
                      >
                        <div className="flex items-center gap-3">
                          <div
                            className={cn(
                              "w-8 h-8 rounded-md flex items-center justify-center shrink-0 transition-colors duration-200",
                              selected
                                ? "bg-primary text-primary-foreground"
                                : "bg-muted text-muted-foreground",
                            )}
                          >
                            <Icon className="w-4 h-4" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium">{t.label}</p>
                            <p className="text-xs text-muted-foreground mt-0.5">
                              {t.description}
                            </p>
                          </div>
                          {selected && (
                            <Check className="w-4 h-4 text-primary shrink-0" />
                          )}
                        </div>
                      </button>
                    );
                  })}
                </div>
                {errors.type && (
                  <p className="text-xs text-destructive">
                    {errors.type.message}
                  </p>
                )}
              </div>
            )}

            {/* ── Step 3: Contact & location ── */}
            {step === 3 && (
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-1.5">
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
                      {...register("phone")}
                    />
                    {errors.phone && (
                      <p className="text-xs text-destructive">
                        {errors.phone.message}
                      </p>
                    )}
                  </div>

                  <div className="space-y-1.5">
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
                      {...register("website")}
                    />
                    {errors.website && (
                      <p className="text-xs text-destructive">
                        {errors.website.message}
                      </p>
                    )}
                  </div>
                </div>

                <div className="space-y-1.5">
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
                    {...register("address")}
                  />
                  {errors.address && (
                    <p className="text-xs text-destructive">
                      {errors.address.message}
                    </p>
                  )}
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-1.5">
                    <label
                      htmlFor="city"
                      className="text-sm font-medium leading-none"
                    >
                      City
                    </label>
                    <Input
                      id="city"
                      placeholder="Kathmandu"
                      {...register("city")}
                    />
                    {errors.city && (
                      <p className="text-xs text-destructive">
                        {errors.city.message}
                      </p>
                    )}
                  </div>

                  <div className="space-y-1.5">
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
                          onChange={(country) => field.onChange(country.alpha3)}
                          placeholder="Select country"
                        />
                      )}
                    />
                    {errors.country && (
                      <p className="text-xs text-destructive">
                        {errors.country.message}
                      </p>
                    )}
                  </div>
                </div>
              </div>
            )}

            {/* ── Navigation ── */}
            <div className="flex items-center justify-between pt-1">
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={handleBack}
                disabled={step === 1}
                className="gap-1.5"
              >
                <ChevronLeft className="w-4 h-4" />
                Back
              </Button>

              {step < STEPS.length ? (
                <Button
                  type="button"
                  variant="default"
                  onClick={handleNext}
                  className="gap-1.5"
                >
                  Continue
                  <ChevronRight className="w-4 h-4" />
                </Button>
              ) : (
                <Button
                  type="submit"
                  variant="default"
                  disabled={isSubmitting}
                  className="gap-1.5 min-w-36"
                >
                  {isSubmitting ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Creating...
                    </>
                  ) : (
                    <>
                      Create business
                      <Check className="w-4 h-4" />
                    </>
                  )}
                </Button>
              )}
            </div>
          </form>
        </Card>

        <p className="text-center text-xs text-muted-foreground mt-4">
          You can update all of this later in Settings
        </p>
      </div>
    </div>
  );
}
