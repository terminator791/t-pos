## Role-Based Access Control (RBAC)

### Role Definitions

- ACL WITH CASBIN

| Role             | Access Level   | Shop Access          | Sync Capabilities   | Domain              |
| ---------------- | -------------- | -------------------- | ------------------- | ------------------- |
| `super_admin`    | Global         | All shops            | Full sync access    | \*                  |
| `admin`          | Global         | All shops            | Full sync access    | \*                  |
| `owner_business` | License-scoped | License shops only   | Business data sync  | users -> license_id |
| `cashier`        | Shop-scoped    | Single assigned shop | Limited entity sync | users -> shop_id    |

## AUTH WITH JWT

## GORM FOR ORM
