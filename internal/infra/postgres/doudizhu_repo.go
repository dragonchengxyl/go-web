package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/doudizhu"
	"github.com/studio/platform/internal/pkg/apperr"
)

type DoudizhuRepository struct {
	pool *pgxpool.Pool
}

func NewDoudizhuRepository(pool *pgxpool.Pool) *DoudizhuRepository {
	return &DoudizhuRepository{pool: pool}
}

func (r *DoudizhuRepository) SaveMatch(ctx context.Context, match *doudizhu.Match, results []doudizhu.MatchPlayerResult, events []doudizhu.ActionEvent) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin doudizhu match tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	_, err = tx.Exec(ctx, `
		INSERT INTO doudizhu_matches (
			id, room_id, room_code, room_title, match_mode, started_at, finished_at,
			landlord_seat, winner_side, multiplier, bomb_count, spring, anti_spring, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (id) DO NOTHING
	`, match.ID, match.RoomID, match.RoomCode, match.RoomTitle, match.MatchMode, match.StartedAt, match.FinishedAt, match.LandlordSeat, match.WinnerSide, match.Multiplier, match.BombCount, match.Spring, match.AntiSpring, match.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert doudizhu match: %w", err)
	}

	for _, result := range results {
		_, err = tx.Exec(ctx, `
			INSERT INTO doudizhu_match_players (
				id, match_id, session_id, user_id, is_bot, bot_level, seat, player_name, role,
				bid_score, cards_left, is_winner, score_delta, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			ON CONFLICT (id) DO NOTHING
		`, result.ID, result.MatchID, result.SessionID, result.UserID, result.IsBot, nullIfEmpty(result.BotLevel), result.Seat, result.PlayerName, result.Role, result.BidScore, result.CardsLeft, result.IsWinner, result.ScoreDelta, result.CreatedAt)
		if err != nil {
			return fmt.Errorf("insert doudizhu match player: %w", err)
		}
	}

	for _, event := range events {
		var cardsJSON []byte
		if len(event.Cards) > 0 {
			cardsJSON, err = json.Marshal(event.Cards)
			if err != nil {
				return fmt.Errorf("marshal doudizhu event cards: %w", err)
			}
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO doudizhu_action_events (
				id, match_id, turn_no, action_index, session_id, user_id, player_name, seat,
				action_type, cards_json, combo_type, combo_main_rank, combo_sequence_length,
				combo_total_cards, multiplier_after, occurred_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
			ON CONFLICT (id) DO NOTHING
		`, event.ID, event.MatchID, event.TurnNo, event.ActionIndex, event.SessionID, event.UserID, event.PlayerName, event.Seat, event.ActionType, nullableBytes(cardsJSON), comboTypeValue(event.Combo), comboRankValue(event.Combo), comboSequenceValue(event.Combo), comboTotalCardsValue(event.Combo), event.MultiplierAfter, event.OccurredAt)
		if err != nil {
			return fmt.Errorf("insert doudizhu action event: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit doudizhu match tx: %w", err)
	}
	return nil
}

func (r *DoudizhuRepository) ListLeaderboard(ctx context.Context, limit int) ([]*doudizhu.LeaderboardEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := r.pool.Query(ctx, `
		WITH player_stats AS (
			SELECT
				CASE
					WHEN p.user_id IS NOT NULL THEN p.user_id::text
					ELSE 'guest:' || LOWER(p.player_name)
				END AS player_key,
				p.user_id,
				p.player_name,
				COALESCE(u.username, p.player_name) AS display_name,
				COUNT(*) AS matches,
				SUM(CASE WHEN p.is_winner THEN 1 ELSE 0 END) AS wins,
				SUM(p.score_delta) AS total_score,
				MAX(m.finished_at) AS last_played
			FROM doudizhu_match_players p
			JOIN doudizhu_matches m ON m.id = p.match_id
			LEFT JOIN users u ON u.id = p.user_id
			WHERE p.is_bot = FALSE
			  AND m.match_mode = 'pvp'
			GROUP BY player_key, p.user_id, p.player_name, COALESCE(u.username, p.player_name)
		)
		SELECT
			ROW_NUMBER() OVER (ORDER BY total_score DESC, wins DESC, last_played DESC) AS leaderboard_rank,
			user_id,
			player_name,
			display_name,
			matches,
			wins,
			total_score,
			last_played
		FROM player_stats
		ORDER BY total_score DESC, wins DESC, last_played DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list doudizhu leaderboard: %w", err)
	}
	defer rows.Close()

	entries := make([]*doudizhu.LeaderboardEntry, 0, limit)
	for rows.Next() {
		item := &doudizhu.LeaderboardEntry{}
		if err := rows.Scan(
			&item.Rank,
			&item.UserID,
			&item.PlayerName,
			&item.DisplayName,
			&item.Matches,
			&item.Wins,
			&item.TotalScore,
			&item.LastPlayed,
		); err != nil {
			return nil, fmt.Errorf("scan doudizhu leaderboard: %w", err)
		}
		entries = append(entries, item)
	}
	return entries, nil
}

func (r *DoudizhuRepository) ListRecentMatches(ctx context.Context, limit int) ([]*doudizhu.MatchSummary, error) {
	return r.listRecentMatches(ctx, nil, limit)
}

func (r *DoudizhuRepository) ListUserRecentMatches(ctx context.Context, userID uuid.UUID, limit int) ([]*doudizhu.MatchSummary, error) {
	return r.listRecentMatches(ctx, &userID, limit)
}

func (r *DoudizhuRepository) listRecentMatches(ctx context.Context, userID *uuid.UUID, limit int) ([]*doudizhu.MatchSummary, error) {
	if limit <= 0 || limit > 50 {
		limit = 8
	}

	query := `
		SELECT DISTINCT m.id, m.room_id, m.room_code, m.room_title, m.match_mode, m.finished_at, m.winner_side, m.landlord_seat, m.multiplier
		FROM doudizhu_matches m
	`
	args := []any{}
	if userID != nil {
		query += `
			JOIN doudizhu_match_players p ON p.match_id = m.id
			WHERE p.user_id = $1
		`
		args = append(args, *userID)
	}
	query += `
		ORDER BY m.finished_at DESC
		LIMIT $` + fmt.Sprint(len(args)+1)
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list doudizhu recent matches: %w", err)
	}
	defer rows.Close()

	matchIDs := make([]uuid.UUID, 0, limit)
	items := make([]*doudizhu.MatchSummary, 0, limit)
	indexByMatchID := make(map[uuid.UUID]*doudizhu.MatchSummary, limit)
	for rows.Next() {
		item := &doudizhu.MatchSummary{}
		if err := rows.Scan(
			&item.MatchID,
			&item.RoomID,
			&item.RoomCode,
			&item.RoomTitle,
			&item.MatchMode,
			&item.FinishedAt,
			&item.WinnerSide,
			&item.LandlordSeat,
			&item.Multiplier,
		); err != nil {
			return nil, fmt.Errorf("scan doudizhu recent match: %w", err)
		}
		matchIDs = append(matchIDs, item.MatchID)
		items = append(items, item)
		indexByMatchID[item.MatchID] = item
	}
	if len(matchIDs) == 0 {
		return items, nil
	}

	resultsRows, err := r.pool.Query(ctx, `
		SELECT
			p.id,
			p.match_id,
			p.session_id,
			p.user_id,
			p.is_bot,
			COALESCE(p.bot_level, '') AS bot_level,
			p.seat,
			p.player_name,
			COALESCE(u.username, p.player_name) AS display_name,
			p.role,
			p.bid_score,
			p.cards_left,
			p.is_winner,
			p.score_delta,
			p.created_at
		FROM doudizhu_match_players p
		LEFT JOIN users u ON u.id = p.user_id
		WHERE p.match_id = ANY($1)
		ORDER BY p.match_id ASC, p.score_delta DESC, p.cards_left ASC, p.created_at ASC
	`, matchIDs)
	if err != nil {
		return nil, fmt.Errorf("list doudizhu recent match results: %w", err)
	}
	defer resultsRows.Close()

	countByMatch := make(map[uuid.UUID]int, len(matchIDs))
	for resultsRows.Next() {
		item, err := scanDoudizhuMatchPlayerResult(resultsRows)
		if err != nil {
			return nil, err
		}
		target := indexByMatchID[item.MatchID]
		if target == nil {
			continue
		}
		target.PlayerCount++
		if countByMatch[item.MatchID] < 3 {
			target.TopResults = append(target.TopResults, *item)
			countByMatch[item.MatchID]++
		}
	}

	return items, nil
}

