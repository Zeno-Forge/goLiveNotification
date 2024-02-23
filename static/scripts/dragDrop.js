// Prevent default drag behaviors
['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
    document.addEventListener(eventName, preventDefaults, false)
    document.body.addEventListener(eventName, preventDefaults, false)
})
  
// Highlight drop area when item is dragged over it
var dropArea = document.getElementById('dropArea');
  
['dragenter', 'dragover'].forEach(eventName => {
    dropArea.addEventListener(eventName, highlight, false)
})
  
dropArea.addEventListener('dragleave', unhighlight, false);
dropArea.addEventListener('drop', unhighlight, false);
  
// Prevent default drag behaviors
function preventDefaults (e) {
    e.preventDefault()
    e.stopPropagation()
}
  
// Highlight the drop area
function highlight(e) {
    dropArea.classList.add('highlight') // Add a 'highlight' class or change the style directly
}
  
// Unhighlight the drop area
function unhighlight(e) {
    dropArea.classList.remove('highlight') // Remove the 'highlight' class or change the style directly
}
  
// Handle dropped files
dropArea.addEventListener('drop', handleDrop, false)
  
var droppedFile = null;

function handleDrop(e) {
    let dt = e.dataTransfer
    if (dt.files.length > 0) {
        droppedFile = dt.files[0]
        uploadFile(droppedFile)
    }
}

function uploadFile(file) {
    const reader = new FileReader();
    reader.readAsDataURL(file);
    reader.onloadend = function(e) {
        const img = document.getElementById('image-preview');
        if(img) {
            img.src = reader.result;
        }
        
        const dropArea = document.getElementById('dropArea');
        if(dropArea) {
            dropArea.classList.add('hidden');
        }
        
        const previewImage = document.getElementById('previewImage');
        if(previewImage) {
            previewImage.src = reader.result;
            previewImage.classList.remove('hidden');
        }
    };
}