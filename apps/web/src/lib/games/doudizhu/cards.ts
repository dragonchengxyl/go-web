import type { DoudizhuCard } from "@/lib/api-client";

export function cardKey(card: DoudizhuCard): string {
  return `${card.suit}-${card.rank}`;
}

export function cardSuitSymbol(card: DoudizhuCard): string {
  switch (card.suit) {
    case "spade":
      return "♠";
    case "heart":
      return "♥";
    case "club":
      return "♣";
    case "diamond":
      return "♦";
    default:
      return "J";
  }
}

export function cardRankLabel(card: DoudizhuCard): string {
  const rankMap: Record<number, string> = {
    11: "J",
    12: "Q",
    13: "K",
    14: "A",
    15: "2",
    16: "SJ",
    17: "BJ",
  };
  return rankMap[card.rank] ?? String(card.rank);
}

export function isRedCard(card: DoudizhuCard): boolean {
  return (
    card.suit === "heart" ||
    card.suit === "diamond" ||
    card.suit === "joker"
  );
}

export function cardLabel(card: DoudizhuCard): string {
  return `${cardSuitSymbol(card)}${cardRankLabel(card)}`;
}

export function sortedSelection(cards: DoudizhuCard[]): DoudizhuCard[] {
  return [...cards].sort((a, b) => {
    if (a.rank === b.rank) {
      return a.suit.localeCompare(b.suit);
    }
    return a.rank - b.rank;
  });
}
