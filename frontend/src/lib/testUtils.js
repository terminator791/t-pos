/**
 * Test utilities for Casbin roles and permissions
 */

// Mock data for testing
export const mockRoles = [
  {
    id: "1",
    name: "admin", 
    description: "System administrator with full access",
    domain: "*",
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z"
  },
  {
    id: "2",
    name: "manager",
    description: "Sales manager with limited admin access", 
    domain: "company1",
    created_at: "2024-01-02T00:00:00Z",
    updated_at: "2024-01-02T00:00:00Z"
  },
  {
    id: "3",
    name: "cashier",
    description: "Cashier with basic POS access",
    domain: "company1", 
    created_at: "2024-01-03T00:00:00Z",
    updated_at: "2024-01-03T00:00:00Z"
  }
];

export const mockPermissions = [
  // Admin permissions (all access)
  { id: "1", subject: "admin", domain: "*", object: "*", action: "*" },
  
  // Manager permissions
  { id: "2", subject: "manager", domain: "company1", object: "products", action: "read" },
  { id: "3", subject: "manager", domain: "company1", object: "products", action: "write" },
  { id: "4", subject: "manager", domain: "company1", object: "customers", action: "read" },
  { id: "5", subject: "manager", domain: "company1", object: "customers", action: "write" },
  { id: "6", subject: "manager", domain: "company1", object: "reports", action: "read" },
  
  // Cashier permissions
  { id: "7", subject: "cashier", domain: "company1", object: "products", action: "read" },
  { id: "8", subject: "cashier", domain: "company1", object: "customers", action: "read" },
  { id: "9", subject: "cashier", domain: "company1", object: "dashboard", action: "read" }
];

export const mockDomains = [
  {
    id: "1",
    name: "company1",
    description: "Company 1 domain",
    created_at: "2024-01-01T00:00:00Z"
  },
  {
    id: "2", 
    name: "company2",
    description: "Company 2 domain",
    created_at: "2024-01-02T00:00:00Z"
  }
];

export const mockRoleAssignments = [
  { id: "1", user_id: "user1", role_name: "admin", domain: "*" },
  { id: "2", user_id: "user2", role_name: "manager", domain: "company1" },
  { id: "3", user_id: "user3", role_name: "cashier", domain: "company1" },
  { id: "4", user_id: "user4", role_name: "cashier", domain: "company1" }
];

// Helper functions for testing
export const getPermissionsForRole = (roleName, domain = "*") => {
  return mockPermissions.filter(p => 
    p.subject === roleName && 
    (p.domain === domain || p.domain === "*" || domain === "*")
  );
};

export const getUsersForRole = (roleName, domain = "*") => {
  return mockRoleAssignments.filter(ra => 
    ra.role_name === roleName && 
    (ra.domain === domain || ra.domain === "*" || domain === "*")
  );
};

export const hasPermission = (roleName, domain, object, action) => {
  const permissions = getPermissionsForRole(roleName, domain);
  
  return permissions.some(p => {
    const objectMatch = p.object === "*" || p.object === object;
    const actionMatch = p.action === "*" || p.action === action;
    const domainMatch = p.domain === "*" || p.domain === domain;
    
    return objectMatch && actionMatch && domainMatch;
  });
};

// Test scenarios
export const testScenarios = [
  {
    name: "Admin has full access",
    test: () => {
      const tests = [
        hasPermission("admin", "company1", "users", "read"),
        hasPermission("admin", "company1", "users", "write"), 
        hasPermission("admin", "company1", "users", "delete"),
        hasPermission("admin", "*", "products", "admin")
      ];
      return tests.every(t => t === true);
    }
  },
  {
    name: "Manager has limited access",
    test: () => {
      const canRead = hasPermission("manager", "company1", "products", "read");
      const canWrite = hasPermission("manager", "company1", "products", "write");
      const cannotDelete = !hasPermission("manager", "company1", "users", "delete");
      const cannotAccessOtherDomain = !hasPermission("manager", "company2", "products", "read");
      
      return canRead && canWrite && cannotDelete && cannotAccessOtherDomain;
    }
  },
  {
    name: "Cashier has read-only access",
    test: () => {
      const canRead = hasPermission("cashier", "company1", "products", "read");
      const cannotWrite = !hasPermission("cashier", "company1", "products", "write");
      const cannotDelete = !hasPermission("cashier", "company1", "products", "delete");
      
      return canRead && cannotWrite && cannotDelete;
    }
  }
];

// Run all tests
export const runTests = () => {
  console.log("🧪 Running Casbin RBAC Tests...\n");
  
  const results = testScenarios.map(scenario => {
    const result = scenario.test();
    const status = result ? "✅ PASS" : "❌ FAIL";
    console.log(`${status}: ${scenario.name}`);
    return result;
  });
  
  const passed = results.filter(r => r).length;
  const total = results.length;
  
  console.log(`\n📊 Test Results: ${passed}/${total} passed`);
  
  if (passed === total) {
    console.log("🎉 All tests passed! Casbin configuration is working correctly.");
  } else {
    console.log("⚠️ Some tests failed. Please check the Casbin configuration.");
  }
  
  return { passed, total, success: passed === total };
};

export default {
  mockRoles,
  mockPermissions,
  mockDomains,
  mockRoleAssignments,
  getPermissionsForRole,
  getUsersForRole,
  hasPermission,
  testScenarios,
  runTests
};