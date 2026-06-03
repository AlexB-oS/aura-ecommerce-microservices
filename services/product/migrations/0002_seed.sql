-- 0002_seed.sql — seed the catalog with the demo products so the storefront has
-- something to show immediately. Idempotent: ON CONFLICT (slug) DO NOTHING means
-- re-running the migrations never duplicates rows. Remove this migration once
-- you manage real inventory through the admin API.

INSERT INTO products (id, slug, name, tagline, description, category, price_cents, stock, badge) VALUES
  (gen_random_uuid(), 'halo-ambient-lamp',     'Halo Ambient Lamp',
     'Warm, dimmable glow with a featherlight aluminium body.',
     'A sculptural desk lamp with stepless dimming and a featherlight anodised aluminium body. USB-C powered.',
     'Home',      12900, 40,  'New'),
  (gen_random_uuid(), 'drift-headphones',      'Drift Wireless Headphones',
     '40h playback, adaptive noise cancelling, zero fatigue.',
     'Over-ear wireless headphones with adaptive noise cancelling, 40 hours of playback and memory-foam cushions.',
     'Audio',     21900, 25,  'Best seller'),
  (gen_random_uuid(), 'nimbus-daypack',        'Nimbus Daypack',
     'Weatherproof 22L pack that swallows a 16" laptop whole.',
     'A weatherproof 22-litre daypack with a padded 16-inch laptop sleeve and a magnetic roll-top closure.',
     'Travel',    14500, 60,  ''),
  (gen_random_uuid(), 'ember-temperature-mug', 'Ember Temperature Mug',
     'Keeps coffee at the exact degree you like, all morning.',
     'A smart mug that holds your coffee at a precise temperature for hours, with an app-controlled set point.',
     'Home',       9900,  8,  ''),
  (gen_random_uuid(), 'pulse-smart-watch',     'Pulse Smart Watch',
     'Health tracking and a 9-day battery in 32g of titanium.',
     'A titanium smart watch with continuous health tracking, GPS and a nine-day battery in a 32-gram case.',
     'Wearables', 27900, 18,  'New'),
  (gen_random_uuid(), 'lumen-keyboard',        'Lumen Mechanical Keyboard',
     'Hot-swappable switches and per-key aura backlighting.',
     'A 75% mechanical keyboard with hot-swappable switches, a gasket mount and per-key RGB backlighting.',
     'Desk',      16900, 30,  '')
ON CONFLICT (slug) DO NOTHING;
