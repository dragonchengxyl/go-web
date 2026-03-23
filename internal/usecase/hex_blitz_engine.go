package usecase

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/gameplay"
	"github.com/studio/platform/internal/pkg/apperr"
)

const (
	hexBlitzBoardRadius = 2
	hexBlitzComboWindow = 2500 * time.Millisecond
)

var (
	hexBlitzDirections = [][2]int{
		{1, 0},
		{1, -1},
		{0, -1},
		{-1, 0},
		{-1, 1},
		{0, 1},
	}
	hexBlitzBoardCoords = buildHexBlitzBoardCoords()
)

type hexBlitzBoardCoord struct {
	Q int
	R int
}

type hexBlitzPlayerBoard struct {
	seed        int64
	rng         *rand.Rand
	tiles       []gameplay.HexBlitzTile
	score       int
	combo       int
	bestCombo   int
	moves       int
	lastGain    int
	lastCleared int
	lastMoveAt  *time.Time
	message     string
	updatedAt   time.Time
}

func newHexBlitzPlayerBoard(seed int64, now time.Time) *hexBlitzPlayerBoard {
	rng := rand.New(rand.NewSource(seed))
	board := &hexBlitzPlayerBoard{
		seed:      seed,
		rng:       rng,
		updatedAt: now,
		message:   "服务端棋盘已同步，点击相邻同色六角块开始冲分。",
	}
	board.tiles = hexBlitzCreateRandomBoard(rng)
	return board
}

func (b *hexBlitzPlayerBoard) applyMove(sessionID string, matchID uuid.UUID, tileID string, now time.Time) (*gameplay.HexBlitzMoveResult, error) {
	group := hexBlitzCollectGroup(b.tiles, tileID)
	if len(group) < 2 {
		return nil, apperr.BadRequest("当前选中的连线不足 2 格")
	}

	cleared := hexBlitzExpandBurst(group, b.tiles)
	clearedTiles := make([]gameplay.HexBlitzTile, 0, len(cleared))
	for _, tile := range b.tiles {
		if _, ok := cleared[tile.ID]; ok {
			clearedTiles = append(clearedTiles, tile)
		}
	}

	currentCombo := b.currentCombo(now)
	nextCombo := 1
	if currentCombo > 0 {
		nextCombo = currentCombo + 1
	}

	sparkCount := 0
	burstCount := 0
	for _, tile := range clearedTiles {
		if tile.Special == gameplay.HexBlitzTileSpecialSpark {
			sparkCount++
		}
	}
	for _, tile := range group {
		if tile.Special == gameplay.HexBlitzTileSpecialBurst {
			burstCount++
		}
	}

	base := len(clearedTiles) * len(clearedTiles) * 12
	comboBonus := int(float64(base) * float64(max(0, nextCombo-1)) * 0.35)
	specialBonus := sparkCount*70 + burstCount*40
	gainedScore := base + comboBonus + specialBonus

	b.tiles = hexBlitzReplaceClearedTiles(b.tiles, cleared, b.rng)
	b.score += gainedScore
	b.combo = nextCombo
	if b.combo > b.bestCombo {
		b.bestCombo = b.combo
	}
	b.moves++
	b.lastGain = gainedScore
	b.lastCleared = len(clearedTiles)
	b.updatedAt = now
	b.lastMoveAt = &now
	b.message = fmt.Sprintf(
		"服务端判定：清掉 %d 格，获得 %d 分，当前连击 x%d。",
		len(clearedTiles),
		gainedScore,
		nextCombo,
	)
	return &gameplay.HexBlitzMoveResult{
		SessionID:    sessionID,
		MatchID:      matchID,
		TileID:       tileID,
		ClearedCount: len(clearedTiles),
		GainedScore:  gainedScore,
		Score:        b.score,
		Combo:        b.combo,
		BestCombo:    b.bestCombo,
		Moves:        b.moves,
		Message:      b.message,
		UpdatedAt:    now,
	}, nil
}

func (b *hexBlitzPlayerBoard) snapshot(sessionID string, matchID uuid.UUID, phase gameplay.RoomStatus, now time.Time) *gameplay.HexBlitzBoardState {
	tiles := make([]gameplay.HexBlitzTile, len(b.tiles))
	copy(tiles, b.tiles)

	return &gameplay.HexBlitzBoardState{
		SessionID:   sessionID,
		MatchID:     matchID,
		Phase:       phase,
		Seed:        b.seed,
		Score:       b.score,
		Combo:       b.currentCombo(now),
		BestCombo:   b.bestCombo,
		Moves:       b.moves,
		LastGain:    b.lastGain,
		LastCleared: b.lastCleared,
		Message:     b.message,
		UpdatedAt:   b.updatedAt,
		Tiles:       tiles,
	}
}

func (b *hexBlitzPlayerBoard) currentCombo(now time.Time) int {
	if b.lastMoveAt == nil {
		return 0
	}
	if now.Sub(*b.lastMoveAt) > hexBlitzComboWindow {
		return 0
	}
	return b.combo
}

func buildHexBlitzBoardCoords() []hexBlitzBoardCoord {
	coords := make([]hexBlitzBoardCoord, 0, 19)
	for q := -hexBlitzBoardRadius; q <= hexBlitzBoardRadius; q++ {
		rMin := max(-hexBlitzBoardRadius, -q-hexBlitzBoardRadius)
		rMax := min(hexBlitzBoardRadius, -q+hexBlitzBoardRadius)
		for r := rMin; r <= rMax; r++ {
			coords = append(coords, hexBlitzBoardCoord{Q: q, R: r})
		}
	}
	return coords
}

