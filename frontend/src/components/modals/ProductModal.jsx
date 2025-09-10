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
      barcode: "",
      sale: 0,
      buy: 0,
      stock: 0,
      unit: "pcs",
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

  const unitOptions = [
    { value: "pcs", label: "Pieces" },
    { value: "kg", label: "Kilograms" },
    { value: "gram", label: "Grams" },
    { value: "liter", label: "Liters" },
    { value: "ml", label: "Milliliters" },
    { value: "bottle", label: "Bottles" },
    { value: "box", label: "Boxes" },
    { value: "pack", label: "Packs" },
    { value: "portion", label: "Portions" },
  ];

  return (
    <CrudModal
      isOpen={isOpen}
      onClose={onClose}
      title={isEditing ? "Edit Product" : "Add New Product"}
      description={
        isEditing ? "Update product information" : "Create a new product"
      }
      onSubmit={handleSubmit(onSubmit)}
      isLoading={isLoading}
      submitText={isEditing ? "Update Product" : "Create Product"}
      size="lg"
    >
      <div className="space-y-4">
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
            name="barcode"
            label="Barcode"
            placeholder="Enter barcode"
            register={register}
            error={errors.barcode}
            disabled={isLoading}
          />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormField
            name="sale"
            label="Sale Price"
            type="number"
            step="0.01"
            placeholder="0.00"
            register={register}
            error={errors.sale}
            disabled={isLoading}
          />

          <FormField
            name="buy"
            label="Buy Price"
            type="number"
            step="0.01"
            placeholder="0.00"
            register={register}
            error={errors.buy}
            disabled={isLoading}
          />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormField
            name="stock"
            label="Stock Quantity"
            type="number"
            placeholder="0"
            register={register}
            error={errors.stock}
            disabled={isLoading}
          />

          <FormField
            name="unit"
            label="Unit"
            type="select"
            options={unitOptions}
            register={register}
            error={errors.unit}
            disabled={isLoading}
          />
        </div>

        <div className="bg-blue-50 dark:bg-blue-900/20 p-4 rounded-md">
          <p className="text-sm text-blue-800 dark:text-blue-200">
            <strong>Note:</strong> Make sure to set appropriate sale and buy
            prices. Stock quantity will be tracked automatically.
          </p>
        </div>
      </div>
    </CrudModal>
  );
};

export default ProductModal;
