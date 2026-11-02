// Entry point: wires the guess form, suggestions and hints together.
import { sendGuess } from "./api.js";
import { createHintsElement } from "./badges.js";
import { createHintsPanel } from "./hints.js";
import { setupSuggestions } from "./suggestions.js";

const form = document.getElementById("guessForm");
const input = document.getElementById("guessInput");
const guessList = document.getElementById("guesses");
const statusEl = document.getElementById("status");

const suggestions = setupSuggestions(input);
const hints = createHintsPanel({
  box: document.getElementById("hints"),
  status: document.getElementById("hints-status"),
  dynamic: document.getElementById("hints-dynamic"),
});

function setStatus(text, isError = false) {
  statusEl.textContent = text;
  statusEl.style.color = isError ? "red" : "#e7ecff";
}

function createGuessItem(data) {
  const li = document.createElement("li");
  li.className = "guess";

  const sprite = document.createElement("img");
  sprite.src = data.guess.sprite || "";
  sprite.alt = data.guess.name;

  const info = document.createElement("div");
  const title = document.createElement("div");
  title.className = "name";
  title.textContent = data.guess.name;
  info.append(title, createHintsElement(data.hints));

  if (data.correct && data.reveal) {
    const reveal = document.createElement("div");
    reveal.className = "reveal";
    reveal.textContent = `Congrats! The Pokémon of the day was ${data.guess.name}.`;
    info.appendChild(reveal);
  }

  li.append(sprite, info);
  return li;
}

function endGame() {
  input.disabled = true;
  form.querySelector("button").disabled = true;
  form.style.display = "none";
  hints.hide();
}

form.addEventListener("submit", async (e) => {
  e.preventDefault();
  const guess = input.value.trim();
  if (!guess) return;

  input.value = "";
  suggestions.hide();
  setStatus("Checking...");

  try {
    const data = await sendGuess(guess);
    if (!data.ok) {
      setStatus(data.error || "Erreur.", true);
      return;
    }

    hints.updateStatus(data.guessCounter);
    guessList.prepend(createGuessItem(data));
    if (data.correct && data.reveal) endGame();

    setStatus("");
    await hints.refresh();
  } catch (err) {
    setStatus("Network Error.", true);
    console.error(err);
  }
});
