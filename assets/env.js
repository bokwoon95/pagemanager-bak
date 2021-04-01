"use strict";

document.addEventListener("DOMContentLoaded", function () {
  window.Env = function (name) {
    try {
      const el = document.querySelector(`script[type="application/json"][data-pagemanager-env]`);
      if (!el) {
        return undefined;
      }
      const data = JSON.parse(el.textContent);
      if (name === undefined) {
        return data;
      }
      return data[name];
    } catch (err) {
      return undefined;
    }
  };
});
