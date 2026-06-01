import { ProductEditForm } from '../[id]/page';

export default function NewProductPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-foreground">Add New Product</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Create a new product in your catalog
        </p>
      </div>

      <ProductEditForm isNew />
    </div>
  );
}
