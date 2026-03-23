package usecase

import (
	"sort"

	"github.com/studio/platform/internal/domain/doudizhu"
)

func chooseDoudizhuBid(hand []doudizhu.Card, currentHighest int) int {
	score := 0
	counts := doudizhuRankCounts(hand)

	for rank, count := range counts {
		switch {
		case rank == doudizhu.RankJokerBig:
			score += 3
		case rank == doudizhu.RankJokerSmall:
			score += 2
		case rank == doudizhu.Rank2:
			score += count
		case count == 4:
			score += 3
		case count == 3:
			score += 1
		}
	}

	bid := 0
	switch {
	case score >= 8:
		bid = 3
	case score >= 5:
		bid = 2
	case score >= 2:
		bid = 1
	default:
		bid = 0
	}
	if bid <= currentHighest {
		return 0
	}
	return bid
}

func chooseDoudizhuPlay(
	hand []doudizhu.Card,
	target *doudizhu.Combo,
	targetSeat *doudizhu.Seat,
	selfSeat doudizhu.Seat,
) []doudizhu.Card {
	if len(hand) == 0 {
		return nil
	}

	if target == nil || targetSeat == nil || *targetSeat == selfSeat {
		return chooseDoudizhuLeadPlay(hand)
	}

	if cards := chooseDoudizhuFollowByType(hand, target); len(cards) > 0 {
		return cards
	}
	if cards := chooseDoudizhuBomb(hand, target.MainRank); len(cards) > 0 {
		return cards
	}
	if cards := chooseDoudizhuRocket(hand); len(cards) > 0 {
		return cards
	}
	return nil
}

func chooseDoudizhuLeadPlay(hand []doudizhu.Card) []doudizhu.Card {
	sorted := doudizhuSortedCards(hand)
	counts := doudizhuRankCounts(sorted)

	if cards := chooseDoudizhuStraight(sorted); len(cards) > 0 {
		return cards
	}
	if cards := chooseDoudizhuTripleWithPair(sorted); len(cards) > 0 {
		return cards
	}
	if cards := chooseDoudizhuTripleWithSingle(sorted); len(cards) > 0 {
		return cards
	}
	if cards := chooseDoudizhuPair(sorted, 0); len(cards) > 0 {
		return cards
	}
	for _, rank := range doudizhuSortedRanks(counts) {
		return cardsOfRank(sorted, rank, 1)
	}
	return nil
}

func chooseDoudizhuFollowByType(hand []doudizhu.Card, target *doudizhu.Combo) []doudizhu.Card {
	sorted := doudizhuSortedCards(hand)
	switch target.Type {
	case doudizhu.ComboSingle:
		return chooseDoudizhuSingle(sorted, target.MainRank)
	case doudizhu.ComboPair:
		return chooseDoudizhuPair(sorted, target.MainRank)
	case doudizhu.ComboTriple:
		return chooseDoudizhuTriple(sorted, target.MainRank)
	case doudizhu.ComboTripleWithSingle:
		return chooseDoudizhuTripleWithSingleAbove(sorted, target.MainRank)
	case doudizhu.ComboTripleWithPair:
		return chooseDoudizhuTripleWithPairAbove(sorted, target.MainRank)
	case doudizhu.ComboStraight:
		return chooseDoudizhuStraightAbove(sorted, target.SequenceLength, target.MainRank)
	}
	return nil
}

func chooseDoudizhuSingle(hand []doudizhu.Card, minRank doudizhu.Rank) []doudizhu.Card {
	for _, card := range hand {
		if card.Rank > minRank {
			return []doudizhu.Card{card}
		}
	}
	return nil
}

func chooseDoudizhuPair(hand []doudizhu.Card, minRank doudizhu.Rank) []doudizhu.Card {
	counts := doudizhuRankCounts(hand)
	for _, rank := range doudizhuSortedRanks(counts) {
		if rank > minRank && counts[rank] >= 2 {
			return cardsOfRank(hand, rank, 2)
		}
	}
	return nil
}

func chooseDoudizhuTriple(hand []doudizhu.Card, minRank doudizhu.Rank) []doudizhu.Card {
	counts := doudizhuRankCounts(hand)
	for _, rank := range doudizhuSortedRanks(counts) {
		if rank > minRank && counts[rank] >= 3 {
			return cardsOfRank(hand, rank, 3)
		}
	}
	return nil
}

