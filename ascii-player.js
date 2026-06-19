const preview = window.ASCII_FIRST_FRAME;
const canvas = document.getElementById("ascii");
const loading = document.getElementById("loading");
const themeToggle = document.getElementById("theme-toggle");
const ctx = canvas.getContext("2d", { alpha: false });
const reduceMotionQuery = window.matchMedia("(prefers-reduced-motion: reduce)");

let colors = [];
let canvasBackground = "";
let frames = [Uint8Array.from(preview.frame, (char) => Number.parseInt(char, 16))];
let frameIndex = 0;
let lastTick = 0;
let cellWidth = 0;
let cellHeight = 0;
let fontSize = 7;
let isStarted = false;
let isDataReady = false;
let animation = {
  cols: preview.cols,
  rows: preview.rows,
  fps: 1,
  ramp: preview.ramp,
};

function readThemeColors() {
  const styles = getComputedStyle(document.documentElement);
  colors = [0, 1, 2, 3, 4, 5].map((index) => styles.getPropertyValue(`--ascii-${index}`).trim());
  canvasBackground = styles.getPropertyValue("--ascii-bg").trim();
}

function updateThemeToggle() {
  const isLight = document.documentElement.dataset.theme === "light";
  themeToggle.innerHTML = isLight
    ? `<svg viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
        <path d="M20.6 15.2a8.1 8.1 0 0 1-11.8-10 8.6 8.6 0 1 0 11.8 10Z"></path>
      </svg>`
    : `<svg viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="4"></circle>
        <path d="M12 2v2"></path>
        <path d="M12 20v2"></path>
        <path d="m4.93 4.93 1.41 1.41"></path>
        <path d="m17.66 17.66 1.41 1.41"></path>
        <path d="M2 12h2"></path>
        <path d="M20 12h2"></path>
        <path d="m6.34 17.66-1.41 1.41"></path>
        <path d="m19.07 4.93-1.41 1.41"></path>
      </svg>`;
  themeToggle.setAttribute("aria-label", `Switch to ${isLight ? "dark" : "light"} theme`);
  themeToggle.setAttribute("aria-pressed", String(isLight));
}

function setTheme(theme, persist = true) {
  document.documentElement.dataset.theme = theme;
  if (persist) {
    localStorage.setItem("mar-theme", theme);
  }
  readThemeColors();
  updateThemeToggle();
  buildGlyphSprites();
  render();
}

async function fetchFrameText(url) {
  const response = await fetch(url);
  const stream = response.body.pipeThrough(new DecompressionStream("deflate"));
  return new Response(stream).text();
}

function decodeFrames(encoded) {
  const decoded = [];
  let current = "";

  for (const frame of encoded) {
    if (frame.base) {
      current = frame.base;
    } else {
      const chars = current.split("");
      for (const [start, text] of frame.diff) {
        for (let offset = 0; offset < text.length; offset += 1) {
          chars[start + offset] = text[offset];
        }
      }
      current = chars.join("");
    }
    decoded.push(Uint8Array.from(current, (char) => Number.parseInt(char, 16)));
  }

  return decoded;
}

function prepareCanvas() {
  ctx.fillStyle = canvasBackground;
  ctx.fillRect(0, 0, canvas.width, canvas.height);
  ctx.textBaseline = "top";
  ctx.font = `${fontSize}px AsciiMono, monospace`;
  ctx.fontKerning = "none";
  ctx.fontVariantCaps = "normal";
  ctx.letterSpacing = "0px";
}

// Pre-rendered glyph atlas: one offscreen canvas per ramp level (the glyph in
// its tone colour). Each level maps to exactly one glyph and one colour, so the
// per-frame loop becomes cheap drawImage blits instead of thousands of fillText
// calls — the dominant cost in Safari. Rebuilt whenever the cell size (resize)
// or the colours (theme) change. Empty levels are stored as null and skipped.
let glyphSprites = [];

function buildGlyphSprites() {
  const { ramp } = animation;
  const rampLength = ramp.length;
  const toneCount = colors.length;
  if (!toneCount || !cellWidth) {
    glyphSprites = [];
    return;
  }

  // A square the size of the font comfortably holds any monospace glyph.
  const size = Math.max(1, Math.ceil(fontSize) + 1);
  const sprites = new Array(rampLength);

  for (let level = 0; level < rampLength; level += 1) {
    const char = ramp[level];
    if (char === " ") {
      sprites[level] = null;
      continue;
    }
    const sprite = document.createElement("canvas");
    sprite.width = size;
    sprite.height = size;
    const sctx = sprite.getContext("2d");
    sctx.textBaseline = "top";
    sctx.font = `${fontSize}px AsciiMono, monospace`;
    sctx.fontKerning = "none";
    sctx.letterSpacing = "0px";
    let tone = (level * toneCount / rampLength) | 0;
    if (tone >= toneCount) tone = toneCount - 1;
    sctx.fillStyle = colors[tone];
    sctx.fillText(char, 0, 0);
    sprites[level] = sprite;
  }

  glyphSprites = sprites;
}

