const root = document.documentElement;
const themeToggle = document.querySelector("#theme-toggle");
const themeLabel = document.querySelector("#theme-label");
const themeColor = document.querySelector("meta[name='theme-color']");
const progressNumber = document.querySelector("#progress-number");
const progressBar = document.querySelector("#progress-bar");
const progressTrack = document.querySelector("#progress-track");
const progressLabel = document.querySelector("#progress-label");
const phase = document.querySelector("#phase");
const bootState = document.querySelector("#boot-state");
const worldState = document.querySelector("#world-state");
const creatureCount = document.querySelector("#creature-count");
const speciesCount = document.querySelector("#species-count");
const creatureList = document.querySelector("#creature-list");
const creatureTemplate = document.querySelector("#creature-template");
const systemLog = document.querySelector("#system-log");
const eventCount = document.querySelector("#event-count");
const clearLog = document.querySelector("#clear-log");

const species = new Set();
let receivedEvents = 0;
let displayedProgress = 0;
let progressAnimation;

function preferredTheme() {
  const stored = localStorage.getItem("simulator-theme");
  if (stored === "light" || stored === "dark") return stored;
  return window.matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark";
}

function applyTheme(theme) {
  root.dataset.theme = theme;
  const isDark = theme === "dark";
  themeLabel.textContent = isDark ? "LIGHT MODE" : "DARK MODE";
  themeToggle.setAttribute("aria-label", isDark ? "Use light mode" : "Use dark mode");
  themeColor.setAttribute("content", isDark ? "#080b10" : "#e9ebe5");
}

applyTheme(preferredTheme());

themeToggle.addEventListener("click", () => {
  const nextTheme = root.dataset.theme === "dark" ? "light" : "dark";
  localStorage.setItem("simulator-theme", nextTheme);
  applyTheme(nextTheme);
});

clearLog.addEventListener("click", () => {
  systemLog.replaceChildren();
});

function formatCounter(value) {
  return String(value).padStart(3, "0");
}

function animateProgress(target) {
  cancelAnimationFrame(progressAnimation);
  const start = displayedProgress;
  const startedAt = performance.now();
  const duration = 420;

  function frame(now) {
    const elapsed = Math.min((now - startedAt) / duration, 1);
    const eased = 1 - Math.pow(1 - elapsed, 3);
    displayedProgress = Math.round(start + ((target - start) * eased));
    progressNumber.textContent = formatCounter(displayedProgress);
    if (elapsed < 1) progressAnimation = requestAnimationFrame(frame);
  }

  progressAnimation = requestAnimationFrame(frame);
  progressBar.style.width = `${target}%`;
  progressTrack.setAttribute("aria-valuenow", String(target));
}

function glyphFor(creature) {
  const glyphs = { dog: "D", human: "H", horse: "H", bird: "B" };
  return glyphs[creature.species] ?? creature.species.slice(0, 1).toUpperCase() ?? "?";
}

function readableCapability(capability) {
  return capability.replaceAll("_", " ");
}

function renderCreature(creature) {
  document.querySelector("#waiting-state")?.remove();

  const fragment = creatureTemplate.content.cloneNode(true);
  const card = fragment.querySelector(".creature-card");
  card.dataset.entityId = creature.id;
  fragment.querySelector(".entity-number").textContent = `#${String(creature.sequence).padStart(2, "0")}`;
  fragment.querySelector(".entity-avatar").textContent = glyphFor(creature);
  fragment.querySelector(".entity-name").textContent = creature.name || "Unnamed creature";
  fragment.querySelector(".entity-variant").textContent = creature.variant || "Unclassified variant";
  fragment.querySelector(".species-tag").textContent = creature.species;
  fragment.querySelector(".entity-id").textContent = creature.id;
  fragment.querySelector(".entity-age").textContent = String(creature.age);
  fragment.querySelector(".hunger-max").textContent = String(creature.hunger_max);
  fragment.querySelector(".hunger-rate").textContent = Number(creature.hunger_rate).toFixed(4);

  const capabilityList = fragment.querySelector(".capability-list");
  for (const capability of creature.capabilities) {
    const chip = document.createElement("span");
    chip.className = "capability";
    chip.textContent = readableCapability(capability);
    capabilityList.append(chip);
  }

  for (const previousCard of creatureList.querySelectorAll(".creature-card[open]")) {
    previousCard.open = false;
  }
  card.open = true;
  creatureList.append(fragment);
  card.scrollIntoView({ block: "nearest", behavior: "smooth" });
}

function appendLog(entry) {
  receivedEvents += 1;
  eventCount.textContent = `${formatCounter(receivedEvents)} EVENTS`;

  const line = document.createElement("div");
  line.className = "console-line";

  const time = document.createElement("time");
  time.textContent = entry.time;
  const level = document.createElement("span");
  level.className = `log-level level-${entry.level.toLowerCase()}`;
  level.textContent = entry.level;
  const scope = document.createElement("span");
  scope.className = "log-scope";
  scope.textContent = entry.scope;
  const message = document.createElement("p");
  message.textContent = entry.message;

  line.append(time, level, scope, message);
  systemLog.append(line);
  systemLog.scrollTo({ top: systemLog.scrollHeight, behavior: "smooth" });
}

function appendClientError(message) {
  appendLog({ time: new Date().toLocaleTimeString("en-GB", { hour12: false }) + ".---", level: "ERROR", scope: "CLIENT", message });
}

const stream = new EventSource("/api/world/initialize");

stream.addEventListener("status", ({ data }) => {
  const event = JSON.parse(data);
  phase.textContent = event.message;
});

stream.addEventListener("log", ({ data }) => {
  appendLog(JSON.parse(data));
});

stream.addEventListener("creature", ({ data }) => {
  const creature = JSON.parse(data);
  renderCreature(creature);
  species.add(creature.species);

  const percentage = Math.round((creature.sequence / creature.total) * 100);
  animateProgress(percentage);
  progressLabel.textContent = `${creature.sequence} / ${creature.total} ENTITIES`;
  creatureCount.textContent = String(creature.sequence).padStart(2, "0");
  speciesCount.textContent = String(species.size).padStart(2, "0");
  phase.textContent = `Synchronizing ${creature.name} · ${creature.species}`;
});

stream.addEventListener("complete", ({ data }) => {
  const event = JSON.parse(data);
  stream.close();
  animateProgress(100);
  progressLabel.textContent = `${event.count} / ${event.count} ENTITIES`;
  phase.textContent = "All systems nominal. World ready.";
  bootState.textContent = "COMPLETE";
  worldState.textContent = "ONLINE";
  worldState.classList.remove("pending");
  worldState.classList.add("ready");
  document.title = "World online · Simulator";
});

stream.addEventListener("world-error", ({ data }) => {
  const event = JSON.parse(data);
  stream.close();
  appendClientError(event.message);
  phase.textContent = event.message;
  bootState.textContent = "FAILED";
  worldState.textContent = "FAILED";
  worldState.classList.remove("pending");
  worldState.classList.add("failed");
});

stream.onerror = () => {
  if (worldState.classList.contains("ready") || worldState.classList.contains("failed")) return;
  stream.close();
  appendClientError("Connection to the simulation engine was interrupted");
  phase.textContent = "Connection interrupted";
  bootState.textContent = "HALTED";
  worldState.textContent = "OFFLINE";
  worldState.classList.remove("pending");
  worldState.classList.add("failed");
};
