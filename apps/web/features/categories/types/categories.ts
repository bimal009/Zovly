export type Category = {
  id: string;
  business_id: string;
  name: string;
  description: string | null;
  slug: string | null;
  created_at: string;
  updated_at: string;
};

export type CreateCategoryInput = {
  name: string;
  description?: string;
  slug?: string;
};
