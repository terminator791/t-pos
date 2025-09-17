import React, { useState, useEffect } from "react";
import Modal from "@/components/ui/Modal";
import Button from "@/components/ui/Button";
import Textinput from "@/components/ui/Textinput";
import Select from "@/components/ui/Select";
import { useForm } from "react-hook-form";
import { yupResolver } from "@hookform/resolvers/yup";
import * as yup from "yup";
import { useCreateCategory, useUpdateCategory, useShops } from "@/services/api";

const schema = yup.object({
  name: yup.string().required("Category name is required"),
  description: yup.string(),
  shop_id: yup.string().required("Shop is required"),
});

const CategoryModal = ({ isOpen, onClose, category, isEditing = false }) => {
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  const createCategory = useCreateCategory();
  const updateCategory = useUpdateCategory();
  const { data: shopsData } = useShops();

  const shops = shopsData?.data?.shops || [];
  const shopOptions = shops.map(shop => ({
    value: shop.id,
    label: shop.name,
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
      if (isEditing && category) {
        // Pre-fill form for editing
        setValue("name", category.name || "");
        setValue("description", category.description || "");
        setValue("shop_id", category.shop_id || "");
      } else {
        // Reset form for new category
        reset();
      }
    }
  }, [isOpen, isEditing, category, reset, setValue]);

  const onSubmit = async (data) => {
    setIsSubmitting(true);
    try {
      if (isEditing && category) {
        await updateCategory.mutateAsync({
          id: category.id,
          ...data,
        });
      } else {
        await createCategory.mutateAsync(data);
      }
      onClose();
      reset();
    } catch (error) {
      console.error("Failed to save category:", error);
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
      title={isEditing ? "Edit Category" : "Add New Category"}
      activeModal={isOpen}
      onClose={handleClose}
      className="max-w-lg"
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <Textinput
          name="name"
          label="Category Name"
          placeholder="Enter category name"
          register={register}
          error={errors.name}
          className="relative"
        />

        <div>
          <label className="form-label">Description</label>
          <textarea
            {...register("description")}
            placeholder="Enter category description (optional)"
            className="form-control"
            rows="3"
          />
          {errors.description && (
            <div className="mt-2 text-danger-500 block text-sm">
              {errors.description?.message}
            </div>
          )}
        </div>

        <Select
          name="shop_id"
          label="Shop"
          placeholder="Select shop"
          register={register}
          error={errors.shop_id}
          options={shopOptions}
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
                ? "Update Category"
                : "Create Category"
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

export default CategoryModal;