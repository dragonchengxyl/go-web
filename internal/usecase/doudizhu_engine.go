package usecase

import (
	"math/rand"
	"sort"
	"time"

	"github.com/studio/platform/internal/domain/doudizhu"
)

type DoudizhuDealResult struct {
	Hands  [3][]doudizhu.Card
	Bottom []doudizhu.Card
}

type DoudizhuBidRecord struct {
	Seat  doudizhu.Seat `json:"seat"`
	Score int           `json:"score"`
}

type DoudizhuRoundState struct {
	Seed          int64                  `json:"seed"`
	Phase         doudizhu.RoundPhase    `json:"phase"`
	StartedAt     time.Time              `json:"started_at"`
	FinishedAt    *time.Time             `json:"finished_at,omitempty"`
	Hands         [3][]doudizhu.Card     `json:"hands"`
	BottomCards   []doudizhu.Card        `json:"bottom_cards"`
	CurrentBidder doudizhu.Seat          `json:"current_bidder"`
	HighestBid    int                    `json:"highest_bid"`
	HighestBidder *doudizhu.Seat         `json:"highest_bidder,omitempty"`
	Landlord      *doudizhu.Seat         `json:"landlord,omitempty"`
	CurrentTurn   *doudizhu.Seat         `json:"current_turn,omitempty"`
	Roles         [3]doudizhu.PlayerRole `json:"roles"`
	BidHistory    []DoudizhuBidRecord    `json:"bid_history"`
	MatchMode     doudizhu.MatchMode     `json:"match_mode"`
	LastPlay      *doudizhu.Combo        `json:"last_play,omitempty"`
}

func DoudizhuBuildDeck() []doudizhu.Card {
	deck := make([]doudizhu.Card, 0, 54)
	suits := []doudizhu.Suit{
		doudizhu.SuitSpade,
		doudizhu.SuitHeart,
		doudizhu.SuitClub,
		doudizhu.SuitDiamond,
	}

	for rank := doudizhu.Rank3; rank <= doudizhu.Rank2; rank++ {
		for _, suit := range suits {
			deck = append(deck, doudizhu.Card{
				Suit: suit,
				Rank: rank,
			})
		}
	}

	deck = append(deck,
		doudizhu.Card{Suit: doudizhu.SuitJoker, Rank: doudizhu.RankJokerSmall},
		doudizhu.Card{Suit: doudizhu.SuitJoker, Rank: doudizhu.RankJokerBig},
	)
	return deck
}

func DoudizhuShuffleAndDeal(seed int64) DoudizhuDealResult {
	deck := DoudizhuBuildDeck()
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	result := DoudizhuDealResult{}
	for index, card := range deck[:51] {
		seat := index % 3
		result.Hands[seat] = append(result.Hands[seat], card)
	}
	result.Bottom = append(result.Bottom, deck[51:]...)

	for seat := range result.Hands {
		result.Hands[seat] = doudizhuSortedCards(result.Hands[seat])
	}
	result.Bottom = doudizhuSortedCards(result.Bottom)
	return result
}

func NewDoudizhuRound(seed int64, matchMode doudizhu.MatchMode, startingSeat doudizhu.Seat) (*DoudizhuRoundState, error) {
	if !startingSeat.Valid() {
		return nil, doudizhu.ErrInvalidSeat
	}
	if matchMode == "" {
		matchMode = doudizhu.MatchModePVP
	}

	deal := DoudizhuShuffleAndDeal(seed)
	state := &DoudizhuRoundState{
		Seed:          seed,
		Phase:         doudizhu.RoundPhaseBidding,
		StartedAt:     time.Now(),
		Hands:         deal.Hands,
		BottomCards:   deal.Bottom,
		CurrentBidder: startingSeat,
		MatchMode:     matchMode,
	}
	for seat := range state.Roles {
		state.Roles[seat] = doudizhu.PlayerRoleFarmer
	}
	return state, nil
}

