import React, { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import CrudModal from "@/components/ui/CrudModal";
import FormField from "@/components/ui/FormField";
import { CreateCustomerSchema, UpdateCustomerSchema } from "@/lib/schemas";
import {
  useCreateCustomer,
  useUpdateCustomer,
  useRoles,
  useLicenses,
} from "@/services/api";

const CustomerModal = ({
  isOpen,
  onClose,
  customer = null,
  isEditing = false,
}) => {
  const createCustomer = useCreateCustomer();
  const updateCustomer = useUpdateCustomer();
  const { data: rolesData } = useRoles();
  const { data: licensesData } = useLicenses();

  const schema = isEditing ? UpdateCustomerSchema : CreateCustomerSchema;

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

  // Load customer data when editing
  useEffect(() => {
    if (isEditing && customer) {
      Object.keys(customer).forEach((key) => {
        if (key !== "pin") {
          // Don't pre-fill PIN for security
          setValue(key, customer[key]);
        }
      });
    } else {
      reset();
    }
  }, [isEditing, customer, setValue, reset]);

  const onSubmit = async (data) => {
    try {
      if (isEditing) {
        // Remove pin if empty for update
        if (!data.pin) {
          delete data.pin;
        }
        await updateCustomer.mutateAsync({ id: customer.id, ...data });
      } else {
        await createCustomer.mutateAsync(data);
      }
      reset();
      onClose();
    } catch (error) {
      console.error("Customer operation failed:", error);
    }
  };

  const isLoading = createCustomer.isPending || updateCustomer.isPending;

  // Get customer roles from API
  const roles = rolesData?.data?.roles || [];
  const customerRoles = roles.filter(
    (role) => role.name === "cashier" || role.name === "owner_business"
  );

  const roleOptions = customerRoles.map((role) => ({
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
      title={isEditing ? "Edit Customer" : "Add New Customer"}
      description={
        isEditing ? "Update customer information" : "Create a new customer"
      }
      onSubmit={handleSubmit(onSubmit)}
      isLoading={isLoading}
      submitText={isEditing ? "Update Customer" : "Create Customer"}
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
          label="PIN"
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

        <div className="bg-blue-50 dark:bg-blue-900/20 p-4 rounded-md">
          <p className="text-sm text-blue-800 dark:text-blue-200">
            <strong>Note:</strong> Customers are users with cashier or business
            owner roles who can access the POS system.
          </p>
        </div>
      </div>
    </CrudModal>
  );
};

export default CustomerModal;
