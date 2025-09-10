import React, { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import CrudModal from "@/components/ui/CrudModal";
import FormField from "@/components/ui/FormField";
import { CreateUserSchema, UpdateUserSchema } from "@/lib/schemas";
import { useCreateUser, useUpdateUser } from "@/services/api";

const UserModal = ({
  isOpen,
  onClose,
  user = null,
  isEditing = false,
}) => {
  const createUser = useCreateUser();
  const updateUser = useUpdateUser();

  const schema = isEditing ? UpdateUserSchema : CreateUserSchema;
  
  const {
    register,
    handleSubmit,
    reset,
    setValue,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(schema),
    defaultValues: {
      username: "",
      email: "",
      password: "",
      role_id: "admin",
      serial_number: "",
    },
  });

  // Load user data when editing
  useEffect(() => {
    if (isEditing && user) {
      Object.keys(user).forEach((key) => {
        if (key !== "password") { // Don't pre-fill password for security
          setValue(key, user[key]);
        }
      });
    } else {
      reset();
    }
  }, [isEditing, user, setValue, reset]);

  const onSubmit = async (data) => {
    try {
      if (isEditing) {
        // Remove password if empty for update
        if (!data.password) {
          delete data.password;
        }
        await updateUser.mutateAsync({ id: user.id, ...data });
      } else {
        await createUser.mutateAsync(data);
      }
      reset();
      onClose();
    } catch (error) {
      console.error("User operation failed:", error);
    }
  };

  const isLoading = createUser.isPending || updateUser.isPending;

  const roleOptions = [
    { value: "admin", label: "Admin" },
    { value: "super_admin", label: "Super Admin" },
  ];

  return (
    <CrudModal
      isOpen={isOpen}
      onClose={onClose}
      title={isEditing ? "Edit User" : "Add New User"}
      description={isEditing ? "Update user information" : "Create a new admin user"}
      onSubmit={handleSubmit(onSubmit)}
      isLoading={isLoading}
      submitText={isEditing ? "Update User" : "Create User"}
      size="md"
    >
      <div className="space-y-4">
        <FormField
          name="username"
          label="Username"
          placeholder="Enter username"
          register={register}
          error={errors.username}
          disabled={isLoading}
        />

        <FormField
          name="email"
          label="Email"
          type="email"
          placeholder="Enter email address"
          register={register}
          error={errors.email}
          disabled={isLoading}
        />

        <FormField
          name="password"
          label={isEditing ? "Password (leave blank to keep current)" : "Password"}
          type="password"
          placeholder={isEditing ? "Enter new password" : "Enter password"}
          register={register}
          error={errors.password}
          disabled={isLoading}
        />

        <FormField
          name="role_id"
          label="Role"
          type="select"
          options={roleOptions}
          register={register}
          error={errors.role_id}
          disabled={isLoading}
        />

        <FormField
          name="serial_number"
          label="License Serial Number"
          placeholder="Enter license serial number"
          register={register}
          error={errors.serial_number}
          disabled={isLoading}
        />

        <div className="bg-yellow-50 dark:bg-yellow-900/20 p-4 rounded-md">
          <p className="text-sm text-yellow-800 dark:text-yellow-200">
            <strong>Note:</strong> Admin users have elevated privileges and can manage other users and system settings.
          </p>
        </div>
      </div>
    </CrudModal>
  );
};

export default UserModal;