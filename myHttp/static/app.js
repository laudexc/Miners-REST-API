const state = {
  prices: {},
  equipmentTitles: {},
  stopped: false,
  toastTimer: null,
  detailsLoading: false,
  goalPromptShown: false,
};

const minerClassOrder = ["weak", "normal", "strong"];
const activePreviewLimit = 60;

const elements = {
  balance: document.querySelector("#balance"),
  activeCount: document.querySelector("#active-count"),
  hiredTotal: document.querySelector("#hired-total"),
  hiredBreakdown: document.querySelector("#hired-breakdown"),
  goalProgress: document.querySelector("#goal-progress"),
  goalNote: document.querySelector("#goal-note"),
  hireForm: document.querySelector("#hire-form"),
  minerClass: document.querySelector("#miner-class"),
  minerCount: document.querySelector("#miner-count"),
  minerPrices: document.querySelector("#miner-prices"),
  equipmentList: document.querySelector("#equipment-list"),
  activeMiners: document.querySelector("#active-miners"),
  refreshButton: document.querySelector("#refresh-button"),
  startButton: document.querySelector("#start-button"),
  shutdownButton: document.querySelector("#shutdown-button"),
  appCloseButton: document.querySelector("#app-close-button"),
  toast: document.querySelector("#toast"),
};

async function requestJSON(url, options = {}) {
  const response = await fetch(url, {
    headers: { Accept: "application/json" },
    ...options,
  });
  const data = await response.json().catch(() => ({}));

  if (!response.ok) {
    throw new Error(data.error || `HTTP ${response.status}`);
  }

  return data;
}

function formatNumber(value) {
  return new Intl.NumberFormat("ru-RU").format(value || 0);
}

function showToast(message, type = "info") {
  window.clearTimeout(state.toastTimer);
  elements.toast.textContent = message;
  elements.toast.classList.toggle("error", type === "error");
  elements.toast.classList.add("visible");

  state.toastTimer = window.setTimeout(() => {
    elements.toast.classList.remove("visible");
  }, 2800);
}

function minerClassTitle(className) {
  const titles = {
    weak: "Weak",
    normal: "Normal",
    strong: "Strong",
  };
  return titles[className] || className;
}

function equipmentTitle(type) {
  return state.equipmentTitles[type] || type;
}

function renderPrices() {
  const entries = minerClassOrder
    .filter((className) => state.prices[className])
    .map((className) => [className, state.prices[className]]);
  if (entries.length === 0) {
    elements.minerPrices.innerHTML = '<div class="empty-state">Цены шахтеров загружаются</div>';
    return;
  }

  elements.minerPrices.innerHTML = entries
    .map(([className, profile]) => {
      return `
        <div class="price-row">
          <strong>${minerClassTitle(className)}</strong>
          <span>Цена ${profile.Cost}, энергия ${profile.Energy}, добыча ${profile.CoalPerMine}, интервал ${profile.IntervalSec}s</span>
        </div>
      `;
    })
    .join("");
}

function renderStatus(status) {
  const hired = status.hired_stats || {};
  const weak = hired.weak || 0;
  const normal = hired.normal || 0;
  const strong = hired.strong || 0;
  const total = weak + normal + strong;

  elements.balance.textContent = formatNumber(status.balance);
  elements.activeCount.textContent = formatNumber(status.active_count || 0);
  elements.hiredTotal.textContent = formatNumber(total);
  elements.hiredBreakdown.textContent = `weak ${weak} / normal ${normal} / strong ${strong}`;
  state.stopped = state.stopped || Boolean(status.is_shutdown);
  elements.goalProgress.textContent = `${status.goal_progress || 0}/${status.goal_total || 0}`;
  elements.goalNote.textContent = status.goal_complete
    ? "все цели закрыты"
    : `следующая: ${status.next_goal_title || "оборудование"} (${formatNumber(status.next_goal_price)} угля)`;
  elements.startButton.disabled = !state.stopped;
  elements.shutdownButton.disabled = state.stopped;

  if ("active_preview" in status) {
    renderActiveMiners(status.active_preview || [], status.active_count || 0);
  }

  maybeOfferRunFinish(status);
}

function renderActiveMiners(miners, total) {
  if (miners.length === 0) {
    elements.activeMiners.innerHTML = '<div class="empty-state">Активных шахтеров пока нет</div>';
    return;
  }

  const hiddenCount = Math.max(0, total - miners.length);
  const rows = miners
    .map((miner) => {
      return `
        <div class="miner-row">
          <strong>#${miner.id} ${minerClassTitle(miner.class)}</strong>
          <span>${miner.is_working ? "работает" : "остановлен"}</span>
          <div class="miner-meta">
            <span class="pill">энергия ${miner.energy}</span>
            <span class="pill">добыча ${miner.coal_per_mining}</span>
          </div>
        </div>
      `;
    })
    .join("");

  const hiddenMessage = hiddenCount > 0
    ? `<div class="empty-state">Показаны первые ${miners.length} из ${formatNumber(total)} активных шахтеров</div>`
    : "";

  elements.activeMiners.innerHTML = rows + hiddenMessage;
}

