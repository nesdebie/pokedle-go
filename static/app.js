
const form = document.getElementById("guessForm");
const input = document.getElementById("guessInput");
const list = document.getElementById("guesses");
const statusEl = document.getElementById("status");
const counterEl = document.getElementById("counter");

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
  
    bigHintTier = Math.floor(data.guessCounter / 3);
    switch (bigHintTier) {
      case 0:
        statusEl.textContent = `Attempts : ${data.guessCounter}. First hint coming after ${3 - data.guessCounter} other attempts.`;
        break;
      case 1:
        statusEl.textContent = `Attempts : ${data.guessCounter}. Next hint coming after ${6 - data.guessCounter} other attempts.`;
        break;
      case 2:
        statusEl.textContent = `Attempts : ${data.guessCounter}. Last hint coming after ${9 - data.guessCounter} other attempts.`;          break;
      default:
        statusEl.textContent = `Attempts : ${data.guessCounter}`;
    }
  
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

    const t1 = document.createElement("span");
    t1.className = "badge " + (data.hints.type1Match ? "ok" : "wrong");
    if (data.hints.type1MatchWrongPlace) t1.className = "badge neutral"
    t1.textContent = `${data.hints.type1}`;
    
    const t2 = document.createElement("span");
    t2.className = "badge " + (data.hints.type2Match ? "ok" : "wrong");
    if (data.hints.type2MatchWrongPlace) t2.className = "badge neutral"
    t2.textContent = `${data.hints.type2}`;
  
    const idh = document.createElement("span");
    const guessedGen = data.hints.guessedGen;
    const correctGen = data.hints.correctGen;
    idh.className = "badge ok";
    let idTxt = `${guessedGen}G`;
    if (guessedGen !== correctGen) {
      idh.className = "badge wrong";
      if (guessedGen < correctGen) {
        idTxt = ">" + idTxt;
      } else {
        idTxt = "<" + idTxt;
      }
    }
    idh.textContent = idTxt;
    
    const wh = document.createElement("span");
    wh.className = "badge ok";
    wh.textContent = `${data.hints.weightHint}`;
    if (wh.textContent.startsWith(">") || wh.textContent.startsWith("<"))
      wh.className = "badge wrong";

    const hh = document.createElement("span");
    hh.className = "badge ok";
    hh.textContent = `${data.hints.heightHint}`;
    if (hh.textContent.startsWith(">") || hh.textContent.startsWith("<"))
      hh.className = "badge wrong";
  
    hints.appendChild(t1);
    hints.appendChild(t2);
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
