window.components = window.components || {};

window.components.deviceInfo = {
    loadPage: function() {
        console.log("Device Info component loaded");
        loadDeviceInfo();
    },

    cleanup: function() {
        console.log("Device Info component cleaned up");
    }
}

async function loadDeviceInfo() {
    const loadingCards = document.querySelectorAll('.component-container .device-info-card[data-loading="true"]');
    loadingCards.forEach((card) => card.classList.add('loading'));

    try {
        const data = await ApiService.getHardwareInfo();
        applyHardwareInfo(data);
    } catch (error) {
        console.error("Failed to load hardware info:", error);
        applyHardwareInfo({});
    } finally {
        loadingCards.forEach((card) => card.classList.remove('loading'));
        initializeDeviceInfoDrawers();
    }
}

function applyHardwareInfo(data) {
    const root = document.querySelector('.component-container');
    if (!root) return;

    const fields = root.querySelectorAll('[data-field]');
    fields.forEach((el) => {
        const field = el.dataset.field;
        el.textContent = formatHardwareField(field, getHardwareValue(data, field));
    });

    renderCpuTemperatures(root, data?.host?.temperature);
    renderCpuUsage(root, data?.cpu?.usage);
    updateUsageBars(root, data);
    updateCpuTemperatureIndicator(root, data?.host?.temperature);
}

function initializeDeviceInfoDrawers() {
    const cards = document.querySelectorAll('.component-container .device-info-card');

    cards.forEach((card) => {
        const drawer = card.querySelector('.drawer');
        if (!drawer) return;

        drawer.style.maxHeight = '0px';

        card.addEventListener('click', () => {
            const isOpen = card.classList.toggle('drawer-open');
            drawer.style.maxHeight = isOpen ? `${drawer.scrollHeight}px` : '0px';
        });
    });
}

function getHardwareValue(data, field) {
    if (!field) return undefined;
    if (field === 'disk.usage.summary') {
        return data?.disk?.usage;
    }
    if (field === 'disk.io.writeCount') {
        const counters = data?.disk?.['io-counters'];
        const firstDevice = counters ? Object.values(counters)[0] : undefined;
        return firstDevice?.writeCount;
    }
    return field.split('.').reduce((value, key) => value?.[key], data);
}

function formatHardwareField(field, value) {
    if (field === 'host.os.uptime') {
        return formatUptimeHours(value);
    }
    if (field === 'memory.swap.usedPercent' || field === 'memory.virtual.usedPercent' || field === 'disk.usage.usedPercent' || field === 'cpu.usage.avg') {
        return formatPercent(value);
    }
    if (field.startsWith('memory.') || field === 'disk.usage.free' || field === 'disk.usage.total' || field === 'disk.usage.used') {
        return formatBytes(value);
    }
    if (field === 'cpu.temperature.primary') {
        return formatTemperature(value);
    }
    if (field === 'disk.usage.summary') {
        return formatDiskSummary(value);
    }
    if (field === 'disk.io.writeCount') {
        return value === undefined ? '-' : String(value);
    }
    return value ?? '-';
}

function formatDiskSummary(usage) {
    if (!usage) return '-';
    const used = formatBytes(usage.used);
    const total = formatBytes(usage.total);
    if (used === '-' || total === '-') return '-';
    return `${used} / ${total}`;
}

function formatUptimeHours(seconds) {
    const hours = Number(seconds) / 3600;
    if (!Number.isFinite(hours)) return '-';
    return `${hours.toFixed(2)} hours`;
}

function formatTemperature(value) {
    const temp = Number(value);
    if (!Number.isFinite(temp)) return '-';
    return `${temp.toFixed(2)}°C`;
}

function formatPercent(value) {
    const num = Number(value);
    if (!Number.isFinite(num)) return '-';
    return `${num.toFixed(2)}%`;
}

function formatBytes(value) {
    const num = Number(value);
    if (!Number.isFinite(num)) return '-';
    if (num === 0) return '0 Bytes';

    const units = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    let idx = 0;
    let size = num;
    while (size >= 1024 && idx < units.length - 1) {
        size /= 1024;
        idx += 1;
    }
    const precision = idx === 0 ? 0 : 2;
    return `${size.toFixed(precision)} ${units[idx]}`;
}

function renderCpuTemperatures(root, temps) {
    const container = root.querySelector('[data-section="cpu-temperatures"]');
    const template = document.getElementById('device-info-row-template');
    const headerTemp = root.querySelector('[data-field="cpu.temperature.primary"]');
    if (!container || !template) return;

    container.innerHTML = '';
    const entries = Object.entries(temps ?? {});
    if (!entries.length) {
        if (headerTemp) {
            headerTemp.textContent = '-';
        }
        return;
    }

    const [primaryLabel, primaryValue] = entries[0];
    if (headerTemp) {
        headerTemp.textContent = formatTemperature(primaryValue);
    }

    entries.forEach(([label, value]) => {
        const row = template.content.firstElementChild.cloneNode(true);
        row.querySelector('.sub-heading').textContent = `${label} temperature`;
        row.querySelector('.sub-value').textContent = formatTemperature(value);
        container.appendChild(row);
    });
}

function updateCpuTemperatureIndicator(root, temps) {
    const indicator = root.querySelector('[data-temp-indicator="true"]');
    if (!indicator) return;

    indicator.classList.remove('temp-ok', 'temp-warn', 'temp-hot');
    const entries = Object.values(temps ?? {});
    const primaryTemp = entries.length ? Number(entries[0]) : NaN;

    if (!Number.isFinite(primaryTemp)) {
        return;
    }

    const fill = indicator.querySelector('.temp-indicator-fill');
    if (fill) {
        const clamped = Math.max(0, Math.min(100, primaryTemp));
        fill.style.width = `${clamped.toFixed(2)}%`;
    }

    if (primaryTemp > 80) {
        indicator.classList.add('temp-hot');
    } else if (primaryTemp > 70) {
        indicator.classList.add('temp-warn');
    } else {
        indicator.classList.add('temp-ok');
    }
}

function renderCpuUsage(root, usage) {
    const container = root.querySelector('[data-section="cpu-usage"]');
    const template = document.getElementById('device-info-row-template');
    if (!container || !template) return;

    container.innerHTML = '';
    const entries = Object.entries(usage ?? {});
    if (!entries.length) return;

    entries.sort(([a], [b]) => a.localeCompare(b, undefined, { numeric: true }));
    entries.forEach(([label, value]) => {
        if (label === 'avg') return;
        const row = template.content.firstElementChild.cloneNode(true);
        const displayLabel = label === 'avg' ? 'AVG' : label.toUpperCase();
        row.querySelector('.sub-heading').textContent = displayLabel;
        row.querySelector('.sub-value').textContent = formatPercent(value);
        container.appendChild(row);
    });
}

function updateUsageBars(root, data) {
    const bars = root.querySelectorAll('[data-bar]');
    bars.forEach((bar) => {
        const field = bar.dataset.bar;
        const value = getHardwareValue(data, field);
        const fill = bar.querySelector('.usage-bar-fill');
        if (!fill) return;

        const num = Number(value);
        if (!Number.isFinite(num)) {
            bar.classList.add('is-muted');
            fill.style.width = '0%';
            return;
        }

        bar.classList.remove('is-muted');
        const clamped = Math.max(0, Math.min(100, num));
        fill.style.width = `${clamped.toFixed(2)}%`;
        bar.title = `${clamped.toFixed(2)}%`;
    });
}
