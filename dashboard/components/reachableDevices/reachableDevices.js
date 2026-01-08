window.components = window.components || {};

window.components.reachableDevices = {
    loadPage: function() {
        console.log("Reachable Devices component loaded");
        loadReachableDevices();
    },

    cleanup: function() {
        console.log("Reachable Devices component cleaned up");
    }
}

async function loadReachableDevices() {
    const container = document.querySelector('.component-container .reachable-devices-list');
    const template = document.getElementById('reachable-device-card-template');
    const loadingCard = document.querySelector('.component-container .reachable-devices-loading');

    if (!container || !template) return;

    container.innerHTML = '';
    if (loadingCard) {
        loadingCard.classList.add('loading');
        loadingCard.style.display = '';
    }

    const errors = [];
    const pingPromise = ApiService.getRoomPing().catch((error) => {
        console.error('Error fetching room ping data:', error);
        errors.push(error);
        return new Map();
    });
    const healthPromise = ApiService.getRoomHealth().catch((error) => {
        console.error('Error fetching room health data:', error);
        errors.push(error);
        return new Map();
    });

    const [pingMap, healthMap] = await Promise.all([pingPromise, healthPromise]);
    const devices = new Set([...(pingMap?.keys?.() || []), ...(healthMap?.keys?.() || [])]);

    if (!devices.size) {
        if (errors.length) {
            const errorCard = document.createElement('div');
            errorCard.className = 'content-box';
            errorCard.textContent = 'Unable to load reachable devices.';
            container.appendChild(errorCard);
        } else {
            const emptyCard = document.createElement('div');
            emptyCard.className = 'content-box';
            emptyCard.textContent = 'No reachable devices reported.';
            container.appendChild(emptyCard);
        }
        if (loadingCard) {
            loadingCard.classList.remove('loading');
            loadingCard.remove();
        }
        return;
    }

    const sortedDevices = Array.from(devices).sort((a, b) => a.localeCompare(b));
    sortedDevices.forEach((deviceId) => {
        const ping = pingMap.get(deviceId);
        const health = healthMap.get(deviceId);
        const card = template.content.firstElementChild.cloneNode(true);

        const ip = ping?.ip || '0.0.0.0';
        const sent = parseNumber(ping?.['packets-sent']);
        const received = parseNumber(ping?.['packets-received']);
        const lost = (Number.isFinite(sent) && Number.isFinite(received))
            ? Math.max(0, sent - received)
            : undefined;
        const avgRoundTrip = ping?.['average-round-trip'] || '-';

        setCardField(card, 'device-name', deviceId);
        setCardField(card, 'device-ip', ip);
        setCardField(card, 'drawer-device-id', deviceId);
        setCardField(card, 'drawer-device-ip', ip);
        setCardField(card, 'packets-sent', formatValue(sent));
        setCardField(card, 'packets-received', formatValue(received));
        setCardField(card, 'packets-lost', formatValue(lost));
        setCardField(card, 'average-round-trip', avgRoundTrip);

        updateConnectivityStatus(card, sent, received, lost);

        if (healthMap.has(deviceId)) {
            const healthRow = buildHealthRow(health);
            const ipRow = card.querySelector('[data-field="drawer-device-ip"]')?.closest('.content-box-sub');
            if (healthRow && ipRow?.parentElement) {
                ipRow.parentElement.insertBefore(healthRow, ipRow.nextSibling);
            }
        }
        updateHealthStatus(card, health);

        container.appendChild(card);
    });

    if (loadingCard) {
        loadingCard.classList.remove('loading');
        loadingCard.remove();
    }
    initializeDrawers();
    if (typeof injectIcons === 'function') {
        await injectIcons(container);
    }
}

function initializeDrawers() {
    const cards = document.querySelectorAll('.component-container .device-card');

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

function setCardField(card, field, value) {
    const el = card.querySelector(`[data-field="${field}"]`);
    if (el) {
        el.textContent = value ?? '-';
    }
}

function parseNumber(value) {
    const num = Number(value);
    return Number.isFinite(num) ? num : undefined;
}

function formatValue(value) {
    return Number.isFinite(value) ? String(value) : '-';
}

function buildHealthRow(health) {
    const row = document.createElement('div');
    row.className = 'content-box-sub';

    const heading = document.createElement('p');
    heading.className = 'sub-heading';
    heading.textContent = 'Health Check Response';

    const value = document.createElement('p');
    value.className = 'sub-value';
    value.textContent = health ?? '-';

    row.appendChild(heading);
    row.appendChild(value);
    return row;
}

function updateConnectivityStatus(card, sent, received, lost) {
    const icon = card.querySelector('.connectivity');
    if (!icon) return;

    icon.classList.remove('status-ok', 'status-bad');

    if (Number.isFinite(lost) && lost > 0) {
        icon.classList.add('status-bad');
        icon.dataset.icon = 'thumb_down';
        return;
    }

    if (Number.isFinite(sent) && Number.isFinite(received) && sent === received) {
        icon.classList.add('status-ok');
        icon.dataset.icon = 'thumb_up';
    }
}

function updateHealthStatus(card, health) {
    const icon = card.querySelector('.power');
    if (!icon) return;

    const normalized = typeof health === 'string' ? health.trim().toLowerCase() : '';
    if (!normalized || normalized === 'healthy') {
        icon.classList.remove('status-bad');
        return;
    }

    icon.classList.add('status-bad');
}
