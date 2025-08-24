
const form = document.getElementById("guessForm");
const input = document.getElementById("guessInput");
const list = document.getElementById("guesses");
const statusEl = document.getElementById("status");

form.addEventListener("submit", async (e) => {
  e.preventDefault();
  const guess = input.value.trim();
  if (!guess) return;
  input.value = "";
  statusEl.textContent = "Vérification…";

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
    statusEl.textContent = "";

    const li = document.createElement("li");
    li.className = "guess";
    const sprite = document.createElement("img");
    sprite.src = data.guess.sprite || "";
    sprite.alt = data.guess.name;

    const info = document.createElement("div");
    const title = document.createElement("div");
    title.className = "name";
    title.textContent = `#${data.guess.id} — ${data.guess.name}`;
    info.appendChild(title);

    const hints = document.createElement("div");
    hints.className = "hints";
    const tm = document.createElement("span");
    tm.className = "badge " + (data.hints.typeMatch === 2 ? "ok" : data.hints.typeMatch === 1 ? "neutral" : "wrong");
    tm.textContent = `Shared types: ${data.hints.typeMatch}`;
    const idh = document.createElement("span");
    let idTxt = "same ID";
    if (data.hints.idHint < 0) idTxt = "id too low (#)";
    if (data.hints.idHint > 0) idTxt = "id too high(#)";
    idh.className = "badge neutral";
    idh.textContent = idTxt;

    const wh = document.createElement("span");
    wh.className = "badge neutral";
    wh.textContent = `Weight: ${data.hints.weightHint}`;

    const hh = document.createElement("span");
    hh.className = "badge neutral";
    hh.textContent = `Size: ${data.hints.heightHint}`;

    hints.appendChild(tm);
    hints.appendChild(idh);
    hints.appendChild(wh);
    hints.appendChild(hh);
    info.appendChild(hints);

    if (data.correct && data.reveal) {
      const rev = document.createElement("div");
      rev.className = "reveal";
      rev.textContent = `Congrats ! The Pokémon of the day was #${data.reveal.id} — ${data.reveal.name}.`;
      info.appendChild(rev);
    }

    li.appendChild(sprite);
    li.appendChild(info);
    list.prepend(li);
  } catch (e) {
    statusEl.textContent = "Network Error.";
  }
});
