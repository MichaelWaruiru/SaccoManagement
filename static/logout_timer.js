var timeout;

function resetTimer() {
    clearTimeout(timeout);
    timeout = setTimeout(logout, 30 * 60 * 1000); // 30 minutes
}

function logout() {
    window.location.href = "/logout";
}

document.addEventListener("mousemove", resetTimer);
document.addEventListener("keypress", resetTimer);

resetTimer(); // Initialize the timer when the page loads
