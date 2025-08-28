const form = document.getElementById("guessForm");
const input = document.getElementById("guessInput");
const list = document.getElementById("guesses");
const statusEl = document.getElementById("status");
const counterEl = document.getElementById("counter");

const positionLabels = ["BASIC", "LVL 1", "LVL 2"];

function createBadge(text, type = "ok") {
  const span = document.createElement("span");
  span.className = `badge ${type}`;
  span.textContent = text;
  return span;
}

function createTypeBadges(typeName, match, wrongPlace) {
  const badge = createBadge(typeName, match ? "ok" : "wrong");
  if (wrongPlace) badge.className = "badge neutral";
  return badge;
}

function createGenBadge(guessedGen, correctGen) {
  let text = `${guessedGen}G`;
  let type = "ok";
  if (guessedGen !== correctGen) {
    type = "wrong";
    text = guessedGen < correctGen ? `>${text}` : `<${text}`;
  }
  return createBadge(text, type);
}

function createPositionBadge(guessPos, targetPos) {
  const text = positionLabels[guessPos] || "Unknown";
  const type = guessPos === targetPos ? "ok" : "wrong";
  return createBadge(text, type);
}

function createEvolutionBadge(guessEvo, targetEvo) {
  const text = guessEvo === 1 ? "fully evolved" : "not fully evolved";
  const type = guessEvo === targetEvo ? "ok" : "wrong";
  return createBadge(text, type);
}

function createHintsElement(hints) {
  const container = document.createElement("div");
  container.className = "hints";

  container.appendChild(createTypeBadges(hints.type1, hints.type1Match, hints.type1MatchWrongPlace));
  container.appendChild(createTypeBadges(hints.type2, hints.type2Match, hints.type2MatchWrongPlace));
  container.appendChild(createGenBadge(hints.guessedGen, hints.correctGen));

  const weightBadge = createBadge(hints.weightHint.startsWith(">") || hints.weightHint.startsWith("<") ? hints.weightHint : hints.weightHint, hints.weightHint.startsWith(">") || hints.weightHint.startsWith("<") ? "wrong" : "ok");
  const heightBadge = createBadge(hints.heightHint.startsWith(">") || hints.heightHint.startsWith("<") ? hints.heightHint : hints.heightHint, hints.heightHint.startsWith(">") || hints.heightHint.startsWith("<") ? "wrong" : "ok");
  container.appendChild(weightBadge);
  container.appendChild(heightBadge);

  container.appendChild(createPositionBadge(hints.guessPosition, hints.targetPosition));
  container.appendChild(createEvolutionBadge(hints.guessFullyEvolved, hints.targetFullyEvolved));

  return container;
}

function updateStatus(data) {
  const tier = Math.floor(data.guessCounter / 3);
  switch(tier) {
    case 0:
      statusEl.textContent = `Attempts : ${data.guessCounter}. First hint coming after ${3 - data.guessCounter} other attempts.`;
      break;
    case 1:
      statusEl.textContent = `Attempts : ${data.guessCounter}. Next hint coming after ${6 - data.guessCounter} other attempts.`;
      break;
    case 2:
      statusEl.textContent = `Attempts : ${data.guessCounter}. Last hint coming after ${9 - data.guessCounter} other attempts.`;
      break;
    default:
      statusEl.textContent = `Attempts : ${data.guessCounter}`;
  }
}

form.addEventListener("submit", async (e) => {
  e.preventDefault();
  const guess = input.value.trim();
  if (!guess) return;
  input.value = "";
  statusEl.textContent = "Checking...";

  try {
    const res = await fetch("/api/guess", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ guess }),
    });
    const data = await res.json();
    if (!data.ok) {
      statusEl.textContent = data.error || "Erreur.";
      return;
    }

    updateStatus(data);

    const li = document.createElement("li");
    li.className = "guess";

    const sprite = document.createElement("img");
    sprite.src = data.guess.sprite || "";
    sprite.alt = data.guess.name;
    li.appendChild(sprite);

    const info = document.createElement("div");
    const title = document.createElement("div");
    title.className = "name";
    title.textContent = `#${data.guess.id} — ${data.guess.name}`;
    info.appendChild(title);

    const hintsEl = createHintsElement(data.hints);
    info.appendChild(hintsEl);

    if (data.correct && data.reveal) {
      const rev = document.createElement("div");
      rev.className = "reveal";
      rev.textContent = `Congrats! The Pokémon of the day was #${data.reveal.id} — ${data.reveal.name}.`;
      info.appendChild(rev);

      input.disabled = true;
      const guessButton = form.querySelector("button");
      if (guessButton) guessButton.disabled = true;
      form.style.display = "none";
    }

    li.appendChild(info);
    list.prepend(li);
  } catch (err) {
    statusEl.textContent = "Network Error.";
    console.error(err);
  }
});
