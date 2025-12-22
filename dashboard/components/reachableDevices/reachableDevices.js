window.components = window.components || {};

window.components.reachableDevices = {
    loadPage: function() {
        console.log("Reachable Devices component loaded");
        initializeDrawers();
    },

    cleanup: function() {
        console.log("Reachable Devices component cleaned up");
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
