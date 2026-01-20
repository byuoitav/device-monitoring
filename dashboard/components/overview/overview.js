window.components = window.components || {};

window.components.overview = {
    loadPage: async function() {
        bindOverviewButtons();
        bindOverviewRefreshButtons();
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
        try {
            await ApiService[action]();
        } catch (err) {
            console.error(`Failed to run ${action}:`, err);
        } finally {
            button.disabled = false;
        }
    });
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
