let currentComponent = null;

document.addEventListener('DOMContentLoaded', async () => {
    console.log('Dashboard loaded');

    setupNavigation();

    const defaultNavItem = document.querySelector('.nav-item.nav-item-selected') || document.querySelector('.nav-item');
    if (defaultNavItem?.dataset.component) {
        await switchComponent(defaultNavItem.dataset.component, defaultNavItem);
    }

    await injectIcons();
});

function setupNavigation() {
    const navItems = document.querySelectorAll('.nav-item');

    navItems.forEach((item) => {
        item.addEventListener('click', async () => {
            const componentName = item.dataset.component;
            if (!componentName || componentName === currentComponent) return;
            await switchComponent(componentName, item);
        });
    });
}

async function switchComponent(componentName, navItem) {
    await cleanupCurrentComponent();
    updateNavSelection(navItem);
    removeComponentAssets();
    await loadComponent(componentName);
    await injectIcons(document.querySelector('.component-container'));
}

function updateNavSelection(selectedItem) {
    if (!selectedItem) return;
    document.querySelectorAll('.nav-item').forEach((item) => {
        item.classList.toggle('nav-item-selected', item === selectedItem);
    });
}

function removeComponentAssets() {
    const stylesheet = document.getElementById('component-stylesheet');
    if (stylesheet) {
        stylesheet.remove();
    }

    const script = document.getElementById('component-script');
    if (script) {
        script.remove();
    }
}

async function cleanupCurrentComponent() {
    if (!currentComponent) return;

    const module = window.components?.[currentComponent];
    if (module?.cleanup) {
        module.cleanup();
    }
}

async function injectIcons(root = document) {
    console.log('Injecting icons...');
    const icons = root.querySelectorAll('[data-icon]');

    for (const el of icons) {
        const name = el.dataset.icon;

        try {
            const res = await fetch(`/dashboard/assets/${name}.svg`);
            if (!res.ok) throw new Error();

            el.innerHTML = await res.text();
        } catch {
            console.warn(`Icon not found: ${name}`);
        }
    }
}

async function loadComponent(componentName, divQuerySelector = `.component-container`) {
    console.log(`Loading component: ${componentName} into ${divQuerySelector}`);
    const htmlPath = `./components/${componentName}/${componentName}.html`;
    const jsPath = `./components/${componentName}/${componentName}.js`;
    const cssPath = `./components/${componentName}/${componentName}.css`;

    const stylesheet = document.createElement('link');
    stylesheet.rel = 'stylesheet';
    stylesheet.href = cssPath;
    stylesheet.id = 'component-stylesheet';
    stylesheet.onload = () => {
        const module = window.components?.[componentName];
        if (module?.loadStyles) {
            module.loadStyles();
        }
    };

    document.body.appendChild(stylesheet);

    const componentContainer = document.querySelector(divQuerySelector);
    componentContainer.classList.add('loading');
    const response = await fetch(htmlPath);
    const html = await response.text();
    componentContainer.innerHTML = html;

    const script = document.createElement('script');
    script.src = jsPath;
    script.id = 'component-script';

    await new Promise((resolve, reject) => {
        script.onload = () => {
            const module = window.components?.[componentName];
            if (module?.loadPage) {
                module.loadPage();
                if (divQuerySelector === `.component-container`) {
                    currentComponent = componentName;
                }
            }
            componentContainer.classList.remove('loading');
            resolve();
        };
        script.onerror = reject;
        document.body.appendChild(script);
    });
}
