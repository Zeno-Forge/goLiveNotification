const eventSource = new EventSource("/events");
eventSource.onmessage = function (event) {
  // When a message is received, dispatch a custom event to trigger HTMX to reload the posts list
  document
    .getElementById("postListWrapper")
    .dispatchEvent(new CustomEvent("sseReceived"));
};
document.addEventListener("DOMContentLoaded", function () {
  const settingsBtn = document.getElementById("settingsBtn");
  const slideOutMenu = document.getElementById("slideOutMenu");
  const header = document.querySelector("header");

  // Function to open the menu
  function openMenu() {
    slideOutMenu.style.right = "0";
  }

  // Function to close the menu
  function closeMenu() {
    slideOutMenu.style.right = "-100%";
  }

  // Function to adjust slideOutMenu's top position based on header's height
  function adjustMenuPosition() {
    const headerHeight = header.offsetHeight;
    slideOutMenu.style.top = `${headerHeight}px`;
    slideOutMenu.style.height = `calc(100% - ${headerHeight}px)`;
  }

  // Adjust menu position initially and on window resize
  adjustMenuPosition();
  window.addEventListener("resize", adjustMenuPosition);

  settingsBtn.addEventListener("click", function (event) {
    const isMenuHidden =
      slideOutMenu.style.right === "-100%" || slideOutMenu.style.right === "";
    if (isMenuHidden) {
      openMenu();
    } else {
      closeMenu();
    }
    event.stopPropagation(); // Prevent click from reaching the document listener
  });

  document.addEventListener("click", function (event) {
    const isClickInsideMenu = slideOutMenu.contains(event.target);
    const isClickSettingsBtn = settingsBtn.contains(event.target);

    if (!isClickInsideMenu && !isClickSettingsBtn) {
      closeMenu();
    }
  });

  // Prevent menu from closing when clicking inside it
  slideOutMenu.addEventListener("click", function (event) {
    event.stopPropagation();
  });
});
