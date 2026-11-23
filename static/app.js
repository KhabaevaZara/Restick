(function () {
  var splash = document.getElementById('splash');
  if (!splash) return;

  var HOLD = 3000; // сколько заставка висит (мс)
  var FADE = 1300; // длительность исчезновения (совпадает с transition в CSS)
  var done = false;

  document.documentElement.classList.add('is-loading');

  function hideSplash() {
    if (done) return;
    done = true;

    splash.classList.add('is-hidden');
    document.documentElement.classList.remove('is-loading');

    window.setTimeout(function () {
      if (splash.parentNode) splash.parentNode.removeChild(splash);
    }, FADE + 100);
  }

  window.addEventListener('load', function () {
    window.setTimeout(hideSplash, HOLD);
  });

  window.setTimeout(hideSplash, HOLD + 4000);

  if (performance.getEntriesByType && performance.getEntriesByType('navigation')[0] &&
      performance.getEntriesByType('navigation')[0].type === 'back_forward') {
    hideSplash();
  }
})();