import { z } from 'zod';

// Product schemas
export const ProductSchema = z.object({
  id: z.string().uuid().optional(),
  name: z.string().min(1, 'Product name is required').max(255, 'Product name too long'),
  description: z.string().optional(),
  sku: z.string().min(1, 'SKU is required').max(100, 'SKU too long'),
  category_id: z.string().uuid('Invalid category ID'),
  price: z.number().min(0, 'Price must be positive'),
  cost: z.number().min(0, 'Cost must be positive').optional(),
  stock_quantity: z.number().int().min(0, 'Stock quantity must be non-negative'),
  min_stock_level: z.number().int().min(0, 'Minimum stock level must be non-negative').optional(),
  unit: z.string().optional(),
  status: z.enum(['active', 'inactive']).default('active'),
  created_at: z.string().optional(),
  updated_at: z.string().optional(),
});

export const CreateProductSchema = ProductSchema.omit({
  id: true,
  created_at: true,
  updated_at: true,
});

export const UpdateProductSchema = ProductSchema.partial().omit({
  id: true,
  created_at: true,
  updated_at: true,
});

// License schemas
export const LicenseSchema = z.object({
  id: z.string().uuid().optional(),
  serial_number: z.string().min(1, 'Serial number is required').max(255, 'Serial number too long'),
  created_at: z.string().optional(),
  updated_at: z.string().optional(),
});

export const CreateLicenseSchema = LicenseSchema.omit({
  id: true,
  created_at: true,
  updated_at: true,
});

// Customer schemas
export const CustomerSchema = z.object({
  id: z.string().uuid().optional(),
  username: z.string().min(1, 'Username is required').max(100, 'Username too long'),
  pin: z.string().min(4, 'PIN must be at least 4 characters').max(255, 'PIN too long'),
  role_id: z.enum(['cashier', 'owner_business']),
  serial_number: z.string().min(1, 'License serial number is required'),
  created_at: z.string().optional(),
  updated_at: z.string().optional(),
});

export const CreateCustomerSchema = CustomerSchema.omit({
  id: true,
  created_at: true,
  updated_at: true,
});

export const UpdateCustomerSchema = CustomerSchema.partial().omit({
  id: true,
  created_at: true,
  updated_at: true,
});

// User schemas
export const UserSchema = z.object({
  id: z.string().uuid().optional(),
  username: z.string().min(1, 'Username is required').max(100, 'Username too long'),
  email: z.string().email('Invalid email'),
  password: z.string().min(8, 'Password must be at least 8 characters').optional(),
  role_id: z.enum(['admin', 'super_admin']),
  serial_number: z.string().min(1, 'License serial number is required'),
  created_at: z.string().optional(),
  updated_at: z.string().optional(),
});

export const CreateUserSchema = UserSchema.omit({
  id: true,
  created_at: true,
  updated_at: true,
}).extend({
  password: z.string().min(8, 'Password must be at least 8 characters'),
});

export const UpdateUserSchema = UserSchema.partial().omit({
  id: true,
  created_at: true,
  updated_at: true,
});

// Role schema (Casbin RBAC with domain support)
export const RoleSchema = z.object({
  id: z.string().uuid().optional(),
  name: z.string().min(1, 'Role name is required').max(100, 'Role name too long'),
  description: z.string().optional(),
  domain: z.string().min(1, 'Domain is required'),
  created_at: z.string().optional(),
  updated_at: z.string().optional(),
});

export const CreateRoleSchema = RoleSchema.omit({
  id: true,
  created_at: true,
  updated_at: true,
});

export const UpdateRoleSchema = RoleSchema.partial().omit({
  id: true,
  created_at: true,
  updated_at: true,
});

// Permission schema (Casbin policy: sub, dom, obj, act)
export const PermissionSchema = z.object({
  id: z.string().uuid().optional(),
  subject: z.string().min(1, 'Subject is required'), // role or user
  domain: z.string().min(1, 'Domain is required'),
  object: z.string().min(1, 'Object is required'), // resource like users, products, etc.
  action: z.string().min(1, 'Action is required'), // read, write, delete, etc.
  created_at: z.string().optional(),
  updated_at: z.string().optional(),
});

export const CreatePermissionSchema = PermissionSchema.omit({
  id: true,
  created_at: true,
  updated_at: true,
});

export const UpdatePermissionSchema = PermissionSchema.partial().omit({
  id: true,
  created_at: true,
  updated_at: true,
});

// Role assignment schema (Casbin grouping: user, role, domain)
export const RoleAssignmentSchema = z.object({
  id: z.string().uuid().optional(),
  user_id: z.string().uuid('Invalid user ID'),
  role_name: z.string().min(1, 'Role name is required'),
  domain: z.string().min(1, 'Domain is required'),
  created_at: z.string().optional(),
  updated_at: z.string().optional(),
});

export const CreateRoleAssignmentSchema = RoleAssignmentSchema.omit({
  id: true,
  created_at: true,
  updated_at: true,
});

// Domain schema
export const DomainSchema = z.object({
  id: z.string().uuid().optional(),
  name: z.string().min(1, 'Domain name is required').max(100, 'Domain name too long'),
  description: z.string().optional(),
  created_at: z.string().optional(),
  updated_at: z.string().optional(),
});

export const CreateDomainSchema = DomainSchema.omit({
  id: true,
  created_at: true,
  updated_at: true,
});

// Auth schemas
export const LoginSchema = z.object({
  email: z.string().email('Invalid email'),
  password: z.string().min(1, 'Password is required'),
  domain: z.string().min(1, 'Domain/License is required'),
});

export const RegisterSchema = z.object({
  username: z.string().min(1, 'Username is required'),
  email: z.string().email('Invalid email'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
  domain: z.string().min(1, 'Domain/License is required'),
});