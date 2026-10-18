(function () {
  let position = null;

  function isDashRequest(event) {
    const path = event.detail.pathInfo && event.detail.pathInfo.requestPath;
    if (!path) return false;
    return new URL(path, window.location.href).pathname === "/dash/fragment";
  }

  function badge() { return document.getElementById("dash-observation-status"); }

  function markObservation(event, failed) {
    const status = badge();
    if (!status) return;
    status.removeAttribute("aria-busy");
    const observed = event.detail.xhr && event.detail.xhr.getResponseHeader("X-CHR-Observed-At");
    if (!failed && observed) status.dataset.observedAt = observed;
    status.dataset.state = failed ? (status.dataset.observedAt ? "stale" : "unavailable") : "fresh";
    status.textContent = failed
      ? (status.dataset.observedAt ? "Live work is unavailable. Showing the last observation from " + status.dataset.observedAt + "." : "Work data is unavailable. No observation has loaded yet.")
      : (status.dataset.observedAt ? "Observed at " + status.dataset.observedAt : "Work observation is current.");
  }

  function selectionPosition() {
    const selection = window.getSelection();
    if (!selection || selection.isCollapsed || !selection.anchorNode) return null;
    const parent = selection.anchorNode.parentElement;
    const container = parent && parent.closest(".dash-context, .dash-queue-item, .dash-event");
    if (!container || !container.contains(selection.focusNode)) return null;
    const attribute = container.hasAttribute("data-event-id") ? "data-event-id" : (container.hasAttribute("data-dash-key") ? "data-dash-key" : "data-context-id");
    return {attribute: attribute, value: container.getAttribute(attribute), text: selection.toString()};
  }

  function savePosition() {
    const active = document.activeElement;
    return {
      x: window.scrollX, y: window.scrollY,
      queue: document.querySelector(".dash-queue__items")?.scrollTop || 0,
      events: document.querySelector(".dash-events")?.scrollTop || 0,
      href: active && active.closest("#workspace") && active.matches("a[href]") ? active.getAttribute("href") : null,
      selection: selectionPosition()
    };
  }

  function restoreSelection(saved) {
    if (!saved || !saved.text) return;
    const container = document.querySelector("[" + saved.attribute + "='" + CSS.escape(saved.value) + "']");
    if (!container) return;
    const start = container.textContent.indexOf(saved.text);
    if (start < 0) return;
    const walker = document.createTreeWalker(container, NodeFilter.SHOW_TEXT);
    let node, offset = 0, first = null, last = null;
    while ((node = walker.nextNode())) {
      const end = offset + node.textContent.length;
      if (!first && start >= offset && start < end) first = {node: node, offset: start - offset};
      if (first && start + saved.text.length <= end) { last = {node: node, offset: start + saved.text.length - offset}; break; }
      offset = end;
    }
    if (!first || !last) return;
    const range = document.createRange();
    range.setStart(first.node, first.offset); range.setEnd(last.node, last.offset);
    const selection = window.getSelection(); selection.removeAllRanges(); selection.addRange(range);
  }

  function restorePosition() {
    if (!position) return;
    const queue = document.querySelector(".dash-queue__items"), events = document.querySelector(".dash-events");
    if (queue) queue.scrollTop = position.queue;
    if (events) events.scrollTop = position.events;
    if (position.href) {
      const link = Array.from(document.querySelectorAll("#workspace a[href]")).find(function (item) { return item.getAttribute("href") === position.href; });
      if (link) link.focus({preventScroll: true});
    }
    restoreSelection(position.selection);
    window.scrollTo(position.x, position.y);
    position = null;
  }

  document.addEventListener("htmx:configRequest", function (event) {
    if (event.detail.target?.id !== "workspace" || !document.querySelector("[data-dash-observation]")) return;
    const marker = document.querySelector("#workspace .state-marker");
    if (marker) event.detail.headers["If-None-Match"] = '"' + marker.dataset.revision + '"';
  });
  document.addEventListener("htmx:beforeRequest", function (event) {
    if (isDashRequest(event) && badge()) badge().setAttribute("aria-busy", "true");
  });
  document.addEventListener("htmx:beforeSwap", function (event) {
    if (!isDashRequest(event)) return;
    if (event.detail.xhr.status === 304) { event.detail.shouldSwap = false; markObservation(event, false); return; }
    if (event.detail.xhr.status >= 400) { event.detail.shouldSwap = false; markObservation(event, true); return; }
    position = savePosition();
  });
  document.addEventListener("htmx:afterSwap", function (event) {
    if (!isDashRequest(event)) return;
    restorePosition(); markObservation(event, false);
  });
  document.addEventListener("htmx:afterRequest", function (event) {
    if (isDashRequest(event)) markObservation(event, event.detail.failed);
  });
  ["htmx:responseError", "htmx:sendError", "htmx:timeout"].forEach(function (name) {
    document.addEventListener(name, function (event) { if (isDashRequest(event)) markObservation(event, true); });
  });
})();
