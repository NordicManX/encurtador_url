document.addEventListener('DOMContentLoaded', () => {
    // Obtém referências para os elementos HTML que vamos manipular
    const shortenForm = document.getElementById('shortenForm');
    const longUrlInput = document.getElementById('longUrl');
    const resultContainer = document.getElementById('resultContainer');
    const shortUrlInput = document.getElementById('shortUrl');
    const copyButton = document.getElementById('copyButton');
    const copySuccessAlert = document.getElementById('copySuccessAlert');

    // Adiciona um "listener" para o evento de submit do formulário de encurtar URL
    shortenForm.addEventListener('submit', async (e) => {
        e.preventDefault(); // Impede o comportamento padrão do formulário (recarregar a página)

        const longUrl = longUrlInput.value; // Pega o valor da URL longa digitada pelo usuário

        try {
            // Faz uma requisição POST assíncrona para o endpoint '/shorten' do nosso servidor Go
            const response = await fetch('http://localhost:8080/shorten', {
                method: 'POST', // Método HTTP POST
                headers: {
                    // Define o tipo de conteúdo do corpo da requisição
                    'Content-Type': 'application/x-www-form-urlencoded',
                },
                // Corpo da requisição: a URL codificada para ser enviada como dado de formulário
                body: `url=${encodeURIComponent(longUrl)}`,
            });

            // Verifica se a requisição foi bem-sucedida (status 2xx)
            if (!response.ok) {
                const errorMessage = await response.text(); // Pega a mensagem de erro do servidor
                alert(`Erro ao encurtar URL: ${errorMessage}`); // Mostra um alerta com o erro
                resultContainer.style.display = 'none'; // Esconde o container de resultados
                return; // Sai da função
            }

            // Se a requisição foi bem-sucedida, pega a URL encurtada da resposta
            const shortUrl = await response.text();
            shortUrlInput.value = shortUrl; // Define o valor do campo de URL encurtada
            resultContainer.style.display = 'block'; // Torna o container de resultados visível
            copySuccessAlert.classList.add('d-none'); // Esconde o alerta de cópia, caso esteja visível
            shortUrlInput.select(); // Seleciona o texto no campo da URL encurtada (útil para o usuário copiar)
        } catch (error) {
            // Captura e trata erros que podem ocorrer durante a requisição
            console.error('Erro na requisição:', error);
            alert('Ocorreu um erro ao tentar encurtar a URL. Tente novamente.');
            resultContainer.style.display = 'none'; // Esconde o container de resultados em caso de erro
        }
    });

    // Adiciona um "listener" para o evento de clique do botão de copiar
    copyButton.addEventListener('click', () => {
        shortUrlInput.select(); // Seleciona o texto no campo da URL encurtada
        // Para dispositivos móveis, garante que todo o texto seja selecionado
        shortUrlInput.setSelectionRange(0, 99999);
        document.execCommand('copy'); // Executa o comando de copiar para a área de transferência

        copySuccessAlert.classList.remove('d-none'); // Mostra o alerta de sucesso da cópia
        // Esconde o alerta de sucesso após 2 segundos
        setTimeout(() => {
            copySuccessAlert.classList.add('d-none');
        }, 2000);
    });
});