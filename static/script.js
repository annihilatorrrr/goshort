document
  .getElementById("shortenForm")
  .addEventListener("submit", async function (event) {
    event.preventDefault(); // Prevent form submission
    const longURL = document.getElementById("longURL").value;
    if (longURL.trim() === "") {
      alert("Please enter a URL to shorten.");
      return;
    }
    const res = await fetch("/sort", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ url: longURL }),
    });
    if (!res.ok) {
      alert("Try again later, unable to reach the backend at the moment!");
      return;
    }
    const data = await res.json();
    if (data.err) {
      alert('Error: ' + data.err);
      return;
    }
    if (data.hash) {
      const baseUrl = window.location.origin;
      const shortenedLink = `${baseUrl}/r/${data.hash}`;
      showPopup(shortenedLink);
    }
    document.getElementById("longURL").value = "";
  });

function closePopup() {
  const popup = document.getElementById("customPopup");
  popup.style.display = "none";
}

function showPopup(shortenedLink) {
  const popup = document.getElementById("customPopup");
  const shortLinkElement = document.getElementById("shortLink");
  shortLinkElement.textContent = shortenedLink;
  shortLinkElement.href = shortenedLink;
  popup.style.display = "block";
}

function copyToClipboard() {
  const shortLinkElement = document.getElementById("shortLink");
  const shortenedLink = shortLinkElement.href;
  navigator.clipboard
    .writeText(shortenedLink)
    .then(function () {
      alert("Copied!");
    })
    .catch(function (err) {
      console.error("Failed to copy: ", err);
    });
}