func (s *DoudizhuRoundState) ApplyBid(seat doudizhu.Seat, score int) error {
	if s == nil || s.Phase != doudizhu.RoundPhaseBidding {
		return doudizhu.ErrRoundNotBidding
	}
	if !seat.Valid() {
		return doudizhu.ErrInvalidSeat
	}
	if seat != s.CurrentBidder {
		return doudizhu.ErrNotCurrentBidder
	}
	if score < 0 || score > 3 {
		return doudizhu.ErrInvalidBidScore
	}
	if score > 0 && score <= s.HighestBid {
		return doudizhu.ErrInvalidBidScore
	}

	s.BidHistory = append(s.BidHistory, DoudizhuBidRecord{
		Seat:  seat,
		Score: score,
	})
	if score > s.HighestBid {
		s.HighestBid = score
		winner := seat
		s.HighestBidder = &winner
	}

	if score == 3 || len(s.BidHistory) >= 3 {
		s.finalizeBidding()
		return nil
	}

	s.CurrentBidder = seat.Next()
	return nil
}

func (s *DoudizhuRoundState) finalizeBidding() {
	if s.HighestBidder == nil {
		s.Phase = doudizhu.RoundPhaseRedeal
		return
	}

	landlord := *s.HighestBidder
	s.Landlord = &landlord
	s.CurrentTurn = &landlord
	s.Phase = doudizhu.RoundPhasePlaying
	s.Roles[landlord] = doudizhu.PlayerRoleLandlord
	s.Hands[landlord] = append(append([]doudizhu.Card(nil), s.Hands[landlord]...), s.BottomCards...)
	s.Hands[landlord] = doudizhuSortedCards(s.Hands[landlord])
}

func DoudizhuEvaluateCombo(cards []doudizhu.Card) (*doudizhu.Combo, error) {
	if len(cards) == 0 {
		return nil, doudizhu.ErrInvalidCombo
	}

	sorted := doudizhuSortedCards(cards)
	counts := doudizhuRankCounts(sorted)
	uniqueRanks := doudizhuSortedRanks(counts)
	total := len(sorted)

	switch total {
	case 1:
		return &doudizhu.Combo{
			Type:           doudizhu.ComboSingle,
			MainRank:       sorted[0].Rank,
			SequenceLength: 1,
			TotalCards:     total,
		}, nil
	case 2:
		if sorted[0].Rank == doudizhu.RankJokerSmall && sorted[1].Rank == doudizhu.RankJokerBig {
			return &doudizhu.Combo{
				Type:           doudizhu.ComboRocket,
				MainRank:       doudizhu.RankJokerBig,
				SequenceLength: 1,
				TotalCards:     total,
			}, nil
		}
		if len(uniqueRanks) == 1 {
			return &doudizhu.Combo{
				Type:           doudizhu.ComboPair,
				MainRank:       uniqueRanks[0],
				SequenceLength: 1,
				TotalCards:     total,
			}, nil
		}
	case 3:
		if len(uniqueRanks) == 1 {
			return &doudizhu.Combo{
				Type:           doudizhu.ComboTriple,
				MainRank:       uniqueRanks[0],
				SequenceLength: 1,
				TotalCards:     total,
			}, nil
		}
	case 4:
		if len(uniqueRanks) == 1 {
			return &doudizhu.Combo{
				Type:           doudizhu.ComboBomb,
				MainRank:       uniqueRanks[0],
				SequenceLength: 1,
				TotalCards:     total,
			}, nil
		}
		if tripleRank, ok := doudizhuRankWithCount(counts, 3); ok {
			return &doudizhu.Combo{
				Type:           doudizhu.ComboTripleWithSingle,
				MainRank:       tripleRank,
				SequenceLength: 1,
				TotalCards:     total,
			}, nil
		}
	case 5:
		if tripleRank, ok := doudizhuRankWithCount(counts, 3); ok && doudizhuHasCount(counts, 2) {
			return &doudizhu.Combo{
				Type:           doudizhu.ComboTripleWithPair,
				MainRank:       tripleRank,
				SequenceLength: 1,
				TotalCards:     total,
			}, nil
		}
	}

	if doudizhuIsStraight(counts) {
		return &doudizhu.Combo{
			Type:           doudizhu.ComboStraight,
			MainRank:       uniqueRanks[len(uniqueRanks)-1],
			SequenceLength: len(uniqueRanks),
			TotalCards:     total,
		}, nil
	}

	if doudizhuIsStraightPairs(counts) {
		return &doudizhu.Combo{
			Type:           doudizhu.ComboStraightPairs,
			MainRank:       uniqueRanks[len(uniqueRanks)-1],
			SequenceLength: len(uniqueRanks),
			TotalCards:     total,
		}, nil
	}

	if airplane := doudizhuFindAirplane(counts, total); airplane != nil {
		return airplane, nil
	}

	if total == 6 {
		if fourRank, ok := doudizhuRankWithCount(counts, 4); ok && doudizhuCountPattern(counts, 4, 1, 1) {
			return &doudizhu.Combo{
				Type:           doudizhu.ComboFourWithTwoSingle,
				MainRank:       fourRank,
				SequenceLength: 1,
				TotalCards:     total,
			}, nil
		}
	}
	if total == 8 {
		if fourRank, ok := doudizhuRankWithCount(counts, 4); ok && doudizhuCountPattern(counts, 4, 2, 2) {
			return &doudizhu.Combo{
				Type:           doudizhu.ComboFourWithTwoPair,
				MainRank:       fourRank,
				SequenceLength: 1,
				TotalCards:     total,
			}, nil
		}
	}

	return nil, doudizhu.ErrInvalidCombo
}

