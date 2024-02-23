
// Set user local timezone with htmx request
document.querySelector('form').addEventListener('htmx:configRequest', function(evt) {
  const formData = new FormData(evt.detail.formData)
  evt.detail.parameters['timezone'] = Intl.DateTimeFormat().resolvedOptions().timeZone;

  if (droppedFile !== null) {
    evt.detail.parameters['imageUpload'] = droppedFile;
  }
});

// Color picker event listener
document.getElementById('colorPicker').addEventListener('input', function(e) {
  const color = e.target.value;
  document.getElementById('colorInput').value = parseInt(color.replace('#', ''), 16);
  document.documentElement.style.setProperty('--background-tertiary', color);
});

// Title input event listener
document.getElementById('titleInput').addEventListener('input', function() {
  document.getElementById('previewTitle').textContent = this.value;
});

// Description input event listener
document.getElementById('descriptionInput').addEventListener('input', function() {
  document.getElementById('previewDesc').textContent = this.value;
});

// URL input event listener
document.getElementById('urlInput').addEventListener('input', function() {
  document.getElementById('previewURL').href = this.value;
});

// Advanced section toggle event listener
document.getElementById('advancedToggle').addEventListener('click', function() {
  const advancedFields = document.getElementById('advancedFields');
  advancedFields.classList.toggle('hidden');
});

// Image upload event listener
document.getElementById('imageUpload').addEventListener('change', function(event) {
  const reader = new FileReader();
  reader.onload = function(e) {
    const previewImage = document.getElementById('previewImage');
    previewImage.src = e.target.result;
    previewImage.classList.remove('hidden');
    const img = document.getElementById('image-preview');
    if(img) {
      img.src = e.target.result;
    }
        
    const dropArea = document.getElementById('dropArea');
    if(dropArea) {
      dropArea.classList.add('hidden');
    }
  };
  reader.readAsDataURL(event.target.files[0]);
});

// Preview toggle checkbox event listener
document.getElementById('togglePreviewCheckbox').addEventListener('change', function() {
  const discordPreview = document.getElementById('discordPreview');
  discordPreview.style.display = this.checked ? 'block' : 'none';
});

// Listen for the HTMX afterOnLoad event on the form
document.querySelector('form').addEventListener('htmx:afterOnLoad', function(event) {
    // Check if the response status is 200 (OK)
    if (event.detail.xhr.status === 200) {
        // Hide the modal
        document.getElementById('editModal').style.display = 'none';
    } else {
        // Handle non-OK responses if needed
        console.error('Request failed with status:', event.detail.xhr.status);
    }
});

document.body.addEventListener('click', function(event) {
  if (event.target.id === 'cancelButton') {
    document.getElementById('editModal').innerHTML = '';
    document.getElementById('editModal').style.display = 'none';
  }
});