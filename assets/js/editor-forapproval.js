// THIS FILE IS SET FOR REWRITING
// it was used to pull a download from Demozoo, but needs to be rewritten.
// See: handler/app/internal/str/str.go, DemozooGetLink()
(() => {
  "use strict";
  const buttons = document.getElementsByName("editorGetDemozoo");
  const workQueue = [];
  for (const button of buttons) {
    button.addEventListener("click", handleClick, false);
    const uuid = button.dataset.uid;
    if (!uuid) console.error("No UUID found");
    else workQueue.push(uuid);
  }
  console.log("workQueue", workQueue);

  function removeClick(event) {
    console.log("event listener removed");
    event.target.removeEventListener("click", handleClick, false);
  }

  function handleClick(event) {
    const button = event.target;
    const id = event.target.dataset.id;
    const uuid = event.target.dataset.uid;
    button.classList.add("btn-outline-primary");
    button.classList.remove(
      "btn-outline-warning",
      "btn-outline-danger",
      "btn-outline-success"
    );
    const feedback = document.getElementById("editorFeedback" + id);
    feedback.classList.remove("text-danger-emphasis", "text-success-emphasis");
    feedback.classList.add("text-primary-emphasis");
    feedback.textContent = `Fetching download from Demozoo ID ${id}`;
    fetch("/editor/get/demozoo/download/" + id + "?uuid=" + uuid, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        //'X-CSRFToken': getCookie('csrftoken')
      },
    })
      .then((response) => response.json())
      .then((data) => {
        console.log(JSON.stringify(data, null, 2));
        if (data.success) {
          button.classList.add("btn-outline-success");
          button.classList.remove("btn-outline-primary");
          feedback.classList.remove("text-primary-emphasis");
          feedback.classList.add("text-success-emphasis");
          feedback.textContent = `Success, got ${data.filename}, refresh the page to see changes`;
          button.disabled = true;
          removeClick(event);
        } else {
          button.classList.add("btn-outline-danger");
          button.classList.remove("btn-outline-primary");
          feedback.classList.remove("text-primary-emphasis");
          feedback.classList.add("text-danger-emphasis");
          feedback.textContent = `Problem, ${data.error}`;
        }
      });
  }
})();