func (r *DoudizhuRepository) GetReplay(ctx context.Context, matchID uuid.UUID) (*doudizhu.MatchReplay, error) {
	match := &doudizhu.Match{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, room_id, room_code, room_title, match_mode, started_at, finished_at,
		       landlord_seat, winner_side, multiplier, bomb_count, spring, anti_spring, created_at
		FROM doudizhu_matches
		WHERE id = $1
	`, matchID).Scan(
		&match.ID,
		&match.RoomID,
		&match.RoomCode,
		&match.RoomTitle,
		&match.MatchMode,
		&match.StartedAt,
		&match.FinishedAt,
		&match.LandlordSeat,
		&match.WinnerSide,
		&match.Multiplier,
		&match.BombCount,
		&match.Spring,
		&match.AntiSpring,
		&match.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrNotFound
		}
		return nil, fmt.Errorf("get doudizhu replay match: %w", err)
	}

	resultsRows, err := r.pool.Query(ctx, `
		SELECT
			p.id,
			p.match_id,
			p.session_id,
			p.user_id,
			p.is_bot,
			COALESCE(p.bot_level, '') AS bot_level,
			p.seat,
			p.player_name,
			COALESCE(u.username, p.player_name) AS display_name,
			p.role,
			p.bid_score,
			p.cards_left,
			p.is_winner,
			p.score_delta,
			p.created_at
		FROM doudizhu_match_players p
		LEFT JOIN users u ON u.id = p.user_id
		WHERE p.match_id = $1
		ORDER BY p.score_delta DESC, p.cards_left ASC, p.created_at ASC
	`, matchID)
	if err != nil {
		return nil, fmt.Errorf("get doudizhu replay results: %w", err)
	}
	defer resultsRows.Close()

	results := make([]doudizhu.MatchPlayerResult, 0)
	for resultsRows.Next() {
		item, err := scanDoudizhuMatchPlayerResult(resultsRows)
		if err != nil {
			return nil, err
		}
		results = append(results, *item)
	}

	eventsRows, err := r.pool.Query(ctx, `
		SELECT
			e.id,
			e.match_id,
			e.turn_no,
			e.action_index,
			e.session_id,
			e.user_id,
			e.player_name,
			COALESCE(u.username, e.player_name) AS display_name,
			e.seat,
			e.action_type,
			e.cards_json,
			e.combo_type,
			e.combo_main_rank,
			e.combo_sequence_length,
			e.combo_total_cards,
			e.multiplier_after,
			e.occurred_at
		FROM doudizhu_action_events e
		LEFT JOIN users u ON u.id = e.user_id
		WHERE e.match_id = $1
		ORDER BY e.action_index ASC, e.occurred_at ASC
	`, matchID)
	if err != nil {
		return nil, fmt.Errorf("get doudizhu replay events: %w", err)
	}
	defer eventsRows.Close()

	events := make([]doudizhu.ActionEvent, 0)
	for eventsRows.Next() {
		item, err := scanDoudizhuActionEvent(eventsRows)
		if err != nil {
			return nil, err
		}
		events = append(events, *item)
	}

	match.Results = append(match.Results, results...)
	return &doudizhu.MatchReplay{
		Match:   match,
		Results: results,
		Events:  events,
	}, nil
}

func scanDoudizhuMatchPlayerResult(row pgx.Row) (*doudizhu.MatchPlayerResult, error) {
	item := &doudizhu.MatchPlayerResult{}
	var botLevel string
	if err := row.Scan(
		&item.ID,
		&item.MatchID,
		&item.SessionID,
		&item.UserID,
		&item.IsBot,
		&botLevel,
		&item.Seat,
		&item.PlayerName,
		&item.DisplayName,
		&item.Role,
		&item.BidScore,
		&item.CardsLeft,
		&item.IsWinner,
		&item.ScoreDelta,
		&item.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("scan doudizhu match player result: %w", err)
	}
	item.BotLevel = botLevel
	return item, nil
}

func scanDoudizhuActionEvent(row pgx.Row) (*doudizhu.ActionEvent, error) {
	item := &doudizhu.ActionEvent{}
	var cardsJSON []byte
	var comboType *string
	var comboMainRank *int
	var comboSequenceLength *int
	var comboTotalCards *int
	if err := row.Scan(
		&item.ID,
		&item.MatchID,
		&item.TurnNo,
		&item.ActionIndex,
		&item.SessionID,
		&item.UserID,
		&item.PlayerName,
		&item.DisplayName,
		&item.Seat,
		&item.ActionType,
		&cardsJSON,
		&comboType,
		&comboMainRank,
		&comboSequenceLength,
		&comboTotalCards,
		&item.MultiplierAfter,
		&item.OccurredAt,
	); err != nil {
		return nil, fmt.Errorf("scan doudizhu action event: %w", err)
	}
	if len(cardsJSON) > 0 {
		if err := json.Unmarshal(cardsJSON, &item.Cards); err != nil {
			return nil, fmt.Errorf("unmarshal doudizhu action cards: %w", err)
		}
	}
	if comboType != nil && comboMainRank != nil && comboSequenceLength != nil && comboTotalCards != nil {
		item.Combo = &doudizhu.Combo{
			Type:           doudizhu.ComboType(*comboType),
			MainRank:       doudizhu.Rank(*comboMainRank),
			SequenceLength: *comboSequenceLength,
			TotalCards:     *comboTotalCards,
		}
	}
	return item, nil
}

func nullableBytes(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return value
}

func nullIfEmpty(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func comboTypeValue(combo *doudizhu.Combo) any {
	if combo == nil {
		return nil
	}
	return string(combo.Type)
}

func comboRankValue(combo *doudizhu.Combo) any {
	if combo == nil {
		return nil
	}
	return int(combo.MainRank)
}

func comboSequenceValue(combo *doudizhu.Combo) any {
	if combo == nil {
		return nil
	}
	return combo.SequenceLength
}

func comboTotalCardsValue(combo *doudizhu.Combo) any {
	if combo == nil {
		return nil
	}
	return combo.TotalCards
}

var _ doudizhu.Repository = (*DoudizhuRepository)(nil)
