import React, { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import CrudModal from "@/components/ui/CrudModal";
import FormField from "@/components/ui/FormField";
import { CreateProductSchema, UpdateProductSchema } from "@/lib/schemas";
import { useCreateProduct, useUpdateProduct } from "@/services/api";

const ProductModal = ({
  isOpen,
  onClose,
  product = null,
  isEditing = false,
}) => {
  const createProduct = useCreateProduct();
  const updateProduct = useUpdateProduct();

  const schema = isEditing ? UpdateProductSchema : CreateProductSchema;
  
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
      sku: "",
      category_id: "",
      price: 0,
      cost: 0,
      stock_quantity: 0,
      min_stock_level: 0,
      unit: "",
      status: "active",
    },
  });

  // Load product data when editing
  useEffect(() => {
    if (isEditing && product) {
      Object.keys(product).forEach((key) => {
        setValue(key, product[key]);
      });
    } else {
      reset();
    }
  }, [isEditing, product, setValue, reset]);

  const onSubmit = async (data) => {
    try {
      if (isEditing) {
        await updateProduct.mutateAsync({ id: product.id, ...data });
      } else {
        await createProduct.mutateAsync(data);
      }
      reset();
      onClose();
    } catch (error) {
      console.error("Product operation failed:", error);
    }
  };

  const isLoading = createProduct.isPending || updateProduct.isPending;

  // Mock categories - in real app, fetch from API
  const categoryOptions = [
    { value: "electronics", label: "Electronics" },
    { value: "clothing", label: "Clothing" },
    { value: "home-garden", label: "Home & Garden" },
    { value: "books", label: "Books" },
    { value: "sports", label: "Sports" },
  ];

  const statusOptions = [
    { value: "active", label: "Active" },
    { value: "inactive", label: "Inactive" },
  ];

  return (
    <CrudModal
      isOpen={isOpen}
      onClose={onClose}
      title={isEditing ? "Edit Product" : "Add New Product"}
      description={isEditing ? "Update product information" : "Create a new product"}
      onSubmit={handleSubmit(onSubmit)}
      isLoading={isLoading}
      submitText={isEditing ? "Update Product" : "Create Product"}
      size="lg"
    >
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <FormField
          name="name"
          label="Product Name"
          placeholder="Enter product name"
          register={register}
          error={errors.name}
          disabled={isLoading}
        />

        <FormField
          name="sku"
          label="SKU"
          placeholder="Enter product SKU"
          register={register}
          error={errors.sku}
          disabled={isLoading}
        />

        <FormField
          name="category_id"
          label="Category"
          type="select"
          placeholder="Select category"
          options={categoryOptions}
          register={register}
          error={errors.category_id}
          disabled={isLoading}
        />

        <FormField
          name="status"
          label="Status"
          type="select"
          options={statusOptions}
          register={register}
          error={errors.status}
          disabled={isLoading}
        />

        <FormField
          name="price"
          label="Price"
          type="number"
          placeholder="0.00"
          register={register}
          error={errors.price}
          disabled={isLoading}
        />

        <FormField
          name="cost"
          label="Cost"
          type="number"
          placeholder="0.00"
          register={register}
          error={errors.cost}
          disabled={isLoading}
        />

        <FormField
          name="stock_quantity"
          label="Stock Quantity"
          type="number"
          placeholder="0"
          register={register}
          error={errors.stock_quantity}
          disabled={isLoading}
        />

        <FormField
          name="min_stock_level"
          label="Minimum Stock Level"
          type="number"
          placeholder="0"
          register={register}
          error={errors.min_stock_level}
          disabled={isLoading}
        />

        <FormField
          name="unit"
          label="Unit"
          placeholder="e.g., pcs, kg, liter"
          register={register}
          error={errors.unit}
          disabled={isLoading}
        />
      </div>

      <FormField
        name="description"
        label="Description"
        type="textarea"
        placeholder="Enter product description"
        register={register}
        error={errors.description}
        disabled={isLoading}
      />
    </CrudModal>
  );
};

export default ProductModal;