function render() {
  if (!frames.length) {
    prepareCanvas();
    return;
  }

  const { cols, rows } = animation;
  const frame = frames[frameIndex % frames.length];
  prepareCanvas();

  if (glyphSprites.length !== animation.ramp.length) {
    buildGlyphSprites();
  }

  // Blit each cell's pre-rendered glyph. drawImage is far cheaper than fillText.
  for (let row = 0; row < rows; row += 1) {
    const rowOffset = row * cols;
    const y = Math.round(row * cellHeight);
    for (let col = 0; col < cols; col += 1) {
      const sprite = glyphSprites[frame[rowOffset + col]];
      if (sprite) ctx.drawImage(sprite, Math.round(col * cellWidth), y);
    }
  }
}

function fitToViewport() {
  const stage = canvas.parentElement.getBoundingClientRect();
  // Cap DPR: this canvas is a faint (0.16 opacity) ambient backdrop, so
  // rendering at full Retina resolution is wasted work — and glyph rasterizing
  // is the dominant cost in Safari.
  const dpr = Math.min(window.devicePixelRatio || 1, 1.5);
  canvas.width = Math.max(1, Math.ceil(stage.width * dpr));
  canvas.height = Math.max(1, Math.ceil(stage.height * dpr));
  canvas.style.width = `${stage.width}px`;
  canvas.style.height = `${stage.height}px`;

  ctx.font = "7px AsciiMono, monospace";
  const baseCellWidth = ctx.measureText("M").width;
  const baseCellHeight = 7 * 0.79;
  const cssScale = Math.max(
    stage.width / (animation.cols * baseCellWidth),
    stage.height / (animation.rows * baseCellHeight),
  );
  const drawScale = cssScale * dpr;
  fontSize = 7 * drawScale;
  cellWidth = baseCellWidth * drawScale;
  cellHeight = baseCellHeight * drawScale;
  document.documentElement.style.setProperty("--ascii-scale", cssScale);
  buildGlyphSprites();
  render();
}

function tick(now) {
  if (isStarted && !reduceMotionQuery.matches && now - lastTick >= 1000 / animation.fps) {
    frameIndex = (frameIndex + 1) % frames.length;
    lastTick = now;
    render();
  }
  requestAnimationFrame(tick);
}

function loadFullDataScript() {
  return new Promise((resolve, reject) => {
    const script = document.createElement("script");
    script.src = "shore1-ascii-data.js";
    script.onload = resolve;
    script.onerror = reject;
    document.body.appendChild(script);
  });
}

async function loadAnimationData() {
  await loadFullDataScript();
  const data = window.ASCII_ANIMATION_DATA;
  const text = await fetchFrameText(data.bin);
  frames = decodeFrames(JSON.parse(text));
  animation = {
    cols: data.cols,
    rows: data.rows,
    fps: data.fps,
    ramp: data.ramp,
  };
  frameIndex = 0;
  isDataReady = true;
  fitToViewport();
  startAnimation();
}

function startAnimation() {
  if (isStarted || !isDataReady) {
    return;
  }

  isStarted = true;
  lastTick = performance.now();
  window.clearInterval(window.loadingSpinnerTimer);
  fitToViewport();
  loading.classList.add("is-fading");
  loading.addEventListener("transitionend", () => {
    loading.hidden = true;
  }, { once: true });
}

readThemeColors();
updateThemeToggle();

document.fonts.ready.then(() => {
  fitToViewport();
  requestAnimationFrame(tick);
  loadAnimationData();
});

themeToggle.addEventListener("click", () => {
  const nextTheme = document.documentElement.dataset.theme === "light" ? "dark" : "light";
  setTheme(nextTheme);
});

window.matchMedia("(prefers-color-scheme: light)").addEventListener("change", (event) => {
  if (localStorage.getItem("mar-theme")) {
    return;
  }
  setTheme(event.matches ? "light" : "dark", false);
});

window.addEventListener("resize", fitToViewport);
