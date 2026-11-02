// Autocomplete dropdown under the guess input, grouped by language.
import { fetchSuggestions } from "./api.js";

export function setupSuggestions(input) {
  const box = document.createElement("ul");
  box.id = "suggestions";
  input.parentNode.appendChild(box);

  const hide = () => { box.style.display = "none"; };

  function render(groups) {
    box.innerHTML = "";
    for (const { lang, names } of groups) {
      const header = document.createElement("li");
      header.className = "suggest-lang";
      header.textContent = lang.toUpperCase();
      box.appendChild(header);

      for (const name of names || []) {
        const li = document.createElement("li");
        li.textContent = name;
        li.addEventListener("click", () => {
          input.value = name;
          hide();
          input.focus();
        });
        box.appendChild(li);
      }
    }
    box.style.display = "block";
  }

  input.addEventListener("input", async () => {
    const query = input.value.trim();
    if (!query) return hide();

    try {
      const groups = await fetchSuggestions(query);
      if (!Array.isArray(groups) || groups.length === 0) return hide();
      render(groups);
    } catch (err) {
      console.error("suggest error", err);
      hide();
    }
  });

  document.addEventListener("click", (e) => {
    if (!box.contains(e.target) && e.target !== input) hide();
  });

  return { hide };
}
