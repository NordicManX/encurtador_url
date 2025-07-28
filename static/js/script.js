document.addEventListener('DOMContentLoaded', () => {
    // Obtém referências para os elementos HTML que vamos manipular
    const shortenForm = document.getElementById('shortenForm');
    const longUrlInput = document.getElementById('longUrl');
    const resultContainer = document.getElementById('resultContainer');
    const shortUrlInput = document.getElementById('shortUrl');
    const copyButton = document.getElementById('copyButton');
    const copySuccessAlert = document.getElementById('copySuccessAlert');

    shortenForm.addEventListener('submit', async (e) => {
        e.preventDefault(); 

        const longUrl = longUrlInput.value; 
        try {
            
            const response = await fetch('/api/shorten', { // <-- MUDANÇA AQUI!
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                },
                body: `url=${encodeURIComponent(longUrl)}`,
            });
           
            if (!response.ok) {
                const errorMessage = await response.text();
                alert(`Erro ao encurtar URL: ${errorMessage}`);
                resultContainer.style.display = 'none';
                return;
            }

            const shortUrl = await response.text();
            shortUrlInput.value = shortUrl;
            resultContainer.style.display = 'block';
            copySuccessAlert.classList.add('d-none');
            shortUrlInput.select();
        } catch (error) {
            console.error('Erro na requisição:', error);
            alert('Ocorreu um erro ao tentar encurtar a URL. Tente novamente.');
            resultContainer.style.display = 'none';
        }
    });

    copyButton.addEventListener('click', () => {
        shortUrlInput.select();
        shortUrlInput.setSelectionRange(0, 99999);
        document.execCommand('copy');

        copySuccessAlert.classList.remove('d-none');
        setTimeout(() => {
            copySuccessAlert.classList.add('d-none');
        }, 2000);
    });
});