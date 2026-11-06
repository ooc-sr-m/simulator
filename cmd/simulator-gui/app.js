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
const bootScreen = document.querySelector("#boot-screen");
const worldScreen = document.querySelector("#world-screen");
const launchPanel = document.querySelector("#launch-panel");
const startWorld = document.querySelector("#start-world");
const runtimeCreatureList = document.querySelector("#runtime-creature-list");
const runtimeCreatureCount = document.querySelector("#runtime-creature-count");
const selectedSpecies = document.querySelector("#selected-species");
const selectedAvatar = document.querySelector("#selected-avatar");
const selectedName = document.querySelector("#selected-name");
const selectedVariant = document.querySelector("#selected-variant");
const selectedID = document.querySelector("#selected-id");
const selectedAge = document.querySelector("#selected-age");
const selectedHungerMax = document.querySelector("#selected-hunger-max");
const selectedHungerRate = document.querySelector("#selected-hunger-rate");
const selectedCapabilities = document.querySelector("#selected-capabilities");
const logCreatureName = document.querySelector("#log-creature-name");
const creatureLog = document.querySelector("#creature-log");
const creatureEventCount = document.querySelector("#creature-event-count");
const clearCreatureLog = document.querySelector("#clear-creature-log");

const species = new Set();
const creatures = [];
const creaturesByID = new Map();
const logsByCreature = new Map();
let receivedEvents = 0;
let displayedProgress = 0;
let progressAnimation;
let selectedCreatureID = "";
let creatureStream;

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

