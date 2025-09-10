import React, { useState } from "react";
import Card from "@/components/ui/Card";
import Icon from "@/components/ui/Icon";
import Button from "@/components/ui/Button";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import ErrorDisplay from "@/components/ui/ErrorDisplay";
import ProductModal from "@/components/modals/ProductModal";
import DeleteConfirmModal from "@/components/ui/DeleteConfirmModal";
import { useProducts, useDeleteProduct } from "@/services/api";

const ProductsPage = () => {
  const [showModal, setShowModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [selectedProduct, setSelectedProduct] = useState(null);
  const [isEditing, setIsEditing] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");

  // API hooks
  const { data: productsData, isLoading, error, refetch } = useProducts();
  const deleteProduct = useDeleteProduct();

  const products = productsData?.data?.products || [];
  const totalProducts = productsData?.data?.count || 0;

  // Filter products based on search term
  const filteredProducts = products.filter(
    (product) =>
      product.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      product.barcode?.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const handleAddProduct = () => {
    setSelectedProduct(null);
    setIsEditing(false);
    setShowModal(true);
  };

  const handleEditProduct = (product) => {
    setSelectedProduct(product);
    setIsEditing(true);
    setShowModal(true);
  };

  const handleDeleteProduct = (product) => {
    setSelectedProduct(product);
    setShowDeleteModal(true);
  };

  const confirmDelete = async () => {
    if (selectedProduct) {
      try {
        await deleteProduct.mutateAsync(selectedProduct.id);
        setShowDeleteModal(false);
        setSelectedProduct(null);
      } catch (error) {
        console.error("Delete failed:", error);
      }
    }
  };

  if (isLoading) {
    return <LoadingSpinner message="Loading products..." />;
  }

  if (error) {
    return (
      <ErrorDisplay
        message="Failed to load products. Please try again."
        onRetry={refetch}
      />
    );
  }

  // Calculate stats
  const activeProducts = products.filter(
    (p) => p.status === "active" || !p.status
  ).length; // default to active if no status
  const outOfStockProducts = products.filter((p) => p.stock === 0).length; // use 'stock' instead of 'stock_quantity'
  const totalValue = products.reduce((sum, p) => sum + p.sale * p.stock, 0); // use 'sale' instead of 'price'

  return (
    <div className="space-y-5">
      {/* Stats Cards */}
      <div className="grid xl:grid-cols-4 sm:grid-cols-2 grid-cols-1 gap-5">
        <Card>
          <div>
            <div className="flex">
              <div className="flex-1 text-base font-medium">Total Products</div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-indigo-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:package" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                {totalProducts}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-indigo-500/10 text-indigo-500">
                  +12%
                </span>
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  Since last month
                </span>
              </span>
            </div>
          </div>
        </Card>

        <Card>
          <div>
            <div className="flex">
              <div className="flex-1 text-base font-medium">
                Active Products
              </div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-green-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:check-circle" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                {activeProducts}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-green-500/10 text-green-500">
                  +5%
                </span>
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  Since last week
                </span>
              </span>
            </div>
          </div>
        </Card>

        <Card>
          <div>
            <div className="flex">
              <div className="flex-1 text-base font-medium">Out of Stock</div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-red-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:warning" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                {outOfStockProducts}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-red-500/10 text-red-500">-2%</span>
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  Since yesterday
                </span>
              </span>
            </div>
          </div>
        </Card>

        <Card>
          <div>
            <div className="flex">
              <div className="flex-1 text-base font-medium">Total Value</div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-yellow-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:currency-dollar" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                ${totalValue.toFixed(2)}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-yellow-500/10 text-yellow-500">
                  +8%
                </span>
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  Since last month
                </span>
              </span>
            </div>
          </div>
        </Card>
      </div>

      {/* Products Table */}
      <Card title="Product Management">
        <div className="flex justify-between items-center mb-4">
          <div className="flex space-x-2">
            <Button
              icon="ph:plus"
              className="btn-primary"
              onClick={handleAddProduct}
            >
              Add Product
            </Button>
            <Button icon="ph:funnel" className="btn-secondary">
              Filter
            </Button>
          </div>
          <div className="flex items-center space-x-2">
            <input
              type="text"
              placeholder="Search products..."
              className="form-control"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead className="bg-gray-50 dark:bg-gray-800">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Product
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Barcode
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Price
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Stock
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
              {filteredProducts.map((product) => (
                <tr key={product.id}>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <div className="flex-shrink-0 h-10 w-10">
                        <div className="h-10 w-10 rounded-full bg-gray-300 dark:bg-gray-600 flex items-center justify-center">
                          <Icon
                            icon="ph:package"
                            className="text-gray-500 dark:text-gray-400"
                          />
                        </div>
                      </div>
                      <div className="ml-4">
                        <div className="text-sm font-medium text-gray-900 dark:text-white">
                          {product.name}
                        </div>
                        <div className="text-sm text-gray-500 dark:text-gray-400">
                          {product.barcode}
                        </div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                    {product.barcode || "N/A"}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                    ${product.sale?.toFixed(2) || "0.00"}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                    {product.stock} {product.unit}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span
                      className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                        product.stock > 0
                          ? "bg-green-100 text-green-800"
                          : "bg-red-100 text-red-800"
                      }`}
                    >
                      {product.stock > 0 ? "In Stock" : "Out of Stock"}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                    <div className="flex space-x-2">
                      <Button
                        size="sm"
                        className="btn-secondary"
                        onClick={() => handleEditProduct(product)}
                      >
                        <Icon icon="ph:pencil" />
                      </Button>
                      <Button
                        size="sm"
                        className="btn-danger"
                        onClick={() => handleDeleteProduct(product)}
                      >
                        <Icon icon="ph:trash" />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {filteredProducts.length === 0 && (
            <div className="text-center py-8">
              <p className="text-gray-500 dark:text-gray-400">
                {searchTerm
                  ? "No products found matching your search."
                  : "No products available."}
              </p>
            </div>
          )}
        </div>
      </Card>

      {/* Product Modal */}
      <ProductModal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        product={selectedProduct}
        isEditing={isEditing}
      />

      {/* Delete Confirmation Modal */}
      <DeleteConfirmModal
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        onConfirm={confirmDelete}
        title="Delete Product"
        message="Are you sure you want to delete this product? This action cannot be undone."
        itemName={selectedProduct?.name}
        isLoading={deleteProduct.isPending}
      />
    </div>
  );
};

export default ProductsPage;
