(() => {
  "use strict";

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
          feedback.textContent = "Success, refresh the page to see changes";
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

  const buttons = document.getElementsByName("editorGetDemozoo");
  console.log(buttons.length);
  for (let button of buttons) {
    // /get/demozoo/download/:id?uuid=:uuid
    button.addEventListener("click", handleClick, false);

    // button.addEventListener('click', () => {
    //     fetch('/editor/getdemozoo/' + id, {
    //         method: 'POST',
    //         headers: {
    //             'Content-Type': 'application/json',
    //             'X-CSRFToken': getCookie('csrftoken')
    //         },
    //         body: JSON.stringify({
    //             'uuid': uuid
    //         })
    //     })
    //     .then(response => response.json())
    //     .then(data => {
    //         console.log(data);
    //         if (data.success) {
    //             button.classList.add('btn-success');
    //             button.classList.remove('btn-primary');
    //             button.innerHTML = 'Got it!';
    //         } else {
    //             button.classList.add('btn-danger');
    //             button.classList.remove('btn-primary');
    //             button.innerHTML = 'Error';
    //         }
    //     });
    // });
  }
})();