-- Migration 024: Ensure comments support 'post' commentable_type
-- The comment entity already defines CommentableTypePost = "post"
-- This migration verifies the constraint allows 'post' values.
-- If there's a CHECK constraint, update it; otherwise this is a no-op.

-- Add 'post' to commentable_type if there's a constraint (check first)
DO $$
BEGIN
    -- Update any existing constraint to include 'post'
    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'comments_commentable_type_check'
    ) THEN
        ALTER TABLE comments DROP CONSTRAINT comments_commentable_type_check;
        ALTER TABLE comments ADD CONSTRAINT comments_commentable_type_check
            CHECK (commentable_type IN ('game', 'track', 'post', 'album'));
    END IF;
END $$;
