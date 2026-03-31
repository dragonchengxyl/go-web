CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_table TEXT NOT NULL,
    aggregate_type TEXT NOT NULL,
    aggregate_id UUID,
    event_type TEXT NOT NULL,
    topic TEXT NOT NULL,
    partition_key TEXT NOT NULL DEFAULT '',
    payload JSONB NOT NULL,
    headers JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'published', 'dead_lettered')),
    attempt_count INT NOT NULL DEFAULT 0,
    last_error TEXT,
    next_retry_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_outbox_events_status_retry
    ON outbox_events (status, next_retry_at, created_at);

CREATE INDEX IF NOT EXISTS idx_outbox_events_topic_created_at
    ON outbox_events (topic, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_outbox_events_event_type
    ON outbox_events (event_type, created_at DESC);

CREATE OR REPLACE FUNCTION enqueue_outbox_event(
    p_source_table TEXT,
    p_aggregate_type TEXT,
    p_aggregate_id UUID,
    p_event_type TEXT,
    p_topic TEXT,
    p_partition_key TEXT,
    p_payload JSONB
) RETURNS VOID AS $$
BEGIN
    INSERT INTO outbox_events (
        source_table,
        aggregate_type,
        aggregate_id,
        event_type,
        topic,
        partition_key,
        payload,
        headers
    )
    VALUES (
        p_source_table,
        p_aggregate_type,
        p_aggregate_id,
        p_event_type,
        p_topic,
        COALESCE(p_partition_key, ''),
        p_payload,
        jsonb_build_object('event_version', 1, 'occurred_at', NOW())
    );
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION trg_posts_outbox_created() RETURNS TRIGGER AS $$
BEGIN
    PERFORM enqueue_outbox_event(
        'posts',
        'post',
        NEW.id,
        'post.created',
        'content',
        NEW.id::text,
        jsonb_build_object(
            'post_id', NEW.id::text,
            'author_id', NEW.author_id::text,
            'content', NEW.content,
            'media_urls', COALESCE(NEW.media_urls, '[]'::jsonb)
        )
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION trg_posts_outbox_moderated() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.moderation_status IS DISTINCT FROM OLD.moderation_status THEN
        PERFORM enqueue_outbox_event(
            'posts',
            'post',
            NEW.id,
            'post.moderated',
            'content',
            NEW.id::text,
            jsonb_build_object(
                'post_id', NEW.id::text,
                'author_id', NEW.author_id::text,
                'status', NEW.moderation_status
            )
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION trg_post_likes_outbox_created() RETURNS TRIGGER AS $$
DECLARE
    v_author_id UUID;
BEGIN
    SELECT author_id INTO v_author_id FROM posts WHERE id = NEW.post_id;
    IF v_author_id IS NULL THEN
        RETURN NEW;
    END IF;

    PERFORM enqueue_outbox_event(
        'post_likes',
        'post',
        NEW.post_id,
        'post.liked',
        'social',
        NEW.post_id::text,
        jsonb_build_object(
            'post_id', NEW.post_id::text,
            'actor_id', NEW.user_id::text,
            'author_id', v_author_id::text
        )
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION trg_user_follows_outbox_created() RETURNS TRIGGER AS $$
BEGIN
    PERFORM enqueue_outbox_event(
        'user_follows',
        'follow',
        NEW.followee_id,
        'user.followed',
        'social',
        NEW.followee_id::text,
        jsonb_build_object(
            'follower_id', NEW.follower_id::text,
            'followee_id', NEW.followee_id::text
        )
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION trg_comments_outbox_created() RETURNS TRIGGER AS $$
DECLARE
    v_target_user_id UUID;
BEGIN
    IF NEW.commentable_type <> 'post' THEN
        RETURN NEW;
    END IF;

    IF NEW.parent_id IS NOT NULL THEN
        SELECT user_id INTO v_target_user_id FROM comments WHERE id = NEW.parent_id;
    ELSE
        SELECT author_id INTO v_target_user_id FROM posts WHERE id = NEW.commentable_id;
    END IF;

    IF v_target_user_id IS NULL THEN
        RETURN NEW;
    END IF;

    PERFORM enqueue_outbox_event(
        'comments',
        'comment',
        NEW.id,
        'comment.created',
        'social',
        NEW.commentable_id::text,
        jsonb_build_object(
            'comment_id', NEW.id::text,
            'post_id', NEW.commentable_id::text,
            'commentable_id', NEW.commentable_id::text,
            'author_id', NEW.user_id::text,
            'target_user_id', v_target_user_id::text
        )
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION trg_orders_outbox_tip_paid() RETURNS TRIGGER AS $$
DECLARE
    v_type TEXT;
    v_receiver_id TEXT;
BEGIN
    v_type := COALESCE(NEW.metadata->>'type', '');
    v_receiver_id := COALESCE(NEW.metadata->>'to_user_id', '');

    IF v_type = 'tip'
       AND NEW.status IN ('paid', 'fulfilled')
       AND (OLD.status IS DISTINCT FROM NEW.status)
       AND v_receiver_id <> '' THEN
        PERFORM enqueue_outbox_event(
            'orders',
            'order',
            NEW.id,
            'tip.sent',
            'social',
            NEW.id::text,
            jsonb_build_object(
                'tip_id', NEW.id::text,
                'sender_id', NEW.user_id::text,
                'receiver_id', v_receiver_id,
                'amount_cents', NEW.total_cents
            )
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION trg_audio_jobs_outbox_created() RETURNS TRIGGER AS $$
BEGIN
    PERFORM enqueue_outbox_event(
        'audio_jobs',
        'audio_job',
        NEW.id,
        'audio.job.created',
        'audio',
        NEW.id::text,
        jsonb_build_object(
            'job_id', NEW.id::text
        )
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS posts_outbox_created ON posts;
CREATE TRIGGER posts_outbox_created
AFTER INSERT ON posts
FOR EACH ROW
EXECUTE FUNCTION trg_posts_outbox_created();

DROP TRIGGER IF EXISTS posts_outbox_moderated ON posts;
CREATE TRIGGER posts_outbox_moderated
AFTER UPDATE OF moderation_status ON posts
FOR EACH ROW
EXECUTE FUNCTION trg_posts_outbox_moderated();

DROP TRIGGER IF EXISTS post_likes_outbox_created ON post_likes;
CREATE TRIGGER post_likes_outbox_created
AFTER INSERT ON post_likes
FOR EACH ROW
EXECUTE FUNCTION trg_post_likes_outbox_created();

DROP TRIGGER IF EXISTS user_follows_outbox_created ON user_follows;
CREATE TRIGGER user_follows_outbox_created
AFTER INSERT ON user_follows
FOR EACH ROW
EXECUTE FUNCTION trg_user_follows_outbox_created();

DROP TRIGGER IF EXISTS comments_outbox_created ON comments;
CREATE TRIGGER comments_outbox_created
AFTER INSERT ON comments
FOR EACH ROW
EXECUTE FUNCTION trg_comments_outbox_created();

DROP TRIGGER IF EXISTS orders_outbox_tip_paid ON orders;
CREATE TRIGGER orders_outbox_tip_paid
AFTER UPDATE OF status ON orders
FOR EACH ROW
EXECUTE FUNCTION trg_orders_outbox_tip_paid();

DROP TRIGGER IF EXISTS audio_jobs_outbox_created ON audio_jobs;
CREATE TRIGGER audio_jobs_outbox_created
AFTER INSERT ON audio_jobs
FOR EACH ROW
EXECUTE FUNCTION trg_audio_jobs_outbox_created();
