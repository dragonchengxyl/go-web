package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/group"
)

type GroupRepository struct {
	pool *pgxpool.Pool
}

func NewGroupRepository(pool *pgxpool.Pool) *GroupRepository {
	return &GroupRepository{pool: pool}
}

func (r *GroupRepository) Create(ctx context.Context, g *group.Group) error {
	tags, _ := json.Marshal(g.Tags)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO groups (id, owner_id, name, description, announcement, rules, avatar_key, featured_post_id, tags, privacy, member_count, post_count, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`, g.ID, g.OwnerID, g.Name, g.Description, g.Announcement, g.Rules, g.AvatarKey, g.FeaturedPostID, tags, g.Privacy, g.MemberCount, g.PostCount, g.CreatedAt, g.UpdatedAt)
	return err
}

func (r *GroupRepository) GetByID(ctx context.Context, id uuid.UUID) (*group.Group, error) {
	var g group.Group
	var tags []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, owner_id, name, description, announcement, rules, avatar_key, featured_post_id, tags, privacy, member_count, post_count, created_at, updated_at
		FROM groups WHERE id=$1
	`, id).Scan(
		&g.ID, &g.OwnerID, &g.Name, &g.Description, &g.Announcement, &g.Rules, &g.AvatarKey, &g.FeaturedPostID, &tags,
		&g.Privacy, &g.MemberCount, &g.PostCount, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, group.ErrNotFound
		}
		return nil, err
	}
	_ = json.Unmarshal(tags, &g.Tags)
	return &g, nil
}

func (r *GroupRepository) Update(ctx context.Context, g *group.Group) error {
	tags, _ := json.Marshal(g.Tags)
	_, err := r.pool.Exec(ctx, `
		UPDATE groups SET name=$1, description=$2, announcement=$3, rules=$4, avatar_key=$5, featured_post_id=$6, tags=$7, privacy=$8, updated_at=$9
		WHERE id=$10
	`, g.Name, g.Description, g.Announcement, g.Rules, g.AvatarKey, g.FeaturedPostID, tags, g.Privacy, g.UpdatedAt, g.ID)
	return err
}

func (r *GroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM groups WHERE id=$1`, id)
	return err
}

func (r *GroupRepository) List(ctx context.Context, filter group.ListFilter) ([]*group.Group, int64, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	rows, err := r.pool.Query(ctx, `
		SELECT g.id, g.owner_id, g.name, g.description, g.announcement, g.rules, g.avatar_key, g.featured_post_id, g.tags, g.privacy, g.member_count, g.post_count, g.created_at, g.updated_at
		FROM groups g
		WHERE ($1::text IS NULL OR g.privacy = $1)
		  AND ($2::text IS NULL OR g.name ILIKE '%' || $2 || '%')
		ORDER BY g.member_count DESC
		LIMIT $3 OFFSET $4
	`, filter.Privacy, filter.Search, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var groups []*group.Group
	for rows.Next() {
		var g group.Group
		var tags []byte
		if err := rows.Scan(
			&g.ID, &g.OwnerID, &g.Name, &g.Description, &g.Announcement, &g.Rules, &g.AvatarKey, &g.FeaturedPostID, &tags,
			&g.Privacy, &g.MemberCount, &g.PostCount, &g.CreatedAt, &g.UpdatedAt,
		); err != nil {
			continue
		}
		_ = json.Unmarshal(tags, &g.Tags)
		groups = append(groups, &g)
	}

	var total int64
	_ = r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM groups
		WHERE ($1::text IS NULL OR privacy = $1)
		  AND ($2::text IS NULL OR name ILIKE '%' || $2 || '%')
	`, filter.Privacy, filter.Search).Scan(&total)

	return groups, total, nil
}

func (r *GroupRepository) AddMember(ctx context.Context, m *group.GroupMember) error {
	result, err := r.pool.Exec(ctx, `
		INSERT INTO group_members (group_id, user_id, role, joined_at)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (group_id, user_id) DO NOTHING
	`, m.GroupID, m.UserID, m.Role, m.JoinedAt)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return group.ErrAlreadyMember
	}
	return nil
}

