import React, { useState } from "react";
import Card from "@/components/ui/Card";
import Icon from "@/components/ui/Icon";
import Button from "@/components/ui/Button";

const usersData = [
  {
    id: 1,
    name: "Admin User",
    email: "admin@tpos.com",
    role: "Administrator",
    department: "IT",
    status: "Active",
    lastLogin: "2024-01-15 14:30",
    permissions: ["Read", "Write", "Delete", "Admin"],
  },
  {
    id: 2,
    name: "Sales Manager",
    email: "sales@tpos.com",
    role: "Manager",
    department: "Sales",
    status: "Active",
    lastLogin: "2024-01-15 09:15",
    permissions: ["Read", "Write"],
  },
  {
    id: 3,
    name: "Support Agent",
    email: "support@tpos.com",
    role: "Agent",
    department: "Support",
    status: "Active",
    lastLogin: "2024-01-15 11:45",
    permissions: ["Read"],
  },
  {
    id: 4,
    name: "Marketing Lead",
    email: "marketing@tpos.com",
    role: "Lead",
    department: "Marketing",
    status: "Inactive",
    lastLogin: "2024-01-10 16:20",
    permissions: ["Read", "Write"],
  },
  {
    id: 5,
    name: "Finance Officer",
    email: "finance@tpos.com",
    role: "Officer",
    department: "Finance",
    status: "Active",
    lastLogin: "2024-01-15 08:00",
    permissions: ["Read", "Write"],
  },
];

const UsersPage = () => {
  const [users, setUsers] = useState(usersData);

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
                {users.length}
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
                {users.filter(u => u.status === "Active").length}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-green-500/10 text-green-500">
                  80%
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
                {users.filter(u => u.role === "Administrator").length}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-purple-500/10 text-purple-500">
                  Admin
                </span>
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  Super users
                </span>
              </span>
            </div>
          </div>
        </Card>

        <Card>
          <div>
            <div className="flex">
              <div className="flex-1 text-base font-medium">Departments</div>
              <div className="flex-none">
                <div className="h-10 w-10 rounded-full bg-yellow-500 text-white text-2xl flex items-center justify-center">
                  <Icon icon="ph:buildings" />
                </div>
              </div>
            </div>
            <div>
              <span className="text-2xl font-medium text-gray-800 dark:text-white">
                {new Set(users.map(u => u.department)).size}
              </span>
              <span className="space-x-2 block mt-4">
                <span className="badge bg-yellow-500/10 text-yellow-500">
                  5
                </span>
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  Active departments
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
            >
              Add User
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
            <select className="form-control">
              <option value="">All Departments</option>
              <option value="IT">IT</option>
              <option value="Sales">Sales</option>
              <option value="Support">Support</option>
              <option value="Marketing">Marketing</option>
              <option value="Finance">Finance</option>
            </select>
            <input
              type="text"
              placeholder="Search users..."
              className="form-control"
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
                  Role & Department
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Permissions
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Last Login
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
              {users.map((user) => (
                <tr key={user.id}>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <div className="flex-shrink-0 h-10 w-10">
                        <div className="h-10 w-10 rounded-full bg-gray-300 dark:bg-gray-600 flex items-center justify-center">
                          <Icon icon="ph:user" className="text-gray-500 dark:text-gray-400" />
                        </div>
                      </div>
                      <div className="ml-4">
                        <div className="text-sm font-medium text-gray-900 dark:text-white">
                          {user.name}
                        </div>
                        <div className="text-sm text-gray-500 dark:text-gray-400">
                          {user.email}
                        </div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-900 dark:text-white">
                      {user.role}
                    </div>
                    <div className="text-sm text-gray-500 dark:text-gray-400">
                      {user.department}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex flex-wrap gap-1">
                      {user.permissions.map((permission, index) => (
                        <span 
                          key={index}
                          className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                            permission === "Admin" 
                              ? "bg-purple-100 text-purple-800"
                              : permission === "Delete"
                              ? "bg-red-100 text-red-800"
                              : permission === "Write"
                              ? "bg-blue-100 text-blue-800"
                              : "bg-green-100 text-green-800"
                          }`}
                        >
                          {permission}
                        </span>
                      ))}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                    {user.lastLogin}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                      user.status === "Active" 
                        ? "bg-green-100 text-green-800" 
                        : "bg-red-100 text-red-800"
                    }`}>
                      {user.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                    <div className="flex space-x-2">
                      <Button size="sm" className="btn-secondary">
                        <Icon icon="ph:eye" />
                      </Button>
                      <Button size="sm" className="btn-secondary">
                        <Icon icon="ph:pencil" />
                      </Button>
                      <Button size="sm" className="btn-warning">
                        <Icon icon="ph:lock" />
                      </Button>
                      <Button size="sm" className="btn-danger">
                        <Icon icon="ph:trash" />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  );
};

export default UsersPage;