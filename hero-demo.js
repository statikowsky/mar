// First-screen demo: types real mar commands on the left and drives the
// matching browser screen on the right, on one looping, synchronized timeline.
// Respects prefers-reduced-motion (static transcript + static doc screen) and
// only starts once the demo scrolls into view.
(() => {
  const term = document.getElementById("hd-term");
  const screenRoot = document.getElementById("hd-screen");
  const addr = document.getElementById("hd-addr");
  if (!term || !screenRoot || !addr) return;

  const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

  // Single source of truth for both columns. Each beat: a typed command + its
  // short output (left), and the address + screen it drives (right).
  const beats = [
    { cmd: "mar init", out: "Initialized MAR store in .mar/", addr: "127.0.0.1:7777/docs", screen: "empty" },
    { cmd: 'mar doc create --code AUTH --title "Auth design" --type design --body design.md', out: "Created DOC-AUTH", addr: "127.0.0.1:7777/doc/DOC-AUTH", screen: "doc" },
    { cmd: "mar doc list", out: "DOC-AUTH   design   Auth design", addr: "127.0.0.1:7777/docs", screen: "docs" },
    { cmd: "mar board show", out: "To do\nIn progress\nDone", addr: "127.0.0.1:7777/board", screen: "board" },
    { cmd: 'mar task create --title "Wire up auth" --column "To do"', out: "Created T-WIRE-UP-AUTH", addr: "127.0.0.1:7777/board", screen: "board", action: "create" },
    { cmd: "mar task show T-WIRE-UP-AUTH", out: "T-WIRE-UP-AUTH  Wire up auth", addr: "127.0.0.1:7777/task/T-WIRE-UP-AUTH", screen: "task" },
  ];

  const screens = new Map(
    [...screenRoot.querySelectorAll(".scr")].map((el) => [el.dataset.screen, el]),
  );

  function showScreen(name) {
    for (const [key, el] of screens) el.classList.toggle("is-active", key === name);
  }

  function setAddr(text) { addr.textContent = text; }

  function makeLine() {
    const line = document.createElement("div");
    line.className = "hd-line";
    const prompt = document.createElement("span");
    prompt.className = "hd-prompt";
    prompt.textContent = "$";
    const cmd = document.createElement("span");
    cmd.className = "hd-cmd";
    line.append(prompt, document.createTextNode(" "), cmd);
    term.appendChild(line);
    term.scrollTop = term.scrollHeight;
    return cmd;
  }

  function addOutput(text) {
    if (!text) return;
    const out = document.createElement("pre");
    out.className = "hd-out";
    out.textContent = text;
    term.appendChild(out);
    term.scrollTop = term.scrollHeight;
  }

  function typeInto(el, text) {
    return new Promise((resolve) => {
      let i = 0;
      const tick = () => {
        el.textContent = text.slice(0, (i += 1));
        term.scrollTop = term.scrollHeight;
        if (i < text.length) setTimeout(tick, 18);
        else resolve();
      };
      tick();
    });
  }

  function insertCreatedCard() {
    const col = screenRoot.querySelector('[data-col="todo"]');
    if (!col || col.querySelector(".bcard--new")) return;
    const card = document.createElement("article");
    card.className = "bcard bcard--new";
    card.innerHTML = '<p class="bcode">T-WIRE-UP-AUTH</p><p class="btitle">Wire up auth</p>';
    col.appendChild(card);
  }

  function resetBoard() {
    screenRoot.querySelectorAll(".bcard--new").forEach((c) => c.remove());
  }

  function clearTerminal() { term.textContent = ""; }

  function renderStatic() {
    clearTerminal();
    resetBoard();
    insertCreatedCard();
    for (const b of beats) {
      makeLine().textContent = b.cmd;
      addOutput(b.out);
    }
    setAddr(beats[1].addr);
    showScreen("doc");
  }

  let running = false;
  async function runLoop() {
    if (running) return;
    running = true;
    // eslint-disable-next-line no-constant-condition
    while (true) {
      clearTerminal();
      resetBoard();
      const cursor = document.createElement("span");
      cursor.className = "hd-cursor";
      for (const b of beats) {
        const cmd = makeLine();
        cmd.after(cursor); // cursor trails the command being typed
        await typeInto(cmd, b.cmd);
        await sleep(150);
        setAddr(b.addr);
        if (b.action === "create") {
          showScreen("board");
          await sleep(120);
          insertCreatedCard();
        } else {
          showScreen(b.screen);
        }
        addOutput(b.out);
        await sleep(1150);
      }
      // Park the cursor on a fresh prompt during the inter-cycle pause.
      makeLine().after(cursor);
      await sleep(1600);
    }
  }

  showScreen("empty");

  if (reduceMotion) {
    renderStatic();
    return;
  }

  if (!("IntersectionObserver" in window)) {
    runLoop();
    return;
  }
  const io = new IntersectionObserver(
    (records) => {
      for (const r of records) {
        if (r.isIntersecting) {
          runLoop();
          io.disconnect();
        }
      }
    },
    { threshold: 0.3 },
  );
  io.observe(screenRoot);
})();
