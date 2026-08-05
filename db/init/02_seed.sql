TRUNCATE orders, products, customers RESTART IDENTITY CASCADE;

-- customers --
WITH first_names AS (
    SELECT unnest(ARRAY[
        'Emma','Liam','Olivia','Noah','Ava','Lucas','Sophia','Mason','Isabella','Ethan',
        'Mia','Alexander','Charlotte','James','Amelia','Benjamin','Harper','Elijah','Evelyn','Daniel',
        'Camille','Louis','Chloé','Hugo','Léa','Nathan','Manon','Gabriel','Inès','Arthur',
        'Sofia','Matteo','Giulia','Marco','Anna','Luca','Elena','Andrea','Francesca','Alessandro',
        'Hans','Greta','Lukas','Anja','Felix','Mia','Jonas','Lena','Paul','Marie'
    ]) AS name
),
last_names AS (
    SELECT unnest(ARRAY[
        'Smith','Johnson','Williams','Brown','Jones','Garcia','Miller','Davis','Rodriguez','Martinez',
        'Dubois','Lefevre','Moreau','Bernard','Petit','Durand','Leroy','Rousseau','Fontaine','Girard',
        'Rossi','Russo','Ferrari','Esposito','Bianchi','Romano','Colombo','Ricci','Marino','Greco',
        'Müller','Schmidt','Schneider','Fischer','Weber','Meyer','Wagner','Becker','Hoffmann','Schulz',
        'Rakoto','Andria','Ravelo','Randria','Rasoa','Ranaivo','Rabe','Andria','Rafara','Zafy'
    ]) AS name
),
picked AS (
    SELECT
        f.name AS first_name,
        l.name AS last_name,
        row_number() OVER () AS rn
    FROM (SELECT name, floor(random()*1000)::int AS ord FROM first_names, generate_series(1,10)) f
    CROSS JOIN LATERAL (SELECT name FROM last_names ORDER BY random() LIMIT 1) l
    ORDER BY random()
    LIMIT 500
)
INSERT INTO customers (name, email, country, created_at)
SELECT
    first_name || ' ' || last_name,
    lower(first_name) || '.' || lower(last_name) || rn || '@gmail.com',
    (ARRAY['FR','US','DE','MG','CA','BE','GB','ES','IT','NL'])[1 + floor(random() * 10)::int],
    NOW() - (random() * INTERVAL '730 days')
FROM picked;

-- products --
INSERT INTO products (name, category, price)
VALUES
    ('Wireless Mechanical Keyboard', 'Electronics', 89.99),
    ('Noise-Cancelling Headphones', 'Electronics', 149.99),
    ('4K Ultra HD Webcam', 'Electronics', 64.50),
    ('Portable SSD 1TB', 'Electronics', 109.00),
    ('Smart Home Speaker', 'Electronics', 79.99),
    ('USB-C Docking Station', 'Electronics', 59.90),
    ('Ergonomic Wireless Mouse', 'Electronics', 34.99),
    ('27-inch 4K Monitor', 'Electronics', 329.00),
    ('The Silent Patient', 'Books', 12.99),
    ('Atomic Habits', 'Books', 15.50),
    ('Clean Architecture', 'Books', 34.90),
    ('Designing Data-Intensive Applications', 'Books', 42.00),
    ('The Pragmatic Programmer', 'Books', 29.99),
    ('Sapiens: A Brief History of Humankind', 'Books', 13.75),
    ('Project Hail Mary', 'Books', 11.99),
    ('The Go Programming Language', 'Books', 38.50),
    ('Organic Cotton T-Shirt', 'Clothing', 24.99),
    ('Merino Wool Sweater', 'Clothing', 79.00),
    ('Waterproof Rain Jacket', 'Clothing', 119.99),
    ('Classic Denim Jeans', 'Clothing', 59.90),
    ('Running Shoes', 'Clothing', 89.50),
    ('Leather Crossbody Bag', 'Clothing', 145.00),
    ('Wool Beanie', 'Clothing', 19.99),
    ('Ceramic Non-Stick Pan Set', 'Home', 74.99),
    ('Stainless Steel Water Bottle', 'Home', 22.50),
    ('Scented Soy Candle', 'Home', 18.99),
    ('Linen Bedding Set', 'Home', 129.00),
    ('Cast Iron Dutch Oven', 'Home', 89.99),
    ('Espresso Machine', 'Home', 249.00),
    ('Bamboo Cutting Board', 'Home', 27.99),
    ('Air-Purifying Plant Pot', 'Home', 34.50),
    ('Wooden Building Blocks Set', 'Toys', 39.99),
    ('Remote Control Drone', 'Toys', 89.00),
    ('STEM Robotics Kit', 'Toys', 64.99),
    ('Board Game: Strategy Classics', 'Toys', 29.50),
    ('Plush Teddy Bear', 'Toys', 19.99),
    ('Puzzle 1000 Pieces', 'Toys', 14.99),
    ('Yoga Mat', 'Sports', 34.99),
    ('Adjustable Dumbbell Set', 'Sports', 179.00),
    ('Insulated Sports Bottle', 'Sports', 24.50),
    ('Trail Running Backpack', 'Sports', 69.99),
    ('Resistance Bands Set', 'Sports', 19.99),
    ('Foldable Camping Chair', 'Sports', 44.90),
    ('Bike Repair Kit', 'Sports', 27.50);

-- orders --
INSERT INTO orders (customer_id, product_id, quantity, order_date, status)
SELECT
    1 + floor(random() * 500)::int,
    1 + floor(random() * 44)::int,
    1 + floor(random() * 5)::int,
    NOW() - (random() * INTERVAL '365 days'),
    (ARRAY['pending','shipped','delivered','cancelled'])[1 + floor(random() * 4)::int]
FROM generate_series(1, 100000) AS s(i);