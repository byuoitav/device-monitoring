window.components = window.components || {};

window.components.overview = {
    loadPage: async function() {
        bindOverviewButtons();
        bindOverviewRefreshButtons();
        bindSystemActionPopup();
        console.log("Overview component loaded");
        await loadOverviewDeviceInfo();
        await loadOverviewRoomPing();
    },

    cleanup: function() {
        console.log("Overview component cleaned up");
        // Additional cleanup code can go here
    }
}

function bindOverviewRefreshButtons() {
    const buttons = document.querySelectorAll('.component-container .card-refresh[data-refresh]');
    buttons.forEach((button) => {
        button.addEventListener('click', async () => {
            const target = button.dataset.refresh;
            if (target === "device-info") {
                document.querySelector('.component-container .system-info')?.classList.add('loading');
                await loadOverviewDeviceInfo();
            } else if (target === "room-status") {
                document.querySelector('.component-container .room-status')?.classList.add('loading');
                await loadOverviewRoomPing();
            }
        });
    });
}

async function loadOverviewRoomPing() {
    const roomStatus = document.querySelector('.component-container .room-status');
    try {
        const pingResults = await ApiService.getRoomPing();
        const counts = countRoomPingResults(pingResults);
        applyOverviewRoomStatus(counts);
    } catch (error) {
        console.error("Failed to load room ping:", error);
        applyOverviewRoomStatus({ reachable: 0, unreachable: 0 });
    } finally {
        roomStatus?.classList.remove('loading');
    }
}

function countRoomPingResults(pingResults) {
    const entries = pingResults instanceof Map
        ? Array.from(pingResults.entries())
        : Object.entries(pingResults ?? {});

    let reachable = 0;
    let unreachable = 0;

    for (const [, info] of entries) {
        const received = Number(info?.["packets-received"] ?? 0);
        if (received > 0) {
            reachable += 1;
        } else {
            unreachable += 1;
        }
    }

    return { reachable, unreachable };
}

function applyOverviewRoomStatus(counts) {
    const reachableEl = document.querySelector('.component-container [data-field="room.reachable"]');
    const unreachableEl = document.querySelector('.component-container [data-field="room.unreachable"]');
    if (reachableEl) reachableEl.textContent = counts.reachable ?? 0;
    if (unreachableEl) unreachableEl.textContent = counts.unreachable ?? 0;
}

function bindOverviewButtons() {
    const container = document.querySelector('.component-container');
    if (!container) return;

    if (container.dataset.buttonsBound === 'true') return;
    container.dataset.buttonsBound = 'true';

    container.addEventListener('click', async (event) => {
        const button = event.target.closest('.system-button[data-action]');
        if (!button) return;

        const action = button.dataset.action;
        if (typeof ApiService?.[action] !== "function") return;

        if (button.disabled) return;

        button.disabled = true;
        const commandLabel = getSystemActionLabel(button, action);
        showSystemActionPopup(commandLabel, "Sending...", "");
        try {
            const result = await ApiService[action]();
            const response = formatSystemActionResponse(result);
            const status = deriveSystemActionStatus(result);
            showSystemActionPopup(commandLabel, status, response);
        } catch (err) {
            console.error(`Failed to run ${action}:`, err);
            const response = formatSystemActionResponse(err?.message || err);
            showSystemActionPopup(commandLabel, "Failed", response);
        } finally {
            button.disabled = false;
        }
    });
}

function bindSystemActionPopup() {
    const popup = document.querySelector('.system-action-popup');
    if (!popup || popup.dataset.bound === 'true') return;
    popup.dataset.bound = 'true';

    const closeButton = popup.querySelector('.system-action-popup__close');
    const backdrop = popup.querySelector('.system-action-popup__backdrop');
    const content = popup.querySelector('.system-action-popup__content');

    const closePopup = () => {
        popup.classList.remove('is-visible');
        popup.setAttribute('aria-hidden', 'true');
    };

    const stopScrollDrag = (event) => {
        event.stopPropagation();
    };

    popup.addEventListener('pointerdown', stopScrollDrag);
    content?.addEventListener('pointerdown', stopScrollDrag);

    closeButton?.addEventListener('click', (event) => {
        event.stopPropagation();
        closePopup();
    });

    backdrop?.addEventListener('click', (event) => {
        event.stopPropagation();
        closePopup();
    });

    document.addEventListener('keydown', (event) => {
        if (event.key !== 'Escape') return;
        if (!popup.classList.contains('is-visible')) return;
        closePopup();
    });
}

function showSystemActionPopup(command, status, response) {
    const popup = document.querySelector('.system-action-popup');
    if (!popup) return;

    const commandEl = popup.querySelector('[data-field="system-action-command"]');
    const statusEl = popup.querySelector('[data-field="system-action-status"]');
    const responseEl = popup.querySelector('[data-field="system-action-response"]');
    if (commandEl) commandEl.textContent = command || "Unknown command";
    if (statusEl) statusEl.textContent = status || "Pending";
    if (responseEl) responseEl.textContent = response || "No response";

    popup.classList.add('is-visible');
    popup.setAttribute('aria-hidden', 'false');
}

function getSystemActionLabel(button, action) {
    const label = button.querySelector('p')?.textContent?.trim()
        || button.textContent?.trim()
        || action;
    return label;
}

function formatSystemActionResponse(response) {
    if (response === undefined || response === null || response === "") {
        return "No response";
    }
    if (typeof response === "string") {
        return response;
    }
    try {
        return JSON.stringify(response);
    } catch (error) {
        console.warn("Failed to stringify response:", error);
        return String(response);
    }
}

function deriveSystemActionStatus(result) {
    if (typeof result === "string" && result.toLowerCase().includes("success")) {
        return "Success";
    }
    return "Request sent";
}


async function loadOverviewDeviceInfo() {
    const container = document.querySelector('.component-container');
    if (!container) return;

    const systemInfo = container.querySelector('.system-info');
    try {
        const deviceInfo = await ApiService.getDeviceInfo();
        applyOverviewDeviceInfo(deviceInfo);
    } catch (error) {
        console.error("Failed to load overview device info:", error);
        applyOverviewDeviceInfo({});
    } finally {
        systemInfo?.classList.remove('loading');
    }
}

function applyOverviewDeviceInfo(deviceInfo) {
    const systemInfo = document.querySelector('.component-container .system-info');
    const fields = systemInfo?.querySelectorAll('[data-field]') ?? [];
    fields.forEach((el) => {
        const field = el.dataset.field;
        const value = getOverviewFieldValue(deviceInfo, field);
        el.textContent = formatOverviewFieldValue(field, value);
    });
}

function getOverviewFieldValue(deviceInfo, field) {
    if (!deviceInfo || !field) return undefined;
    return field.split('.').reduce((value, key) => value?.[key], deviceInfo);
}

function formatOverviewFieldValue(field, value) {
    if (field === "internet-connectivity") {
        return value ? "Connected" : "Disconnected";
    }
    if (typeof value === "boolean") {
        return value ? "Yes" : "No";
    }
    if (value === undefined || value === null || value === "") {
        return "Unknown";
    }
    return value;
}
