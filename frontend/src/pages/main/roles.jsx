import React, { useState } from "react";
import Card from "@/components/ui/Card";
import Icon from "@/components/ui/Icon";
import Button from "@/components/ui/Button";

const rolesData = [
  {
    id: 1,
    name: "Administrator",
    description: "Full access to all system features and settings",
    users: 1,
    permissions: [
      "User Management",
      "Product Management", 
      "License Management",
      "Customer Management",
      "System Settings",
      "Reports & Analytics",
      "Data Export",
      "Backup & Restore"
    ],
    createdDate: "2024-01-01",
    status: "Active",
  },
  {
    id: 2,
    name: "Sales Manager",
    description: "Manage sales operations and customer relationships",
    users: 3,
    permissions: [
      "Product Management",
      "Customer Management", 
      "Sales Reports",
      "Lead Management"
    ],
    createdDate: "2024-01-05",
    status: "Active",
  },
  {
    id: 3,
    name: "Support Agent",
    description: "Handle customer support and basic operations",
    users: 5,
    permissions: [
      "Customer Management",
      "Ticket Management",
      "Knowledge Base"
    ],
    createdDate: "2024-01-10",
    status: "Active",
  },
  {
    id: 4,
    name: "Viewer",
    description: "Read-only access to essential information",
    users: 8,
    permissions: [
      "View Dashboard",
      "View Reports",
      "View Customer Info"
    ],
    createdDate: "2024-01-15",
    status: "Active",
  },
];

const allPermissions = [
  "User Management",
  "Product Management", 
  "License Management",
  "Customer Management",
  "System Settings",
  "Reports & Analytics",
  "Data Export",
  "Backup & Restore",
  "Sales Reports",
  "Lead Management",
  "Ticket Management",
  "Knowledge Base",
  "View Dashboard",
  "View Reports",
  "View Customer Info"
];

const RolesPage = () => {
  const [roles, setRoles] = useState(rolesData);
  const [selectedRole, setSelectedRole] = useState(null);

  return (
    <div className="space-y-5">
      {/* Stats Cards */}
      <div className="grid xl:grid-cols-4 sm:grid-cols-2 grid-cols-1 gap-5">
        <Card>
          <div>
            <div className="flex">
              <div className="flex-1 text-base font-medium">Total Roles</div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-indigo-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:lock-key" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                {roles.length}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-indigo-500/10 text-indigo-500">
                  +1
                </span>
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  New role added
                </span>
              </span>
            </div>
          </div>
        </Card>
        
        <Card>
          <div>
            <div className="flex">
              <div className="flex-1 text-base font-medium">Active Roles</div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-green-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:check-circle" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                {roles.filter(r => r.status === "Active").length}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-green-500/10 text-green-500">
                  100%
                </span>
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  All roles active
                </span>
              </span>
            </div>
          </div>
        </Card>

        <Card>
          <div>
            <div className="flex">
              <div className="flex-1 text-base font-medium">Total Users</div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-yellow-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:users" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                {roles.reduce((sum, r) => sum + r.users, 0)}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-yellow-500/10 text-yellow-500">
                  Assigned
                </span>
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  Users with roles
                </span>
              </span>
            </div>
          </div>
        </Card>

        <Card>
          <div>
            <div className="flex">
              <div className="flex-1 text-base font-medium">Permissions</div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-purple-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:shield-check" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                {allPermissions.length}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-purple-500/10 text-purple-500">
                  Available
                </span>
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  System permissions
                </span>
              </span>
            </div>
          </div>
        </Card>
      </div>

      {/* Roles Grid */}
      <div className="grid xl:grid-cols-2 gap-5">
        {/* Roles List */}
        <Card title="Roles Management">
          <div className="flex justify-between items-center mb-4">
            <div className="flex space-x-2">
              <Button 
                icon="ph:plus" 
                className="btn-primary"
              >
                Add Role
              </Button>
            </div>
            <input
              type="text"
              placeholder="Search roles..."
              className="form-control max-w-xs"
            />
          </div>
          
          <div className="space-y-4">
            {roles.map((role) => (
              <div 
                key={role.id} 
                className={`p-4 border rounded-lg cursor-pointer transition-all ${
                  selectedRole?.id === role.id 
                    ? "border-indigo-500 bg-indigo-50 dark:bg-indigo-900/20" 
                    : "border-gray-200 dark:border-gray-700 hover:border-gray-300"
                }`}
                onClick={() => setSelectedRole(role)}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    <div className="h-10 w-10 rounded-full bg-indigo-500 text-white flex items-center justify-center">
                      <Icon icon="ph:lock-key" />
                    </div>
                    <div>
                      <h3 className="text-sm font-medium text-gray-900 dark:text-white">
                        {role.name}
                      </h3>
                      <p className="text-xs text-gray-500 dark:text-gray-400">
                        {role.users} users assigned
                      </p>
                    </div>
                  </div>
                  <div className="flex space-x-2">
                    <Button size="sm" className="btn-secondary">
                      <Icon icon="ph:pencil" />
                    </Button>
                    <Button size="sm" className="btn-danger">
                      <Icon icon="ph:trash" />
                    </Button>
                  </div>
                </div>
                <p className="mt-2 text-sm text-gray-600 dark:text-gray-300">
                  {role.description}
                </p>
                <div className="mt-2 flex items-center justify-between">
                  <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                    role.status === "Active" 
                      ? "bg-green-100 text-green-800" 
                      : "bg-red-100 text-red-800"
                  }`}>
                    {role.status}
                  </span>
                  <span className="text-xs text-gray-500 dark:text-gray-400">
                    {role.permissions.length} permissions
                  </span>
                </div>
              </div>
            ))}
          </div>
        </Card>

        {/* Permissions Details */}
        <Card title={selectedRole ? `${selectedRole.name} Permissions` : "Select a Role"}>
          {selectedRole ? (
            <div>
              <div className="mb-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <h3 className="font-medium text-gray-900 dark:text-white mb-2">
                  {selectedRole.name}
                </h3>
                <p className="text-sm text-gray-600 dark:text-gray-300 mb-2">
                  {selectedRole.description}
                </p>
                <div className="flex items-center space-x-4 text-xs text-gray-500 dark:text-gray-400">
                  <span>Created: {selectedRole.createdDate}</span>
                  <span>Users: {selectedRole.users}</span>
                  <span>Status: {selectedRole.status}</span>
                </div>
              </div>

              <div className="space-y-3">
                <div className="flex justify-between items-center">
                  <h4 className="font-medium text-gray-900 dark:text-white">Permissions</h4>
                  <Button size="sm" className="btn-secondary">
                    <Icon icon="ph:pencil" className="mr-1" />
                    Edit Permissions
                  </Button>
                </div>
                
                <div className="grid grid-cols-1 gap-2">
                  {allPermissions.map((permission) => (
                    <div 
                      key={permission}
                      className="flex items-center justify-between p-2 border rounded-lg"
                    >
                      <span className="text-sm text-gray-700 dark:text-gray-300">
                        {permission}
                      </span>
                      <div className="flex items-center">
                        <input
                          type="checkbox"
                          checked={selectedRole.permissions.includes(permission)}
                          readOnly
                          className="form-checkbox h-4 w-4 text-indigo-600"
                        />
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          ) : (
            <div className="text-center py-12">
              <Icon icon="ph:lock-key" className="mx-auto h-12 w-12 text-gray-400" />
              <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-white">
                No role selected
              </h3>
              <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                Select a role from the list to view its permissions.
              </p>
            </div>
          )}
        </Card>
      </div>
    </div>
  );
};

export default RolesPage;