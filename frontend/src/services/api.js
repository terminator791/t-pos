import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import api from "../lib/api";
import { toast } from "react-toastify";

// Products API
export const useProducts = (params = {}) => {
  return useQuery({
    queryKey: ["products", params],
    queryFn: async () => {
      const response = await api.get("/products", { params });
      return response.data;
    },
  });
};

export const useProduct = (id) => {
  return useQuery({
    queryKey: ["products", id],
    queryFn: async () => {
      const response = await api.get(`/products/${id}`);
      return response.data;
    },
    enabled: !!id,
  });
};

export const useProductByBarcode = (barcode) => {
  return useQuery({
    queryKey: ["products", "barcode", barcode],
    queryFn: async () => {
      const response = await api.get(`/products/barcode/${barcode}`);
      return response.data;
    },
    enabled: !!barcode,
  });
};

export const useSearchProducts = (query, shopId) => {
  return useQuery({
    queryKey: ["products", "search", query, shopId],
    queryFn: async () => {
      const response = await api.get("/products/search", {
        params: { q: query, shop_id: shopId },
      });
      return response.data;
    },
    enabled: !!query && !!shopId,
  });
};

export const useLowStockProducts = (shopId) => {
  return useQuery({
    queryKey: ["products", "low-stock", shopId],
    queryFn: async () => {
      const response = await api.get("/products/low-stock", {
        params: { shop_id: shopId },
      });
      return response.data;
    },
    enabled: !!shopId,
  });
};

export const useCreateProduct = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (productData) => {
      const response = await api.post("/products", productData);
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["products"] });
      toast.success("Product created successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to create product";
      toast.error(message);
    },
  });
};

export const useCreateProductWithFile = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (formData) => {
      const response = await api.post("/products/upload", formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      });
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["products"] });
      toast.success("Product created successfully with file upload");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to create product with file";
      toast.error(message);
    },
  });
};

export const useUpdateProduct = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, ...productData }) => {
      const response = await api.put(`/products/${id}`, productData);
      return response.data;
    },
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries({ queryKey: ["products"] });
      queryClient.invalidateQueries({ queryKey: ["products", variables.id] });
      toast.success("Product updated successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to update product";
      toast.error(message);
    },
  });
};

export const useDeleteProduct = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id) => {
      const response = await api.delete(`/products/${id}`);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["products"] });
      toast.success("Product deleted successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to delete product";
      toast.error(message);
    },
  });
};

// Categories API (for product categories)
export const useCategories = (shopId) => {
  return useQuery({
    queryKey: ["categories", shopId],
    queryFn: async () => {
      const response = await api.get("/categories", {
        params: shopId ? { shop_id: shopId } : {},
      });
      return response.data;
    },
  });
};

// Shops API (for product shop selection)
export const useShops = (params = {}) => {
  return useQuery({
    queryKey: ["shops", params],
    queryFn: async () => {
      const response = await api.get("/shops", { params });
      return response.data;
    },
  });
};

// Licenses API
export const useLicenses = (params = {}) => {
  return useQuery({
    queryKey: ["licenses", params],
    queryFn: async () => {
      const response = await api.get("/licenses", { params });
      return response.data;
    },
  });
};

export const useLicense = (id) => {
  return useQuery({
    queryKey: ["licenses", id],
    queryFn: async () => {
      const response = await api.get(`/licenses/${id}`);
      return response.data;
    },
    enabled: !!id,
  });
};

export const useCreateLicense = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (licenseData) => {
      const response = await api.post("/licenses", licenseData);
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["licenses"] });
      toast.success("License created successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to create license";
      toast.error(message);
    },
  });
};

export const useDeleteLicense = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id) => {
      const response = await api.delete(`/licenses/${id}`);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["licenses"] });
      toast.success("License deleted successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to delete license";
      toast.error(message);
    },
  });
};

// Customers API
export const useCustomers = (params = {}) => {
  return useQuery({
    queryKey: ["customers", params],
    queryFn: async () => {
      const response = await api.get("/customers", { params });
      return response.data;
    },
  });
};

export const useCustomer = (id) => {
  return useQuery({
    queryKey: ["customers", id],
    queryFn: async () => {
      const response = await api.get(`/customers/${id}`);
      return response.data;
    },
    enabled: !!id,
  });
};

export const useCreateCustomer = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (customerData) => {
      const response = await api.post("/customers", customerData);
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["customers"] });
      toast.success("Customer created successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to create customer";
      toast.error(message);
    },
  });
};

export const useUpdateCustomer = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, ...customerData }) => {
      const response = await api.put(`/customers/${id}`, customerData);
      return response.data;
    },
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries({ queryKey: ["customers"] });
      queryClient.invalidateQueries({ queryKey: ["customers", variables.id] });
      toast.success("Customer updated successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to update customer";
      toast.error(message);
    },
  });
};

export const useDeleteCustomer = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id) => {
      const response = await api.delete(`/customers/${id}`);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["customers"] });
      toast.success("Customer deleted successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to delete customer";
      toast.error(message);
    },
  });
};