clearCreatureLog.addEventListener("click", () => {
  if (!selectedCreatureID) return;
  logsByCreature.set(selectedCreatureID, []);
  renderSelectedCreatureLogs();
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

function createConsoleLine(entry) {
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
  return line;
}

function appendLog(entry) {
  receivedEvents += 1;
  eventCount.textContent = `${formatCounter(receivedEvents)} EVENTS`;

  const line = createConsoleLine(entry);
  systemLog.append(line);
  systemLog.scrollTo({ top: systemLog.scrollHeight, behavior: "smooth" });
}

function appendClientError(message) {
  appendLog({ time: new Date().toLocaleTimeString("en-GB", { hour12: false }) + ".---", level: "ERROR", scope: "CLIENT", message });
}

function createCapabilityChip(capability) {
  const chip = document.createElement("span");
  chip.className = "capability";
  chip.textContent = readableCapability(capability);
  return chip;
}

function renderRuntimeRoster() {
  runtimeCreatureList.replaceChildren();
  runtimeCreatureCount.textContent = String(creatures.length).padStart(2, "0");

  for (const creature of creatures) {
    const button = document.createElement("button");
    button.className = "runtime-creature-item";
    button.type = "button";
    button.dataset.entityId = creature.id;
    button.setAttribute("role", "option");
    button.setAttribute("aria-selected", "false");

    const avatar = document.createElement("span");
    avatar.className = "runtime-entity-avatar";
    avatar.textContent = glyphFor(creature);

    const identity = document.createElement("span");
    identity.className = "runtime-entity-heading";
    const name = document.createElement("strong");
    name.textContent = creature.name || "Unnamed creature";
    const variant = document.createElement("small");
    variant.textContent = `${creature.species} · ${creature.variant || "unclassified"}`;
    identity.append(name, variant);

    const activity = document.createElement("span");
    activity.className = "entity-activity";
    activity.title = "Waiting for activity";
    activity.setAttribute("aria-label", "Waiting for activity");

    button.append(avatar, identity, activity);
    button.addEventListener("click", () => selectCreature(creature.id));
    runtimeCreatureList.append(button);
  }
}

function selectCreature(creatureID) {
  const creature = creaturesByID.get(creatureID);
  if (!creature) return;

  selectedCreatureID = creatureID;
  for (const item of runtimeCreatureList.querySelectorAll(".runtime-creature-item")) {
    const isSelected = item.dataset.entityId === creatureID;
    item.classList.toggle("selected", isSelected);
    item.setAttribute("aria-selected", String(isSelected));
  }

  selectedSpecies.textContent = creature.species;
  selectedAvatar.textContent = glyphFor(creature);
  selectedName.textContent = creature.name || "Unnamed creature";
  selectedVariant.textContent = creature.variant || "Unclassified variant";
  selectedID.textContent = creature.id;
  selectedAge.textContent = String(creature.age);
  selectedHungerMax.textContent = String(creature.hunger_max);
  selectedHungerRate.textContent = Number(creature.hunger_rate).toFixed(4);
  logCreatureName.textContent = creature.name || "Unnamed creature";

  selectedCapabilities.replaceChildren(...creature.capabilities.map(createCapabilityChip));
  renderSelectedCreatureLogs();
}

function renderSelectedCreatureLogs() {
  creatureLog.replaceChildren();
  const logs = logsByCreature.get(selectedCreatureID) ?? [];
  creatureEventCount.textContent = `${formatCounter(logs.length)} EVENTS`;

  if (logs.length === 0) {
    const empty = document.createElement("div");
    empty.className = "console-empty";
    const icon = document.createElement("span");
    icon.setAttribute("aria-hidden", "true");
    icon.textContent = "⌁";
    const title = document.createElement("strong");
    title.textContent = "WAITING FOR CREATURE ACTIVITY";
    const hint = document.createElement("small");
    hint.textContent = "Updates will arrive once the selected entity completes its next tick.";
    empty.append(icon, title, hint);
    creatureLog.append(empty);
    return;
  }

  creatureLog.append(...logs.map(createConsoleLine));
  creatureLog.scrollTop = creatureLog.scrollHeight;
}

function appendCreatureLog(entry) {
  const logs = logsByCreature.get(entry.creature_id) ?? [];
  logs.push(entry);
  if (logs.length > 250) logs.splice(0, logs.length - 250);
  logsByCreature.set(entry.creature_id, logs);

  const item = runtimeCreatureList.querySelector(`[data-entity-id="${CSS.escape(entry.creature_id)}"]`);
  const activity = item?.querySelector(".entity-activity");
  if (activity) {
    activity.classList.remove("pulse");
    requestAnimationFrame(() => activity.classList.add("pulse"));
    activity.title = `Last update ${entry.time}`;
    activity.setAttribute("aria-label", `Last update ${entry.time}`);
  }

  if (entry.creature_id !== selectedCreatureID) return;
  creatureLog.querySelector(".console-empty")?.remove();
  creatureLog.append(createConsoleLine(entry));
  creatureEventCount.textContent = `${formatCounter(logs.length)} EVENTS`;
  creatureLog.scrollTo({ top: creatureLog.scrollHeight, behavior: "smooth" });
}

function openCreatureStream() {
  if (creatureStream) creatureStream.close();
  creatureStream = new EventSource("/api/world/logs");
  creatureStream.addEventListener("creature-log", ({ data }) => {
    appendCreatureLog(JSON.parse(data));
  });
  creatureStream.onerror = () => {
    if (creatureStream.readyState === EventSource.CLOSED) return;
    creatureEventCount.textContent = "RECONNECTING";
  };
}

function enterWorld() {
  renderRuntimeRoster();
  bootScreen.hidden = true;
  worldScreen.hidden = false;
  if (creatures.length > 0) selectCreature(creatures[0].id);
  document.title = "World online · Simulator";
  if (window.location.pathname !== "/world") history.pushState({ view: "world" }, "", "/world");
  openCreatureStream();
  window.scrollTo({ top: 0, behavior: "instant" });
}

startWorld.addEventListener("click", async () => {
  startWorld.disabled = true;
  startWorld.classList.add("loading");
  startWorld.querySelector("span").textContent = "STARTING WORLD";

  try {
    const response = await fetch("/api/world/start", { method: "POST" });
    if (!response.ok) throw new Error((await response.text()).trim() || "The simulation could not be started");
    enterWorld();
  } catch (error) {
    startWorld.disabled = false;
    startWorld.classList.remove("loading");
    startWorld.querySelector("span").textContent = "START WORLD";
    appendClientError(error.message);
  }
});

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
  if (!creaturesByID.has(creature.id)) creatures.push(creature);
  creaturesByID.set(creature.id, creature);
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
  launchPanel.hidden = false;
  startWorld.disabled = false;
  document.title = "World ready · Simulator";

  if (event.running && window.location.pathname === "/world") enterWorld();
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
