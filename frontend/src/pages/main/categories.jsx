import React, { useState } from "react";
import Card from "@/components/ui/Card";
import Icon from "@/components/ui/Icon";
import Button from "@/components/ui/Button";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import ErrorDisplay from "@/components/ui/ErrorDisplay";
import CategoryModal from "@/components/modals/CategoryModal";
import DeleteConfirmModal from "@/components/ui/DeleteConfirmModal";
import { useCategories, useDeleteCategory, useShops } from "@/services/api";

const CategoriesPage = () => {
  const [showModal, setShowModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [selectedCategory, setSelectedCategory] = useState(null);
  const [isEditing, setIsEditing] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedShop, setSelectedShop] = useState("");

  // API hooks
  const { data: categoriesData, isLoading, error, refetch } = useCategories(selectedShop);
  const { data: shopsData } = useShops();
  const deleteCategory = useDeleteCategory();

  const categories = categoriesData?.data?.categories || [];
  const totalCategories = categoriesData?.data?.count || 0;
  const shops = shopsData?.data?.shops || [];

  // Filter categories based on search term
  const filteredCategories = categories.filter((category) =>
    category.name?.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const handleAddCategory = () => {
    setSelectedCategory(null);
    setIsEditing(false);
    setShowModal(true);
  };

  const handleEditCategory = (category) => {
    setSelectedCategory(category);
    setIsEditing(true);
    setShowModal(true);
  };

  const handleDeleteCategory = (category) => {
    setSelectedCategory(category);
    setShowDeleteModal(true);
  };

  const confirmDelete = async () => {
    if (selectedCategory) {
      try {
        await deleteCategory.mutateAsync(selectedCategory.id);
        setShowDeleteModal(false);
        setSelectedCategory(null);
      } catch (error) {
        console.error("Delete failed:", error);
      }
    }
  };

  const getShopName = (shopId) => {
    const shop = shops.find(s => s.id === shopId);
    return shop?.name || "Unknown Shop";
  };

  return (
    <div className="space-y-5">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
            Categories
          </h2>
          <p className="text-gray-500 dark:text-gray-400">
            Manage product categories for your shops
          </p>
        </div>
        <Button
          icon="ph:plus"
          className="btn-primary"
          onClick={handleAddCategory}
        >
          Add Category
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
        <Card>
          <div className="flex items-center space-x-3">
            <div className="flex-none">
              <div className="w-12 h-12 rounded-full bg-primary-500/10 flex items-center justify-center">
                <Icon
                  icon="ph:tag"
                  className="text-2xl text-primary-500"
                />
              </div>
            </div>
            <div className="flex-1">
              <div className="text-slate-900 dark:text-white text-lg font-medium">
                {totalCategories}
              </div>
              <div className="text-slate-500 dark:text-slate-400 text-sm">
                Total Categories
              </div>
            </div>
          </div>
        </Card>

        <Card>
          <div className="flex items-center space-x-3">
            <div className="flex-none">
              <div className="w-12 h-12 rounded-full bg-success-500/10 flex items-center justify-center">
                <Icon
                  icon="ph:storefront"
                  className="text-2xl text-success-500"
                />
              </div>
            </div>
            <div className="flex-1">
              <div className="text-slate-900 dark:text-white text-lg font-medium">
                {shops.length}
              </div>
              <div className="text-slate-500 dark:text-slate-400 text-sm">
                Total Shops
              </div>
            </div>
          </div>
        </Card>

        <Card>
          <div className="flex items-center space-x-3">
            <div className="flex-none">
              <div className="w-12 h-12 rounded-full bg-warning-500/10 flex items-center justify-center">
                <Icon
                  icon="ph:list-dashes"
                  className="text-2xl text-warning-500"
                />
              </div>
            </div>
            <div className="flex-1">
              <div className="text-slate-900 dark:text-white text-lg font-medium">
                {filteredCategories.length}
              </div>
              <div className="text-slate-500 dark:text-slate-400 text-sm">
                Filtered Results
              </div>
            </div>
          </div>
        </Card>
      </div>

      {/* Filters */}
      <Card>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Search Categories
            </label>
            <input
              type="text"
              placeholder="Search by name..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="form-control"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Filter by Shop
            </label>
            <select
              value={selectedShop}
              onChange={(e) => setSelectedShop(e.target.value)}
              className="form-control"
            >
              <option value="">All Shops</option>
              {shops.map((shop) => (
                <option key={shop.id} value={shop.id}>
                  {shop.name}
                </option>
              ))}
            </select>
          </div>
        </div>
      </Card>

      {/* Categories Table */}
      <Card>
        {isLoading && (
          <div className="text-center py-8">
            <LoadingSpinner />
          </div>
        )}

        {error && (
          <div className="text-center py-8">
            <ErrorDisplay
              message="Failed to load categories."
              onRetry={refetch}
            />
          </div>
        )}

        {!isLoading && !error && (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
              <thead className="bg-gray-50 dark:bg-gray-800">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Category
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Shop
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Description
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Created
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
                {filteredCategories.length === 0 ? (
                  <tr>
                    <td
                      colSpan="5"
                      className="px-6 py-12 text-center text-gray-500 dark:text-gray-400"
                    >
                      <Icon
                        icon="ph:tag"
                        className="mx-auto h-12 w-12 text-gray-400 mb-4"
                      />
                      <h3 className="text-sm font-medium text-gray-900 dark:text-white">
                        No categories found
                      </h3>
                      <p className="text-sm text-gray-500 dark:text-gray-400">
                        Get started by creating a new category.
                      </p>
                    </td>
                  </tr>
                ) : (
                  filteredCategories.map((category) => (
                    <tr
                      key={category.id}
                      className="hover:bg-gray-50 dark:hover:bg-gray-800/50"
                    >
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center">
                          <div className="flex-shrink-0 h-10 w-10">
                            <div className="h-10 w-10 rounded-lg bg-primary-100 dark:bg-primary-900/20 flex items-center justify-center">
                              <Icon
                                icon="ph:tag"
                                className="h-5 w-5 text-primary-600 dark:text-primary-400"
                              />
                            </div>
                          </div>
                          <div className="ml-4">
                            <div className="text-sm font-medium text-gray-900 dark:text-white">
                              {category.name}
                            </div>
                            <div className="text-sm text-gray-500 dark:text-gray-400">
                              ID: {category.id}
                            </div>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                        {getShopName(category.shop_id)}
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                        {category.description || "No description"}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                        {category.created_at
                          ? new Date(category.created_at).toLocaleDateString()
                          : "N/A"}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium space-x-2">
                        <Button
                          icon="ph:pencil"
                          size="sm"
                          className="btn-outline-secondary"
                          onClick={() => handleEditCategory(category)}
                        />
                        <Button
                          icon="ph:trash"
                          size="sm"
                          className="btn-outline-danger"
                          onClick={() => handleDeleteCategory(category)}
                        />
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      {/* Category Modal */}
      <CategoryModal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        category={selectedCategory}
        isEditing={isEditing}
      />

      {/* Delete Confirmation Modal */}
      <DeleteConfirmModal
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        onConfirm={confirmDelete}
        title="Delete Category"
        message={`Are you sure you want to delete "${selectedCategory?.name}"? This action cannot be undone.`}
        isLoading={deleteCategory.isPending}
      />
    </div>
  );
};

export default CategoriesPage;