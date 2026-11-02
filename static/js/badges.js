// Badges displayed under each guess, comparing it to the Pokémon of the day.
// "ok" = green (match), "wrong" = red, "neutral" = right type, wrong slot.

const POSITION_LABELS = ["BASIC", "LVL 1", "LVL 2"];

export function createBadge(text, type = "ok") {
  const span = document.createElement("span");
  span.className = `badge ${type}`;
  span.textContent = text;
  return span;
}

function typeBadge(typeName, match, wrongPlace) {
  if (wrongPlace) return createBadge(typeName, "neutral");
  return createBadge(typeName, match ? "ok" : "wrong");
}

function generationBadge(guessed, correct) {
  if (guessed === correct) return createBadge(`${guessed}G`, "ok");
  const arrow = guessed < correct ? ">" : "<";
  return createBadge(`${arrow}${guessed}G`, "wrong");
}

/** Weight / height hints start with ">" or "<" when they don't match. */
function measureBadge(hint) {
  const mismatch = hint.startsWith(">") || hint.startsWith("<");
  return createBadge(hint, mismatch ? "wrong" : "ok");
}

function positionBadge(guessPos, targetPos) {
  const text = POSITION_LABELS[guessPos] || "Unknown";
  return createBadge(text, guessPos === targetPos ? "ok" : "wrong");
}

function evolutionBadge(guessEvo, targetEvo) {
  const text = guessEvo === 1 ? "fully evolved" : "not fully evolved";
  return createBadge(text, guessEvo === targetEvo ? "ok" : "wrong");
}

export function createHintsElement(hints) {
  const container = document.createElement("div");
  container.className = "hints";
  container.append(
    typeBadge(hints.type1, hints.type1Match, hints.type1MatchWrongPlace),
    typeBadge(hints.type2, hints.type2Match, hints.type2MatchWrongPlace),
    generationBadge(hints.guessedGen, hints.correctGen),
    measureBadge(hints.weightHint),
    measureBadge(hints.heightHint),
    positionBadge(hints.guessPosition, hints.targetPosition),
    evolutionBadge(hints.guessFullyEvolved, hints.targetFullyEvolved),
  );
  return container;
}
