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

  function closeModal() {
    const root = document.querySelector("#modal-root");
    if (root) root.innerHTML = "";
  }

  function openTemplateModal(templateID) {
    const root = document.querySelector("#modal-root");
    const template = document.getElementById(templateID);
    if (!root || !template) return;

    root.innerHTML = "";
    root.appendChild(template.content.cloneNode(true));
    if (window.htmx) window.htmx.process(root);
  }

  function setupModalEvents() {
    if (document.body.dataset.modalEventsBound === "1") return;
    document.body.dataset.modalEventsBound = "1";

    document.body.addEventListener("click", function (event) {
      const modalTrigger = event.target.closest("[data-modal-template]");
      if (modalTrigger) {
        event.preventDefault();
        openTemplateModal(modalTrigger.dataset.modalTemplate);
        return;
      }

      if (event.target.closest("[data-modal-close]")) {
        closeModal();
        return;
      }

      const backdrop = event.target.closest("[data-modal-backdrop]");
      if (backdrop && event.target === backdrop) closeModal();
    });

    document.addEventListener("keydown", function (event) {
      if (event.key === "Escape") closeModal();
    });
  }

  document.addEventListener("DOMContentLoaded", setupTopChrome);
  document.addEventListener("DOMContentLoaded", setupModalEvents);
  document.addEventListener("htmx:afterSwap", setupTopChrome);
  document.addEventListener("htmx:afterSwap", setupModalEvents);
})();