func DoudizhuCompareCombos(candidate, target *doudizhu.Combo) (bool, error) {
	if candidate == nil || target == nil {
		return false, doudizhu.ErrInvalidCombo
	}

	if candidate.Type == doudizhu.ComboRocket {
		return true, nil
	}
	if target.Type == doudizhu.ComboRocket {
		return false, nil
	}
	if candidate.Type == doudizhu.ComboBomb && target.Type != doudizhu.ComboBomb {
		return true, nil
	}
	if target.Type == doudizhu.ComboBomb && candidate.Type != doudizhu.ComboBomb {
		return false, nil
	}
	if candidate.Type != target.Type {
		return false, doudizhu.ErrCannotBeatTargetPlay
	}
	if candidate.TotalCards != target.TotalCards || candidate.SequenceLength != target.SequenceLength {
		return false, doudizhu.ErrCannotBeatTargetPlay
	}
	return candidate.MainRank > target.MainRank, nil
}

func DoudizhuCanBeat(candidateCards, targetCards []doudizhu.Card) (bool, error) {
	candidate, err := DoudizhuEvaluateCombo(candidateCards)
	if err != nil {
		return false, err
	}
	target, err := DoudizhuEvaluateCombo(targetCards)
	if err != nil {
		return false, err
	}
	return DoudizhuCompareCombos(candidate, target)
}

func doudizhuSortedCards(cards []doudizhu.Card) []doudizhu.Card {
	sorted := append([]doudizhu.Card(nil), cards...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Rank == sorted[j].Rank {
			return sorted[i].Suit < sorted[j].Suit
		}
		return sorted[i].Rank < sorted[j].Rank
	})
	return sorted
}

func doudizhuRankCounts(cards []doudizhu.Card) map[doudizhu.Rank]int {
	counts := make(map[doudizhu.Rank]int, len(cards))
	for _, card := range cards {
		counts[card.Rank]++
	}
	return counts
}

func doudizhuSortedRanks(counts map[doudizhu.Rank]int) []doudizhu.Rank {
	ranks := make([]doudizhu.Rank, 0, len(counts))
	for rank := range counts {
		ranks = append(ranks, rank)
	}
	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i] < ranks[j]
	})
	return ranks
}

func doudizhuRankWithCount(counts map[doudizhu.Rank]int, expected int) (doudizhu.Rank, bool) {
	for rank, count := range counts {
		if count == expected {
			return rank, true
		}
	}
	return 0, false
}

