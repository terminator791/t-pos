import React, { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import CrudModal from "@/components/ui/CrudModal";
import FormField from "@/components/ui/FormField";
import { CreateRoleSchema, UpdateRoleSchema } from "@/lib/schemas";
import { useCreateRole, useUpdateRole, useDomains } from "@/services/api";

const RoleModal = ({
  isOpen,
  onClose,
  role = null,
  isEditing = false,
}) => {
  const createRole = useCreateRole();
  const updateRole = useUpdateRole();
  const { data: domains = [] } = useDomains();

  const schema = isEditing ? UpdateRoleSchema : CreateRoleSchema;
  
  const {
    register,
    handleSubmit,
    reset,
    setValue,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      description: "",
      domain: "",
    },
  });

  // Load role data when editing
  useEffect(() => {
    if (isEditing && role) {
      Object.keys(role).forEach((key) => {
        setValue(key, role[key]);
      });
    } else {
      reset();
    }
  }, [isEditing, role, setValue, reset]);

  const onSubmit = async (data) => {
    try {
      if (isEditing) {
        await updateRole.mutateAsync({ id: role.id, ...data });
      } else {
        await createRole.mutateAsync(data);
      }
      onClose();
      reset();
    } catch (error) {
      console.error("Error saving role:", error);
    }
  };

  const isLoading = createRole.isPending || updateRole.isPending;

  return (
    <CrudModal
      isOpen={isOpen}
      onClose={onClose}
      title={isEditing ? "Edit Role" : "Create New Role"}
      onSubmit={handleSubmit(onSubmit)}
      isLoading={isLoading}
      isEditing={isEditing}
    >
      <div className="space-y-4">
        <FormField
          label="Role Name"
          name="name"
          register={register}
          error={errors.name}
          placeholder="Enter role name"
          required
        />

        <FormField
          label="Description"
          name="description"
          register={register}
          error={errors.description}
          placeholder="Enter role description"
          type="textarea"
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

        <div className="bg-blue-50 dark:bg-blue-900/20 p-4 rounded-lg">
          <h4 className="text-sm font-medium text-blue-900 dark:text-blue-200 mb-2">
            Casbin RBAC Model
          </h4>
          <p className="text-xs text-blue-700 dark:text-blue-300">
            This role will be used in Casbin's RBAC model with domain support. 
            After creating the role, you can assign specific permissions (subject, domain, object, action) to it.
          </p>
        </div>
      </div>
    </CrudModal>
  );
};

export default RoleModal;