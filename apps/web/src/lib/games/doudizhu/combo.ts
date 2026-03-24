import type {
  DoudizhuCard,
  DoudizhuCombo,
  DoudizhuComboType,
} from "@/lib/api-client";
import { sortedSelection } from "@/lib/games/doudizhu/cards";

export interface EvaluatedDoudizhuCombo {
  combo: DoudizhuCombo | null;
  error: string;
}

const comboTypeLabelMap: Record<DoudizhuComboType, string> = {
  single: "单张",
  pair: "对子",
  triple: "三张",
  triple_with_single: "三带一",
  triple_with_pair: "三带二",
  straight: "顺子",
  straight_pairs: "连对",
  airplane: "飞机",
  airplane_with_single: "飞机带单",
  airplane_with_pair: "飞机带对",
  four_with_two_single: "四带二",
  four_with_two_pair: "四带两对",
  bomb: "炸弹",
  rocket: "王炸",
};

export function comboTypeLabel(type?: DoudizhuComboType | string): string {
  if (!type) {
    return "普通操作";
  }
  return comboTypeLabelMap[type as DoudizhuComboType] ?? type;
}

export function comboLabel(combo?: DoudizhuCombo | null): string {
  if (!combo) {
    return "等待首家出牌";
  }
  return `${comboTypeLabel(combo.type)} · 主值 ${combo.main_rank}`;
}

function rankCounts(cards: DoudizhuCard[]): Map<number, number> {
  const counts = new Map<number, number>();
  for (const card of cards) {
    counts.set(card.rank, (counts.get(card.rank) ?? 0) + 1);
  }
  return counts;
}

function sortedRanks(counts: Map<number, number>): number[] {
  return [...counts.keys()].sort((a, b) => a - b);
}

function rankWithCount(
  counts: Map<number, number>,
  expected: number,
): number | null {
  for (const [rank, count] of counts.entries()) {
    if (count === expected) {
      return rank;
    }
  }
  return null;
}

function hasCount(counts: Map<number, number>, expected: number): boolean {
  return [...counts.values()].some((count) => count === expected);
}

function countPattern(counts: Map<number, number>, pattern: number[]): boolean {
  const found = [...counts.values()].sort((a, b) => a - b);
  const expected = [...pattern].sort((a, b) => a - b);
  return (
    found.length === expected.length &&
    found.every((value, index) => value === expected[index])
  );
}

function areConsecutive(ranks: number[]): boolean {
  for (let index = 1; index < ranks.length; index += 1) {
    if (ranks[index] !== ranks[index - 1] + 1) {
      return false;
    }
  }
  return true;
}

function isStraight(counts: Map<number, number>): boolean {
  if (counts.size < 5) {
    return false;
  }
  const ranks = sortedRanks(counts);
  return ranks.every((rank, index) => {
    if (rank > 14 || counts.get(rank) !== 1) {
      return false;
    }
    return index === 0 || rank === ranks[index - 1] + 1;
  });
}

function isStraightPairs(counts: Map<number, number>): boolean {
  if (counts.size < 3) {
    return false;
  }
  const ranks = sortedRanks(counts);
  return ranks.every((rank, index) => {
    if (rank > 14 || counts.get(rank) !== 2) {
      return false;
    }
    return index === 0 || rank === ranks[index - 1] + 1;
  });
}

function remainingPattern(
  counts: Map<number, number>,
  expectedCount: number,
  expectedKinds: number,
): boolean {
  let kinds = 0;
  for (const count of counts.values()) {
    if (count === 0) {
      continue;
    }
    if (count !== expectedCount) {
      return false;
    }
    kinds += 1;
  }
  return kinds === expectedKinds;
}

function airplaneCombo(
  counts: Map<number, number>,
  segment: number[],
  total: number,
): DoudizhuCombo | null {
  const remaining = new Map(counts);
  for (const rank of segment) {
    remaining.set(rank, (remaining.get(rank) ?? 0) - 3);
  }
  const runLength = segment.length;
  const remainingCards = total - runLength * 3;

  let type: DoudizhuCombo["type"] | null = null;
  if (remainingCards === 0) {
    type = "airplane";
  } else if (
    remainingCards === runLength &&
    remainingPattern(remaining, 1, runLength)
  ) {
    type = "airplane_with_single";
  } else if (
    remainingCards === runLength * 2 &&
    remainingPattern(remaining, 2, runLength)
  ) {
    type = "airplane_with_pair";
  }

  if (!type) {
    return null;
  }
  return {
    type,
    main_rank: segment[segment.length - 1],
    sequence_length: runLength,
    total_cards: total,
  };
}

