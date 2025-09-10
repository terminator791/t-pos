import React, { useState } from "react";
import Card from "@/components/ui/Card";
import Icon from "@/components/ui/Icon";
import Button from "@/components/ui/Button";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import ErrorDisplay from "@/components/ui/ErrorDisplay";
import ProductModal from "@/components/modals/ProductModal";
import DeleteConfirmModal from "@/components/ui/DeleteConfirmModal";
import {
  useProducts,
  useDeleteProduct,
  useSearchProducts,
  useLowStockProducts,
  useProductByBarcode,
} from "@/services/api";

const ProductsPage = () => {
  const [showModal, setShowModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [selectedProduct, setSelectedProduct] = useState(null);
  const [isEditing, setIsEditing] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [barcodeSearch, setBarcodeSearch] = useState("");
  const [activeTab, setActiveTab] = useState("all"); // all, search, low-stock, barcode
  const [currentShopId, setCurrentShopId] = useState(
    "fdb1b3bc-1fd7-49f6-acbe-982a8b4aa680"
  ); // Default shop ID

  // API hooks
  const {
    data: productsData,
    isLoading: productsLoading,
    error: productsError,
    refetch: refetchProducts,
  } = useProducts();

  const {
    data: searchData,
    isLoading: searchLoading,
    error: searchError,
  } = useSearchProducts(searchTerm, currentShopId, {
    enabled: activeTab === "search" && !!searchTerm && !!currentShopId,
  });

  const {
    data: lowStockData,
    isLoading: lowStockLoading,
    error: lowStockError,
  } = useLowStockProducts(currentShopId, {
    enabled: activeTab === "low-stock" && !!currentShopId,
  });

  const {
    data: barcodeData,
    isLoading: barcodeLoading,
    error: barcodeError,
  } = useProductByBarcode(barcodeSearch, {
    enabled: activeTab === "barcode" && !!barcodeSearch,
  });

  const deleteProduct = useDeleteProduct();

  // Determine which data to display
  let displayData, isLoading, error;

  switch (activeTab) {
    case "search":
      displayData = searchData;
      isLoading = searchLoading;
      error = searchError;
      break;
    case "low-stock":
      displayData = lowStockData;
      isLoading = lowStockLoading;
      error = lowStockError;
      break;
    case "barcode":
      displayData = barcodeData
        ? { data: { products: [barcodeData.data] } }
        : null;
      isLoading = barcodeLoading;
      error = barcodeError;
      break;
    default:
      displayData = productsData;
      isLoading = productsLoading;
      error = productsError;
      break;
  }

  const products = displayData?.data?.products || [];
  const totalProducts = productsData?.data?.count || 0;

  // Filter products based on search term for "all" tab
  const filteredProducts =
    activeTab === "all" && searchTerm
      ? products.filter(
          (product) =>
            product.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
            product.barcode?.toLowerCase().includes(searchTerm.toLowerCase())
        )
      : products;

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

  const handleTabChange = (tab) => {
    setActiveTab(tab);
    setSearchTerm("");
    setBarcodeSearch("");
  };

  const handleSearchProducts = () => {
    if (searchTerm.trim()) {
      setActiveTab("search");
    }
  };

  const handleBarcodeSearch = () => {
    if (barcodeSearch.trim()) {
      setActiveTab("barcode");
    }
  };

  if (isLoading && activeTab === "all") {
    return <LoadingSpinner message="Loading products..." />;
  }

  if (error && activeTab === "all") {
    return (
      <ErrorDisplay
        message="Failed to load products. Please try again."
        onRetry={refetchProducts}
      />
    );
  }

  // Calculate stats from all products
  const allProducts = productsData?.data?.products || [];
  const activeProducts = allProducts.filter(
    (p) => p.status === "active" || !p.status
  ).length;
  const outOfStockProducts = allProducts.filter((p) => p.stock === 0).length;
  const totalValue = allProducts.reduce(
    (sum, p) => sum + (p.sale || 0) * (p.stock || 0),
    0
  );

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
        {/* Action Bar */}
        <div className="flex flex-col md:flex-row justify-between items-start md:items-center mb-6 space-y-4 md:space-y-0">
          <div className="flex flex-wrap gap-2">
            <Button
              icon="ph:plus"
              className="btn-primary"
              onClick={handleAddProduct}
            >
              Add Product
            </Button>
            <Button
              icon="ph:list"
              className={`${
                activeTab === "all" ? "btn-primary" : "btn-secondary"
              }`}
              onClick={() => handleTabChange("all")}
            >
              All Products
            </Button>
            <Button
              icon="ph:warning"
              className={`${
                activeTab === "low-stock" ? "btn-warning" : "btn-secondary"
              }`}
              onClick={() => handleTabChange("low-stock")}
            >
              Low Stock
            </Button>
          </div>

          <div className="flex flex-col sm:flex-row items-stretch sm:items-center space-y-2 sm:space-y-0 sm:space-x-2 w-full md:w-auto">
            {/* Search Bar */}
            <div className="flex">
              <input
                type="text"
                placeholder="Search products..."
                className="form-control rounded-r-none"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                onKeyPress={(e) => e.key === "Enter" && handleSearchProducts()}
              />
              <Button
                icon="ph:magnifying-glass"
                className="btn-primary rounded-l-none"
                onClick={handleSearchProducts}
                disabled={!searchTerm.trim()}
              >
                Search
              </Button>
            </div>

            {/* Barcode Search */}
            <div className="flex">
              <input
                type="text"
                placeholder="Search by barcode..."
                className="form-control rounded-r-none"
                value={barcodeSearch}
                onChange={(e) => setBarcodeSearch(e.target.value)}
                onKeyPress={(e) => e.key === "Enter" && handleBarcodeSearch()}
              />
              <Button
                icon="ph:barcode"
                className="btn-secondary rounded-l-none"
                onClick={handleBarcodeSearch}
                disabled={!barcodeSearch.trim()}
              >
                Barcode
              </Button>
            </div>
          </div>
        </div>

        {/* Active Tab Indicator */}
        <div className="mb-4">
          <div className="flex items-center space-x-2">
            <span className="text-sm text-gray-500">Active filter:</span>
            <span className="badge bg-blue-500/10 text-blue-500 capitalize">
              {activeTab === "low-stock" ? "Low Stock" : activeTab}
              {activeTab === "search" && searchTerm && ` (${searchTerm})`}
              {activeTab === "barcode" &&
                barcodeSearch &&
                ` (${barcodeSearch})`}
            </span>
            {activeTab !== "all" && (
              <Button
                icon="ph:x"
                size="sm"
                className="btn-secondary"
                onClick={() => handleTabChange("all")}
              >
                Clear
              </Button>
            )}
          </div>
        </div>

        {/* Loading State for specific operations */}
        {isLoading && activeTab !== "all" && (
          <div className="text-center py-8">
            <LoadingSpinner
              message={`Loading ${
                activeTab === "low-stock" ? "low stock" : activeTab
              } products...`}
            />
          </div>
        )}

        {/* Error State for specific operations */}
        {error && activeTab !== "all" && (
          <div className="text-center py-8">
            <ErrorDisplay
              message={`Failed to load ${
                activeTab === "low-stock" ? "low stock" : activeTab
              } products.`}
              onRetry={() => {
                if (activeTab === "search") handleSearchProducts();
                else if (activeTab === "barcode") handleBarcodeSearch();
                else if (activeTab === "low-stock") setActiveTab("low-stock");
              }}
            />
          </div>
        )}

        {/* Products Table */}
        {!isLoading && !error && (
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
                    Prices
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Stock
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Profit
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
                  <tr
                    key={product.id}
                    className="hover:bg-gray-50 dark:hover:bg-gray-800"
                  >
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center">
                        <div className="flex-shrink-0 h-10 w-10">
                          {product.photo ? (
                            <img
                              className="h-10 w-10 rounded-lg object-cover"
                              src={product.photo}
                              alt={product.name}
                              onError={(e) => {
                                e.target.style.display = "none";
                                e.target.nextSibling.style.display = "flex";
                              }}
                            />
                          ) : null}
                          <div
                            className={`h-10 w-10 rounded-lg bg-gray-300 dark:bg-gray-600 flex items-center justify-center ${
                              product.photo ? "hidden" : ""
                            }`}
                          >
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
                            {product.category?.name || "No Category"}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                      <span className="font-mono bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
                        {product.barcode || "N/A"}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                      <div className="space-y-1">
                        <div>
                          Sale:{" "}
                          <span className="font-semibold text-green-600">
                            ${product.sale?.toFixed(2) || "0.00"}
                          </span>
                        </div>
                        <div>
                          Buy:{" "}
                          <span className="text-gray-500">
                            ${product.buy?.toFixed(2) || "0.00"}
                          </span>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                      <span
                        className={`font-semibold ${
                          product.stock <= 10
                            ? "text-red-600"
                            : "text-gray-900 dark:text-white"
                        }`}
                      >
                        {product.stock} {product.unit}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      <span
                        className={`font-semibold ${
                          (product.profit || 0) > 0
                            ? "text-green-600"
                            : "text-red-600"
                        }`}
                      >
                        ${(product.profit || 0).toFixed(2)}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span
                        className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                          product.stock > 0
                            ? "bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-400"
                            : "bg-red-100 text-red-800 dark:bg-red-900/20 dark:text-red-400"
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
                          title="Edit Product"
                        >
                          <Icon icon="ph:pencil" />
                        </Button>
                        <Button
                          size="sm"
                          className="btn-danger"
                          onClick={() => handleDeleteProduct(product)}
                          title="Delete Product"
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
                <Icon
                  icon="ph:package"
                  className="mx-auto h-12 w-12 text-gray-400"
                />
                <p className="text-gray-500 dark:text-gray-400 mt-2">
                  {activeTab === "search" && searchTerm
                    ? `No products found matching "${searchTerm}".`
                    : activeTab === "barcode" && barcodeSearch
                    ? `No product found with barcode "${barcodeSearch}".`
                    : activeTab === "low-stock"
                    ? "No low stock products found."
                    : searchTerm
                    ? "No products found matching your search."
                    : "No products available."}
                </p>
              </div>
            )}
          </div>
        )}
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
