(function () {
  function setupTopChrome() {
    const chrome = document.querySelector(".top-chrome");
    if (!chrome || chrome.dataset.chromeBound === "1") return;

    const pin = chrome.querySelector(".nav-pin");
    const handle = chrome.querySelector(".top-chrome-handle");
    if (!pin) return;

    chrome.dataset.chromeBound = "1";

    function setExpanded(expanded) {
      chrome.dataset.navState = expanded ? "expanded" : "collapsed";
      if (handle) handle.setAttribute("aria-expanded", expanded ? "true" : "false");
    }

    function syncPinned() {
      setExpanded(pin.checked);
    }

    pin.addEventListener("change", syncPinned);

    chrome.addEventListener("mouseenter", function () {
      if (!pin.checked) setExpanded(true);
    });

    chrome.addEventListener("mouseleave", function () {
      if (!pin.checked) setExpanded(false);
    });

    if (handle) {
      handle.addEventListener("click", function () {
        if (!pin.checked) setExpanded(true);
      });
    }

    syncPinned();
  }

  document.addEventListener("DOMContentLoaded", setupTopChrome);
  document.addEventListener("htmx:afterSwap", setupTopChrome);
})();