func hexBlitzSeedFromMatchID(matchID uuid.UUID) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(matchID.String()))
	return int64(h.Sum64())
}

func hexBlitzCreateRandomBoard(rng *rand.Rand) []gameplay.HexBlitzTile {
	tiles := make([]gameplay.HexBlitzTile, 0, len(hexBlitzBoardCoords))
	for _, coord := range hexBlitzBoardCoords {
		tiles = append(tiles, gameplay.HexBlitzTile{
			ID:      hexBlitzCoordKey(coord.Q, coord.R),
			Q:       coord.Q,
			R:       coord.R,
			Color:   hexBlitzRandomColor(rng),
			Special: hexBlitzRandomSpecial(rng),
		})
	}
	return hexBlitzEnsurePlayableBoard(tiles, rng)
}

func hexBlitzRandomColor(rng *rand.Rand) gameplay.HexBlitzTileColor {
	colors := []gameplay.HexBlitzTileColor{
		gameplay.HexBlitzTileColorEmber,
		gameplay.HexBlitzTileColorLagoon,
		gameplay.HexBlitzTileColorMint,
		gameplay.HexBlitzTileColorSun,
		gameplay.HexBlitzTileColorViolet,
	}
	return colors[rng.Intn(len(colors))]
}

func hexBlitzRandomSpecial(rng *rand.Rand) gameplay.HexBlitzTileSpecial {
	roll := rng.Float64()
	if roll < 0.08 {
		return gameplay.HexBlitzTileSpecialBurst
	}
	if roll < 0.20 {
		return gameplay.HexBlitzTileSpecialSpark
	}
	return gameplay.HexBlitzTileSpecialNone
}

func hexBlitzCoordKey(q, r int) string {
	return fmt.Sprintf("%d:%d", q, r)
}

func hexBlitzCollectGroup(tiles []gameplay.HexBlitzTile, startID string) []gameplay.HexBlitzTile {
	tileMap := make(map[string]gameplay.HexBlitzTile, len(tiles))
	for _, tile := range tiles {
		tileMap[tile.ID] = tile
	}

	start, ok := tileMap[startID]
	if !ok {
		return nil
	}

	queue := []gameplay.HexBlitzTile{start}
	visited := map[string]struct{}{start.ID: {}}
	group := make([]gameplay.HexBlitzTile, 0, 4)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		group = append(group, current)
		for _, direction := range hexBlitzDirections {
			neighborID := hexBlitzCoordKey(current.Q+direction[0], current.R+direction[1])
			neighbor, exists := tileMap[neighborID]
			if !exists {
				continue
			}
			if _, seen := visited[neighbor.ID]; seen {
				continue
			}
			if neighbor.Color != start.Color {
				continue
			}
			visited[neighbor.ID] = struct{}{}
			queue = append(queue, neighbor)
		}
	}

	return group
}

func hexBlitzHasPlayableMove(tiles []gameplay.HexBlitzTile) bool {
	for _, tile := range tiles {
		if len(hexBlitzCollectGroup(tiles, tile.ID)) >= 2 {
			return true
		}
	}
	return false
}

func hexBlitzEnsurePlayableBoard(tiles []gameplay.HexBlitzTile, rng *rand.Rand) []gameplay.HexBlitzTile {
	next := cloneHexBlitzTiles(tiles)
	attempts := 0
	for !hexBlitzHasPlayableMove(next) && attempts < 8 {
		for index := range next {
			next[index].Color = hexBlitzRandomColor(rng)
			next[index].Special = hexBlitzRandomSpecial(rng)
		}
		attempts++
	}
	return next
}

func hexBlitzExpandBurst(group []gameplay.HexBlitzTile, tiles []gameplay.HexBlitzTile) map[string]struct{} {
	tileMap := make(map[string]gameplay.HexBlitzTile, len(tiles))
	for _, tile := range tiles {
		tileMap[tile.ID] = tile
	}

	cleared := make(map[string]struct{}, len(group))
	for _, tile := range group {
		cleared[tile.ID] = struct{}{}
	}

	for _, tile := range group {
		if tile.Special != gameplay.HexBlitzTileSpecialBurst {
			continue
		}
		for _, direction := range hexBlitzDirections {
			neighborID := hexBlitzCoordKey(tile.Q+direction[0], tile.R+direction[1])
			if neighbor, ok := tileMap[neighborID]; ok {
				cleared[neighbor.ID] = struct{}{}
			}
		}
	}

	return cleared
}

func hexBlitzReplaceClearedTiles(tiles []gameplay.HexBlitzTile, cleared map[string]struct{}, rng *rand.Rand) []gameplay.HexBlitzTile {
	next := cloneHexBlitzTiles(tiles)
	for index := range next {
		if _, ok := cleared[next[index].ID]; !ok {
			continue
		}
		next[index].Color = hexBlitzRandomColor(rng)
		next[index].Special = hexBlitzRandomSpecial(rng)
	}
	return hexBlitzEnsurePlayableBoard(next, rng)
}

func cloneHexBlitzTiles(tiles []gameplay.HexBlitzTile) []gameplay.HexBlitzTile {
	next := make([]gameplay.HexBlitzTile, len(tiles))
	copy(next, tiles)
	return next
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
