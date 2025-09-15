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

export const useCategory = (id) => {
  return useQuery({
    queryKey: ["categories", id],
    queryFn: async () => {
      const response = await api.get(`/categories/${id}`);
      return response.data;
    },
    enabled: !!id,
  });
};

export const useCreateCategory = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (categoryData) => {
      const response = await api.post("/categories", categoryData);
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["categories"] });
      toast.success("Category created successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to create category";
      toast.error(message);
    },
  });
};

export const useUpdateCategory = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, ...categoryData }) => {
      const response = await api.put(`/categories/${id}`, categoryData);
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["categories"] });
      toast.success("Category updated successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to update category";
      toast.error(message);
    },
  });
};

export const useDeleteCategory = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (categoryId) => {
      const response = await api.delete(`/categories/${categoryId}`);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["categories"] });
      toast.success("Category deleted successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to delete category";
      toast.error(message);
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

export const useShop = (id) => {
  return useQuery({
    queryKey: ["shops", id],
    queryFn: async () => {
      const response = await api.get(`/shops/${id}`);
      return response.data;
    },
    enabled: !!id,
  });
};

export const useCreateShop = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (shopData) => {
      const response = await api.post("/shops", shopData);
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["shops"] });
      toast.success("Shop created successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to create shop";
      toast.error(message);
    },
  });
};

export const useUpdateShop = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, ...shopData }) => {
      const response = await api.put(`/shops/${id}`, shopData);
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["shops"] });
      toast.success("Shop updated successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to update shop";
      toast.error(message);
    },
  });
};

export const useDeleteShop = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (shopId) => {
      const response = await api.delete(`/shops/${shopId}`);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["shops"] });
      toast.success("Shop deleted successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to delete shop";
      toast.error(message);
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
      try {
        const response = await api.get("/permissions", { params });
        return response.data;
      } catch (error) {
        // Return fallback data if endpoint doesn't exist
        if (error.response?.status === 404) {
          console.warn("Permissions endpoint not found, using fallback data");
          return { data: [] };
        }
        throw error;
      }
    },
    retry: 1,
    retryDelay: 1000,
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
      try {
        const response = await api.get("/role-assignments", { params });
        return response.data;
      } catch (error) {
        // Return fallback data if endpoint doesn't exist
        if (error.response?.status === 404) {
          console.warn(
            "Role assignments endpoint not found, using fallback data"
          );
          return { data: [] };
        }
        throw error;
      }
    },
    retry: 1,
    retryDelay: 1000,
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
      try {
        const response = await api.get("/domains", { params });
        return response.data;
      } catch (error) {
        // Return fallback data if endpoint doesn't exist
        if (error.response?.status === 404) {
          console.warn("Domains endpoint not found, using fallback data");
          return {
            data: [
              { id: "global", name: "*" },
              { id: "demo", name: "LIC-001-DEMO" },
            ],
          };
        }
        throw error;
      }
    },
    retry: 1,
    retryDelay: 1000,
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
      try {
        const response = await api.get("/permissions/check", {
          params: { object, action, domain },
        });
        return response.data?.allowed || false;
      } catch (error) {
        // Return fallback permission for development
        if (error.response?.status === 404) {
          console.warn(
            "Permission check endpoint not found, allowing access for development"
          );
          return true;
        }
        throw error;
      }
    },
    enabled: !!object && !!action && !!domain,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: 1,
    retryDelay: 1000,
  });
};

// ACL Management API (matches backend routes)
// Policy management
export const useGetAllPolicies = () => {
  return useQuery({
    queryKey: ["acl", "policies"],
    queryFn: async () => {
      const response = await api.get("/acl/policies");
      return response.data;
    },
  });
};

export const useAddPolicy = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (policyData) => {
      const response = await api.post("/acl/policies", policyData);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["acl", "policies"] });
      toast.success("Policy added successfully");
    },
    onError: (error) => {
      const message = error.response?.data?.message || "Failed to add policy";
      toast.error(message);
    },
  });
};

export const useRemovePolicy = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (policyData) => {
      const response = await api.delete("/acl/policies", { data: policyData });
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["acl", "policies"] });
      toast.success("Policy removed successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to remove policy";
      toast.error(message);
    },
  });
};

export const useGetSystemPolicies = () => {
  return useQuery({
    queryKey: ["acl", "policies", "system"],
    queryFn: async () => {
      const response = await api.get("/acl/policies/system");
      return response.data;
    },
  });
};

// Role assignment management
export const useGetAllACLRoles = () => {
  return useQuery({
    queryKey: ["acl", "roles"],
    queryFn: async () => {
      const response = await api.get("/acl/roles");
      return response.data;
    },
  });
};

