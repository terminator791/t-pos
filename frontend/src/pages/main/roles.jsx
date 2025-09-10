import React, { useState, useMemo } from "react";
import Card from "@/components/ui/Card";
import Icon from "@/components/ui/Icon";
import Button from "@/components/ui/Button";
import RoleModal from "@/components/modals/RoleModal";
import PermissionModal from "@/components/modals/PermissionModal";
import PermissionMatrix from "@/components/PermissionMatrix";
import DeleteConfirmModal from "@/components/ui/DeleteConfirmModal";
import { PermissionGate } from "@/hooks/usePermissionCheck";
import {
  useRoles,
  useDeleteRole,
  usePermissions,
  useDomains,
  useRoleAssignments,
} from "@/services/api";

const RolesPage = () => {
  const [selectedRole, setSelectedRole] = useState(null);
  const [selectedDomain, setSelectedDomain] = useState("*");
  const [searchTerm, setSearchTerm] = useState("");
  const [isRoleModalOpen, setIsRoleModalOpen] = useState(false);
  const [isPermissionModalOpen, setIsPermissionModalOpen] = useState(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [editingRole, setEditingRole] = useState(null);
  const [roleToDelete, setRoleToDelete] = useState(null);

  // API hooks
  const { data: rolesData, isLoading: rolesLoading } = useRoles();
  const { data: permissionsData } = usePermissions();
  const { data: domainsData } = useDomains();
  const { data: roleAssignmentsData } = useRoleAssignments();
  const deleteRole = useDeleteRole();

  // Safely extract data arrays
  const roles = Array.isArray(rolesData?.data)
    ? rolesData.data
    : Array.isArray(rolesData)
    ? rolesData
    : [];
  const permissions = Array.isArray(permissionsData?.data)
    ? permissionsData.data
    : Array.isArray(permissionsData)
    ? permissionsData
    : [];
  const domains = Array.isArray(domainsData?.data)
    ? domainsData.data
    : Array.isArray(domainsData)
    ? domainsData
    : [];
  const roleAssignments = Array.isArray(roleAssignmentsData?.data)
    ? roleAssignmentsData.data
    : Array.isArray(roleAssignmentsData)
    ? roleAssignmentsData
    : [];

  // Filter roles based on search term and domain
  const filteredRoles = useMemo(() => {
    if (!Array.isArray(roles)) return [];
    return roles.filter((role) => {
      if (!role || !role.name) return false;
      const matchesSearch =
        role.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        (role.description &&
          role.description.toLowerCase().includes(searchTerm.toLowerCase()));
      const matchesDomain =
        selectedDomain === "*" || role.domain === selectedDomain;
      return matchesSearch && matchesDomain;
    });
  }, [roles, searchTerm, selectedDomain]);

  // Get statistics
  const stats = useMemo(() => {
    const totalRoles = roles.length;
    const activeRoles = roles.filter((r) => r.status !== "inactive").length;
    const totalUsers = roleAssignments.length;
    const totalPermissions = permissions.length;

    return {
      totalRoles,
      activeRoles,
      totalUsers,
      totalPermissions,
    };
  }, [roles, roleAssignments, permissions]);

  // Get role assignment count
  const getRoleUserCount = (roleName, domain) => {
    return roleAssignments.filter(
      (assignment) =>
        assignment.role_name === roleName &&
        (domain === "*" || assignment.domain === domain)
    ).length;
  };

  // Get role permission count
  const getRolePermissionCount = (roleName, domain) => {
    return permissions.filter(
      (permission) =>
        permission.subject === roleName &&
        (domain === "*" || permission.domain === domain)
    ).length;
  };

  const handleEditRole = (role) => {
    setEditingRole(role);
    setIsRoleModalOpen(true);
  };

  const handleDeleteRole = (role) => {
    setRoleToDelete(role);
    setIsDeleteModalOpen(true);
  };

  const confirmDelete = async () => {
    if (roleToDelete) {
      try {
        await deleteRole.mutateAsync(roleToDelete.id);
        setIsDeleteModalOpen(false);
        setRoleToDelete(null);
        if (selectedRole?.id === roleToDelete.id) {
          setSelectedRole(null);
        }
      } catch (error) {
        console.error("Error deleting role:", error);
      }
    }
  };

  const closeRoleModal = () => {
    setIsRoleModalOpen(false);
    setEditingRole(null);
  };

  return (
    <div className="space-y-5">
      {/* Permission Check Warning */}
      <PermissionGate
        object="roles"
        action="read"
        fallback={
          <Card>
            <div className="text-center py-8">
              <Icon
                icon="ph:shield-warning"
                className="mx-auto h-12 w-12 text-red-400"
              />
              <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-white">
                Access Restricted
              </h3>
              <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                You don't have permission to view roles and permissions.
              </p>
            </div>
          </Card>
        }
      >
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
                  {stats.totalRoles}
                </span>
                <span className="space-x-2 block mt-4">
                  <span className="badge bg-indigo-500/10 text-indigo-500">
                    Active
                  </span>
                  <span className="text-sm text-gray-500 dark:text-gray-400">
                    Casbin RBAC roles
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
                  {stats.activeRoles}
                </span>
                <span className="space-x-2 block mt-4">
                  <span className="badge bg-green-500/10 text-green-500">
                    {stats.totalRoles > 0
                      ? Math.round((stats.activeRoles / stats.totalRoles) * 100)
                      : 0}
                    %
                  </span>
                  <span className="text-sm text-gray-500 dark:text-gray-400">
                    Operational
                  </span>
                </span>
              </div>
            </div>
          </Card>

          <Card>
            <div>
              <div className="flex">
                <div className="flex-1 text-base font-medium">
                  Role Assignments
                </div>
                <div className="flex-none">
                  <div className="h-10 w-10 rounded-full bg-yellow-500 text-white text-2xl flex items-center justify-center">
                    <Icon icon="ph:users" />
                  </div>
                </div>
              </div>
              <div>
                <span className="text-2xl font-medium text-gray-800 dark:text-white">
                  {stats.totalUsers}
                </span>
                <span className="space-x-2 block mt-4">
                  <span className="badge bg-yellow-500/10 text-yellow-500">
                    Assigned
                  </span>
                  <span className="text-sm text-gray-500 dark:text-gray-400">
                    User-role mappings
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
                  {stats.totalPermissions}
                </span>
                <span className="space-x-2 block mt-4">
                  <span className="badge bg-purple-500/10 text-purple-500">
                    Policies
                  </span>
                  <span className="text-sm text-gray-500 dark:text-gray-400">
                    Casbin policies
                  </span>
                </span>
              </div>
            </div>
          </Card>
        </div>

        {/* Main Content */}
        <div className="grid xl:grid-cols-2 gap-5">
          {/* Roles List */}
          <Card title="Roles Management">
            <div className="flex justify-between items-center mb-4">
              <div className="flex space-x-2">
                <PermissionGate object="roles" action="write">
                  <Button
                    icon="ph:plus"
                    className="btn-primary"
                    onClick={() => setIsRoleModalOpen(true)}
                  >
                    Add Role
                  </Button>
                </PermissionGate>
                <PermissionGate object="permissions" action="write">
                  <Button
                    icon="ph:shield-plus"
                    className="btn-secondary"
                    onClick={() => setIsPermissionModalOpen(true)}
                    disabled={!selectedRole}
                  >
                    Grant Permission
                  </Button>
                </PermissionGate>
              </div>
            </div>

            {/* Filters */}
            <div className="flex space-x-2 mb-4">
              <input
                type="text"
                placeholder="Search roles..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="form-control flex-1"
              />
              <select
                value={selectedDomain}
                onChange={(e) => setSelectedDomain(e.target.value)}
                className="form-control w-40"
              >
                <option value="*">All Domains</option>
                {domains.map((domain) => (
                  <option key={domain.id} value={domain.name}>
                    {domain.name}
                  </option>
                ))}
              </select>
            </div>

            <div className="space-y-4 max-h-96 overflow-y-auto">
              {rolesLoading ? (
                <div className="flex items-center justify-center py-8">
                  <Icon
                    icon="ph:spinner"
                    className="animate-spin h-6 w-6 text-gray-400"
                  />
                  <span className="ml-2 text-gray-500">Loading roles...</span>
                </div>
              ) : filteredRoles.length === 0 ? (
                <div className="text-center py-8">
                  <Icon
                    icon="ph:lock-key"
                    className="mx-auto h-12 w-12 text-gray-400"
                  />
                  <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-white">
                    No roles found
                  </h3>
                  <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                    {searchTerm || selectedDomain !== "*"
                      ? "Try adjusting your search or filter."
                      : "Get started by creating a new role."}
                  </p>
                </div>
              ) : (
                filteredRoles.map((role) => (
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
                            {getRoleUserCount(role.name, role.domain)} users •{" "}
                            {role.domain === "*" ? "All domains" : role.domain}
                          </p>
                        </div>
                      </div>
                      <div className="flex space-x-2">
                        <PermissionGate object="roles" action="write">
                          <Button
                            size="sm"
                            className="btn-secondary"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleEditRole(role);
                            }}
                          >
                            <Icon icon="ph:pencil" />
                          </Button>
                        </PermissionGate>
                        <PermissionGate object="roles" action="delete">
                          <Button
                            size="sm"
                            className="btn-danger"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleDeleteRole(role);
                            }}
                          >
                            <Icon icon="ph:trash" />
                          </Button>
                        </PermissionGate>
                      </div>
                    </div>
                    <p className="mt-2 text-sm text-gray-600 dark:text-gray-300">
                      {role.description || "No description provided"}
                    </p>
                    <div className="mt-2 flex items-center justify-between">
                      <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
                        Active
                      </span>
                      <span className="text-xs text-gray-500 dark:text-gray-400">
                        {getRolePermissionCount(role.name, role.domain)}{" "}
                        permissions
                      </span>
                    </div>
                  </div>
                ))
              )}
            </div>
          </Card>

          {/* Permission Matrix */}
          <Card
            title={
              selectedRole
                ? `${selectedRole.name} Permissions`
                : "Select a Role"
            }
          >
            {selectedRole ? (
              <div>
                <div className="mb-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                  <h3 className="font-medium text-gray-900 dark:text-white mb-2">
                    {selectedRole.name}
                  </h3>
                  <p className="text-sm text-gray-600 dark:text-gray-300 mb-2">
                    {selectedRole.description || "No description provided"}
                  </p>
                  <div className="flex items-center space-x-4 text-xs text-gray-500 dark:text-gray-400">
                    <span>Domain: {selectedRole.domain}</span>
                    <span>
                      Users:{" "}
                      {getRoleUserCount(selectedRole.name, selectedRole.domain)}
                    </span>
                    <span>
                      Permissions:{" "}
                      {getRolePermissionCount(
                        selectedRole.name,
                        selectedRole.domain
                      )}
                    </span>
                  </div>
                </div>

                <PermissionGate
                  object="permissions"
                  action="read"
                  fallback={
                    <div className="text-center py-8">
                      <Icon
                        icon="ph:shield-warning"
                        className="mx-auto h-8 w-8 text-orange-400"
                      />
                      <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
                        No permission to view role permissions
                      </p>
                    </div>
                  }
                >
                  <PermissionMatrix
                    subject={selectedRole.name}
                    domain={selectedRole.domain}
                    isEditable={true}
                  />
                </PermissionGate>
              </div>
            ) : (
              <div className="text-center py-12">
                <Icon
                  icon="ph:lock-key"
                  className="mx-auto h-12 w-12 text-gray-400"
                />
                <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-white">
                  No role selected
                </h3>
                <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  Select a role from the list to view and manage its
                  permissions.
                </p>
              </div>
            )}
          </Card>
        </div>

        {/* Casbin Information */}
        <Card title="Casbin RBAC Configuration">
          <div className="grid md:grid-cols-2 gap-6">
            <div>
              <h4 className="text-sm font-medium text-gray-900 dark:text-white mb-3">
                Current Configuration
              </h4>
              <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 font-mono text-xs">
                <div className="space-y-1 text-gray-600 dark:text-gray-400">
                  <div>
                    <span className="text-blue-600 dark:text-blue-400">
                      [request_definition]
                    </span>
                  </div>
                  <div>r = sub, dom, obj, act</div>
                  <div className="mt-2">
                    <span className="text-green-600 dark:text-green-400">
                      [policy_definition]
                    </span>
                  </div>
                  <div>p = sub, dom, obj, act</div>
                  <div className="mt-2">
                    <span className="text-purple-600 dark:text-purple-400">
                      [role_definition]
                    </span>
                  </div>
                  <div>g = _, _, _</div>
                  <div className="mt-2">
                    <span className="text-orange-600 dark:text-orange-400">
                      [matchers]
                    </span>
                  </div>
                  <div>
                    m = g(r.sub, p.sub, r.dom) && (r.dom == p.dom || p.dom ==
                    "*") && keyMatch2(r.obj, p.obj) && (r.act == p.act ||
                    regexMatch(r.act, p.act))
                  </div>
                </div>
              </div>
            </div>
            <div>
              <h4 className="text-sm font-medium text-gray-900 dark:text-white mb-3">
                Model Explanation
              </h4>
              <div className="space-y-3 text-sm text-gray-600 dark:text-gray-400">
                <div>
                  <span className="font-medium text-gray-900 dark:text-white">
                    Request:
                  </span>
                  <span className="ml-2">subject, domain, object, action</span>
                </div>
                <div>
                  <span className="font-medium text-gray-900 dark:text-white">
                    Policy:
                  </span>
                  <span className="ml-2">subject, domain, object, action</span>
                </div>
                <div>
                  <span className="font-medium text-gray-900 dark:text-white">
                    Grouping:
                  </span>
                  <span className="ml-2">user, role, domain</span>
                </div>
                <div className="pt-2 border-t border-gray-200 dark:border-gray-700">
                  <p className="text-xs">
                    This configuration supports multi-tenant RBAC with domain
                    isolation, wildcard permissions, and flexible object/action
                    matching.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </Card>

        {/* Modals */}
        <RoleModal
          isOpen={isRoleModalOpen}
          onClose={closeRoleModal}
          role={editingRole}
          isEditing={!!editingRole}
        />

        <PermissionModal
          isOpen={isPermissionModalOpen}
          onClose={() => setIsPermissionModalOpen(false)}
          defaultSubject={selectedRole?.name || ""}
          defaultDomain={selectedRole?.domain || ""}
        />

        <DeleteConfirmModal
          isOpen={isDeleteModalOpen}
          onClose={() => setIsDeleteModalOpen(false)}
          onConfirm={confirmDelete}
          title="Delete Role"
          message={`Are you sure you want to delete the role "${roleToDelete?.name}"? This will remove all associated permissions and role assignments.`}
          isLoading={deleteRole.isPending}
        />
      </PermissionGate>
    </div>
  );
};

export default RolesPage;
