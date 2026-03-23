package usecase

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studio/platform/internal/domain/doudizhu"
)

func TestDoudizhuShuffleAndDeal(t *testing.T) {
	result := DoudizhuShuffleAndDeal(42)

	require.Len(t, result.Hands[0], 17)
	require.Len(t, result.Hands[1], 17)
	require.Len(t, result.Hands[2], 17)
	require.Len(t, result.Bottom, 3)

	seen := make(map[string]struct{}, 54)
	for _, hand := range result.Hands {
		for _, card := range hand {
			seen[card.String()] = struct{}{}
		}
	}
	for _, card := range result.Bottom {
		seen[card.String()] = struct{}{}
	}
	require.Len(t, seen, 54)
}

func TestDoudizhuEvaluateCombo(t *testing.T) {
	tests := []struct {
		name           string
		cards          []doudizhu.Card
		expectedType   doudizhu.ComboType
		expectedRank   doudizhu.Rank
		expectedLength int
	}{
		{
			name:           "single",
			cards:          testDouDizhuCards(doudizhu.RankA),
			expectedType:   doudizhu.ComboSingle,
			expectedRank:   doudizhu.RankA,
			expectedLength: 1,
		},
		{
			name:           "pair",
			cards:          testDouDizhuCards(doudizhu.Rank8, doudizhu.Rank8),
			expectedType:   doudizhu.ComboPair,
			expectedRank:   doudizhu.Rank8,
			expectedLength: 1,
		},
		{
			name:           "triple with single",
			cards:          testDouDizhuCards(doudizhu.Rank6, doudizhu.Rank6, doudizhu.Rank6, doudizhu.RankQ),
			expectedType:   doudizhu.ComboTripleWithSingle,
			expectedRank:   doudizhu.Rank6,
			expectedLength: 1,
		},
		{
			name:           "triple with pair",
			cards:          testDouDizhuCards(doudizhu.Rank9, doudizhu.Rank9, doudizhu.Rank9, doudizhu.Rank5, doudizhu.Rank5),
			expectedType:   doudizhu.ComboTripleWithPair,
			expectedRank:   doudizhu.Rank9,
			expectedLength: 1,
		},
		{
			name:           "straight",
			cards:          testDouDizhuCards(doudizhu.Rank6, doudizhu.Rank7, doudizhu.Rank8, doudizhu.Rank9, doudizhu.Rank10),
			expectedType:   doudizhu.ComboStraight,
			expectedRank:   doudizhu.Rank10,
			expectedLength: 5,
		},
		{
			name:           "straight pairs",
			cards:          testDouDizhuCards(doudizhu.Rank7, doudizhu.Rank7, doudizhu.Rank8, doudizhu.Rank8, doudizhu.Rank9, doudizhu.Rank9),
			expectedType:   doudizhu.ComboStraightPairs,
			expectedRank:   doudizhu.Rank9,
			expectedLength: 3,
		},
		{
			name:           "airplane",
			cards:          testDouDizhuCards(doudizhu.Rank4, doudizhu.Rank4, doudizhu.Rank4, doudizhu.Rank5, doudizhu.Rank5, doudizhu.Rank5),
			expectedType:   doudizhu.ComboAirplane,
			expectedRank:   doudizhu.Rank5,
			expectedLength: 2,
		},
		{
			name:           "airplane with singles",
			cards:          testDouDizhuCards(doudizhu.Rank4, doudizhu.Rank4, doudizhu.Rank4, doudizhu.Rank5, doudizhu.Rank5, doudizhu.Rank5, doudizhu.Rank9, doudizhu.RankJ),
			expectedType:   doudizhu.ComboAirplaneWithSingle,
			expectedRank:   doudizhu.Rank5,
			expectedLength: 2,
		},
		{
			name:           "airplane with pairs",
			cards:          testDouDizhuCards(doudizhu.Rank6, doudizhu.Rank6, doudizhu.Rank6, doudizhu.Rank7, doudizhu.Rank7, doudizhu.Rank7, doudizhu.Rank9, doudizhu.Rank9, doudizhu.RankJ, doudizhu.RankJ),
			expectedType:   doudizhu.ComboAirplaneWithPair,
			expectedRank:   doudizhu.Rank7,
			expectedLength: 2,
		},
		{
			name:           "four with two pairs",
			cards:          testDouDizhuCards(doudizhu.RankQ, doudizhu.RankQ, doudizhu.RankQ, doudizhu.RankQ, doudizhu.Rank5, doudizhu.Rank5, doudizhu.Rank8, doudizhu.Rank8),
			expectedType:   doudizhu.ComboFourWithTwoPair,
			expectedRank:   doudizhu.RankQ,
			expectedLength: 1,
		},
		{
			name:           "bomb",
			cards:          testDouDizhuCards(doudizhu.Rank10, doudizhu.Rank10, doudizhu.Rank10, doudizhu.Rank10),
			expectedType:   doudizhu.ComboBomb,
			expectedRank:   doudizhu.Rank10,
			expectedLength: 1,
		},
		{
			name:           "rocket",
			cards:          []doudizhu.Card{{Suit: doudizhu.SuitJoker, Rank: doudizhu.RankJokerSmall}, {Suit: doudizhu.SuitJoker, Rank: doudizhu.RankJokerBig}},
			expectedType:   doudizhu.ComboRocket,
			expectedRank:   doudizhu.RankJokerBig,
			expectedLength: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			combo, err := DoudizhuEvaluateCombo(tt.cards)
			require.NoError(t, err)
			require.Equal(t, tt.expectedType, combo.Type)
			require.Equal(t, tt.expectedRank, combo.MainRank)
			require.Equal(t, tt.expectedLength, combo.SequenceLength)
			require.Equal(t, len(tt.cards), combo.TotalCards)
		})
	}
}