export const useGetSystemRoles = () => {
  return useQuery({
    queryKey: ["acl", "roles", "system"],
    queryFn: async () => {
      const response = await api.get("/acl/roles/system");
      return response.data;
    },
  });
};

export const useGetUserRoles = (userId) => {
  return useQuery({
    queryKey: ["acl", "users", userId, "roles"],
    queryFn: async () => {
      const response = await api.get(`/acl/users/${userId}/roles`);
      return response.data;
    },
    enabled: !!userId,
  });
};

export const useGetRoleUsers = (role) => {
  return useQuery({
    queryKey: ["acl", "roles", role, "users"],
    queryFn: async () => {
      const response = await api.get(`/acl/roles/${role}/users`);
      return response.data;
    },
    enabled: !!role,
  });
};

export const useAddRoleForUser = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (roleData) => {
      const response = await api.post("/acl/users/roles", roleData);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["acl"] });
      toast.success("Role assigned successfully");
    },
    onError: (error) => {
      const message = error.response?.data?.message || "Failed to assign role";
      toast.error(message);
    },
  });
};

export const useRemoveRoleForUser = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (roleData) => {
      const response = await api.delete("/acl/users/roles", { data: roleData });
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["acl"] });
      toast.success("Role removed successfully");
    },
    onError: (error) => {
      const message = error.response?.data?.message || "Failed to remove role";
      toast.error(message);
    },
  });
};

// Permission checking and reloading
export const useCheckPermission = () => {
  return useMutation({
    mutationFn: async (permissionData) => {
      const response = await api.post("/acl/check", permissionData);
      return response.data;
    },
  });
};

export const useReloadPolicies = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const response = await api.post("/acl/reload");
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["acl"] });
      queryClient.invalidateQueries({ queryKey: ["permission-check"] });
      toast.success("Policies reloaded successfully");
    },
    onError: (error) => {
      const message =
        error.response?.data?.message || "Failed to reload policies";
      toast.error(message);
    },
  });
};

// Role management hooks
export const useCreateRole = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (roleData) => {
      const response = await api.post("/roles", roleData);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["roles"] });
      queryClient.invalidateQueries({ queryKey: ["acl"] });
      toast.success("Role created successfully");
    },
    onError: (error) => {
      const message = error.response?.data?.message || "Failed to create role";
      toast.error(message);
    },
  });
};

export const useUpdateRole = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, ...roleData }) => {
      const response = await api.put(`/roles/${id}`, roleData);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["roles"] });
      queryClient.invalidateQueries({ queryKey: ["acl"] });
      toast.success("Role updated successfully");
    },
    onError: (error) => {
      const message = error.response?.data?.message || "Failed to update role";
      toast.error(message);
    },
  });
};

// Additional helper hooks for the roles page
export const useDeleteRole = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (roleId) => {
      // Note: Backend doesn't have delete role endpoint, this is a placeholder
      // You may need to implement this in the backend or use ACL policies removal instead
      throw new Error("Role deletion not implemented in backend");
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["roles"] });
      queryClient.invalidateQueries({ queryKey: ["acl"] });
      toast.success("Role deleted successfully");
    },
    onError: (error) => {
      const message = error.response?.data?.message || "Failed to delete role";
      toast.error(message);
    },
  });
};

// Transaction Histories API
export const useTransactionHistories = (shopId, params = {}) => {
  return useQuery({
    queryKey: ["histories", shopId, params],
    queryFn: async () => {
      const endpoint = shopId ? `/histories/shop/${shopId}` : "/histories";
      const response = await api.get(endpoint, { params });
      return response.data;
    },
  });
};

export const useTransactionHistory = (id) => {
  return useQuery({
    queryKey: ["histories", id],
    queryFn: async () => {
      const response = await api.get(`/histories/${id}`);
      return response.data;
    },
    enabled: !!id,
  });
};

// Transactions API (for detailed transaction view)
export const useTransactions = (shopId, params = {}) => {
  return useQuery({
    queryKey: ["transactions", shopId, params],
    queryFn: async () => {
      const endpoint = shopId ? `/transactions/shop/${shopId}` : "/transactions";
      const response = await api.get(endpoint, { params });
      return response.data;
    },
  });
};

export const useTransaction = (id) => {
  return useQuery({
    queryKey: ["transactions", id],
    queryFn: async () => {
      const response = await api.get(`/transactions/${id}`);
      return response.data;
    },
    enabled: !!id,
  });
};

export const useTodaysTransactions = (shopId) => {
  return useQuery({
    queryKey: ["transactions", "today", shopId],
    queryFn: async () => {
      const response = await api.get(`/transactions/shop/${shopId}/today`);
      return response.data;
    },
    enabled: !!shopId,
  });
};

