DROP TRIGGER IF EXISTS audio_jobs_outbox_created ON audio_jobs;
DROP TRIGGER IF EXISTS orders_outbox_tip_paid ON orders;
DROP TRIGGER IF EXISTS comments_outbox_created ON comments;
DROP TRIGGER IF EXISTS user_follows_outbox_created ON user_follows;
DROP TRIGGER IF EXISTS post_likes_outbox_created ON post_likes;
DROP TRIGGER IF EXISTS posts_outbox_moderated ON posts;
DROP TRIGGER IF EXISTS posts_outbox_created ON posts;

DROP FUNCTION IF EXISTS trg_audio_jobs_outbox_created();
DROP FUNCTION IF EXISTS trg_orders_outbox_tip_paid();
DROP FUNCTION IF EXISTS trg_comments_outbox_created();
DROP FUNCTION IF EXISTS trg_user_follows_outbox_created();
DROP FUNCTION IF EXISTS trg_post_likes_outbox_created();
DROP FUNCTION IF EXISTS trg_posts_outbox_moderated();
DROP FUNCTION IF EXISTS trg_posts_outbox_created();
DROP FUNCTION IF EXISTS enqueue_outbox_event(TEXT, TEXT, UUID, TEXT, TEXT, TEXT, JSONB);

DROP INDEX IF EXISTS idx_outbox_events_event_type;
DROP INDEX IF EXISTS idx_outbox_events_topic_created_at;
DROP INDEX IF EXISTS idx_outbox_events_status_retry;

DROP TABLE IF EXISTS outbox_events;
