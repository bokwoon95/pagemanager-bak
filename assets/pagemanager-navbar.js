"use strict";

document.addEventListener("DOMContentLoaded", function main() {
  function createElement(tag, attributes, ...children) {
    if (tag.includes("<") && tag.includes(">") && attributes === undefined && children.length === 0) {
      const template = document.createElement("template");
      template.innerHTML = tag;
      return template.content;
    }
    const element = document.createElement(tag);
    for (const [attribute, value] of Object.entries(attributes || {})) {
      if (attribute === "style") {
        Object.assign(element.style, value);
        continue;
      }
      if (attribute.startsWith("on")) {
        element.addEventListener(attribute.slice(2), value);
        continue;
      }
      element.setAttribute(attribute, value);
    }
    element.append(...children);
    return element;
  }
  const navbar = createElement(
    "div",
    {
      class: "pm-navbar-padding flex justify-between",
      style: { height: "32px" } ,
    },
    createElement("div", {}, "testing"),
    createElement("div", {}, "testing"),
  );
  document.body.prepend(navbar);
});
