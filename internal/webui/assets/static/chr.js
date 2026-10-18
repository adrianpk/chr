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

  let modalTrigger = null;
  let selectedKey = "";
  let boardScroll = 0;
  let boardMessage = "";
  let boardFocusKey = "";

  function syncModal() {
    const dialog = document.querySelector("#modal-root [role='dialog']");
    document.body.style.overflow = dialog ? "hidden" : "";
    document.querySelectorAll(".app-main, .top-chrome").forEach(function (element) {
      element.inert = Boolean(dialog);
    });
    if (!dialog || dialog.dataset.focusBound) return;
    dialog.dataset.focusBound = "1";
    const first = dialog.querySelector("#ticket-title:not([disabled])") || dialog.querySelector("input:not([type='hidden']):not([disabled]), textarea:not([disabled]), button:not([disabled]), a[href]");
    if (first) first.focus({preventScroll: true});
  }

  function closeModal() {
    const root = document.querySelector("#modal-root");
    if (root) root.innerHTML = "";
    syncModal();
    const card = Array.from(document.querySelectorAll("[data-ticket-key]")).find(function (element) {
      return element.dataset.ticketKey === selectedKey && element.matches("a");
    });
    const trigger = modalTrigger && modalTrigger.isConnected ? modalTrigger : card;
    if (trigger) trigger.focus({preventScroll: true});
  }

  function openTemplateModal(templateID) {
    const root = document.querySelector("#modal-root");
    const template = document.getElementById(templateID);
    if (!root || !template) return;
    root.innerHTML = "";
    root.appendChild(template.content.cloneNode(true));
    root.querySelectorAll("[data-dump-capture]").forEach(function (form) {
      form.elements.request_id.value = newDumpCaptureID();
    });
    if (window.htmx) window.htmx.process(root);
    syncModal();
  }

  function newDumpCaptureID() {
    const bytes = new Uint8Array(16);
    window.crypto.getRandomValues(bytes);
    return "web-dump-" + Array.from(bytes, function (value) { return value.toString(16).padStart(2, "0"); }).join("");
  }

  function bindDumpCapture(form) {
    const fingerprint = JSON.stringify([form.elements.title.value, form.elements.content.value, form.elements.case_slug.value]);
    if (form.dataset.captureArguments && form.dataset.captureArguments !== fingerprint) {
      form.elements.request_id.value = newDumpCaptureID();
    }
    form.dataset.captureArguments = fingerprint;
  }

  function dumpCaptureResponse(event) {
    const parameters = event.detail.requestConfig && event.detail.requestConfig.parameters;
    if (!parameters || parameters.action !== "dump.create" || event.detail.target.id !== "modal-root") return;
    if (document.querySelector("#modal-root [role='dialog']")) return;
    closeModal();
    const workspace = document.querySelector("#workspace");
    if (workspace && window.htmx) window.htmx.ajax("GET", workspace.getAttribute("hx-get"), {source: workspace, target: workspace, swap: "innerHTML"});
  }

  function refreshBoard() {
    const workspace = document.querySelector("#workspace");
    if (!workspace || !workspace.querySelector(".board-grid") || !window.htmx) return;
    window.htmx.ajax("GET", workspace.getAttribute("hx-get"), {source: workspace, target: workspace, swap: "innerHTML"});
  }

  function selectCard() {
    document.querySelectorAll("a[data-ticket-key]").forEach(function (card) {
      if (card.dataset.ticketKey === selectedKey) card.setAttribute("aria-current", "true");
      else card.removeAttribute("aria-current");
    });
  }

  function ticketResponse() {
    const posted = document.querySelector("[data-ticket-commented]");
    if (posted) {
      document.querySelectorAll("[data-ticket-editor] input[name='revision']").forEach(function (input) {
        if (input.value === posted.dataset.previousRevision) input.value = posted.dataset.revision;
      });
      const form = document.querySelector(".ticket-comments__form");
      if (form) {
        form.querySelector("[name='content']").value = "";
        form.querySelector("[name='request_id']").value = posted.dataset.requestId;
      }
      document.querySelector("#ticket-comment-result").innerHTML = "";
      posted.removeAttribute("data-ticket-commented");
      refreshBoard();
    }
    const result = document.querySelector("[data-ticket-write]:not([data-handled])");
    if (!result || result.dataset.handled) return;
    result.dataset.handled = "1";
    if (result.dataset.writeStatus !== "ok") return;
    if (result.dataset.ticketAction === "ticket.move") return;
    closeModal();
    refreshBoard();
  }

  function setupModalEvents() {
    if (document.body.dataset.modalEventsBound === "1") return;
    document.body.dataset.modalEventsBound = "1";
    document.body.addEventListener("click", function (event) {
      const trigger = event.target.closest("[hx-target='#modal-root'], [data-modal-template]");
      if (trigger && !trigger.closest("#modal-root")) {
        modalTrigger = trigger;
        selectedKey = trigger.dataset.ticketKey || "";
        selectCard();
      }
      const templateTrigger = event.target.closest("[data-modal-template]");
      if (templateTrigger) {
        event.preventDefault();
        openTemplateModal(templateTrigger.dataset.modalTemplate);
        return;
      }
      if (event.target.closest("[data-modal-close]")) { closeModal(); return; }
      const maximize = event.target.closest("[data-ticket-maximize]");
      if (maximize) {
        const expanded = maximize.closest(".ticket-modal").classList.toggle("ticket-modal--maximized");
        maximize.setAttribute("aria-pressed", String(expanded));
        maximize.textContent = expanded ? "Restore" : "Maximize";
      }
      if (event.target.matches("[data-modal-backdrop]")) closeModal();
    });
    document.addEventListener("keydown", function (event) {
      const dialog = document.querySelector("#modal-root [role='dialog']");
      if (!dialog) return;
      if (event.key === "Escape") { closeModal(); return; }
      if (event.key !== "Tab") return;
      const focusable = Array.from(dialog.querySelectorAll("a[href], button, input:not([type='hidden']), textarea, select, [tabindex='0']"))
        .filter(function (element) { return !element.disabled && element.getClientRects().length > 0; });
      const first = focusable[0], last = focusable[focusable.length - 1];
      if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last.focus(); }
      else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first.focus(); }
    });
    document.addEventListener("submit", function (event) {
      const form = event.target.closest("[data-dump-capture]");
      if (form) bindDumpCapture(form);
    }, true);
    setupBoardEvents();
  }

  function setupBoardEvents() {
    let dragged = null;
    document.body.addEventListener("dragstart", function (event) {
      const card = event.target.closest("a[data-ticket-key]");
      if (!card) return;
      dragged = {key: card.dataset.ticketKey, revision: card.dataset.ticketRevision};
      event.dataTransfer.setData("text/plain", dragged.key);
      event.dataTransfer.effectAllowed = "move";
    });
    document.body.addEventListener("dragover", function (event) {
      if (dragged && event.target.closest("[data-board-column]")) event.preventDefault();
    });
    document.body.addEventListener("dragend", function () { dragged = null; });
    document.body.addEventListener("drop", function (event) {
      const column = event.target.closest("[data-board-column]");
      if (!dragged || !column) return;
      event.preventDefault();
      const values = {action: "ticket.move", id: dragged.key, revision: dragged.revision, status: column.dataset.boardColumn, confirm: "true"};
      dragged = null;
      boardMessage = "";
      window.htmx.ajax("POST", "/board/ticket/save", {target: "#board-notice", values: values, swap: "innerHTML"}).then(function () {
        const notice = document.querySelector("#board-notice");
        if (notice) boardMessage = notice.innerHTML;
        refreshBoard();
      });
    });
    document.addEventListener("htmx:configRequest", function (event) {
      if (event.detail.elt.id !== "workspace") return;
      const marker = event.detail.elt.querySelector(".state-marker");
      if (marker) event.detail.headers["If-None-Match"] = '"' + marker.dataset.revision + '"';
    });
    document.addEventListener("htmx:beforeSwap", function (event) {
      if (event.detail.xhr.status === 304) { event.detail.shouldSwap = false; return; }
      if (event.detail.target.id !== "workspace") return;
      const grid = document.querySelector(".board-grid");
      if (grid) boardScroll = grid.scrollLeft;
      const focused = document.activeElement;
      boardFocusKey = focused && focused.matches("a[data-ticket-key]") ? focused.dataset.ticketKey : "";
    });
    document.addEventListener("htmx:afterSwap", function (event) {
      if (event.detail.target.id === "workspace") {
        const grid = document.querySelector(".board-grid");
        if (grid) grid.scrollLeft = boardScroll;
        const notice = document.querySelector("#board-notice");
        if (notice && boardMessage) { notice.innerHTML = boardMessage; window.htmx.process(notice); }
        selectCard();
        if (boardFocusKey) {
          const card = Array.from(document.querySelectorAll("a[data-ticket-key]")).find(function (element) { return element.dataset.ticketKey === boardFocusKey; });
          if (card) card.focus({preventScroll: true});
        }
      }
      syncModal();
      ticketResponse();
      dumpCaptureResponse(event);
    });
    ["htmx:responseError", "htmx:sendError", "htmx:timeout"].forEach(function (name) {
      document.addEventListener(name, function (event) {
        const form = event.detail.elt.closest(".ticket-edit-form, .ticket-comments__form, [data-dump-capture]");
        if (!form) return;
        const target = document.querySelector(form.getAttribute("hx-target"));
        if (target) target.textContent = form.matches("[data-dump-capture]") ? "The request failed. Your context is retained. Retry this capture." : "The request failed. Your input is retained. Retry or reload the ticket.";
      });
    });
  }

  document.addEventListener("DOMContentLoaded", setupTopChrome);
  document.addEventListener("DOMContentLoaded", setupModalEvents);
  document.addEventListener("htmx:afterSwap", setupTopChrome);
})();
