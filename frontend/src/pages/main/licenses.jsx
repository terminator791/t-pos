import React, { useState } from "react";
import Card from "@/components/ui/Card";
import Icon from "@/components/ui/Icon";
import Button from "@/components/ui/Button";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import ErrorDisplay from "@/components/ui/ErrorDisplay";
import LicenseModal from "@/components/modals/LicenseModal";
import DeleteConfirmModal from "@/components/ui/DeleteConfirmModal";
import { useLicenses, useDeleteLicense } from "@/services/api";

const LicensesPage = () => {
  const [showModal, setShowModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [selectedLicense, setSelectedLicense] = useState(null);
  const [isEditing, setIsEditing] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");

  // API hooks
  const { data: licensesData, isLoading, error, refetch } = useLicenses();
  const deleteLicense = useDeleteLicense();

  const licenses = licensesData?.data?.licenses || [];
  const totalLicenses = licensesData?.data?.count || 0;

  // Filter licenses based on search term
  const filteredLicenses = licenses.filter(license =>
    license.serial_number?.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const handleAddLicense = () => {
    setSelectedLicense(null);
    setIsEditing(false);
    setShowModal(true);
  };

  const handleViewLicense = (license) => {
    setSelectedLicense(license);
    setIsEditing(true);
    setShowModal(true);
  };

  const handleDeleteLicense = (license) => {
    setSelectedLicense(license);
    setShowDeleteModal(true);
  };

  const confirmDelete = async () => {
    if (selectedLicense) {
      try {
        await deleteLicense.mutateAsync(selectedLicense.id);
        setShowDeleteModal(false);
        setSelectedLicense(null);
      } catch (error) {
        console.error("Delete failed:", error);
      }
    }
  };

  if (isLoading) {
    return <LoadingSpinner message="Loading licenses..." />;
  }

  if (error) {
    return (
      <ErrorDisplay
        message="Failed to load licenses. Please try again."
        onRetry={refetch}
      />
    );
  }

  // Calculate stats
  const activeLicenses = licenses.length; // All fetched licenses are considered active
  const expiredLicenses = 0; // Backend doesn't provide expired status in this simplified version
  const totalUsers = licenses.length * 10; // Estimate for display

  return (
    <div className="space-y-5">
      {/* Stats Cards */}
      <div className="grid xl:grid-cols-4 sm:grid-cols-2 grid-cols-1 gap-5">
        <Card>
          <div>
            <div className="flex">
              <div className="flex-1 text-base font-medium">Total Licenses</div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-indigo-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:certificate" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                {totalLicenses}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-indigo-500/10 text-indigo-500">
                  +2%
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
              <div className="flex-1 text-base font-medium">Active Licenses</div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-green-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:check-circle" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                {activeLicenses}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-green-500/10 text-green-500">
                  +1
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
              <div className="flex-1 text-base font-medium">Expired</div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-red-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:warning" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                {expiredLicenses}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-red-500/10 text-red-500">
                  0
                </span>
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  Needs renewal
                </span>
              </span>
            </div>
          </div>
        </Card>

        <Card>
          <div>
            <div className="flex">
              <div className="flex-1 text-base font-medium">Estimated Users</div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-yellow-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:users" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                {totalUsers}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-yellow-500/10 text-yellow-500">
                  +15%
                </span>
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  Licensed users
                </span>
              </span>
            </div>
          </div>
        </Card>
      </div>

      {/* Licenses Table */}
      <Card title="License Management">
        <div className="flex justify-between items-center mb-4">
          <div className="flex space-x-2">
            <Button 
              icon="ph:plus" 
              className="btn-primary"
              onClick={handleAddLicense}
            >
              Add License
            </Button>
            <Button 
              icon="ph:funnel" 
              className="btn-secondary"
            >
              Filter
            </Button>
          </div>
          <div className="flex items-center space-x-2">
            <input
              type="text"
              placeholder="Search licenses..."
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
                  License
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Serial Number
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Created Date
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
              {filteredLicenses.map((license) => (
                <tr key={license.id}>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <div className="flex-shrink-0 h-10 w-10">
                        <div className="h-10 w-10 rounded-full bg-gray-300 dark:bg-gray-600 flex items-center justify-center">
                          <Icon icon="ph:certificate" className="text-gray-500 dark:text-gray-400" />
                        </div>
                      </div>
                      <div className="ml-4">
                        <div className="text-sm font-medium text-gray-900 dark:text-white">
                          License #{license.id?.slice(0, 8)}
                        </div>
                        <div className="text-sm text-gray-500 dark:text-gray-400">
                          Active License
                        </div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm font-mono text-gray-900 dark:text-white bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
                      {license.serial_number}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                    {new Date(license.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800">
                      Active
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                    <div className="flex space-x-2">
                      <Button 
                        size="sm" 
                        className="btn-secondary"
                        onClick={() => handleViewLicense(license)}
                      >
                        <Icon icon="ph:eye" />
                      </Button>
                      <Button 
                        size="sm" 
                        className="btn-danger"
                        onClick={() => handleDeleteLicense(license)}
                      >
                        <Icon icon="ph:trash" />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          
          {filteredLicenses.length === 0 && (
            <div className="text-center py-8">
              <p className="text-gray-500 dark:text-gray-400">
                {searchTerm ? "No licenses found matching your search." : "No licenses available."}
              </p>
            </div>
          )}
        </div>
      </Card>

      {/* License Modal */}
      <LicenseModal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        license={selectedLicense}
        isEditing={isEditing}
      />

      {/* Delete Confirmation Modal */}
      <DeleteConfirmModal
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        onConfirm={confirmDelete}
        title="Delete License"
        message="Are you sure you want to delete this license? This action cannot be undone and may affect all associated users and data."
        itemName={selectedLicense?.serial_number}
        isLoading={deleteLicense.isPending}
      />
    </div>
  );
};

export default LicensesPage;