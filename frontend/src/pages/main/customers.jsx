import React, { useState } from "react";
import Card from "@/components/ui/Card";
import Icon from "@/components/ui/Icon";
import Button from "@/components/ui/Button";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import ErrorDisplay from "@/components/ui/ErrorDisplay";
import CustomerModal from "@/components/modals/CustomerModal";
import DeleteConfirmModal from "@/components/ui/DeleteConfirmModal";
import { useCustomers, useDeleteCustomer } from "@/services/api";

const CustomersPage = () => {
  const [showModal, setShowModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [selectedCustomer, setSelectedCustomer] = useState(null);
  const [isEditing, setIsEditing] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");

  // API hooks
  const { data: customersData, isLoading, error, refetch } = useCustomers();
  const deleteCustomer = useDeleteCustomer();

  const customers = customersData?.data?.customers || [];
  const totalCustomers = customersData?.data?.count || 0;

  // Filter customers based on search term
  const filteredCustomers = customers.filter(customer =>
    customer.username?.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const handleAddCustomer = () => {
    setSelectedCustomer(null);
    setIsEditing(false);
    setShowModal(true);
  };

  const handleEditCustomer = (customer) => {
    setSelectedCustomer(customer);
    setIsEditing(true);
    setShowModal(true);
  };

  const handleDeleteCustomer = (customer) => {
    setSelectedCustomer(customer);
    setShowDeleteModal(true);
  };

  const confirmDelete = async () => {
    if (selectedCustomer) {
      try {
        await deleteCustomer.mutateAsync(selectedCustomer.id);
        setShowDeleteModal(false);
        setSelectedCustomer(null);
      } catch (error) {
        console.error("Delete failed:", error);
      }
    }
  };

  if (isLoading) {
    return <LoadingSpinner message="Loading customers..." />;
  }

  if (error) {
    return (
      <ErrorDisplay
        message="Failed to load customers. Please try again."
        onRetry={refetch}
      />
    );
  }

  // Calculate stats
  const activeCustomers = customers.filter(c => c.role_id === "cashier" || c.role_id === "owner_business").length;
  const cashiers = customers.filter(c => c.role_id === "cashier").length;
  const businessOwners = customers.filter(c => c.role_id === "owner_business").length;

  return (
    <div className="space-y-5">
      {/* Stats Cards */}
      <div className="grid xl:grid-cols-4 sm:grid-cols-2 grid-cols-1 gap-5">
        <Card>
          <div>
            <div className="flex">
              <div className="flex-1 text-base font-medium">Total Customers</div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-indigo-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:users" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                {totalCustomers}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-indigo-500/10 text-indigo-500">
                  +18%
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
              <div className="flex-1 text-base font-medium">Active Customers</div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-green-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:check-circle" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                {activeCustomers}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-green-500/10 text-green-500">
                  +12%
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
              <div className="flex-1 text-base font-medium">Cashiers</div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-yellow-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:cash-register" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                {cashiers}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-yellow-500/10 text-yellow-500">
                  +25%
                </span>
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  Active cashiers
                </span>
              </span>
            </div>
          </div>
        </Card>

        <Card>
          <div>
            <div className="flex">
              <div className="flex-1 text-base font-medium">Business Owners</div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-purple-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:briefcase" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                {businessOwners}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-purple-500/10 text-purple-500">
                  +10%
                </span>
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  Business owners
                </span>
              </span>
            </div>
          </div>
        </Card>
      </div>

      {/* Customers Table */}
      <Card title="Customer Management">
        <div className="flex justify-between items-center mb-4">
          <div className="flex space-x-2">
            <Button 
              icon="ph:plus" 
              className="btn-primary"
              onClick={handleAddCustomer}
            >
              Add Customer
            </Button>
            <Button 
              icon="ph:funnel" 
              className="btn-secondary"
            >
              Filter
            </Button>
            <Button 
              icon="ph:export" 
              className="btn-secondary"
            >
              Export
            </Button>
          </div>
          <div className="flex items-center space-x-2">
            <input
              type="text"
              placeholder="Search customers..."
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
                  Customer
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Role
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  License
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
              {filteredCustomers.map((customer) => (
                <tr key={customer.id}>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <div className="flex-shrink-0 h-10 w-10">
                        <div className="h-10 w-10 rounded-full bg-gray-300 dark:bg-gray-600 flex items-center justify-center">
                          <Icon icon="ph:user" className="text-gray-500 dark:text-gray-400" />
                        </div>
                      </div>
                      <div className="ml-4">
                        <div className="text-sm font-medium text-gray-900 dark:text-white">
                          {customer.username}
                        </div>
                        <div className="text-sm text-gray-500 dark:text-gray-400">
                          ID: {customer.id?.slice(0, 8)}
                        </div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                      customer.role_id === "owner_business" 
                        ? "bg-purple-100 text-purple-800" 
                        : "bg-blue-100 text-blue-800"
                    }`}>
                      {customer.role_id === "owner_business" ? "Business Owner" : "Cashier"}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm font-mono text-gray-900 dark:text-white bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded text-center">
                      {customer.serial_number}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                    {new Date(customer.created_at).toLocaleDateString()}
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
                        onClick={() => handleEditCustomer(customer)}
                      >
                        <Icon icon="ph:pencil" />
                      </Button>
                      <Button 
                        size="sm" 
                        className="btn-danger"
                        onClick={() => handleDeleteCustomer(customer)}
                      >
                        <Icon icon="ph:trash" />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          
          {filteredCustomers.length === 0 && (
            <div className="text-center py-8">
              <p className="text-gray-500 dark:text-gray-400">
                {searchTerm ? "No customers found matching your search." : "No customers available."}
              </p>
            </div>
          )}
        </div>
      </Card>

      {/* Customer Modal */}
      <CustomerModal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        customer={selectedCustomer}
        isEditing={isEditing}
      />

      {/* Delete Confirmation Modal */}
      <DeleteConfirmModal
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        onConfirm={confirmDelete}
        title="Delete Customer"
        message="Are you sure you want to delete this customer? This action cannot be undone."
        itemName={selectedCustomer?.username}
        isLoading={deleteCustomer.isPending}
      />
    </div>
  );
};

export default CustomersPage;