func doudizhuHasCount(counts map[doudizhu.Rank]int, expected int) bool {
	for _, count := range counts {
		if count == expected {
			return true
		}
	}
	return false
}

func doudizhuIsStraight(counts map[doudizhu.Rank]int) bool {
	if len(counts) < 5 {
		return false
	}
	ranks := doudizhuSortedRanks(counts)
	for index, rank := range ranks {
		if rank > doudizhu.RankA || counts[rank] != 1 {
			return false
		}
		if index > 0 && rank != ranks[index-1]+1 {
			return false
		}
	}
	return true
}

func doudizhuIsStraightPairs(counts map[doudizhu.Rank]int) bool {
	if len(counts) < 3 {
		return false
	}
	ranks := doudizhuSortedRanks(counts)
	for index, rank := range ranks {
		if rank > doudizhu.RankA || counts[rank] != 2 {
			return false
		}
		if index > 0 && rank != ranks[index-1]+1 {
			return false
		}
	}
	return true
}

func doudizhuFindAirplane(counts map[doudizhu.Rank]int, total int) *doudizhu.Combo {
	tripleRanks := make([]doudizhu.Rank, 0)
	for rank, count := range counts {
		if count >= 3 && rank <= doudizhu.RankA {
			tripleRanks = append(tripleRanks, rank)
		}
	}
	sort.Slice(tripleRanks, func(i, j int) bool {
		return tripleRanks[i] < tripleRanks[j]
	})
	if len(tripleRanks) < 2 {
		return nil
	}

	for runLength := len(tripleRanks); runLength >= 2; runLength-- {
		for start := 0; start+runLength <= len(tripleRanks); start++ {
			segment := tripleRanks[start : start+runLength]
			if !doudizhuAreConsecutive(segment) {
				continue
			}
			if combo := doudizhuBuildAirplaneCombo(counts, segment, total); combo != nil {
				return combo
			}
		}
	}
	return nil
}

func doudizhuAreConsecutive(ranks []doudizhu.Rank) bool {
	for index := 1; index < len(ranks); index++ {
		if ranks[index] != ranks[index-1]+1 {
			return false
		}
	}
	return true
}

func doudizhuBuildAirplaneCombo(counts map[doudizhu.Rank]int, segment []doudizhu.Rank, total int) *doudizhu.Combo {
	remaining := make(map[doudizhu.Rank]int, len(counts))
	for rank, count := range counts {
		remaining[rank] = count
	}
	for _, rank := range segment {
		remaining[rank] -= 3
	}

	runLength := len(segment)
	remainingCards := total - runLength*3

	var comboType doudizhu.ComboType
	switch {
	case remainingCards == 0:
		comboType = doudizhu.ComboAirplane
	case remainingCards == runLength && doudizhuRemainingPattern(remaining, 1, runLength):
		comboType = doudizhu.ComboAirplaneWithSingle
	case remainingCards == runLength*2 && doudizhuRemainingPattern(remaining, 2, runLength):
		comboType = doudizhu.ComboAirplaneWithPair
	default:
		return nil
	}

	return &doudizhu.Combo{
		Type:           comboType,
		MainRank:       segment[len(segment)-1],
		SequenceLength: runLength,
		TotalCards:     total,
	}
}

func doudizhuRemainingPattern(counts map[doudizhu.Rank]int, expectedCount, expectedKinds int) bool {
	kinds := 0
	for _, count := range counts {
		switch {
		case count == 0:
			continue
		case count != expectedCount:
			return false
		default:
			kinds++
		}
	}
	return kinds == expectedKinds
}

func doudizhuCountPattern(counts map[doudizhu.Rank]int, pattern ...int) bool {
	found := make([]int, 0, len(counts))
	for _, count := range counts {
		found = append(found, count)
	}
	sort.Ints(found)
	sort.Ints(pattern)
	if len(found) != len(pattern) {
		return false
	}
	for index := range found {
		if found[index] != pattern[index] {
			return false
		}
	}
	return true
}
