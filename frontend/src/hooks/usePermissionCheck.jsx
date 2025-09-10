import React from 'react';
import { useHasPermission } from '@/services/api';
import Icon from '@/components/ui/Icon';

/**
 * Custom hook for checking permissions in components
 * Follows Casbin RBAC model: subject, domain, object, action
 */
export const usePermissionCheck = () => {
  const checkPermission = (object, action, domain = "*") => {
    const { data: hasPermission = false } = useHasPermission(object, action, domain);
    return hasPermission;
  };

  const checkMultiplePermissions = (permissions) => {
    return permissions.every(({ object, action, domain = "*" }) => {
      return checkPermission(object, action, domain);
    });
  };

  return {
    checkPermission,
    checkMultiplePermissions,
    // Common permission checks for T-POS
    canManageUsers: (domain = "*") => checkPermission("users", "admin", domain),
    canCreateUsers: (domain = "*") => checkPermission("users", "write", domain),
    canViewUsers: (domain = "*") => checkPermission("users", "read", domain),
    canDeleteUsers: (domain = "*") => checkPermission("users", "delete", domain),
    
    canManageProducts: (domain = "*") => checkPermission("products", "admin", domain),
    canCreateProducts: (domain = "*") => checkPermission("products", "write", domain),
    canViewProducts: (domain = "*") => checkPermission("products", "read", domain),
    canDeleteProducts: (domain = "*") => checkPermission("products", "delete", domain),
    
    canManageRoles: (domain = "*") => checkPermission("roles", "admin", domain),
    canCreateRoles: (domain = "*") => checkPermission("roles", "write", domain),
    canViewRoles: (domain = "*") => checkPermission("roles", "read", domain),
    canDeleteRoles: (domain = "*") => checkPermission("roles", "delete", domain),
    
    canManagePermissions: (domain = "*") => checkPermission("permissions", "admin", domain),
    canGrantPermissions: (domain = "*") => checkPermission("permissions", "write", domain),
    canViewPermissions: (domain = "*") => checkPermission("permissions", "read", domain),
    canRevokePermissions: (domain = "*") => checkPermission("permissions", "delete", domain),
    
    canViewDashboard: (domain = "*") => checkPermission("dashboard", "read", domain),
    canViewReports: (domain = "*") => checkPermission("reports", "read", domain),
    canAccessAdmin: (domain = "*") => checkPermission("admin", "read", domain),
  };
};

/**
 * Higher-order component for permission-based rendering
 */
export const withPermission = (WrappedComponent, requiredPermissions) => {
  return function PermissionWrappedComponent(props) {
    const { checkMultiplePermissions } = usePermissionCheck();
    
    if (!checkMultiplePermissions(requiredPermissions)) {
      return (
        <div className="text-center py-8">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-red-100 rounded-full mb-4">
            <Icon icon="ph:warning" className="w-8 h-8 text-red-600" />
          </div>
          <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
            Access Denied
          </h3>
          <p className="text-gray-500 dark:text-gray-400">
            You don't have permission to access this feature.
          </p>
        </div>
      );
    }
    
    return <WrappedComponent {...props} />;
  };
};

/**
 * Component for conditional rendering based on permissions
 */
export const PermissionGate = ({ 
  object, 
  action, 
  domain = "*", 
  children, 
  fallback = null,
  multiple = false,
  permissions = []
}) => {
  const { checkPermission, checkMultiplePermissions } = usePermissionCheck();
  
  let hasPermission = false;
  
  if (multiple && permissions.length > 0) {
    hasPermission = checkMultiplePermissions(permissions);
  } else if (object && action) {
    hasPermission = checkPermission(object, action, domain);
  }
  
  return hasPermission ? children : fallback;
};

export default usePermissionCheck;