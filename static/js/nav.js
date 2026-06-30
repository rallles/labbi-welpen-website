(function () {
  var header = document.querySelector('.site-header');
  var menuButton = document.querySelector('.nav-toggle');
  var navigation = document.getElementById('primary-navigation');
  var navigationLinks = navigation ? navigation.querySelectorAll('.site-nav__link') : [];
  var desktopQuery = window.matchMedia('(min-width: 60rem)');

  if (!menuButton || !navigation) {
    return;
  }

  function setMenuOpen(isOpen) {
    document.body.classList.toggle('mobile-menu-open', isOpen);
    menuButton.setAttribute('aria-expanded', isOpen ? 'true' : 'false');
    menuButton.setAttribute('aria-label', isOpen ? 'Menü schließen' : 'Menü öffnen');
  }

  function normalizedPath(path) {
    if (path.length > 1 && path.endsWith('/')) {
      return path.slice(0, -1);
    }
    return path;
  }

  var currentPath = normalizedPath(window.location.pathname);
  navigationLinks.forEach(function (link) {
    if (normalizedPath(new URL(link.href).pathname) === currentPath) {
      link.classList.add('is-active');
      link.setAttribute('aria-current', 'page');
    }
  });

  menuButton.addEventListener('click', function () {
    var isOpen = menuButton.getAttribute('aria-expanded') === 'true';
    setMenuOpen(!isOpen);
  });

  navigation.addEventListener('click', function (event) {
    if (event.target.closest('a')) {
      setMenuOpen(false);
    }
  });

  document.addEventListener('keydown', function (event) {
    if (event.key === 'Escape' && menuButton.getAttribute('aria-expanded') === 'true') {
      setMenuOpen(false);
      menuButton.focus();
    }
  });

  document.addEventListener('click', function (event) {
    if (!desktopQuery.matches && menuButton.getAttribute('aria-expanded') === 'true' &&
        header && !header.contains(event.target)) {
      setMenuOpen(false);
    }
  });

  function closeMenuOnDesktop(event) {
    if (event.matches) {
      setMenuOpen(false);
    }
  }

  if (desktopQuery.addEventListener) {
    desktopQuery.addEventListener('change', closeMenuOnDesktop);
  } else {
    desktopQuery.addListener(closeMenuOnDesktop);
  }
})();
