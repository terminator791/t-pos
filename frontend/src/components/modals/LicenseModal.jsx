import React, { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import CrudModal from "@/components/ui/CrudModal";
import FormField from "@/components/ui/FormField";
import { CreateLicenseSchema } from "@/lib/schemas";
import { useCreateLicense } from "@/services/api";

const LicenseModal = ({
  isOpen,
  onClose,
  license = null,
  isEditing = false,
}) => {
  const createLicense = useCreateLicense();

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(CreateLicenseSchema),
    defaultValues: {
      serial_number: "",
    },
  });

  // Load license data when editing (though licenses typically can't be edited)
  useEffect(() => {
    if (isEditing && license) {
      setValue("serial_number", license.serial_number);
    } else {
      reset();
    }
  }, [isEditing, license, setValue, reset]);

  const onSubmit = async (data) => {
    try {
      await createLicense.mutateAsync(data);
      reset();
      onClose();
    } catch (error) {
      console.error("License operation failed:", error);
    }
  };

  const isLoading = createLicense.isPending;

  return (
    <CrudModal
      isOpen={isOpen}
      onClose={onClose}
      title={isEditing ? "View License" : "Add New License"}
      description={isEditing ? "License information (read-only)" : "Create a new license"}
      onSubmit={handleSubmit(onSubmit)}
      isLoading={isLoading}
      submitText={isEditing ? "Close" : "Create License"}
      size="md"
    >
      <FormField
        name="serial_number"
        label="Serial Number"
        placeholder="Enter license serial number"
        register={register}
        error={errors.serial_number}
        disabled={isLoading || isEditing}
      />
      
      {isEditing && (
        <div className="bg-gray-50 dark:bg-gray-800 p-4 rounded-md">
          <p className="text-sm text-gray-600 dark:text-gray-400">
            <strong>Note:</strong> License serial numbers cannot be modified after creation for security reasons.
          </p>
        </div>
      )}
    </CrudModal>
  );
};

export default LicenseModal;