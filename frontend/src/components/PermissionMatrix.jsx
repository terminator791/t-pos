import React, { useState, useMemo } from "react";
import Button from "@/components/ui/Button";
import Icon from "@/components/ui/Icon";
import { usePermissions, useBulkUpdatePermissions, useDeletePermission } from "@/services/api";
import { toast } from "react-toastify";

const SYSTEM_OBJECTS = [
  "users", "products", "customers", "licenses", "roles", "permissions",
  "dashboard", "reports", "settings", "admin"
];

const SYSTEM_ACTIONS = ["read", "write", "delete", "admin"];

const PermissionMatrix = ({ subject, domain, isEditable = true }) => {
  const [localPermissions, setLocalPermissions] = useState({});
  const [hasChanges, setHasChanges] = useState(false);
  
  const { data: permissions = [], isLoading, refetch } = usePermissions({
    subject,
    domain
  });
  
  const bulkUpdate = useBulkUpdatePermissions();
  const deletePermission = useDeletePermission();

  // Create permission matrix from API data
  const permissionMatrix = useMemo(() => {
    const matrix = {};
    SYSTEM_OBJECTS.forEach(obj => {
      matrix[obj] = {};
      SYSTEM_ACTIONS.forEach(action => {
        matrix[obj][action] = false;
      });
    });

    // Populate with existing permissions
    permissions.forEach(permission => {
      if (permission.object === "*") {
        // Grant all objects for this action
        SYSTEM_OBJECTS.forEach(obj => {
          if (permission.action === "*") {
            // Grant all actions for all objects
            SYSTEM_ACTIONS.forEach(action => {
              matrix[obj][action] = true;
            });
          } else {
            matrix[obj][permission.action] = true;
          }
        });
      } else if (matrix[permission.object]) {
        if (permission.action === "*") {
          // Grant all actions for this object
          SYSTEM_ACTIONS.forEach(action => {
            matrix[permission.object][action] = true;
          });
        } else {
          matrix[permission.object][permission.action] = true;
        }
      }
    });

    return matrix;
  }, [permissions]);

  // Merge with local changes
  const displayMatrix = useMemo(() => {
    const merged = { ...permissionMatrix };
    Object.keys(localPermissions).forEach(key => {
      const [obj, action] = key.split('.');
      if (merged[obj]) {
        merged[obj][action] = localPermissions[key];
      }
    });
    return merged;
  }, [permissionMatrix, localPermissions]);

  const handlePermissionChange = (object, action, checked) => {
    if (!isEditable) return;
    
    const key = `${object}.${action}`;
    setLocalPermissions(prev => ({
      ...prev,
      [key]: checked
    }));
    setHasChanges(true);
  };

  const handleSave = async () => {
    try {
      const permissionsToSave = [];
      
      Object.keys(displayMatrix).forEach(object => {
        Object.keys(displayMatrix[object]).forEach(action => {
          if (displayMatrix[object][action]) {
            permissionsToSave.push({ object, action });
          }
        });
      });

      await bulkUpdate.mutateAsync({
        subject,
        domain,
        permissions: permissionsToSave
      });

      setLocalPermissions({});
      setHasChanges(false);
      refetch();
    } catch (error) {
      console.error("Error saving permissions:", error);
    }
  };

  const handleReset = () => {
    setLocalPermissions({});
    setHasChanges(false);
  };

  const handleRevokePermission = async (permission) => {
    try {
      await deletePermission.mutateAsync(permission.id);
      refetch();
    } catch (error) {
      console.error("Error revoking permission:", error);
    }
  };

  const getPermissionCount = (object) => {
    return Object.values(displayMatrix[object] || {}).filter(Boolean).length;
  };

  const getTotalPermissions = () => {
    return Object.values(displayMatrix).reduce((total, obj) => {
      return total + Object.values(obj).filter(Boolean).length;
    }, 0);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Icon icon="ph:spinner" className="animate-spin h-8 w-8 text-gray-400" />
        <span className="ml-2 text-gray-500">Loading permissions...</span>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-medium text-gray-900 dark:text-white">
            Permission Matrix
          </h3>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            {subject} in domain "{domain}" - {getTotalPermissions()} permissions granted
          </p>
        </div>
        
        {isEditable && (
          <div className="flex space-x-2">
            {hasChanges && (
              <>
                <Button
                  size="sm"
                  className="btn-secondary"
                  onClick={handleReset}
                >
                  Reset
                </Button>
                <Button
                  size="sm"
                  className="btn-primary"
                  onClick={handleSave}
                  disabled={bulkUpdate.isPending}
                >
                  {bulkUpdate.isPending ? "Saving..." : "Save Changes"}
                </Button>
              </>
            )}
          </div>
        )}
      </div>

      {/* Permission Matrix Table */}
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead className="bg-gray-50 dark:bg-gray-800">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Object/Resource
              </th>
              {SYSTEM_ACTIONS.map(action => (
                <th
                  key={action}
                  className="px-6 py-3 text-center text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
                >
                  {action}
                </th>
              ))}
              <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Count
              </th>
            </tr>
          </thead>
          <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
            {SYSTEM_OBJECTS.map(object => (
              <tr key={object} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white capitalize">
                  {object}
                </td>
                {SYSTEM_ACTIONS.map(action => {
                  const isChecked = displayMatrix[object]?.[action] || false;
                  const hasLocalChange = localPermissions[`${object}.${action}`] !== undefined;
                  
                  return (
                    <td key={action} className="px-6 py-4 whitespace-nowrap text-center">
                      <input
                        type="checkbox"
                        checked={isChecked}
                        onChange={(e) => handlePermissionChange(object, action, e.target.checked)}
                        disabled={!isEditable}
                        className={`form-checkbox h-4 w-4 text-indigo-600 rounded ${
                          hasLocalChange ? "ring-2 ring-yellow-400" : ""
                        }`}
                      />
                    </td>
                  );
                })}
                <td className="px-6 py-4 whitespace-nowrap text-center text-sm text-gray-500 dark:text-gray-400">
                  <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                    getPermissionCount(object) > 0 
                      ? "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200"
                      : "bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300"
                  }`}>
                    {getPermissionCount(object)}/{SYSTEM_ACTIONS.length}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Raw Permissions List */}
      {permissions.length > 0 && (
        <div className="mt-6">
          <h4 className="text-sm font-medium text-gray-900 dark:text-white mb-3">
            Raw Casbin Policies ({permissions.length})
          </h4>
          <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 max-h-40 overflow-y-auto">
            <div className="space-y-1">
              {permissions.map((permission, index) => (
                <div
                  key={permission.id || index}
                  className="flex items-center justify-between text-xs font-mono text-gray-600 dark:text-gray-400"
                >
                  <span>
                    p, {permission.subject}, {permission.domain}, {permission.object}, {permission.action}
                  </span>
                  {isEditable && (
                    <Button
                      size="sm"
                      className="btn-danger"
                      onClick={() => handleRevokePermission(permission)}
                      disabled={deletePermission.isPending}
                    >
                      <Icon icon="ph:trash" className="h-3 w-3" />
                    </Button>
                  )}
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {hasChanges && (
        <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4">
          <div className="flex">
            <Icon icon="ph:warning" className="h-5 w-5 text-yellow-400" />
            <div className="ml-3">
              <h3 className="text-sm font-medium text-yellow-800 dark:text-yellow-200">
                Unsaved Changes
              </h3>
              <p className="mt-1 text-sm text-yellow-700 dark:text-yellow-300">
                You have made changes to the permission matrix. Don't forget to save your changes.
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default PermissionMatrix;