-- 插入测试游戏数据
INSERT INTO games (id, slug, title, subtitle, description, genre, tags, engine, status, developer_id, created_at, updated_at)
VALUES
  (gen_random_uuid(), 'pixel-adventure', '像素冒险', '经典像素风格冒险游戏', '在这个充满挑战的像素世界中展开冒险，探索未知的领域，击败强大的敌人。', ARRAY['冒险', '动作'], ARRAY['像素', '单人', '冒险'], 'Unity', 'active', (SELECT id FROM users LIMIT 1), NOW(), NOW()),
  (gen_random_uuid(), 'space-shooter', '星际射击', '激烈的太空射击游戏', '驾驶你的战机在浩瀚的宇宙中战斗，消灭外星入侵者，保卫地球。', ARRAY['射击', '动作'], ARRAY['太空', '射击', '单人'], 'Unreal', 'active', (SELECT id FROM users LIMIT 1), NOW(), NOW()),
  (gen_random_uuid(), 'puzzle-master', '解谜大师', '烧脑的解谜游戏', '挑战你的智力极限，解开一个又一个精心设计的谜题。', ARRAY['解谜', '休闲'], ARRAY['解谜', '益智', '单人'], 'Custom', 'active', (SELECT id FROM users LIMIT 1), NOW(), NOW()),
  (gen_random_uuid(), 'rpg-legend', 'RPG传说', '史诗级角色扮演游戏', '在这个庞大的幻想世界中，创造你的英雄，书写属于你的传奇故事。', ARRAY['RPG', '冒险'], ARRAY['RPG', '多人', '开放世界'], 'Unity', 'active', (SELECT id FROM users LIMIT 1), NOW(), NOW()),
  (gen_random_uuid(), 'racing-fury', '狂飙赛车', '极速竞速游戏', '体验极致的速度与激情，在各种赛道上展现你的驾驶技术。', ARRAY['竞速', '体育'], ARRAY['赛车', '竞速', '多人'], 'Unreal', 'active', (SELECT id FROM users LIMIT 1), NOW(), NOW());

-- 插入测试音乐专辑数据
INSERT INTO albums (id, slug, title, subtitle, artist, composer, album_type, created_at, updated_at)
VALUES
  (gen_random_uuid(), 'pixel-adventure-ost', '像素冒险 原声音乐', '游戏原声带', 'Studio Music Team', 'John Composer', 'ost', NOW(), NOW()),
  (gen_random_uuid(), 'space-shooter-ost', '星际射击 原声音乐', '游戏原声带', 'Studio Music Team', 'Jane Composer', 'ost', NOW(), NOW()),
  (gen_random_uuid(), 'rpg-legend-ost', 'RPG传说 原声音乐', '游戏原声带', 'Studio Orchestra', 'Mike Composer', 'ost', NOW(), NOW());

-- 为游戏创建商品
INSERT INTO products (id, sku, product_type, entity_id, name, description, price_cents, is_active, created_at, updated_at)
SELECT
  gen_random_uuid(),
  'GAME-' || UPPER(SUBSTRING(MD5(RANDOM()::TEXT) FROM 1 FOR 8)),
  'game',
  id,
  title,
  description,
  CASE
    WHEN title LIKE '%像素%' THEN 3900
    WHEN title LIKE '%星际%' THEN 4900
    WHEN title LIKE '%解谜%' THEN 2900
    WHEN title LIKE '%RPG%' THEN 5900
    WHEN title LIKE '%赛车%' THEN 4500
    ELSE 3900
  END,
  true,
  NOW(),
  NOW()
FROM games
WHERE slug IN ('pixel-adventure', 'space-shooter', 'puzzle-master', 'rpg-legend', 'racing-fury');

-- 为音乐专辑创建商品
INSERT INTO products (id, sku, product_type, entity_id, name, description, price_cents, is_active, created_at, updated_at)
SELECT
  gen_random_uuid(),
  'OST-' || UPPER(SUBSTRING(MD5(RANDOM()::TEXT) FROM 1 FOR 8)),
  'ost',
  id,
  title,
  COALESCE(subtitle, '游戏原声音乐'),
  2900,
  true,
  NOW(),
  NOW()
FROM albums
WHERE slug IN ('pixel-adventure-ost', 'space-shooter-ost', 'rpg-legend-ost')
ON CONFLICT (sku) DO NOTHING;

-- 查看插入的数据
SELECT 'Games:' as type, COUNT(*) as count FROM games WHERE slug IN ('pixel-adventure', 'space-shooter', 'puzzle-master', 'rpg-legend', 'racing-fury')
UNION ALL
SELECT 'Albums:', COUNT(*) FROM albums WHERE slug IN ('pixel-adventure-ost', 'space-shooter-ost', 'rpg-legend-ost')
UNION ALL
SELECT 'Products:', COUNT(*) FROM products WHERE product_type IN ('game', 'album');
