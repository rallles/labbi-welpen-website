(function () {
  var menuButton = document.querySelector('.hamburger');
  var navigation = document.getElementById('primary-navigation');
  var desktopQuery = window.matchMedia('(min-width: 56rem)');

  if (!menuButton || !navigation) {
    return;
  }

  function setMenuOpen(isOpen) {
    document.body.classList.toggle('mobile-menu-open', isOpen);
    menuButton.setAttribute('aria-expanded', isOpen ? 'true' : 'false');
    menuButton.setAttribute('aria-label', isOpen ? 'Menü schließen' : 'Menü öffnen');
  }

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
    if (event.key === 'Escape') {
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