export const useTransactionsByStatus = (shopId, status) => {
  return useQuery({
    queryKey: ["transactions", "shop", shopId, "status", status],
    queryFn: async () => {
      const response = await api.get(`/transactions/shop/${shopId}/status/${status}`);
      return response.data;
    },
    enabled: !!shopId && !!status,
  });
};

// Transaction Products API (for transaction details)
export const useTransactionProducts = (transactionId) => {
  return useQuery({
    queryKey: ["transaction-products", transactionId],
    queryFn: async () => {
      const response = await api.get(`/transaction-products/transaction/${transactionId}`);
      return response.data;
    },
    enabled: !!transactionId,
  });
};

export const useTransactionProductsByShop = (shopId, params = {}) => {
  return useQuery({
    queryKey: ["transaction-products", "shop", shopId, params],
    queryFn: async () => {
      const response = await api.get(`/transaction-products/shop/${shopId}`, { params });
      return response.data;
    },
    enabled: !!shopId,
  });
};

// Payments API (for transaction payment details)
export const usePayments = (shopId, params = {}) => {
  return useQuery({
    queryKey: ["payments", shopId, params],
    queryFn: async () => {
      const endpoint = shopId ? `/payments/shop/${shopId}` : "/payments";
      const response = await api.get(endpoint, { params });
      return response.data;
    },
  });
};

export const usePayment = (id) => {
  return useQuery({
    queryKey: ["payments", id],
    queryFn: async () => {
      const response = await api.get(`/payments/${id}`);
      return response.data;
    },
    enabled: !!id,
  });
};

export const usePaymentsByStatus = (shopId, status) => {
  return useQuery({
    queryKey: ["payments", "shop", shopId, "status", status],
    queryFn: async () => {
      const response = await api.get(`/payments/shop/${shopId}/status/${status}`);
      return response.data;
    },
    enabled: !!shopId && !!status,
  });
};

// Cart Management API
export const useCarts = () => {
  return useQuery({
    queryKey: ["carts"],
    queryFn: async () => {
      const response = await api.get("/carts/all");
      return response.data;
    },
  });
};

export const useCart = (id) => {
  return useQuery({
    queryKey: ["carts", id],
    queryFn: async () => {
      const response = await api.get(`/carts/${id}`);
      return response.data;
    },
    enabled: !!id,
  });
};

export const useAddToCart = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (cartData) => {
      const response = await api.post("/carts", cartData);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["carts"] });
      toast.success("Product added to cart");
    },
    onError: (error) => {
      const message = error.response?.data?.message || "Failed to add product to cart";
      toast.error(message);
    },
  });
};

export const useUpdateCartItem = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, ...cartData }) => {
      const response = await api.put(`/carts/${id}`, cartData);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["carts"] });
      toast.success("Cart updated");
    },
    onError: (error) => {
      const message = error.response?.data?.message || "Failed to update cart item";
      toast.error(message);
    },
  });
};

export const useRemoveFromCart = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id) => {
      const response = await api.delete(`/carts/${id}`);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["carts"] });
      toast.success("Item removed from cart");
    },
    onError: (error) => {
      const message = error.response?.data?.message || "Failed to remove item from cart";
      toast.error(message);
    },
  });
};

export const useClearCart = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const response = await api.delete("/carts");
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["carts"] });
      toast.success("Cart cleared");
    },
    onError: (error) => {
      const message = error.response?.data?.message || "Failed to clear cart";
      toast.error(message);
    },
  });
};

// Transaction Creation and Management API
export const useCreateTransaction = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (transactionData) => {
      const response = await api.post("/transactions", transactionData);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["transactions"] });
      queryClient.invalidateQueries({ queryKey: ["carts"] });
      toast.success("Transaction created successfully");
    },
    onError: (error) => {
      const message = error.response?.data?.message || "Failed to create transaction";
      toast.error(message);
    },
  });
};

export const usePayTransaction = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ transactionId, amount }) => {
      const response = await api.post(`/transactions/${transactionId}/pay`, { amount });
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["transactions"] });
      queryClient.invalidateQueries({ queryKey: ["payments"] });
      toast.success("Payment processed successfully");
    },
    onError: (error) => {
      const message = error.response?.data?.message || "Failed to process payment";
      toast.error(message);
    },
  });
};

export const useCancelTransaction = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (transactionId) => {
      const response = await api.post(`/transactions/${transactionId}/cancel`);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["transactions"] });
      toast.success("Transaction cancelled");
    },
    onError: (error) => {
      const message = error.response?.data?.message || "Failed to cancel transaction";
      toast.error(message);
    },
  });
};