func (r *GroupRepository) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM group_members WHERE group_id=$1 AND user_id=$2`, groupID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return group.ErrNotMember
	}
	return nil
}

func (r *GroupRepository) GetMember(ctx context.Context, groupID, userID uuid.UUID) (*group.GroupMember, error) {
	var m group.GroupMember
	err := r.pool.QueryRow(ctx, `
		SELECT gm.group_id, gm.user_id, gm.role, gm.joined_at,
		       u.username, u.furry_name, u.avatar_key
		FROM group_members gm
		JOIN users u ON u.id = gm.user_id
		WHERE gm.group_id=$1 AND gm.user_id=$2
	`, groupID, userID).Scan(&m.GroupID, &m.UserID, &m.Role, &m.JoinedAt, &m.Username, &m.FurryName, &m.AvatarKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *GroupRepository) UpdateMemberRole(ctx context.Context, groupID, userID uuid.UUID, role group.GroupRole) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE group_members SET role=$1 WHERE group_id=$2 AND user_id=$3
	`, role, groupID, userID)
	return err
}

func (r *GroupRepository) ListMembers(ctx context.Context, groupID uuid.UUID, page, pageSize int) ([]*group.GroupMember, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	rows, err := r.pool.Query(ctx, `
		SELECT gm.group_id, gm.user_id, gm.role, gm.joined_at,
		       u.username, u.furry_name, u.avatar_key
		FROM group_members gm
		JOIN users u ON u.id = gm.user_id
		WHERE gm.group_id=$1 ORDER BY gm.joined_at ASC LIMIT $2 OFFSET $3
	`, groupID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var members []*group.GroupMember
	for rows.Next() {
		var m group.GroupMember
		if err := rows.Scan(&m.GroupID, &m.UserID, &m.Role, &m.JoinedAt, &m.Username, &m.FurryName, &m.AvatarKey); err != nil {
			continue
		}
		members = append(members, &m)
	}

	var total int64
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM group_members WHERE group_id=$1`, groupID).Scan(&total)
	return members, total, nil
}

func (r *GroupRepository) CreateAnnouncement(ctx context.Context, item *group.GroupAnnouncement) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO group_announcements (id, group_id, author_id, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, item.ID, item.GroupID, item.AuthorID, item.Content, item.CreatedAt)
	return err
}

func (r *GroupRepository) ListAnnouncements(ctx context.Context, groupID uuid.UUID, page, pageSize int) ([]*group.GroupAnnouncement, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM group_announcements WHERE group_id=$1`, groupID).Scan(&total)

	rows, err := r.pool.Query(ctx, `
		SELECT ga.id, ga.group_id, ga.author_id, ga.content, ga.created_at,
		       u.username, u.furry_name, u.avatar_key
		FROM group_announcements ga
		JOIN users u ON u.id = ga.author_id
		WHERE ga.group_id = $1
		ORDER BY ga.created_at DESC
		LIMIT $2 OFFSET $3
	`, groupID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]*group.GroupAnnouncement, 0, pageSize)
	for rows.Next() {
		var item group.GroupAnnouncement
		if err := rows.Scan(&item.ID, &item.GroupID, &item.AuthorID, &item.Content, &item.CreatedAt, &item.AuthorName, &item.FurryName, &item.AvatarKey); err != nil {
			return nil, 0, err
		}
		items = append(items, &item)
	}
	return items, total, rows.Err()
}

func (r *GroupRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]*group.Group, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	rows, err := r.pool.Query(ctx, `
		SELECT id, owner_id, name, description, announcement, rules, avatar_key, featured_post_id, tags, privacy, member_count, post_count, created_at, updated_at
		FROM groups
		WHERE owner_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`, ownerID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]*group.Group, 0, pageSize)
	for rows.Next() {
		var g group.Group
		var tags []byte
		if err := rows.Scan(&g.ID, &g.OwnerID, &g.Name, &g.Description, &g.Announcement, &g.Rules, &g.AvatarKey, &g.FeaturedPostID, &tags, &g.Privacy, &g.MemberCount, &g.PostCount, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, 0, err
		}
		_ = json.Unmarshal(tags, &g.Tags)
		items = append(items, &g)
	}

	var total int64
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM groups WHERE owner_id = $1`, ownerID).Scan(&total)
	return items, total, rows.Err()
}

func (r *GroupRepository) ListByRole(ctx context.Context, userID uuid.UUID, role group.GroupRole, page, pageSize int) ([]*group.Group, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	rows, err := r.pool.Query(ctx, `
		SELECT g.id, g.owner_id, g.name, g.description, g.announcement, g.rules, g.avatar_key, g.featured_post_id, g.tags, g.privacy, g.member_count, g.post_count, g.created_at, g.updated_at
		FROM groups g
		JOIN group_members gm ON gm.group_id = g.id
		WHERE gm.user_id = $1 AND gm.role = $2
		ORDER BY g.updated_at DESC
		LIMIT $3 OFFSET $4
	`, userID, role, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]*group.Group, 0, pageSize)
	for rows.Next() {
		var g group.Group
		var tags []byte
		if err := rows.Scan(&g.ID, &g.OwnerID, &g.Name, &g.Description, &g.Announcement, &g.Rules, &g.AvatarKey, &g.FeaturedPostID, &tags, &g.Privacy, &g.MemberCount, &g.PostCount, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, 0, err
		}
		_ = json.Unmarshal(tags, &g.Tags)
		items = append(items, &g)
	}

	var total int64
	_ = r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM group_members
		WHERE user_id = $1 AND role = $2
	`, userID, role).Scan(&total)
	return items, total, rows.Err()
}

