window.components = window.components || {};

window.components.dividerSensor = {
    loadPage: function() {
        console.log("Divider Sensor component loaded");
        loadDividerSensorInfo();
    },

    cleanup: function() {
        console.log("Divider Sensor component cleaned up");
    }
}

async function loadDividerSensorInfo() {
    const card = document.querySelector('.component-container .divider-sensor-card');
    if (card) {
        card.classList.add('loading');
    }

    try {
        const info = await ApiService.getDividerSensorInfo();
        applyDividerSensorInfo(info);
    } catch (error) {
        console.error("Failed to load divider sensor info:", error);
        applyDividerSensorInfo({});
    } finally {
        if (card) {
            card.classList.remove('loading');
        }
    }
}

function applyDividerSensorInfo(info) {
    const root = document.querySelector('.component-container');
    if (!root) return;

    setDividerField(root, 'divider.address', info?.address);
    setDividerField(root, 'divider.status', info?.status);
    setDividerField(root, 'divider.preset', info?.preset);
    setDividerField(root, 'divider.pin', info?.pin);
}

function setDividerField(root, field, value) {
    const el = root.querySelector(`[data-field="${field}"]`);
    if (!el) return;
    el.textContent = value ?? '-';
}
