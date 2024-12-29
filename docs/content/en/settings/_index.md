---
menu:
  main:
    weight: 50
title: Settings
# TODO(jeremy): Should I create a new layout for this page? Using the docs type we wind up with the Feedback form at the bottom of the page.
# which is a bit weird
type: docs
---

<div id="settings">
    <h2>BasePath</h2>
    If you set `BasePath` to the path where you have cloned the Foyle repository, then you will be able to click
    on the link "Open In VSCode" to open the current page in VSCode.
</div>
<div>
    <label for="basePath">Base Path:</label>
    <input type="text" id="basePath" name="basePath" />
    <button onclick="saveBasePath()">Save</button>
</div>

<script>
    // Load the saved BasePath value when the page loads
    document.addEventListener('DOMContentLoaded', function() {
        const savedBasePath = localStorage.getItem('BasePath');
        if (savedBasePath) {
            document.getElementById('basePath').value = savedBasePath;
        }
    });

    // Function to save the BasePath value to local storage
    function saveBasePath() {
        const basePath = document.getElementById('basePath').value;
        localStorage.setItem('BasePath', basePath);
        alert('BasePath saved: ' + basePath);
    }
</script>