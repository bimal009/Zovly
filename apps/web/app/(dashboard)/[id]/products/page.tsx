import { SiteHeader } from "@repo/ui/components/ui/site-header";
import { ProductsTable } from "@/features/products/components/products-table";

export default function ProductsPage() {
  return (
    <>
      <SiteHeader title="Products" />
      <div className="flex flex-1 flex-col">
        <div className="@container/main flex flex-1 flex-col gap-2">
          <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
            <ProductsTable />
          </div>
        </div>
      </div>
    </>
  );
}
