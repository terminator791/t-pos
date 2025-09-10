import React, { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import CrudModal from "@/components/ui/CrudModal";
import FormField from "@/components/ui/FormField";
import { CreatePermissionSchema } from "@/lib/schemas";
import { useCreatePermission, useDomains, useRoles } from "@/services/api";

// Common objects and actions for T-POS system
const SYSTEM_OBJECTS = [
  { value: "users", label: "Users" },
  { value: "products", label: "Products" },
  { value: "customers", label: "Customers" },
  { value: "licenses", label: "Licenses" },
  { value: "roles", label: "Roles" },
  { value: "permissions", label: "Permissions" },
  { value: "dashboard", label: "Dashboard" },
  { value: "reports", label: "Reports" },
  { value: "settings", label: "Settings" },
  { value: "admin", label: "Admin Panel" },
];

const SYSTEM_ACTIONS = [
  { value: "read", label: "Read/View" },
  { value: "write", label: "Create/Update" },
  { value: "delete", label: "Delete" },
  { value: "admin", label: "Admin Access" },
  { value: "*", label: "All Actions" },
];

const PermissionModal = ({
  isOpen,
  onClose,
  permission = null,
  defaultSubject = "",
  defaultDomain = "",
}) => {
  const createPermission = useCreatePermission();
  const { data: domains = [] } = useDomains();
  const { data: roles = [] } = useRoles();
  
  const {
    register,
    handleSubmit,
    reset,
    setValue,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(CreatePermissionSchema),
    defaultValues: {
      subject: defaultSubject,
      domain: defaultDomain,
      object: "",
      action: "",
    },
  });

  // Load permission data when editing
  useEffect(() => {
    if (permission) {
      Object.keys(permission).forEach((key) => {
        setValue(key, permission[key]);
      });
    } else {
      reset({
        subject: defaultSubject,
        domain: defaultDomain,
        object: "",
        action: "",
      });
    }
  }, [permission, defaultSubject, defaultDomain, setValue, reset]);

  const onSubmit = async (data) => {
    try {
      await createPermission.mutateAsync(data);
      onClose();
      reset();
    } catch (error) {
      console.error("Error creating permission:", error);
    }
  };

  const isLoading = createPermission.isPending;

  return (
    <CrudModal
      isOpen={isOpen}
      onClose={onClose}
      title="Grant Permission"
      onSubmit={handleSubmit(onSubmit)}
      isLoading={isLoading}
      isEditing={false}
    >
      <div className="space-y-4">
        <FormField
          label="Subject (Role/User)"
          name="subject"
          register={register}
          error={errors.subject}
          placeholder="Select or enter subject"
          type="select"
          options={[
            { value: "", label: "Select Subject" },
            ...roles.map(role => ({
              value: role.name,
              label: `Role: ${role.name}`
            })),
            { value: "custom", label: "Custom Subject..." }
          ]}
          required
        />

        <FormField
          label="Domain"
          name="domain"
          register={register}
          error={errors.domain}
          placeholder="Select domain"
          type="select"
          options={[
            { value: "", label: "Select Domain" },
            { value: "*", label: "All Domains" },
            ...domains.map(domain => ({
              value: domain.name,
              label: domain.name
            }))
          ]}
          required
        />

        <FormField
          label="Object (Resource)"
          name="object"
          register={register}
          error={errors.object}
          placeholder="Select object"
          type="select"
          options={[
            { value: "", label: "Select Object" },
            ...SYSTEM_OBJECTS,
            { value: "*", label: "All Objects" }
          ]}
          required
        />

        <FormField
          label="Action"
          name="action"
          register={register}
          error={errors.action}
          placeholder="Select action"
          type="select"
          options={[
            { value: "", label: "Select Action" },
            ...SYSTEM_ACTIONS
          ]}
          required
        />

        <div className="bg-green-50 dark:bg-green-900/20 p-4 rounded-lg">
          <h4 className="text-sm font-medium text-green-900 dark:text-green-200 mb-2">
            Casbin Policy Format
          </h4>
          <p className="text-xs text-green-700 dark:text-green-300 mb-2">
            This will create a Casbin policy with the format:
          </p>
          <code className="text-xs bg-green-100 dark:bg-green-800 px-2 py-1 rounded text-green-800 dark:text-green-200">
            p, subject, domain, object, action
          </code>
        </div>
      </div>
    </CrudModal>
  );
};

export default PermissionModal;