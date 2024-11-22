<?php

$q1 = "SELECT * FROM users WHERE status = 'active'";
$q2 = "SELECT id, username, email FROM customers ORDER BY created_at DESC";
$q3 = "SELECT * FROM users WHERE status = 'active'";  // Duplicate of q1
$q4 = "SELECT product_name, price FROM products WHERE category_id = 5";
$q5 = "SELECT COUNT(*) as total FROM orders WHERE order_date >= '2024-01-01'";

// Joins and more complex queries
$q6 = "SELECT o.order_id, c.customer_name, p.product_name
       FROM orders o
       JOIN customers c ON o.customer_id = c.id
       JOIN products p ON o.product_id = p.id
       WHERE o.status = 'pending'";

$q6v2 = "SELECT o.order_id, c.customer_name, p.product_name
      FROM orders o
      JOIN customers c ON o.customer_id = c.id
      JOIN products p ON o.product_id = p.id
      WHERE o.status = 'pending'";

$q7 = "SELECT id, username, email FROM customers ORDER BY created_at DESC";  // Duplicate of q2
$q8 = "SELECT COUNT(*) as total FROM orders WHERE order_date >= '2024-01-01'";  // Duplicate of q5

// Group by and aggregates
$q9 = "SELECT category_id, COUNT(*) as product_count, AVG(price) as avg_price
       FROM products
       GROUP BY category_id
       HAVING avg_price > 100";

$q10 = "SELECT department, SUM(salary) as total_payroll
        FROM employees
        WHERE active = 1
        GROUP BY department";

// Updates and inserts
$q11 = "UPDATE users SET last_login = NOW() WHERE id = 42";
$q12 = "INSERT INTO logs (user_id, action, timestamp) VALUES (123, 'login', NOW())";
$q13 = "UPDATE users SET last_login = NOW() WHERE id = 42";  // Duplicate of q11