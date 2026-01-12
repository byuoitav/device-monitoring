window.components = window.components || {};

window.components.dockerContainers = {
    loadPage: function() {
        console.log("Docker Containers component loaded");
        loadDockerContainers();
    },

    cleanup: function() {
        console.log("Docker Containers component cleaned up");
    }
}

async function loadDockerContainers() {
    const root = document.querySelector('.component-container .docker-containers-root');
    if (root) {
        root.classList.add('loading');
    }

    try {
        const data = await ApiService.getHardwareInfo();
        renderDockerContainers(data);
    } catch (error) {
        console.error("Failed to load docker containers:", error);
        renderDockerContainers({});
    } finally {
        if (root) {
            root.classList.remove('loading');
        }
    }
}

function renderDockerContainers(data) {
    const list = Array.isArray(data?.docker?.stats) ? data.docker.stats : [];
    const count = data?.docker?.['docker-containers'];
    const container = document.querySelector('.component-container .docker-container-list');
    const template = document.getElementById('docker-container-row-template');
    const countEl = document.querySelector('.component-container [data-field="docker.count"]');

    if (countEl) {
        countEl.textContent = Number.isFinite(Number(count)) ? String(count) : '-';
    }
    if (!container || !template) return;

    container.innerHTML = '';

    if (!list.length) {
        const row = template.content.firstElementChild.cloneNode(true);
        row.querySelector('.sub-heading').textContent = 'No containers reported';
        row.querySelector('.sub-value').textContent = '-';
        container.appendChild(row);
        return;
    }

    list.forEach((item) => {
        const row = template.content.firstElementChild.cloneNode(true);
        const status = [item?.running ? 'running' : 'stopped', item?.status].filter(Boolean).join(' ');
        row.querySelector('.sub-heading').textContent = item?.name ?? 'unknown';
        row.querySelector('.sub-value').textContent = status || '-';
        container.appendChild(row);
    });
}
