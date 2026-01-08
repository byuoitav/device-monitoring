window.components = window.components || {};

window.components.deviceInfo = {
    loadPage: function() {
        console.log("Device Info component loaded");
        initializeDeviceInfoDrawers();
    },

    cleanup: function() {
        console.log("Device Info component cleaned up");
    }
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
