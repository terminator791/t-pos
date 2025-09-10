import React, { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import CrudModal from "@/components/ui/CrudModal";
import FormField from "@/components/ui/FormField";
import { CreateUserSchema, UpdateUserSchema } from "@/lib/schemas";
import {
  useCreateUser,
  useUpdateUser,
  useRoles,
  useLicenses,
} from "@/services/api";

const UserModal = ({ isOpen, onClose, user = null, isEditing = false }) => {
  const createUser = useCreateUser();
  const updateUser = useUpdateUser();
  const { data: rolesData } = useRoles();
  const { data: licensesData } = useLicenses();

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
      pin: "",
      role_id: "",
      serial_number: "",
    },
  });

  // Load user data when editing
  useEffect(() => {
    if (isEditing && user) {
      Object.keys(user).forEach((key) => {
        if (key !== "pin") {
          // Don't pre-fill PIN for security
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
        // Remove pin if empty for update
        if (!data.pin) {
          delete data.pin;
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

  // Get admin roles from API
  const roles = rolesData?.data?.roles || [];
  const adminRoles = roles.filter(
    (role) => role.name === "admin" || role.name === "super_admin"
  );

  const roleOptions = adminRoles.map((role) => ({
    value: role.id,
    label: role.display_name || role.name,
  }));

  // Get licenses from API
  const licenses = licensesData?.data?.licenses || [];
  const licenseOptions = licenses.map((license) => ({
    value: license.serial_number,
    label: license.serial_number,
  }));

  return (
    <CrudModal
      isOpen={isOpen}
      onClose={onClose}
      title={isEditing ? "Edit User" : "Add New User"}
      description={
        isEditing ? "Update user information" : "Create a new admin user"
      }
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
          name="pin"
          label={isEditing ? "PIN (leave blank to keep current)" : "PIN"}
          type="password"
          placeholder={
            isEditing
              ? "Enter new PIN (6 characters minimum)"
              : "Enter PIN (6 characters minimum)"
          }
          register={register}
          error={errors.pin}
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
          type="select"
          options={licenseOptions}
          register={register}
          error={errors.serial_number}
          disabled={isLoading}
        />

        <div className="bg-yellow-50 dark:bg-yellow-900/20 p-4 rounded-md">
          <p className="text-sm text-yellow-800 dark:text-yellow-200">
            <strong>Note:</strong> Admin users have elevated privileges and can
            manage other users and system settings.
          </p>
        </div>
      </div>
    </CrudModal>
  );
};

export default UserModal;