function renderEquipment(data) {
  const items = (data.items || []).slice().sort((a, b) => {
    return (a.order || 0) - (b.order || 0);
  });
  if (items.length === 0) {
    elements.equipmentList.innerHTML = '<div class="empty-state">Оборудование загружается</div>';
    return;
  }

  state.equipmentTitles = {};
  items.forEach((item) => {
    state.equipmentTitles[item.type] = item.title;
  });

  elements.equipmentList.innerHTML = items
    .map((item) => {
      const className = item.purchased ? "purchased" : item.can_buy_now ? "" : "locked";
      const label = item.purchased
        ? "Куплено"
        : item.can_buy_now
          ? "Купить"
          : item.is_next_goal
            ? "Недостаточно угля"
            : "Ждет очереди";
      return `
        <div class="equipment-item ${className}">
          <div>
            <strong>${item.order}. ${item.title}</strong>
            <span>${item.description}</span>
            <span>Цена ${formatNumber(item.price)} угля</span>
          </div>
          <button type="button" data-buy="${item.type}" ${item.purchased || !item.can_buy_now || state.stopped ? "disabled" : ""}>
            ${label}
          </button>
        </div>
      `;
    })
    .join("");
}

function maybeOfferRunFinish(status) {
  if (!status.goal_complete || state.stopped || state.goalPromptShown) {
    return;
  }

  state.goalPromptShown = true;
  window.setTimeout(() => {
    const shouldStop = window.confirm("Все цели предприятия выполнены. Остановить предприятие и зафиксировать финальный результат?");
    if (shouldStop) {
      shutdownEnterprise({ skipConfirm: true });
    }
  }, 100);
}

async function refreshAll() {
  try {
    const [summary, active, equipment] = await Promise.all([
      requestJSON("/enterprise/summary"),
      requestJSON(`/miners/active?limit=${activePreviewLimit}`),
      requestJSON("/equipment"),
    ]);
    renderStatus({ ...summary, active_preview: active });
    renderEquipment(equipment);
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function refreshDetails(summary) {
  if (state.detailsLoading) return;
  state.detailsLoading = true;
  try {
    const [active, equipment] = await Promise.all([
      requestJSON(`/miners/active?limit=${activePreviewLimit}`),
      requestJSON("/equipment"),
    ]);
    renderStatus({ ...summary, active_preview: active });
    renderEquipment(equipment);
  } catch (error) {
    showToast(error.message, "error");
  } finally {
    state.detailsLoading = false;
  }
}

function connectEvents() {
  if (!window.EventSource) {
    window.setInterval(refreshAll, 2000);
    return;
  }

  const source = new EventSource("/events");
  source.addEventListener("summary", (event) => {
    const summary = JSON.parse(event.data);
    renderStatus(summary);
    refreshDetails(summary);
  });
  source.onerror = () => {
    source.close();
    showToast("SSE connection lost, switched to periodic refresh", "error");
    window.setInterval(refreshAll, 2000);
  };
}

async function loadPrices() {
  try {
    state.prices = await requestJSON("/miners/prices");
    renderPrices();
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function hireMiner(event) {
  event.preventDefault();
  const className = elements.minerClass.value;
  const count = Math.max(1, Number.parseInt(elements.minerCount.value, 10) || 1);

  try {
    const result = await requestJSON(`/miners/hire?class=${encodeURIComponent(className)}&count=${count}`, {
      method: "POST",
    });
    showToast(`Нанято шахтеров: ${result.miners?.length || count}`);
    await refreshAll();
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function buyEquipment(type) {
  try {
    await requestJSON(`/equipment/${encodeURIComponent(type)}/buy`, { method: "POST" });
    showToast(`${equipmentTitle(type)} куплено`);
    await refreshAll();
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function shutdownEnterprise(options = {}) {
  if (!options.skipConfirm && !window.confirm("Остановить предприятие? Добыча будет остановлена, но приложение продолжит работать.")) {
    return;
  }

  try {
    const result = await requestJSON("/enterprise/shutdown", { method: "POST" });
    state.stopped = true;
    elements.shutdownButton.disabled = true;
    showToast(`Предприятие остановлено. Финальный баланс: ${formatNumber(result.final_balance)}`);
    await refreshAll();
  } catch (error) {
    if (error.message.includes("already stopped")) {
      state.stopped = true;
      elements.shutdownButton.disabled = true;
    }
    showToast(error.message, "error");
    await refreshAll();
  }
}

async function startEnterprise() {
  try {
    const summary = await requestJSON("/enterprise/start", { method: "POST" });
    state.stopped = false;
    state.goalPromptShown = false;
    showToast("Предприятие запущено заново");
    renderStatus({ ...summary, active_preview: [] });
    await refreshAll();
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function closeApplication() {
  if (!window.confirm("Завершить приложение? HTTP API остановится, а страница потеряет соединение.")) {
    return;
  }

  try {
    await requestJSON("/app/close", { method: "PUT" });
    showToast("Приложение корректно завершает работу");
    elements.appCloseButton.disabled = true;
    elements.shutdownButton.disabled = true;
    elements.startButton.disabled = true;
  } catch (error) {
    showToast(error.message, "error");
  }
}

elements.hireForm.addEventListener("submit", hireMiner);
elements.refreshButton.addEventListener("click", refreshAll);
elements.startButton.addEventListener("click", startEnterprise);
elements.shutdownButton.addEventListener("click", shutdownEnterprise);
elements.appCloseButton.addEventListener("click", closeApplication);
elements.equipmentList.addEventListener("click", (event) => {
  const button = event.target.closest("[data-buy]");
  if (!button) return;
  buyEquipment(button.dataset.buy);
});

renderPrices();
refreshAll();
loadPrices();
connectEvents();
