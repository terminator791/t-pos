import React, { useState } from "react";
import Card from "@/components/ui/Card";
import Icon from "@/components/ui/Icon";
import Button from "@/components/ui/Button";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import ErrorDisplay from "@/components/ui/ErrorDisplay";
import ShopModal from "@/components/modals/ShopModal";
import DeleteConfirmModal from "@/components/ui/DeleteConfirmModal";
import { useShops, useDeleteShop, useLicenses } from "@/services/api";

const ShopsPage = () => {
  const [showModal, setShowModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [selectedShop, setSelectedShop] = useState(null);
  const [isEditing, setIsEditing] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedLicense, setSelectedLicense] = useState("");

  // API hooks
  const { data: shopsData, isLoading, error, refetch } = useShops();
  const { data: licensesData } = useLicenses();
  const deleteShop = useDeleteShop();

  const shops = shopsData?.data?.shops || [];
  const totalShops = shopsData?.data?.count || 0;
  const licenses = licensesData?.data?.licenses || [];

  // Filter shops based on search term and license
  const filteredShops = shops.filter((shop) => {
    const matchesSearch = shop.name?.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesLicense = !selectedLicense || shop.license_id === selectedLicense;
    return matchesSearch && matchesLicense;
  });

  const handleAddShop = () => {
    setSelectedShop(null);
    setIsEditing(false);
    setShowModal(true);
  };

  const handleEditShop = (shop) => {
    setSelectedShop(shop);
    setIsEditing(true);
    setShowModal(true);
  };

  const handleDeleteShop = (shop) => {
    setSelectedShop(shop);
    setShowDeleteModal(true);
  };

  const confirmDelete = async () => {
    if (selectedShop) {
      try {
        await deleteShop.mutateAsync(selectedShop.id);
        setShowDeleteModal(false);
        setSelectedShop(null);
      } catch (error) {
        console.error("Delete failed:", error);
      }
    }
  };

  const getLicenseInfo = (licenseId) => {
    const license = licenses.find(l => l.id === licenseId);
    return license ? `${license.serial_number} (${license.license_type})` : "Unknown License";
  };

  return (
    <div className="space-y-5">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
            Shops
          </h2>
          <p className="text-gray-500 dark:text-gray-400">
            Manage your business shops and locations
          </p>
        </div>
        <Button
          icon="ph:plus"
          className="btn-primary"
          onClick={handleAddShop}
        >
          Add Shop
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
        <Card>
          <div className="flex items-center space-x-3">
            <div className="flex-none">
              <div className="w-12 h-12 rounded-full bg-primary-500/10 flex items-center justify-center">
                <Icon
                  icon="ph:storefront"
                  className="text-2xl text-primary-500"
                />
              </div>
            </div>
            <div className="flex-1">
              <div className="text-slate-900 dark:text-white text-lg font-medium">
                {totalShops}
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
              <div className="w-12 h-12 rounded-full bg-success-500/10 flex items-center justify-center">
                <Icon
                  icon="ph:certificate"
                  className="text-2xl text-success-500"
                />
              </div>
            </div>
            <div className="flex-1">
              <div className="text-slate-900 dark:text-white text-lg font-medium">
                {licenses.length}
              </div>
              <div className="text-slate-500 dark:text-slate-400 text-sm">
                Available Licenses
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
                {filteredShops.length}
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
              Search Shops
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
              Filter by License
            </label>
            <select
              value={selectedLicense}
              onChange={(e) => setSelectedLicense(e.target.value)}
              className="form-control"
            >
              <option value="">All Licenses</option>
              {licenses.map((license) => (
                <option key={license.id} value={license.id}>
                  {license.serial_number} ({license.license_type})
                </option>
              ))}
            </select>
          </div>
        </div>
      </Card>

      {/* Shops Table */}
      <Card>
        {isLoading && (
          <div className="text-center py-8">
            <LoadingSpinner />
          </div>
        )}

        {error && (
          <div className="text-center py-8">
            <ErrorDisplay
              message="Failed to load shops."
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
                    Shop
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    License
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Contact
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
                {filteredShops.length === 0 ? (
                  <tr>
                    <td
                      colSpan="5"
                      className="px-6 py-12 text-center text-gray-500 dark:text-gray-400"
                    >
                      <Icon
                        icon="ph:storefront"
                        className="mx-auto h-12 w-12 text-gray-400 mb-4"
                      />
                      <h3 className="text-sm font-medium text-gray-900 dark:text-white">
                        No shops found
                      </h3>
                      <p className="text-sm text-gray-500 dark:text-gray-400">
                        Get started by creating a new shop.
                      </p>
                    </td>
                  </tr>
                ) : (
                  filteredShops.map((shop) => (
                    <tr
                      key={shop.id}
                      className="hover:bg-gray-50 dark:hover:bg-gray-800/50"
                    >
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center">
                          <div className="flex-shrink-0 h-10 w-10">
                            <div className="h-10 w-10 rounded-lg bg-primary-100 dark:bg-primary-900/20 flex items-center justify-center">
                              <Icon
                                icon="ph:storefront"
                                className="h-5 w-5 text-primary-600 dark:text-primary-400"
                              />
                            </div>
                          </div>
                          <div className="ml-4">
                            <div className="text-sm font-medium text-gray-900 dark:text-white">
                              {shop.name}
                            </div>
                            <div className="text-sm text-gray-500 dark:text-gray-400">
                              {shop.description || "No description"}
                            </div>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                        {getLicenseInfo(shop.license_id)}
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                        <div>
                          {shop.phone && (
                            <div className="flex items-center">
                              <Icon icon="ph:phone" className="h-4 w-4 mr-1" />
                              {shop.phone}
                            </div>
                          )}
                          {shop.address && (
                            <div className="flex items-center mt-1">
                              <Icon icon="ph:map-pin" className="h-4 w-4 mr-1" />
                              <span className="truncate max-w-xs">{shop.address}</span>
                            </div>
                          )}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                        {shop.created_at
                          ? new Date(shop.created_at).toLocaleDateString()
                          : "N/A"}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium space-x-2">
                        <Button
                          icon="ph:pencil"
                          size="sm"
                          className="btn-outline-secondary"
                          onClick={() => handleEditShop(shop)}
                        />
                        <Button
                          icon="ph:trash"
                          size="sm"
                          className="btn-outline-danger"
                          onClick={() => handleDeleteShop(shop)}
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

      {/* Shop Modal */}
      <ShopModal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        shop={selectedShop}
        isEditing={isEditing}
      />

      {/* Delete Confirmation Modal */}
      <DeleteConfirmModal
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        onConfirm={confirmDelete}
        title="Delete Shop"
        message={`Are you sure you want to delete "${selectedShop?.name}"? This action cannot be undone.`}
        isLoading={deleteShop.isPending}
      />
    </div>
  );
};

export default ShopsPage;