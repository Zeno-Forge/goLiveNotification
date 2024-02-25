// Select elements
var dropArea = document.getElementById("dropArea");
var imgUpload = document.getElementById("imgUploadBtn");
var imageUpload = document.getElementById("imageUpload");
var previewImage = document.getElementById("previewImage");
var imgDelete = document.getElementById("imgDelete");
var imgURL = document.getElementById("imageURLInput");
var imgGroup = document.getElementById("image-preview-group");
var imgPrev = document.getElementById("image-preview");

// Set user local timezone with htmx request
document
  .querySelector("form")
  .addEventListener("htmx:configRequest", function (evt) {
    const formData = new FormData(evt.detail.formData);
    evt.detail.parameters["timezone"] =
      Intl.DateTimeFormat().resolvedOptions().timeZone;

    if (droppedFile !== null) {
      evt.detail.parameters["imageUpload"] = droppedFile;
    }
  });

// Listen for the HTMX afterOnLoad event on the form
document
  .querySelector("form")
  .addEventListener("htmx:afterOnLoad", function (event) {
    // Check if the response status is 200 (OK)
    if (event.detail.xhr.status === 200) {
      // Hide the modal
      document.getElementById("editModal").style.display = "none";
    } else {
      console.error("Request failed with status:", event.detail.xhr.status);
    }
  });

document.getElementById("cancelButton").addEventListener("click", function () {
  document.getElementById("editModal").innerHTML = "";
  document.getElementById("editModal").style.display = "none";
});

// Color picker event listener
document.getElementById("colorPicker").addEventListener("input", function (e) {
  const color = e.target.value;
  document.getElementById("colorInput").value = parseInt(
    color.replace("#", ""),
    16
  );
  document.documentElement.style.setProperty("--background-tertiary", color);
});

// Title input event listener
document.getElementById("titleInput").addEventListener("input", function () {
  document.getElementById("previewTitle").textContent = this.value;
});

// Description input event listener
document
  .getElementById("descriptionInput")
  .addEventListener("input", function () {
    document.getElementById("previewDesc").textContent = this.value;
  });

// URL input event listener
document.getElementById("urlInput").addEventListener("input", function () {
  document.getElementById("previewURL").href = this.value;
});

// Footer input event listener
document
  .getElementById("footerTextInput")
  .addEventListener("input", function () {
    document.getElementById("prevFootText").textContent = this.value;
  });

document
  .getElementById("footerIconInput")
  .addEventListener("input", function () {
    document.getElementById("prevFootIcon").src = this.value;
  });

// Advanced section toggle event listener
document
  .getElementById("advancedToggle")
  .addEventListener("click", function () {
    const advancedFields = document.getElementById("advancedFields");
    advancedFields.classList.toggle("hidden");
  });

// Image upload event listener
document
  .getElementById("imageUpload")
  .addEventListener("change", function (event) {
    const reader = new FileReader();
    reader.onload = function (e) {
      imgURL.value = e.target.result;
      previewImage.src = e.target.result;
      previewImage.classList.remove("hidden");

      imgPrev.src = e.target.result;
      imgGroup.classList.remove("hidden");
      if (dropArea) {
        dropArea.classList.add("hidden");
      }
    };
    reader.readAsDataURL(event.target.files[0]);
  });

document
  .getElementById("thumbImgUpload")
  .addEventListener("change", function (event) {
    const reader = new FileReader();
    reader.onload = function (e) {
      const prevThumb = document.getElementById("previewThumbnail");
      prevThumb.src = e.target.result;

      const thumbURL = document.getElementById("thumbnailInput");
      thumbURL.value = e.target.result;
    };
    reader.readAsDataURL(event.target.files[0]);
  });

// Preview toggle checkbox event listener
document
  .getElementById("togglePreviewCheckbox")
  .addEventListener("change", function () {
    const discordPreview = document.getElementById("discordPreview");
    discordPreview.style.display = this.checked ? "block" : "none";
  });

// Trigger file input when button is clicked
imgUpload.addEventListener("click", function () {
  imageUpload.click();
});

imgDelete.addEventListener("click", function () {
  const previewImage = document.getElementById("previewImage");
  previewImage.src = "";
  previewImage.classList.add("hidden");
  imgPrev.src = "";
  imgURL.value = "";
  dropArea.classList.remove("hidden");
  imgGroup.classList.add("hidden");
});

if (imgURL.value != "") {
  imgGroup.classList.remove("hidden");
  if (dropArea) {
    dropArea.classList.add("hidden");
  }
}

if (document.getElementById("colorPicker").getAttribute("value")) {
  document.documentElement.style.setProperty(
    "--background-tertiary",
    document.getElementById("colorPicker").value
  );
}