function findAirplane(
  counts: Map<number, number>,
  total: number,
): DoudizhuCombo | null {
  const triples = [...counts.entries()]
    .filter(([rank, count]) => count >= 3 && rank <= 14)
    .map(([rank]) => rank)
    .sort((a, b) => a - b);

  for (let runLength = triples.length; runLength >= 2; runLength -= 1) {
    for (let start = 0; start + runLength <= triples.length; start += 1) {
      const segment = triples.slice(start, start + runLength);
      if (!areConsecutive(segment)) {
        continue;
      }
      const combo = airplaneCombo(counts, segment, total);
      if (combo) {
        return combo;
      }
    }
  }
  return null;
}

export function evaluateSelectedCombo(
  cards: DoudizhuCard[],
): EvaluatedDoudizhuCombo {
  if (cards.length === 0) {
    return { combo: null, error: "" };
  }

  const sorted = sortedSelection(cards);
  const counts = rankCounts(sorted);
  const ranks = sortedRanks(counts);
  const total = sorted.length;

  if (total === 1) {
    return {
      combo: {
        type: "single",
        main_rank: sorted[0].rank,
        sequence_length: 1,
        total_cards: total,
      },
      error: "",
    };
  }
  if (total === 2) {
    if (sorted[0].rank === 16 && sorted[1].rank === 17) {
      return {
        combo: {
          type: "rocket",
          main_rank: 17,
          sequence_length: 1,
          total_cards: total,
        },
        error: "",
      };
    }
    if (ranks.length === 1) {
      return {
        combo: {
          type: "pair",
          main_rank: ranks[0],
          sequence_length: 1,
          total_cards: total,
        },
        error: "",
      };
    }
  }
  if (total === 3 && ranks.length === 1) {
    return {
      combo: {
        type: "triple",
        main_rank: ranks[0],
        sequence_length: 1,
        total_cards: total,
      },
      error: "",
    };
  }
  if (total === 4) {
    if (ranks.length === 1) {
      return {
        combo: {
          type: "bomb",
          main_rank: ranks[0],
          sequence_length: 1,
          total_cards: total,
        },
        error: "",
      };
    }
    const tripleRank = rankWithCount(counts, 3);
    if (tripleRank) {
      return {
        combo: {
          type: "triple_with_single",
          main_rank: tripleRank,
          sequence_length: 1,
          total_cards: total,
        },
        error: "",
      };
    }
  }
  if (total === 5) {
    const tripleRank = rankWithCount(counts, 3);
    if (tripleRank && hasCount(counts, 2)) {
      return {
        combo: {
          type: "triple_with_pair",
          main_rank: tripleRank,
          sequence_length: 1,
          total_cards: total,
        },
        error: "",
      };
    }
  }
  if (isStraight(counts)) {
    return {
      combo: {
        type: "straight",
        main_rank: ranks[ranks.length - 1],
        sequence_length: ranks.length,
        total_cards: total,
      },
      error: "",
    };
  }
  if (isStraightPairs(counts)) {
    return {
      combo: {
        type: "straight_pairs",
        main_rank: ranks[ranks.length - 1],
        sequence_length: ranks.length,
        total_cards: total,
      },
      error: "",
    };
  }
  const airplane = findAirplane(counts, total);
  if (airplane) {
    return { combo: airplane, error: "" };
  }
  if (total === 6) {
    const fourRank = rankWithCount(counts, 4);
    if (fourRank && countPattern(counts, [4, 1, 1])) {
      return {
        combo: {
          type: "four_with_two_single",
          main_rank: fourRank,
          sequence_length: 1,
          total_cards: total,
        },
        error: "",
      };
    }
  }
  if (total === 8) {
    const fourRank = rankWithCount(counts, 4);
    if (fourRank && countPattern(counts, [4, 2, 2])) {
      return {
        combo: {
          type: "four_with_two_pair",
          main_rank: fourRank,
          sequence_length: 1,
          total_cards: total,
        },
        error: "",
      };
    }
  }
  return {
    combo: null,
    error: "当前选择不是可出的合法牌型",
  };
}
