const state = {
  prices: {},
  stopped: false,
  toastTimer: null,
};

const elements = {
  balance: document.querySelector("#balance"),
  activeCount: document.querySelector("#active-count"),
  hiredTotal: document.querySelector("#hired-total"),
  hiredBreakdown: document.querySelector("#hired-breakdown"),
  runState: document.querySelector("#run-state"),
  lastUpdated: document.querySelector("#last-updated"),
  hireForm: document.querySelector("#hire-form"),
  minerClass: document.querySelector("#miner-class"),
  minerCount: document.querySelector("#miner-count"),
  minerPrices: document.querySelector("#miner-prices"),
  equipmentList: document.querySelector("#equipment-list"),
  activeMiners: document.querySelector("#active-miners"),
  notifications: document.querySelector("#notifications"),
  refreshButton: document.querySelector("#refresh-button"),
  shutdownButton: document.querySelector("#shutdown-button"),
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
  const titles = {
    pickaxe: "Кирка",
    ventilation: "Вентиляция",
    wagon: "Вагонетка",
  };
  return titles[type] || type;
}

function renderPrices() {
  const entries = Object.entries(state.prices);
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
  elements.activeCount.textContent = formatNumber(status.active_miners?.length || 0);
  elements.hiredTotal.textContent = formatNumber(total);
  elements.hiredBreakdown.textContent = `weak ${weak} / normal ${normal} / strong ${strong}`;
  elements.runState.textContent = state.stopped ? "Stopped" : "Running";
  elements.lastUpdated.textContent = new Date().toLocaleTimeString("ru-RU");

  renderActiveMiners(status.active_miners || []);
  renderNotifications(status.notifications || []);
}

function renderActiveMiners(miners) {
  if (miners.length === 0) {
    elements.activeMiners.innerHTML = '<div class="empty-state">Активных шахтеров пока нет</div>';
    return;
  }

  elements.activeMiners.innerHTML = miners
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
}

function renderNotifications(notifications) {
  if (notifications.length === 0) {
    elements.notifications.innerHTML = '<div class="empty-state">Уведомления появятся на балансных отметках</div>';
    return;
  }

  elements.notifications.innerHTML = notifications
    .slice()
    .reverse()
    .map((message) => `<div class="notification">${message}</div>`)
    .join("");
}

function renderEquipment(data) {
  const items = data.items || [];
  if (items.length === 0) {
    elements.equipmentList.innerHTML = '<div class="empty-state">Оборудование загружается</div>';
    return;
  }

  elements.equipmentList.innerHTML = items
    .map((item) => {
      const className = item.purchased ? "purchased" : item.can_buy_now ? "" : "locked";
      const label = item.purchased ? "Куплено" : item.can_buy_now ? "Купить" : "Недостаточно угля";
      return `
        <div class="equipment-item ${className}">
          <div>
            <strong>${equipmentTitle(item.type)}</strong>
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

async function refreshAll() {
  try {
    const [status, equipment] = await Promise.all([
      requestJSON("/enterprise/status"),
      requestJSON("/equipment"),
    ]);
    renderStatus(status);
    renderEquipment(equipment);
  } catch (error) {
    showToast(error.message, "error");
  }
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

async function shutdownEnterprise() {
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

elements.hireForm.addEventListener("submit", hireMiner);
elements.refreshButton.addEventListener("click", refreshAll);
elements.shutdownButton.addEventListener("click", shutdownEnterprise);
elements.equipmentList.addEventListener("click", (event) => {
  const button = event.target.closest("[data-buy]");
  if (!button) return;
  buyEquipment(button.dataset.buy);
});

renderPrices();
refreshAll();
loadPrices();
window.setInterval(refreshAll, 2000);
