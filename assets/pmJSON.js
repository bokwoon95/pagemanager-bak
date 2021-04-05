document.addEventListener("DOMContentLoaded", function main() {
  const el = document.querySelector(`script[type="application/json"][data-pm-json]`);
  const s = (el && el.textContent) || "{}";
  Object.defineProperty(window, "pmJSON", {
    value: function (name) {
      try {
        const data = JSON.parse(s);
        return name === undefined ? data : data[name];
      } catch {
        return undefined;
      }
    },
  });
});