func chooseDoudizhuTripleWithSingle(hand []doudizhu.Card) []doudizhu.Card {
	return chooseDoudizhuTripleWithSingleAbove(hand, 0)
}

func chooseDoudizhuTripleWithSingleAbove(hand []doudizhu.Card, minRank doudizhu.Rank) []doudizhu.Card {
	counts := doudizhuRankCounts(hand)
	for _, rank := range doudizhuSortedRanks(counts) {
		if rank <= minRank || counts[rank] < 3 {
			continue
		}
		base := cardsOfRank(hand, rank, 3)
		for _, kicker := range hand {
			if kicker.Rank != rank {
				return append(base, kicker)
			}
		}
	}
	return nil
}

func chooseDoudizhuTripleWithPair(hand []doudizhu.Card) []doudizhu.Card {
	return chooseDoudizhuTripleWithPairAbove(hand, 0)
}

func chooseDoudizhuTripleWithPairAbove(hand []doudizhu.Card, minRank doudizhu.Rank) []doudizhu.Card {
	counts := doudizhuRankCounts(hand)
	ranks := doudizhuSortedRanks(counts)
	for _, rank := range ranks {
		if rank <= minRank || counts[rank] < 3 {
			continue
		}
		for _, pairRank := range ranks {
			if pairRank == rank || counts[pairRank] < 2 {
				continue
			}
			cards := append(cardsOfRank(hand, rank, 3), cardsOfRank(hand, pairRank, 2)...)
			return cards
		}
	}
	return nil
}

func chooseDoudizhuStraight(hand []doudizhu.Card) []doudizhu.Card {
	return chooseDoudizhuStraightAbove(hand, 5, 0)
}

func chooseDoudizhuStraightAbove(hand []doudizhu.Card, length int, minMainRank doudizhu.Rank) []doudizhu.Card {
	if length < 5 {
		return nil
	}
	counts := doudizhuRankCounts(hand)
	ranks := make([]int, 0)
	for _, rank := range doudizhuSortedRanks(counts) {
		if rank <= doudizhu.RankA && counts[rank] >= 1 {
			ranks = append(ranks, int(rank))
		}
	}
	if len(ranks) < length {
		return nil
	}

	for start := 0; start+length <= len(ranks); start++ {
		ok := true
		for index := 1; index < length; index++ {
			if ranks[start+index] != ranks[start+index-1]+1 {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		mainRank := doudizhu.Rank(ranks[start+length-1])
		if mainRank <= minMainRank {
			continue
		}
		selected := make([]doudizhu.Card, 0, length)
		for _, rankValue := range ranks[start : start+length] {
			selected = append(selected, cardsOfRank(hand, doudizhu.Rank(rankValue), 1)...)
		}
		return selected
	}
	return nil
}

func chooseDoudizhuBomb(hand []doudizhu.Card, minRank doudizhu.Rank) []doudizhu.Card {
	counts := doudizhuRankCounts(hand)
	for _, rank := range doudizhuSortedRanks(counts) {
		if rank > minRank && counts[rank] == 4 {
			return cardsOfRank(hand, rank, 4)
		}
	}
	return nil
}

func chooseDoudizhuRocket(hand []doudizhu.Card) []doudizhu.Card {
	var small, big *doudizhu.Card
	for _, card := range hand {
		switch card.Rank {
		case doudizhu.RankJokerSmall:
			c := card
			small = &c
		case doudizhu.RankJokerBig:
			c := card
			big = &c
		}
	}
	if small != nil && big != nil {
		return []doudizhu.Card{*small, *big}
	}
	return nil
}

func cardsOfRank(hand []doudizhu.Card, rank doudizhu.Rank, count int) []doudizhu.Card {
	selected := make([]doudizhu.Card, 0, count)
	for _, card := range hand {
		if card.Rank != rank {
			continue
		}
		selected = append(selected, card)
		if len(selected) == count {
			return selected
		}
	}
	return nil
}

func doudizhuSortedRanksForBots(counts map[doudizhu.Rank]int) []doudizhu.Rank {
	ranks := doudizhuSortedRanks(counts)
	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i] < ranks[j]
	})
	return ranks
}
