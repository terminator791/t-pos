import React, { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import CrudModal from "@/components/ui/CrudModal";
import FormField from "@/components/ui/FormField";
import { CreateCustomerSchema, UpdateCustomerSchema } from "@/lib/schemas";
import { useCreateCustomer, useUpdateCustomer } from "@/services/api";

const CustomerModal = ({
  isOpen,
  onClose,
  customer = null,
  isEditing = false,
}) => {
  const createCustomer = useCreateCustomer();
  const updateCustomer = useUpdateCustomer();

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
      role_id: "cashier",
      serial_number: "",
    },
  });

  // Load customer data when editing
  useEffect(() => {
    if (isEditing && customer) {
      Object.keys(customer).forEach((key) => {
        setValue(key, customer[key]);
      });
    } else {
      reset();
    }
  }, [isEditing, customer, setValue, reset]);

  const onSubmit = async (data) => {
    try {
      if (isEditing) {
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

  const roleOptions = [
    { value: "cashier", label: "Cashier" },
    { value: "owner_business", label: "Business Owner" },
  ];

  return (
    <CrudModal
      isOpen={isOpen}
      onClose={onClose}
      title={isEditing ? "Edit Customer" : "Add New Customer"}
      description={isEditing ? "Update customer information" : "Create a new customer"}
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
          placeholder="Enter PIN (minimum 4 characters)"
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
          placeholder="Enter license serial number"
          register={register}
          error={errors.serial_number}
          disabled={isLoading}
        />

        <div className="bg-blue-50 dark:bg-blue-900/20 p-4 rounded-md">
          <p className="text-sm text-blue-800 dark:text-blue-200">
            <strong>Note:</strong> Customers are users with cashier or business owner roles who can access the POS system.
          </p>
        </div>
      </div>
    </CrudModal>
  );
};

export default CustomerModal;