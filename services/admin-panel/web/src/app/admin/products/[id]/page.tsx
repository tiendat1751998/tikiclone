'use client';

import { useState, useTransition, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { cn } from '@shopee/ui-system';
import { ProductSchema, type ProductFormData } from '@/lib/validations';
import { api } from '@/lib/api-client';
import { formatVND } from '@shopee/shared-utils';

interface ProductEditPageProps {
  initialData?: ProductFormData;
  isNew?: boolean;
}

export function ProductEditForm({ initialData, isNew = false }: ProductEditPageProps) {
  const router = useRouter();
  const [isPending, startTransition] = useTransition();
  const [error, setError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [formData, setFormData] = useState<ProductFormData>(
    initialData || {
      name: '',
      slug: '',
      description: '',
      price: 0,
      sale_price: undefined,
      quantity: 0,
      category_id: '',
      brand_id: '',
      images: [],
      status: 'draft',
      attributes: {},
    }
  );

  const handleChange = useCallback(
    (field: keyof ProductFormData, value: unknown) => {
      setFormData((prev) => ({ ...prev, [field]: value }));
      if (fieldErrors[field]) {
        setFieldErrors((prev) => {
          const updated = { ...prev };
          delete updated[field];
          return updated;
        });
      }
      setError(null);
    },
    [fieldErrors]
  );

  const generateSlug = useCallback((name: string) => {
    return name
      .toLowerCase()
      .normalize('NFD')
      .replace(/[\u0300-\u036f]/g, '')
      .replace(/[đĐ]/g, 'd')
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/(^-|-$)/g, '');
  }, []);

  const handleNameChange = (value: string) => {
    handleChange('name', value);
    if (!initialData?.slug) {
      handleChange('slug', generateSlug(value));
    }
  };

  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      setError(null);

      const result = ProductSchema.safeParse(formData);
      if (!result.success) {
        const errors: Record<string, string> = {};
        result.error.issues.forEach((issue) => {
          const field = issue.path[0] as string;
          if (!errors[field]) {
            errors[field] = issue.message;
          }
        });
        setFieldErrors(errors);
        return;
      }

      startTransition(async () => {
        try {
          if (isNew) {
            await api.post('/api/admin/products', result.data);
            router.push('/admin/products');
          } else {
            await api.put(`/api/admin/products/${initialData?.id}`, result.data);
            router.push('/admin/products');
          }
          router.refresh();
        } catch (err) {
          setError(err instanceof Error ? err.message : 'Failed to save product');
        }
      });
    },
    [formData, isNew, initialData, router, generateSlug]
  );

  return (
    <form onSubmit={handleSubmit} className="space-y-8">
      {error && (
        <div className="p-4 rounded-lg bg-danger-50 dark:bg-danger-900/20 border border-danger-200 dark:border-danger-800">
          <p className="text-sm text-danger-600 dark:text-danger-400">{error}</p>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <div className="lg:col-span-2 space-y-6">
          <div className="rounded-xl border border-border bg-card p-6">
            <h3 className="text-lg font-semibold text-foreground mb-4">Basic Information</h3>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-foreground mb-1.5">
                  Product Name <span className="text-danger-500">*</span>
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => handleNameChange(e.target.value)}
                  className={cn(
                    'w-full px-3 py-2.5 rounded-lg border bg-card text-sm transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500',
                    fieldErrors.name ? 'border-danger-500' : 'border-border'
                  )}
                  placeholder="Enter product name"
                  disabled={isPending}
                />
                {fieldErrors.name && (
                  <p className="mt-1 text-xs text-danger-500">{fieldErrors.name}</p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-foreground mb-1.5">
                  Slug <span className="text-danger-500">*</span>
                </label>
                <input
                  type="text"
                  value={formData.slug}
                  onChange={(e) => handleChange('slug', e.target.value)}
                  className={cn(
                    'w-full px-3 py-2.5 rounded-lg border bg-card text-sm transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 font-mono',
                    fieldErrors.slug ? 'border-danger-500' : 'border-border'
                  )}
                  placeholder="product-slug"
                  disabled={isPending}
                />
                {fieldErrors.slug && (
                  <p className="mt-1 text-xs text-danger-500">{fieldErrors.slug}</p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-foreground mb-1.5">
                  Description <span className="text-danger-500">*</span>
                </label>
                <textarea
                  value={formData.description}
                  onChange={(e) => handleChange('description', e.target.value)}
                  rows={6}
                  className={cn(
                    'w-full px-3 py-2.5 rounded-lg border bg-card text-sm transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 resize-none',
                    fieldErrors.description ? 'border-danger-500' : 'border-border'
                  )}
                  placeholder="Describe your product..."
                  disabled={isPending}
                />
                {fieldErrors.description && (
                  <p className="mt-1 text-xs text-danger-500">{fieldErrors.description}</p>
                )}
              </div>
            </div>
          </div>

          <div className="rounded-xl border border-border bg-card p-6">
            <h3 className="text-lg font-semibold text-foreground mb-4">Pricing & Inventory</h3>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div>
                <label className="block text-sm font-medium text-foreground mb-1.5">
                  Price (VND) <span className="text-danger-500">*</span>
                </label>
                <input
                  type="number"
                  min="0"
                  step="1000"
                  value={formData.price}
                  onChange={(e) => handleChange('price', parseFloat(e.target.value) || 0)}
                  className={cn(
                    'w-full px-3 py-2.5 rounded-lg border bg-card text-sm transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500',
                    fieldErrors.price ? 'border-danger-500' : 'border-border'
                  )}
                  placeholder="0"
                  disabled={isPending}
                />
                {fieldErrors.price && (
                  <p className="mt-1 text-xs text-danger-500">{fieldErrors.price}</p>
                )}
                <p className="mt-1 text-xs text-muted-foreground">
                  {formatVND(formData.price)}
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-foreground mb-1.5">
                  Sale Price (VND)
                </label>
                <input
                  type="number"
                  min="0"
                  step="1000"
                  value={formData.sale_price ?? ''}
                  onChange={(e) =>
                    handleChange(
                      'sale_price',
                      e.target.value ? parseFloat(e.target.value) : undefined
                    )
                  }
                  className="w-full px-3 py-2.5 rounded-lg border border-border bg-card text-sm transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500"
                  placeholder="Optional"
                  disabled={isPending}
                />
                {formData.sale_price && (
                  <p className="mt-1 text-xs text-muted-foreground">
                    {formatVND(formData.sale_price)}
                  </p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-foreground mb-1.5">
                  Quantity <span className="text-danger-500">*</span>
                </label>
                <input
                  type="number"
                  min="0"
                  value={formData.quantity}
                  onChange={(e) => handleChange('quantity', parseInt(e.target.value) || 0)}
                  className={cn(
                    'w-full px-3 py-2.5 rounded-lg border bg-card text-sm transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500',
                    fieldErrors.quantity ? 'border-danger-500' : 'border-border'
                  )}
                  placeholder="0"
                  disabled={isPending}
                />
                {fieldErrors.quantity && (
                  <p className="mt-1 text-xs text-danger-500">{fieldErrors.quantity}</p>
                )}
              </div>
            </div>
          </div>
        </div>

        <div className="space-y-6">
          <div className="rounded-xl border border-border bg-card p-6">
            <h3 className="text-lg font-semibold text-foreground mb-4">Status</h3>

            <select
              value={formData.status}
              onChange={(e) =>
                handleChange('status', e.target.value as 'draft' | 'published' | 'archived')
              }
              className="w-full px-3 py-2.5 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              disabled={isPending}
            >
              <option value="draft">Draft</option>
              <option value="published">Published</option>
              <option value="archived">Archived</option>
            </select>
          </div>

          <div className="rounded-xl border border-border bg-card p-6">
            <h3 className="text-lg font-semibold text-foreground mb-4">Organization</h3>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-foreground mb-1.5">
                  Category <span className="text-danger-500">*</span>
                </label>
                <select
                  value={formData.category_id}
                  onChange={(e) => handleChange('category_id', e.target.value)}
                  className={cn(
                    'w-full px-3 py-2.5 rounded-lg border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500',
                    fieldErrors.category_id ? 'border-danger-500' : 'border-border'
                  )}
                  disabled={isPending}
                >
                  <option value="">Select category</option>
                  <option value="electronics">Electronics</option>
                  <option value="fashion">Fashion</option>
                  <option value="home">Home & Living</option>
                  <option value="beauty">Beauty</option>
                  <option value="sports">Sports</option>
                </select>
                {fieldErrors.category_id && (
                  <p className="mt-1 text-xs text-danger-500">{fieldErrors.category_id}</p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-foreground mb-1.5">
                  Brand
                </label>
                <select
                  value={formData.brand_id || ''}
                  onChange={(e) => handleChange('brand_id', e.target.value || '')}
                  className="w-full px-3 py-2.5 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
                  disabled={isPending}
                >
                  <option value="">Select brand</option>
                  <option value="samsung">Samsung</option>
                  <option value="apple">Apple</option>
                  <option value="nike">Nike</option>
                  <option value="adidas">Adidas</option>
                </select>
              </div>
            </div>
          </div>

          <div className="rounded-xl border border-border bg-card p-6">
            <h3 className="text-lg font-semibold text-foreground mb-4">Product Images</h3>

            <div className="border-2 border-dashed border-border rounded-lg p-6 text-center">
              <svg
                className="w-8 h-8 text-muted-foreground mx-auto mb-2"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
                />
              </svg>
              <p className="text-sm text-muted-foreground mb-2">
                Drag and drop images or click to upload
              </p>
              <button
                type="button"
                className="text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400"
                disabled={isPending}
              >
                Browse files
              </button>
            </div>

            {formData.images.length > 0 && (
              <div className="mt-4 grid grid-cols-3 gap-2">
                {formData.images.map((url, index) => (
                  <div
                    key={index}
                    className="aspect-square rounded-lg bg-muted flex items-center justify-center"
                  >
                    <img
                      src={url}
                      alt={`Product ${index + 1}`}
                      className="w-full h-full object-cover rounded-lg"
                    />
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="flex gap-3">
            <button
              type="button"
              onClick={() => router.back()}
              className="flex-1 py-2.5 px-4 rounded-lg border border-border bg-card text-foreground font-medium text-sm hover:bg-muted transition-colors"
              disabled={isPending}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isPending}
              className="flex-1 py-2.5 px-4 rounded-lg bg-primary-500 text-white font-medium text-sm hover:bg-primary-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {isPending ? (
                <span className="inline-flex items-center gap-2">
                  <svg className="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                    <circle
                      className="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      strokeWidth="4"
                    />
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                    />
                  </svg>
                  Saving...
                </span>
              ) : isNew ? (
                'Create Product'
              ) : (
                'Save Changes'
              )}
            </button>
          </div>
        </div>
      </div>
    </form>
  );
}
