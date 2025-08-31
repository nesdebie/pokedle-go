import lottieWeb from 'https://cdn.skypack.dev/lottie-web';

const form = document.getElementById("guessForm");
const input = document.getElementById("guessInput");
const list = document.getElementById("guesses");
const statusEl = document.getElementById("status");
const statusHints = document.getElementById("hints-status");
const hintsDynamic = document.getElementById("hints-dynamic");
const hintsBox = document.getElementById("hints");


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

  const weightBadge = createBadge(
    hints.weightHint.startsWith(">") || hints.weightHint.startsWith("<") ? hints.weightHint : hints.weightHint,
    hints.weightHint.startsWith(">") || hints.weightHint.startsWith("<") ? "wrong" : "ok"
  );
  const heightBadge = createBadge(
    hints.heightHint.startsWith(">") || hints.heightHint.startsWith("<") ? hints.heightHint : hints.heightHint,
    hints.heightHint.startsWith(">") || hints.heightHint.startsWith("<") ? "wrong" : "ok"
  );
  container.appendChild(weightBadge);
  container.appendChild(heightBadge);

  container.appendChild(createPositionBadge(hints.guessPosition, hints.targetPosition));
  container.appendChild(createEvolutionBadge(hints.guessFullyEvolved, hints.targetFullyEvolved));

  return container;
}

function updateStatus(data) {
  if (data.guessCounter === 0) {
    statusHints.textContent = "";
    return;
  }
  hintsBox.style.display = "block";
  const tier = Math.floor(data.guessCounter / 3);
  switch (tier) {
    case 0:
      statusHints.textContent = `Attempt #${data.guessCounter}. ${3 - data.guessCounter} more guess(es) before Hint #1.`;
      break;
    case 1:
      statusHints.textContent = `Attempt #${data.guessCounter}. ${6 - data.guessCounter} more guess(es) before Hint #2.`;
      break;
    case 2:
      statusHints.textContent = `Attempt #${data.guessCounter}. ${9 - data.guessCounter} more guess(es) before Hint #3.`;
      break;
    default:
      statusHints.textContent = `Attempt #${data.guessCounter}`;
  }
}

async function updateHints() {
  const res = await fetch("/api/hints");
  const data = await res.json();

  hintsDynamic.innerHTML = "";

  if (data.description) {
    const container = document.createElement("div");
    container.id = "single-hint-container";
  
    const descKeys = Object.keys(data.description);
    let currentLang = "en";
    const content = document.createElement("span");
    content.textContent = `"${data.description[currentLang]}"`;
    container.appendChild(content);
  
    const langButtons = document.createElement("div");
    langButtons.className = "lang-buttons";
  
    descKeys.forEach(lang => {
      const btn = document.createElement("button");
      btn.textContent = lang.toUpperCase();
      btn.addEventListener("click", () => {
        currentLang = lang;

        content.textContent = `"${data.description[currentLang]}"`;
      });
      langButtons.appendChild(btn);
    });
  
    container.appendChild(langButtons);
    hintsDynamic.appendChild(container);
  }
  

  if (data.types && data.types.length > 0) {
    const container = document.createElement("div");
    container.id = "single-hint-container";
  
    data.types.forEach(type => {
      const span = document.createElement("span");
      span.className = `badge ${type.toLowerCase()}`;
      span.textContent = type;
      container.appendChild(span);
    });
  
    hintsDynamic.appendChild(container);
  }
  

  if (data.cry) {
    const container = document.createElement("div");
    container.id = "single-hint-container";

    const playIconContainer = document.createElement("button");
    playIconContainer.id = "play-icon";
    container.appendChild(playIconContainer);

    const currentTime = document.createElement("span");
    currentTime.id = "current-time";
    currentTime.className = "time";
    currentTime.textContent = "0:00";
    container.appendChild(currentTime);

    const seekSlider = document.createElement("input");
    seekSlider.type = "range";
    seekSlider.id = "seek-slider";
    seekSlider.max = "100";
    seekSlider.value = "0";
    container.appendChild(seekSlider);

    const audio = document.createElement("audio");
    audio.src = data.cry;
    audio.id = "cry-audio";
    container.appendChild(audio);

    hintsDynamic.appendChild(container);

    let playState = 'play';
    const playAnimation = lottieWeb.loadAnimation({
      container: playIconContainer,
      path: 'https://maxst.icons8.com/vue-static/landings/animated-icons/icons/pause/pause.json',
      renderer: 'svg',
      loop: false,
      autoplay: false,
      name: "Play Animation",
    });
    playAnimation.goToAndStop(14, true);

    playIconContainer.addEventListener('click', () => {
      if (playState === 'play') {
        audio.play();
        playAnimation.playSegments([14, 27], true);
        playState = 'pause';
      } else {
        audio.pause();
        playAnimation.playSegments([0, 14], true);
        playState = 'play';
      }
    });

    const showRangeProgress = (rangeInput) => {
      container.style.setProperty('--seek-before-width', rangeInput.value / rangeInput.max * 100 + '%');
    };

    seekSlider.addEventListener('input', (e) => {
      showRangeProgress(e.target);
      audio.currentTime = (audio.duration * seekSlider.value) / 100;
    });

    audio.addEventListener('timeupdate', () => {
      seekSlider.value = (audio.currentTime / audio.duration) * 100;
      currentTime.textContent = formatTime(audio.currentTime);
      duration.textContent = formatTime(audio.duration);
      showRangeProgress(seekSlider);
    });

    function formatTime(time) {
      if (isNaN(time)) return "0:00";
      const minutes = Math.floor(time / 60);
      const seconds = Math.floor(time % 60);
      return `${minutes}:${seconds < 10 ? "0" + seconds : seconds}`;
    }
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
    title.textContent = `${data.guess.name}`;
    info.appendChild(title);

    const hintsEl = createHintsElement(data.hints);
    info.appendChild(hintsEl);

    if (data.correct && data.reveal) {
      const rev = document.createElement("div");
      rev.className = "reveal";
      rev.textContent = `Congrats! The Pokémon of the day was ${data.guess.name}.`;
      info.appendChild(rev);

      input.disabled = true;
      const guessButton = form.querySelector("button");
      if (guessButton) guessButton.disabled = true;
      form.style.display = "none";

      hintsBox.style.display = "none";
    }

    li.appendChild(info);
    list.prepend(li);

    await updateHints();
  } catch (err) {
    statusEl.textContent = "Network Error.";
    console.error(err);
  }
  statusEl.textContent = "";
});
