// Small audio player (Lottie play/pause icon + seek bar) for the cry hint.

const PLAY_PAUSE_ANIMATION =
  "https://maxst.icons8.com/vue-static/landings/animated-icons/icons/pause/pause.json";

function formatTime(time) {
  if (isNaN(time)) return "0:00";
  const minutes = Math.floor(time / 60);
  const seconds = Math.floor(time % 60);
  return `${minutes}:${String(seconds).padStart(2, "0")}`;
}

export function createCryPlayer(src) {
  const container = document.createElement("div");
  container.className = "hint-card";

  const playButton = document.createElement("button");
  playButton.id = "play-icon";

  const currentTime = document.createElement("span");
  currentTime.className = "time";
  currentTime.textContent = "0:00";

  const seekSlider = document.createElement("input");
  seekSlider.type = "range";
  seekSlider.id = "seek-slider";
  seekSlider.max = "100";
  seekSlider.value = "0";

  const audio = document.createElement("audio");
  audio.id = "cry-audio";
  audio.src = src;

  container.append(playButton, currentTime, seekSlider, audio);

  const animation = window.lottie.loadAnimation({
    container: playButton,
    path: PLAY_PAUSE_ANIMATION,
    renderer: "svg",
    loop: false,
    autoplay: false,
    name: "Play Animation",
  });
  animation.goToAndStop(14, true);

  let playing = false;
  const showPlayIcon = () => animation.playSegments([0, 14], true);
  const showPauseIcon = () => animation.playSegments([14, 27], true);

  const showProgress = () => {
    const percent = (seekSlider.value / seekSlider.max) * 100;
    container.style.setProperty("--seek-before-width", `${percent}%`);
  };

  playButton.addEventListener("click", () => {
    if (!playing) {
      if (audio.currentTime === audio.duration) audio.currentTime = 0;
      audio.play();
      showPauseIcon();
    } else {
      audio.pause();
      showPlayIcon();
    }
    playing = !playing;
  });

  seekSlider.addEventListener("input", () => {
    showProgress();
    audio.currentTime = (audio.duration * seekSlider.value) / 100;
  });

  audio.addEventListener("timeupdate", () => {
    seekSlider.value = (audio.currentTime / audio.duration) * 100;
    currentTime.textContent = formatTime(audio.currentTime);
    showProgress();
  });

  audio.addEventListener("ended", () => {
    setTimeout(() => {
      seekSlider.value = 0;
      showProgress();
      currentTime.textContent = "0:00";
      showPlayIcon();
      playing = false;
      audio.currentTime = 0;
    }, 250);
  });

  audio.addEventListener("play", () => {
    if (audio.currentTime === 0) {
      seekSlider.value = 0;
      showProgress();
    }
  });

  return container;
}