func TestDoudizhuEvaluateComboRejectsInvalidStraight(t *testing.T) {
	_, err := DoudizhuEvaluateCombo(testDouDizhuCards(
		doudizhu.Rank10,
		doudizhu.RankJ,
		doudizhu.RankQ,
		doudizhu.RankK,
		doudizhu.RankA,
		doudizhu.Rank2,
	))
	require.ErrorIs(t, err, doudizhu.ErrInvalidCombo)
}

func TestDoudizhuCompareCombos(t *testing.T) {
	biggerPair, err := DoudizhuEvaluateCombo(testDouDizhuCards(doudizhu.RankK, doudizhu.RankK))
	require.NoError(t, err)
	smallerPair, err := DoudizhuEvaluateCombo(testDouDizhuCards(doudizhu.Rank10, doudizhu.Rank10))
	require.NoError(t, err)

	ok, err := DoudizhuCompareCombos(biggerPair, smallerPair)
	require.NoError(t, err)
	require.True(t, ok)

	bomb, err := DoudizhuEvaluateCombo(testDouDizhuCards(doudizhu.Rank6, doudizhu.Rank6, doudizhu.Rank6, doudizhu.Rank6))
	require.NoError(t, err)
	straight, err := DoudizhuEvaluateCombo(testDouDizhuCards(
		doudizhu.Rank5,
		doudizhu.Rank6,
		doudizhu.Rank7,
		doudizhu.Rank8,
		doudizhu.Rank9,
	))
	require.NoError(t, err)

	ok, err = DoudizhuCompareCombos(bomb, straight)
	require.NoError(t, err)
	require.True(t, ok)

	rocket, err := DoudizhuEvaluateCombo([]doudizhu.Card{
		{Suit: doudizhu.SuitJoker, Rank: doudizhu.RankJokerSmall},
		{Suit: doudizhu.SuitJoker, Rank: doudizhu.RankJokerBig},
	})
	require.NoError(t, err)

	ok, err = DoudizhuCompareCombos(rocket, bomb)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestDoudizhuRoundAssignsLandlordAfterBidding(t *testing.T) {
	round, err := NewDoudizhuRound(7, doudizhu.MatchModePVP, doudizhu.Seat1)
	require.NoError(t, err)
	require.Equal(t, doudizhu.RoundPhaseBidding, round.Phase)
	require.Equal(t, doudizhu.Seat1, round.CurrentBidder)

	err = round.ApplyBid(doudizhu.Seat1, 1)
	require.NoError(t, err)
	require.Equal(t, doudizhu.Seat2, round.CurrentBidder)

	err = round.ApplyBid(doudizhu.Seat2, 2)
	require.NoError(t, err)
	require.Equal(t, doudizhu.Seat0, round.CurrentBidder)

	err = round.ApplyBid(doudizhu.Seat0, 0)
	require.NoError(t, err)
	require.Equal(t, doudizhu.RoundPhasePlaying, round.Phase)
	require.NotNil(t, round.Landlord)
	require.Equal(t, doudizhu.Seat2, *round.Landlord)
	require.Equal(t, doudizhu.PlayerRoleLandlord, round.Roles[doudizhu.Seat2])
	require.Len(t, round.Hands[doudizhu.Seat2], 20)
	require.NotNil(t, round.CurrentTurn)
	require.Equal(t, doudizhu.Seat2, *round.CurrentTurn)
}

func TestDoudizhuRoundAllPassTriggersRedeal(t *testing.T) {
	round, err := NewDoudizhuRound(9, doudizhu.MatchModeDemoAI, doudizhu.Seat0)
	require.NoError(t, err)

	require.NoError(t, round.ApplyBid(doudizhu.Seat0, 0))
	require.NoError(t, round.ApplyBid(doudizhu.Seat1, 0))
	require.NoError(t, round.ApplyBid(doudizhu.Seat2, 0))

	require.Equal(t, doudizhu.RoundPhaseRedeal, round.Phase)
	require.Nil(t, round.Landlord)
	require.Nil(t, round.CurrentTurn)
}

func testDouDizhuCards(ranks ...doudizhu.Rank) []doudizhu.Card {
	suits := []doudizhu.Suit{
		doudizhu.SuitSpade,
		doudizhu.SuitHeart,
		doudizhu.SuitClub,
		doudizhu.SuitDiamond,
	}
	used := make(map[doudizhu.Rank]int)
	cards := make([]doudizhu.Card, 0, len(ranks))
	for _, rank := range ranks {
		if rank == doudizhu.RankJokerSmall || rank == doudizhu.RankJokerBig {
			cards = append(cards, doudizhu.Card{
				Suit: doudizhu.SuitJoker,
				Rank: rank,
			})
			continue
		}
		index := used[rank]
		cards = append(cards, doudizhu.Card{
			Suit: suits[index%len(suits)],
			Rank: rank,
		})
		used[rank]++
	}
	return cards
}
