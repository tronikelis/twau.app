import { Component } from "./types";

import components from "./components";

type Mountable = {
  querySelectorAll<E extends Element = Element>(selectors: string): NodeListOf<E>;
};

type Mounted = {
  element: Element;
  cleanup: (() => void) | undefined;
};

const mounted: Mounted[] = [];

function mount(component: Component<any>, to: Element) {
  mounted.push({
    element: to,
    cleanup: component.mount(to),
  });
}

function mountComponents(on: Mountable) {
  for (const component of components) {
    const elements = on.querySelectorAll(`[data-component-id='${component.id}']`);

    elements.forEach((el) => {
      mount(component, el);
    });
  }
}

// Event - htmx:afterSwap
//
//
// This event is triggered after new content has been swapped into the DOM.
// Details
//
//     detail.elt - the swapped in element
//     detail.xhr - the XMLHttpRequest
//     detail.target - the target of the request
//     detail.requestConfig - the configuration of the AJAX request
["htmx:afterSwap", "htmx:oobAfterSwap"].forEach((ev) =>
  addEventListener(ev, (ev) => {
    // temporary fix for duplicate after swap events
    if ((ev as CustomEvent).detail.__seen) {
      return;
    }

    (ev as CustomEvent).detail.__seen = true;
    const elt: HTMLElement = (ev as CustomEvent).detail.elt;
    mountComponents(elt);
  }),
);

function maybeUnmount(parent: Node) {
  for (const [i, comp] of mounted.entries()) {
    if (parent.contains(comp.element) || parent === comp.element) {
      comp.cleanup?.();
      mounted.splice(i, 1);
    }
  }
}

addEventListener("htmx:beforeCleanupElement", (ev) => {
  const elt = (ev as CustomEvent).detail.elt as Element;
  maybeUnmount(elt);
});

addEventListener("DOMContentLoaded", () => {
  new MutationObserver((records) => {
    for (const record of records) {
      for (const node of record.removedNodes) {
        maybeUnmount(node);
      }
    }
  }).observe(document.body, {
    childList: true,
    subtree: true,
  });
});

addEventListener("DOMContentLoaded", () => {
  mountComponents(document);
});
