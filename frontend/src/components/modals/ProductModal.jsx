import React, { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import CrudModal from "@/components/ui/CrudModal";
import FormField from "@/components/ui/FormField";
import { CreateProductSchema, UpdateProductSchema } from "@/lib/schemas";
import {
  useCreateProduct,
  useUpdateProduct,
  useCreateProductWithFile,
  useCategories,
  useShops,
} from "@/services/api";

const ProductModal = ({
  isOpen,
  onClose,
  product = null,
  isEditing = false,
}) => {
  const [uploadMethod, setUploadMethod] = useState("json"); // "json" or "file"
  const [selectedFile, setSelectedFile] = useState(null);

  const createProduct = useCreateProduct();
  const updateProduct = useUpdateProduct();
  const createProductWithFile = useCreateProductWithFile();

  // Get categories and shops for dropdowns
  const { data: categoriesData, isLoading: categoriesLoading } =
    useCategories();
  const { data: shopsData, isLoading: shopsLoading } = useShops();

  const categories = categoriesData?.data?.categories || [];
  const shops = shopsData?.data?.shops || [];

  const schema = isEditing ? UpdateProductSchema : CreateProductSchema;

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      description: "",
      barcode: "",
      sale: 0,
      buy: 0,
      stock_quantity: 0,
      unit: "pcs",
      ppn: 0,
      category_id: "",
      shop_id: "",
      photo: "",
    },
  });

  // Watch form values for profit calculation
  const salePrice = watch("sale");
  const buyPrice = watch("buy");
  const profit = salePrice && buyPrice ? salePrice - buyPrice : 0;

  // Load product data when editing
  useEffect(() => {
    if (isEditing && product) {
      // Map backend fields to form fields
      setValue("name", product.name || "");
      setValue("description", product.description || "");
      setValue("barcode", product.barcode || "");
      setValue("sale", product.sale || 0);
      setValue("buy", product.buy || 0);
      setValue("stock_quantity", product.stock || 0);
      setValue("unit", product.unit || "pcs");
      setValue("ppn", product.ppn || 0);
      setValue("category_id", product.cat_id || "");
      setValue("shop_id", product.shop_id || "");
      setValue("photo", product.photo || "");
    } else {
      reset();
    }
  }, [isEditing, product, setValue, reset]);

  const handleFileChange = (e) => {
    const file = e.target.files[0];
    setSelectedFile(file);
  };

  const onSubmit = async (data) => {
    try {
      if (isEditing) {
        // For updates, always use JSON
        await updateProduct.mutateAsync({ id: product.id, ...data });
      } else {
        if (uploadMethod === "file" && selectedFile) {
          // Create FormData for file upload
          const formData = new FormData();
          formData.append("name", data.name);
          formData.append("description", data.description || "");
          formData.append("sale", data.sale);
          formData.append("buy", data.buy);
          formData.append("unit", data.unit || "");
          formData.append("ppn", data.ppn || "");
          formData.append("category_id", data.category_id || "");
          formData.append("barcode", data.barcode || "");
          formData.append("stock_quantity", data.stock_quantity || 0);
          formData.append("shop_id", data.shop_id);
          formData.append("photo", selectedFile);

          await createProductWithFile.mutateAsync(formData);
        } else {
          // Use JSON API
          const productData = {
            ...data,
            // Convert stock_quantity to stock for backend
            stock_quantity: data.stock_quantity || 0,
          };
          await createProduct.mutateAsync(productData);
        }
      }
      reset();
      setSelectedFile(null);
      onClose();
    } catch (error) {
      console.error("Product operation failed:", error);
    }
  };

  const isLoading =
    createProduct.isPending ||
    updateProduct.isPending ||
    createProductWithFile.isPending ||
    categoriesLoading ||
    shopsLoading;

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

  const categoryOptions = Array.isArray(categories)
    ? categories
        .filter((cat) => cat && cat.id && cat.name)
        .map((cat) => ({
          value: cat.id,
          label: cat.name,
        }))
    : [];

  const shopOptions = Array.isArray(shops)
    ? shops
        .filter((shop) => shop && shop.id && shop.name)
        .map((shop) => ({
          value: shop.id,
          label: shop.name,
        }))
    : [];

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
      size="xl"
    >
      <div className="space-y-6">
        {/* Upload Method Selection - Only for new products */}
        {!isEditing && (
          <div className="border-b border-gray-200 dark:border-gray-700 pb-4">
            <div className="flex space-x-4">
              <label className="flex items-center">
                <input
                  type="radio"
                  value="json"
                  checked={uploadMethod === "json"}
                  onChange={(e) => setUploadMethod(e.target.value)}
                  className="mr-2"
                />
                JSON Form
              </label>
              <label className="flex items-center">
                <input
                  type="radio"
                  value="file"
                  checked={uploadMethod === "file"}
                  onChange={(e) => setUploadMethod(e.target.value)}
                  className="mr-2"
                />
                With File Upload
              </label>
            </div>
          </div>
        )}

        {/* Basic Information */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormField
            name="name"
            label="Product Name"
            placeholder="Enter product name"
            register={register}
            error={errors.name}
            disabled={isLoading}
            required
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

        <FormField
          name="description"
          label="Description"
          type="textarea"
          placeholder="Enter product description"
          register={register}
          error={errors.description}
          disabled={isLoading}
        />

        {/* Pricing */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <FormField
            name="buy"
            label="Buy Price"
            type="number"
            step="0.01"
            placeholder="0.00"
            register={register}
            error={errors.buy}
            disabled={isLoading}
            required
          />

          <FormField
            name="sale"
            label="Sale Price"
            type="number"
            step="0.01"
            placeholder="0.00"
            register={register}
            error={errors.sale}
            disabled={isLoading}
            required
          />

          <div className="flex flex-col">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Profit
            </label>
            <div className="form-control bg-gray-50 dark:bg-gray-800">
              ${profit.toFixed(2)}
            </div>
          </div>
        </div>

        {/* Stock and Unit */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
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
            name="unit"
            label="Unit"
            type="select"
            options={unitOptions}
            register={register}
            error={errors.unit}
            disabled={isLoading}
          />

          <FormField
            name="ppn"
            label="PPN (%)"
            type="number"
            step="0.01"
            placeholder="0.00"
            register={register}
            error={errors.ppn}
            disabled={isLoading}
          />
        </div>

        {/* Category and Shop */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormField
            name="category_id"
            label="Category"
            type="select"
            options={[
              { value: "", label: "Select Category" },
              ...categoryOptions,
            ]}
            register={register}
            error={errors.category_id}
            disabled={isLoading}
          />

          <FormField
            name="shop_id"
            label="Shop"
            type="select"
            options={[{ value: "", label: "Select Shop" }, ...shopOptions]}
            register={register}
            error={errors.shop_id}
            disabled={isLoading}
            required
          />
        </div>

        {/* Photo Upload */}
        {uploadMethod === "file" && !isEditing ? (
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Product Photo
            </label>
            <input
              type="file"
              accept="image/*"
              onChange={handleFileChange}
              className="form-control"
              disabled={isLoading}
            />
            {selectedFile && (
              <p className="text-sm text-gray-500 mt-1">
                Selected: {selectedFile.name}
              </p>
            )}
          </div>
        ) : (
          <FormField
            name="photo"
            label="Photo URL"
            placeholder="Enter photo URL or path"
            register={register}
            error={errors.photo}
            disabled={isLoading}
          />
        )}

        {/* Information Box */}
        <div className="bg-blue-50 dark:bg-blue-900/20 p-4 rounded-md">
          <p className="text-sm text-blue-800 dark:text-blue-200">
            <strong>Note:</strong> Make sure to set appropriate sale and buy
            prices. Stock quantity will be tracked automatically. The profit
            margin is calculated as Sale Price - Buy Price = $
            {profit.toFixed(2)}.
          </p>
        </div>
      </div>
    </CrudModal>
  );
};

export default ProductModal;
