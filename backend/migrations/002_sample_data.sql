-- Sample data for t-pos system
-- This script inserts initial data for development and testing

-- Insert default admin user
INSERT INTO users (email, username, password, first_name, last_name, role, is_active)
VALUES 
    ('admin@tpos.com', 'admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Admin', 'User', 'admin', true),
    ('manager@tpos.com', 'manager', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Store', 'Manager', 'manager', true),
    ('cashier@tpos.com', 'cashier', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'John', 'Cashier', 'cashier', true)
ON CONFLICT (email) DO NOTHING;

-- Insert default categories
INSERT INTO categories (name, description, is_active, sort_order)
VALUES 
    ('Electronics', 'Electronic products and gadgets', true, 1),
    ('Clothing', 'Clothing and apparel', true, 2),
    ('Food & Beverage', 'Food and drink products', true, 3),
    ('Home & Garden', 'Home and garden products', true, 4),
    ('Books & Media', 'Books, magazines, and media', true, 5),
    ('Sports & Outdoors', 'Sports and outdoor equipment', true, 6),
    ('Health & Beauty', 'Health and beauty products', true, 7),
    ('Toys & Games', 'Toys and games for all ages', true, 8)
ON CONFLICT (name) DO NOTHING;

-- Get category IDs for products
-- Insert sample products
WITH category_electronics AS (SELECT id FROM categories WHERE name = 'Electronics' LIMIT 1),
     category_clothing AS (SELECT id FROM categories WHERE name = 'Clothing' LIMIT 1),
     category_food AS (SELECT id FROM categories WHERE name = 'Food & Beverage' LIMIT 1),
     category_books AS (SELECT id FROM categories WHERE name = 'Books & Media' LIMIT 1)

INSERT INTO products (sku, name, description, category_id, price, cost, barcode, stock, min_stock, max_stock, is_active, is_taxable, tax_rate)
VALUES 
    -- Electronics
    ('ELEC-001', 'Wireless Bluetooth Headphones', 'High-quality wireless headphones with noise cancellation', (SELECT id FROM category_electronics), 99.99, 60.00, '1234567890123', 25, 5, 100, true, true, 0.0825),
    ('ELEC-002', 'Smartphone Charging Cable', 'USB-C to Lightning charging cable', (SELECT id FROM category_electronics), 19.99, 8.00, '1234567890124', 50, 10, 200, true, true, 0.0825),
    ('ELEC-003', 'Portable Power Bank', '10000mAh portable battery charger', (SELECT id FROM category_electronics), 39.99, 20.00, '1234567890125', 30, 5, 80, true, true, 0.0825),
    
    -- Clothing
    ('CLOTH-001', 'Cotton T-Shirt', 'Comfortable cotton t-shirt in various colors', (SELECT id FROM category_clothing), 24.99, 12.00, '2234567890123', 100, 20, 300, true, true, 0.0825),
    ('CLOTH-002', 'Denim Jeans', 'Classic fit denim jeans', (SELECT id FROM category_clothing), 59.99, 30.00, '2234567890124', 40, 10, 120, true, true, 0.0825),
    ('CLOTH-003', 'Winter Jacket', 'Warm winter jacket with hood', (SELECT id FROM category_clothing), 129.99, 70.00, '2234567890125', 15, 3, 50, true, true, 0.0825),
    
    -- Food & Beverage
    ('FOOD-001', 'Organic Coffee Beans', 'Premium organic coffee beans, 1lb bag', (SELECT id FROM category_food), 14.99, 7.00, '3234567890123', 60, 10, 200, true, false, 0.0000),
    ('FOOD-002', 'Energy Drink', 'Sugar-free energy drink, 16oz can', (SELECT id FROM category_food), 2.99, 1.20, '3234567890124', 200, 50, 500, true, false, 0.0000),
    ('FOOD-003', 'Protein Bar', 'High-protein nutrition bar', (SELECT id FROM category_food), 3.49, 1.80, '3234567890125', 150, 30, 400, true, false, 0.0000),
    
    -- Books & Media
    ('BOOK-001', 'Business Strategy Guide', 'Complete guide to business strategy', (SELECT id FROM category_books), 29.99, 15.00, '4234567890123', 20, 5, 80, true, false, 0.0000),
    ('BOOK-002', 'Programming Fundamentals', 'Learn programming from scratch', (SELECT id FROM category_books), 39.99, 20.00, '4234567890124', 25, 5, 100, true, false, 0.0000)
ON CONFLICT (sku) DO NOTHING;

-- Insert sample customers
INSERT INTO customers (first_name, last_name, email, phone, address, city, state, zip_code, country, is_active)
VALUES 
    ('John', 'Smith', 'john.smith@email.com', '+1-555-0101', '123 Main St', 'Anytown', 'CA', '12345', 'US', true),
    ('Jane', 'Doe', 'jane.doe@email.com', '+1-555-0102', '456 Oak Ave', 'Somewhere', 'NY', '67890', 'US', true),
    ('Michael', 'Johnson', 'michael.j@email.com', '+1-555-0103', '789 Pine Rd', 'Elsewhere', 'TX', '54321', 'US', true),
    ('Sarah', 'Williams', 'sarah.w@email.com', '+1-555-0104', '321 Elm St', 'Nowhere', 'FL', '98765', 'US', true),
    ('David', 'Brown', 'david.brown@email.com', '+1-555-0105', '654 Maple Dr', 'Anywhere', 'WA', '13579', 'US', true)
ON CONFLICT (email) DO NOTHING;