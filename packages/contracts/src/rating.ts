import { z } from "zod";

export const RATING_SCORES = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10] as const;

export const RatingScoreSchema = z.literal(RATING_SCORES);

export type RatingScore = z.infer<typeof RatingScoreSchema>;

export const RATING_CATEGORIES = {
  10: "Masterpiece",
  9: "Exceptional",
  8: "Amazing",
  7: "Great",
  6: "Good",
  5: "Decent",
  4: "Mediocre",
  3: "Bad",
  2: "Terrible",
  1: "Razzie",
} as const satisfies Record<RatingScore, string>;

export type RatingCategory = (typeof RATING_CATEGORIES)[RatingScore];

export function getRatingCategory(score: RatingScore): RatingCategory {
  return RATING_CATEGORIES[score];
}
