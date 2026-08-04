-- 10 000 clients
INSERT INTO customers (name, email, country, created_at)
SELECT
    'Customer ' || i,
    'customer' || i || '@example.com',
    (ARRAY['FR','US','DE','MG','CA','BE','GB','ES'])[1 + floor(random() * 8)::int],
    NOW() - (random() * INTERVAL '730 days')
FROM generate_series(1, 10000) AS s(i);

-- 200 produits
INSERT INTO products (name, category, price)
SELECT
    'Product ' || i,
    (ARRAY['Electronics','Books','Clothing','Home','Toys','Sports'])[1 + floor(random() * 6)::int],
    round((random() * 500 + 5)::numeric, 2)
FROM generate_series(1, 200) AS s(i);

-- 100 000 commandes
INSERT INTO orders (customer_id, product_id, quantity, order_date, status)
SELECT
    1 + floor(random() * 10000)::int,
    1 + floor(random() * 200)::int,
    1 + floor(random() * 5)::int,
    NOW() - (random() * INTERVAL '365 days'),
    (ARRAY['pending','shipped','delivered','cancelled'])[1 + floor(random() * 4)::int]
FROM generate_series(1, 100000) AS s(i);