func (r *GroupRepository) ListRecentActiveMembers(ctx context.Context, groupID uuid.UUID, limit int) ([]*group.GroupMember, error) {
	if limit <= 0 {
		limit = 5
	}

	rows, err := r.pool.Query(ctx, `
		WITH recent_posts AS (
			SELECT author_id, MAX(created_at) AS last_active_at
			FROM posts
			WHERE group_id = $1 AND deleted_at IS NULL
			GROUP BY author_id
		)
		SELECT gm.group_id, gm.user_id, gm.role, gm.joined_at,
		       u.username, u.furry_name, u.avatar_key
		FROM group_members gm
		JOIN users u ON u.id = gm.user_id
		LEFT JOIN recent_posts rp ON rp.author_id = gm.user_id
		WHERE gm.group_id = $1
		ORDER BY COALESCE(rp.last_active_at, gm.joined_at) DESC, gm.joined_at DESC
		LIMIT $2
	`, groupID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*group.GroupMember, 0, limit)
	for rows.Next() {
		var m group.GroupMember
		if err := rows.Scan(&m.GroupID, &m.UserID, &m.Role, &m.JoinedAt, &m.Username, &m.FurryName, &m.AvatarKey); err != nil {
			return nil, err
		}
		items = append(items, &m)
	}
	return items, rows.Err()
}

func (r *GroupRepository) IncrementMemberCount(ctx context.Context, groupID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE groups SET member_count=member_count+1 WHERE id=$1`, groupID)
	return err
}

func (r *GroupRepository) DecrementMemberCount(ctx context.Context, groupID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE groups SET member_count=GREATEST(0,member_count-1) WHERE id=$1`, groupID)
	return err
}

func (r *GroupRepository) IncrementPostCount(ctx context.Context, groupID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE groups SET post_count=post_count+1 WHERE id=$1`, groupID)
	return err
}

func (r *GroupRepository) DecrementPostCount(ctx context.Context, groupID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE groups SET post_count=GREATEST(0,post_count-1) WHERE id=$1`, groupID)
	return err
}

func (r *GroupRepository) ListByMember(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*group.Group, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	rows, err := r.pool.Query(ctx, `
		SELECT g.id, g.owner_id, g.name, g.description, g.announcement, g.rules, g.avatar_key, g.featured_post_id, g.tags, g.privacy, g.member_count, g.post_count, g.created_at, g.updated_at
		FROM groups g
		JOIN group_members gm ON gm.group_id = g.id
		WHERE gm.user_id=$1
		ORDER BY gm.joined_at DESC
		LIMIT $2 OFFSET $3
	`, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var groups []*group.Group
	for rows.Next() {
		var g group.Group
		var tags []byte
		if err := rows.Scan(
			&g.ID, &g.OwnerID, &g.Name, &g.Description, &g.Announcement, &g.Rules, &g.AvatarKey, &g.FeaturedPostID, &tags,
			&g.Privacy, &g.MemberCount, &g.PostCount, &g.CreatedAt, &g.UpdatedAt,
		); err != nil {
			continue
		}
		_ = json.Unmarshal(tags, &g.Tags)
		groups = append(groups, &g)
	}

	var total int64
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM group_members WHERE user_id=$1`, userID).Scan(&total)
	return groups, total, nil
}

// Ensure GroupRepository satisfies the ListByMember helper used by group_service.
// We expose it as a method since the Repository interface keeps it minimal.
var _ groupListByMember = (*GroupRepository)(nil)

type groupListByMember interface {
	ListByMember(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*group.Group, int64, error)
}

// Ensure compile-time check.
var _ group.Repository = (*GroupRepository)(nil)
