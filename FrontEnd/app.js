document.addEventListener('DOMContentLoaded', function () {
    fetchMessages();
});

function fetchMessages() {
    fetch('http://localhost:8080/api/messages')
        .then(response => response.json())
        .then(messages => {
            const messagesDiv = document.getElementById('messages');
            messages.forEach(message => {
                const messageCard = document.createElement('div');
                messageCard.classList.add('message-card');
                messageCard.innerHTML = `
                    <p><strong>${message.username}</strong>: ${message.message}</p>
                    <p class="timestamp">${message.timestamp}</p>
                `;
                messagesDiv.appendChild(messageCard);
            });
        })
        .catch(error => {
            console.error('Error fetching messages:', error);
        });
}