import React, { useState } from "react";
import Card from "@/components/ui/Card";
import Icon from "@/components/ui/Icon";
import Button from "@/components/ui/Button";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import ErrorDisplay from "@/components/ui/ErrorDisplay";
import UserModal from "@/components/modals/UserModal";
import DeleteConfirmModal from "@/components/ui/DeleteConfirmModal";
import { useUsers, useDeleteUser } from "@/services/api";

const UsersPage = () => {
  const [showModal, setShowModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [selectedUser, setSelectedUser] = useState(null);
  const [isEditing, setIsEditing] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedRole, setSelectedRole] = useState("");

  // API hooks
  const { data: usersData, isLoading, error, refetch } = useUsers();
  const deleteUser = useDeleteUser();

  const users = usersData?.data?.users || [];
  const totalUsers = usersData?.data?.count || 0;

  // Filter users based on search term and role
  const filteredUsers = users.filter((user) => {
    const matchesSearch =
      user.username?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      user.email?.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesRole = !selectedRole || user.role_id === selectedRole;
    return matchesSearch && matchesRole;
  });

  const handleAddUser = () => {
    setSelectedUser(null);
    setIsEditing(false);
    setShowModal(true);
  };

  const handleEditUser = (user) => {
    setSelectedUser(user);
    setIsEditing(true);
    setShowModal(true);
  };

  const handleDeleteUser = (user) => {
    setSelectedUser(user);
    setShowDeleteModal(true);
  };

  const confirmDelete = async () => {
    if (selectedUser) {
      try {
        await deleteUser.mutateAsync(selectedUser.id);
        setShowDeleteModal(false);
        setSelectedUser(null);
      } catch (error) {
        console.error("Delete failed:", error);
      }
    }
  };

  if (isLoading) {
    return <LoadingSpinner message="Loading users..." />;
  }

  if (error) {
    return (
      <ErrorDisplay
        message="Failed to load users. Please try again."
        onRetry={refetch}
      />
    );
  }

  // Calculate stats
  const activeUsers = users.length; // Assuming all fetched users are active

  // check user roles
  const adminUsers = users.filter((u) => {
    return u.role?.name === "admin" || u.role_id === "admin";
  }).length;

  const superAdminUsers = users.filter((u) => {
    return u.role?.name === "super_admin" || u.role_id === "super_admin";
  }).length;

  return (
    <div className="space-y-5">
      {/* Stats Cards */}
      <div className="grid xl:grid-cols-4 sm:grid-cols-2 grid-cols-1 gap-5">
        <Card>
          <div>
            <div className="flex">
              <div className="flex-1 text-base font-medium">Total Users</div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-indigo-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:user-circle" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                {totalUsers}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-indigo-500/10 text-indigo-500">
                  +5%
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
              <div className="flex-1 text-base font-medium">Active Users</div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-green-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:check-circle" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                {activeUsers}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-green-500/10 text-green-500">
                  100%
                </span>
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  Activity rate
                </span>
              </span>
            </div>
          </div>
        </Card>

        <Card>
          <div>
            <div className="flex">
              <div className="flex-1 text-base font-medium">Administrators</div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-purple-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:crown" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                {adminUsers}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-purple-500/10 text-purple-500">
                  Admin
                </span>
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  Regular admins
                </span>
              </span>
            </div>
          </div>
        </Card>

        <Card>
          <div>
            <div className="flex">
              <div className="flex-1 text-base font-medium">Super Admins</div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-red-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:shield-check" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                {superAdminUsers}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-red-500/10 text-red-500">
                  Critical
                </span>
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  System admins
                </span>
              </span>
            </div>
          </div>
        </Card>
      </div>

      {/* Users Table */}
      <Card title="User Management">
        <div className="flex justify-between items-center mb-4">
          <div className="flex space-x-2">
            <Button
              icon="ph:plus"
              className="btn-primary"
              onClick={handleAddUser}
            >
              Add User
            </Button>
            <Button icon="ph:funnel" className="btn-secondary">
              Filter
            </Button>
            <Button icon="ph:export" className="btn-secondary">
              Export
            </Button>
          </div>
          <div className="flex items-center space-x-2">
            <select
              className="form-control"
              value={selectedRole}
              onChange={(e) => setSelectedRole(e.target.value)}
            >
              <option value="">All Roles</option>
              <option value="admin">Admin</option>
              <option value="super_admin">Super Admin</option>
            </select>
            <input
              type="text"
              placeholder="Search users..."
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
                  User
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Role
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Name
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
              {filteredUsers.map((user) => (
                <tr key={user.id}>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <div className="flex-shrink-0 h-10 w-10">
                        <div className="h-10 w-10 rounded-full bg-gray-300 dark:bg-gray-600 flex items-center justify-center">
                          <Icon
                            icon="ph:user"
                            className="text-gray-500 dark:text-gray-400"
                          />
                        </div>
                      </div>
                      <div className="ml-4">
                        <div className="text-sm font-medium text-gray-900 dark:text-white">
                          {user.username}
                        </div>
                        <div className="text-sm text-gray-500 dark:text-gray-400">
                          {user.email}
                        </div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span
                      className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                        user.role_id === "super_admin"
                          ? "bg-red-100 text-red-800"
                          : "bg-purple-100 text-purple-800"
                      }`}
                    >
                      {user.role_id === "super_admin" ? "Super Admin" : "Admin"}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-900 dark:text-white text-center">
                      {user.name || "N/A"}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                    {new Date(user.created_at).toLocaleDateString()}
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
                        onClick={() => handleEditUser(user)}
                      >
                        <Icon icon="ph:pencil" />
                      </Button>
                      <Button
                        size="sm"
                        className="btn-danger"
                        onClick={() => handleDeleteUser(user)}
                      >
                        <Icon icon="ph:trash" />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {filteredUsers.length === 0 && (
            <div className="text-center py-8">
              <p className="text-gray-500 dark:text-gray-400">
                {searchTerm || selectedRole
                  ? "No users found matching your criteria."
                  : "No users available."}
              </p>
            </div>
          )}
        </div>
      </Card>

      {/* User Modal */}
      <UserModal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        user={selectedUser}
        isEditing={isEditing}
      />

      {/* Delete Confirmation Modal */}
      <DeleteConfirmModal
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        onConfirm={confirmDelete}
        title="Delete User"
        message="Are you sure you want to delete this user? This action cannot be undone and may affect system access."
        itemName={selectedUser?.username}
        isLoading={deleteUser.isPending}
      />
    </div>
  );
};

export default UsersPage;
