// Thin wrappers around the Go backend endpoints.

async function postJSON(url, body) {
  const res = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  return res.json();
}

/** POST /api/guess -> { ok, error, correct, guess, hints, reveal, guessCounter } */
export const sendGuess = (guess) => postJSON("/api/guess", { guess });

/** POST /api/suggest -> [{ lang, names: [] }] */
export const fetchSuggestions = (query) => postJSON("/api/suggest", { query });

/** GET /api/hints -> { tier, cry?, types?, description? } */
export async function fetchHints() {
  const res = await fetch("/api/hints");
  return res.json();
}
