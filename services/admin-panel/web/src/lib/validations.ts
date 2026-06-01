import { z } from 'zod';

export const LoginSchema = z.object({
  email: z.string().email('Invalid email address').min(1, 'Email is required'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
  totp_code: z.string().length(6, 'TOTP code must be 6 digits').optional(),
});

export type LoginFormData = z.infer<typeof LoginSchema>;

export const ProductSchema = z.object({
  id: z.string().optional(),
  name: z.string().min(1, 'Product name is required').max(500),
  slug: z.string().min(1).regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/, 'Invalid slug format'),
  description: z.string().min(1).max(5000),
  price: z.number().positive('Price must be positive'),
  sale_price: z.number().nonnegative().optional(),
  quantity: z.number().int().nonnegative(),
  category_id: z.string().uuid('Invalid category'),
  brand_id: z.string().uuid().optional(),
  images: z.array(z.string().url()).min(1, 'At least one image required'),
  status: z.enum(['draft', 'published', 'archived']),
  attributes: z.record(z.string(), z.string()).optional(),
});

export type ProductFormData = z.infer<typeof ProductSchema>;

export const ProductFilterSchema = z.object({
  search: z.string().optional(),
  category_id: z.string().optional(),
  brand_id: z.string().optional(),
  status: z.enum(['draft', 'published', 'archived', 'all']).optional(),
  min_price: z.coerce.number().optional(),
  max_price: z.coerce.number().optional(),
  low_stock: z.coerce.boolean().optional(),
});

export type ProductFilterParams = z.infer<typeof ProductFilterSchema>;

export const OrderFilterSchema = z.object({
  search: z.string().optional(),
  status: z.enum([
    'pending', 'confirmed', 'processing', 'shipped',
    'delivered', 'cancelled', 'refunded', 'all'
  ]).optional(),
  date_from: z.string().optional(),
  date_to: z.string().optional(),
  min_amount: z.coerce.number().optional(),
  max_amount: z.coerce.number().optional(),
});

export type OrderFilterParams = z.infer<typeof OrderFilterSchema>;

export const PaginationSchema = z.object({
  page: z.coerce.number().positive().default(1),
  per_page: z.coerce.number().positive().max(100).default(20),
  sort_by: z.string().optional(),
  sort_order: z.enum(['asc', 'desc']).default('desc'),
});

export type PaginationParams = z.infer<typeof PaginationSchema>;

export const UserFilterSchema = z.object({
  search: z.string().optional(),
  status: z.enum(['active', 'banned', 'pending', 'all']).optional(),
  role: z.string().optional(),
});

export type UserFilterParams = z.infer<typeof UserFilterSchema>;

export const InventoryUpdateSchema = z.object({
  product_id: z.string().uuid(),
  quantity: z.number().int(),
  reason: z.string().min(1).max(200),
});

export type InventoryUpdateData = z.infer<typeof InventoryUpdateSchema>;

export const OrderStatusUpdateSchema = z.object({
  status: z.enum([
    'confirmed', 'processing', 'shipped', 'delivered', 'cancelled', 'refunded'
  ]),
  note: z.string().max(500).optional(),
});

export type OrderStatusUpdateData = z.infer<typeof OrderStatusUpdateSchema>;