// Users API
export const useUsers = (params = {}) => {
  return useQuery({
    queryKey: ["users", params],
    queryFn: async () => {
      const response = await api.get("/users", { params });
      return response.data;
    },
  });
};

export const useUser = (id) => {
  return useQuery({
    queryKey: ["users", id],
    queryFn: async () => {
      const response = await api.get(`/users/${id}`);
      return response.data;
    },
    enabled: !!id,
  });
};

export const useCreateUser = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (userData) => {
      const response = await api.post("/users", userData);
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      toast.success("User created successfully");
    },
    onError: (error) => {
      const message = error.response?.data?.message || "Failed to create user";
      toast.error(message);
    },
  });
};

export const useUpdateUser = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, ...userData }) => {
      const response = await api.put(`/users/${id}`, userData);
      return response.data;
    },
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      queryClient.invalidateQueries({ queryKey: ["users", variables.id] });
      toast.success("User updated successfully");
    },
    onError: (error) => {
      const message = error.response?.data?.message || "Failed to update user";
      toast.error(message);
    },
  });
};

export const useDeleteUser = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id) => {
      const response = await api.delete(`/users/${id}`);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      toast.success("User deleted successfully");
    },
    onError: (error) => {
      const message = error.response?.data?.message || "Failed to delete user";
      toast.error(message);
    },
  });
};

// Roles API
export const useRoles = (params = {}) => {
  return useQuery({
    queryKey: ["roles", params],
    queryFn: async () => {
      const response = await api.get("/roles", { params });
      return response.data;
    },
  });
};

export const useRole = (id) => {
  return useQuery({
    queryKey: ["roles", id],
    queryFn: async () => {
      const response = await api.get(`/roles/${id}`);
      return response.data;
    },
    enabled: !!id,
  });
};

export const useRoleByName = (name) => {
  return useQuery({
    queryKey: ["roles", "name", name],
    queryFn: async () => {
      const response = await api.get(`/roles/name/${name}`);
      return response.data;
    },
    enabled: !!name,
  });
};

// Permissions API (Casbin policies)
export const usePermissions = (params = {}) => {
  return useQuery({
    queryKey: ["permissions", params],
    queryFn: async () => {
      const response = await api.get("/permissions", { params });
      return response.data;
    },
  });
};

export const useRolePermissions = (roleId, domain) => {
  return useQuery({
    queryKey: ["permissions", "role", roleId, domain],
    queryFn: async () => {
      const response = await api.get(`/permissions/role/${roleId}`, {
        params: { domain },
      });
      return response.data;
    },
    enabled: !!roleId && !!domain,
  });
};

export const useCreatePermission = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (permissionData) => {
      const response = await api.post("/permissions", permissionData);
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["permissions"] });
      toast.success("Permission granted successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to grant permission";
      toast.error(message);
    },
  });
};

export const useDeletePermission = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id) => {
      const response = await api.delete(`/permissions/${id}`);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["permissions"] });
      toast.success("Permission revoked successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to revoke permission";
      toast.error(message);
    },
  });
};

export const useBulkUpdatePermissions = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ subject, domain, permissions }) => {
      const response = await api.put("/permissions/bulk", {
        subject,
        domain,
        permissions,
      });
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["permissions"] });
      toast.success("Permissions updated successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to update permissions";
      toast.error(message);
    },
  });
};

// Role Assignments API (Casbin grouping)
export const useRoleAssignments = (params = {}) => {
  return useQuery({
    queryKey: ["role-assignments", params],
    queryFn: async () => {
      const response = await api.get("/role-assignments", { params });
      return response.data;
    },
  });
};

export const useUserRoles = (userId, domain) => {
  return useQuery({
    queryKey: ["role-assignments", "user", userId, domain],
    queryFn: async () => {
      const response = await api.get(`/role-assignments/user/${userId}`, {
        params: { domain },
      });
      return response.data;
    },
    enabled: !!userId && !!domain,
  });
};

export const useAssignRole = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (assignmentData) => {
      const response = await api.post("/role-assignments", assignmentData);
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["role-assignments"] });
      toast.success("Role assigned successfully");
    },
    onError: (error) => {
      const message = error.response?.data?.message || "Failed to assign role";
      toast.error(message);
    },
  });
};

export const useUnassignRole = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id) => {
      const response = await api.delete(`/role-assignments/${id}`);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["role-assignments"] });
      toast.success("Role unassigned successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to unassign role";
      toast.error(message);
    },
  });
};

// Domains API
export const useDomains = (params = {}) => {
  return useQuery({
    queryKey: ["domains", params],
    queryFn: async () => {
      const response = await api.get("/domains", { params });
      return response.data;
    },
  });
};

export const useCreateDomain = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (domainData) => {
      const response = await api.post("/domains", domainData);
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["domains"] });
      toast.success("Domain created successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to create domain";
      toast.error(message);
    },
  });
};

// Permission check API (for UI authorization)
export const useHasPermission = (object, action, domain) => {
  return useQuery({
    queryKey: ["permission-check", object, action, domain],
    queryFn: async () => {
      const response = await api.get("/permissions/check", {
        params: { object, action, domain },
      });
      return response.data?.allowed || false;
    },
    enabled: !!object && !!action && !!domain,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};
