import React, { useState, useEffect } from "react";
import Modal from "@/components/ui/Modal";
import Button from "@/components/ui/Button";
import Textinput from "@/components/ui/Textinput";
import Select from "@/components/ui/Select";
import { useForm } from "react-hook-form";
import { yupResolver } from "@hookform/resolvers/yup";
import * as yup from "yup";
import { useCreateShop, useUpdateShop, useLicenses } from "@/services/api";

const schema = yup.object({
  name: yup.string().required("Shop name is required"),
  description: yup.string(),
  address: yup.string(),
  phone: yup.string(),
  license_id: yup.string().required("License is required"),
});

const ShopModal = ({ isOpen, onClose, shop, isEditing = false }) => {
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  const createShop = useCreateShop();
  const updateShop = useUpdateShop();
  const { data: licensesData } = useLicenses();

  const licenses = licensesData?.data?.licenses || [];
  const licenseOptions = licenses.map(license => ({
    value: license.id,
    label: `${license.serial_number} (${license.license_type})`,
  }));

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
    setValue,
    watch,
  } = useForm({
    resolver: yupResolver(schema),
    mode: "all",
  });

  useEffect(() => {
    if (isOpen) {
      if (isEditing && shop) {
        // Pre-fill form for editing
        setValue("name", shop.name || "");
        setValue("description", shop.description || "");
        setValue("address", shop.address || "");
        setValue("phone", shop.phone || "");
        setValue("license_id", shop.license_id || "");
      } else {
        // Reset form for new shop
        reset();
      }
    }
  }, [isOpen, isEditing, shop, reset, setValue]);

  const onSubmit = async (data) => {
    setIsSubmitting(true);
    try {
      if (isEditing && shop) {
        await updateShop.mutateAsync({
          id: shop.id,
          ...data,
        });
      } else {
        await createShop.mutateAsync(data);
      }
      onClose();
      reset();
    } catch (error) {
      console.error("Failed to save shop:", error);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleClose = () => {
    onClose();
    reset();
  };

  return (
    <Modal
      title={isEditing ? "Edit Shop" : "Add New Shop"}
      activeModal={isOpen}
      onClose={handleClose}
      className="max-w-lg"
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <Textinput
          name="name"
          label="Shop Name"
          placeholder="Enter shop name"
          register={register}
          error={errors.name}
          className="relative"
        />

        <div>
          <label className="form-label">Description</label>
          <textarea
            {...register("description")}
            placeholder="Enter shop description (optional)"
            className="form-control"
            rows="3"
          />
          {errors.description && (
            <div className="mt-2 text-danger-500 block text-sm">
              {errors.description?.message}
            </div>
          )}
        </div>

        <div>
          <label className="form-label">Address</label>
          <textarea
            {...register("address")}
            placeholder="Enter shop address (optional)"
            className="form-control"
            rows="3"
          />
          {errors.address && (
            <div className="mt-2 text-danger-500 block text-sm">
              {errors.address?.message}
            </div>
          )}
        </div>

        <Textinput
          name="phone"
          label="Phone"
          placeholder="Enter phone number (optional)"
          register={register}
          error={errors.phone}
          className="relative"
        />

        <Select
          name="license_id"
          label="License"
          placeholder="Select license"
          register={register}
          error={errors.license_id}
          options={licenseOptions}
        />

        <div className="flex justify-end space-x-3 pt-4">
          <Button
            text="Cancel"
            className="btn-outline-secondary"
            onClick={handleClose}
            type="button"
          />
          <Button
            text={
              isSubmitting
                ? isEditing
                  ? "Updating..."
                  : "Creating..."
                : isEditing
                ? "Update Shop"
                : "Create Shop"
            }
            type="submit"
            className="btn-primary"
            disabled={isSubmitting}
          />
        </div>
      </form>
    </Modal>
  );
};

export default ShopModal;