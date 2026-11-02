// Hint box: attempt counter + hints unlocked every 3 guesses
// (tier 1: cry, tier 2: types, tier 3: Pokédex description).
import { fetchHints } from "./api.js";
import { createCryPlayer } from "./cryPlayer.js";

const GUESSES_PER_TIER = 3;
const TIER_COUNT = 3;

function hintCard() {
  const card = document.createElement("div");
  card.className = "hint-card";
  return card;
}

function descriptionCard(descriptions) {
  const card = hintCard();
  const langs = Object.keys(descriptions);
  let index = Math.max(0, langs.indexOf("en"));

  const text = document.createElement("span");
  text.id = "pokedex-description";
  const render = () => {
    const lang = langs[index];
    text.textContent = `Pokedex (${lang}): "${descriptions[lang]}"`;
  };
  render();
  card.appendChild(text);

  if (langs.length > 1) {
    const button = document.createElement("button");
    button.className = "lang-toggle";
    const icon = document.createElement("img");
    icon.src = "./images/language.svg";
    icon.alt = "Change language";
    button.appendChild(icon);
    button.addEventListener("click", () => {
      index = (index + 1) % langs.length;
      render();
    });
    card.appendChild(button);
  }
  return card;
}

function typesCard(types) {
  const card = hintCard();
  for (const type of types) {
    const span = document.createElement("span");
    span.className = `badge ${type.toLowerCase()}`;
    span.textContent = type;
    card.appendChild(span);
  }
  return card;
}

export function createHintsPanel({ box, status, dynamic }) {
  function updateStatus(guessCounter) {
    if (guessCounter === 0) {
      status.textContent = "";
      return;
    }
    box.style.display = "block";

    const tier = Math.floor(guessCounter / GUESSES_PER_TIER);
    let text = `Attempt #${guessCounter}`;
    if (tier < TIER_COUNT) {
      const remaining = (tier + 1) * GUESSES_PER_TIER - guessCounter;
      text += `. ${remaining} more guess(es) before Hint #${tier + 1}.`;
    }
    status.textContent = text;
  }

  async function refresh() {
    const data = await fetchHints();
    dynamic.innerHTML = "";

    if (data.description) dynamic.appendChild(descriptionCard(data.description));
    if (data.types && data.types.length > 0) dynamic.appendChild(typesCard(data.types));
    if (data.cry) dynamic.appendChild(createCryPlayer(data.cry));
  }

  const hide = () => { box.style.display = "none"; };

  return { updateStatus, refresh, hide };
}
