document.addEventListener('DOMContentLoaded', function () {
    // DOM elements
    const packSizesContainer = document.getElementById('packSizesContainer');
    const addPackSizeBtn = document.getElementById('addPackSizeBtn');
    const submitPackSizesBtn = document.getElementById('submitPackSizesBtn');
    const orderQuantityInput = document.getElementById('orderQuantity');
    const calculateBtn = document.getElementById('calculateBtn');
    const resultsContainer = document.getElementById('resultsContainer');

    // Initial setup
    loadPackSizes();

    // Event listeners
    addPackSizeBtn.addEventListener('click', addPackSizeInput);
    submitPackSizesBtn.addEventListener('click', submitPackSizes);
    calculateBtn.addEventListener('click', calculatePacks);

    // Functions
    function loadPackSizes() {
        fetch('/api/packs')
            .then(response => {
                if (!response.ok) {
                    return response.json().then(err => {
                        throw new Error('Failed to load pack sizes: ' + err.Message);
                    });
                }
                return response.json();
            })
            .then(data => {
                packSizesContainer.innerHTML = '';
                if (data.packSizes && data.packSizes.length > 0) {
                    data.packSizes.forEach(size => {
                        addPackSizeInput(size);
                    });
                } else {
                    // Add default input if no pack sizes are available
                    addPackSizeInput();
                }
            })
            .catch(error => {
                console.error('Error loading pack sizes:', error);
                // Add default input if error
                addPackSizeInput();
            });
    }

    function addPackSizeInput(value = '') {
        const inputGroup = document.createElement('div');
        inputGroup.className = 'pack-size-input';

        const input = document.createElement('input');
        input.type = 'number';
        input.className = 'pack-size';
        input.min = '1';
        input.placeholder = 'Pack size';
        input.value = value;

        const removeBtn = document.createElement('button');
        removeBtn.className = 'remove-btn';
        removeBtn.textContent = 'Remove';
        removeBtn.addEventListener('click', function () {
            packSizesContainer.removeChild(inputGroup);
        });

        inputGroup.appendChild(input);
        inputGroup.appendChild(removeBtn);
        packSizesContainer.appendChild(inputGroup);
    }

    function submitPackSizes() {
        const inputs = document.querySelectorAll('.pack-size');
        const packSizes = [];

        inputs.forEach(input => {
            const size = parseInt(input.value);
            if (!isNaN(size) && size > 0) {
                packSizes.push(size);
            }
        });

        fetch('/api/packs', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ packSizes: packSizes }),
        })
            .then(response => {
                if (!response.ok) {
                    return response.json().then(err => {
                        throw new Error('Failed to update pack sizes: ' + err.Message);
                    });
                }
                return response.json();
            })
            .then(data => {
                alert('Pack sizes updated successfully!');
                loadPackSizes(); // Reload the pack sizes
            })
            .catch(error => {
                console.error('Error updating pack sizes:', error);
                alert(error.message || 'Error updating pack sizes. Please try again.');
            });
    }

    function calculatePacks() {
        const orderQuantity = parseInt(orderQuantityInput.value);

        fetch('/api/calculate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ orderQuantity: orderQuantity }),
        })
            .then(response => {
                if (!response.ok) {
                    return response.json().then(err => {
                        throw new Error('Failed to calculate pack configurations: ' + err.Message);
                    });
                }
                return response.json();
            })
            .then(data => {
                displayResults(data);
            })
            .catch(error => {
                console.error('Error calculating pack_configurations:', error);
                resultsContainer.innerHTML = `<p class="error">${error.message || 'Error calculating pack configurations. Please try again.'}</p>`;
            });
    }

    function escapeHtml(unsafe) {
        return unsafe
            .toString()
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;");
    }

    function displayResults(data) {
        // Handle missing or error response
        if (!data || data.success === false) {
            resultsContainer.innerHTML = `<p class="error">${data?.errorMessage || 'An error occurred while calculating packs.'}</p>`;
            return;
        }

        // Validate required data
        const orderQuantity = data.orderQuantity || 0;
        const totalItems = data.totalItems || 0;
        const totalPacks = data.totalPacks || 0;
        const packs = Array.isArray(data.pack_configurations) ? data.pack_configurations : [];

        // Handle case with no packs
        if (packs.length === 0) {
            resultsContainer.innerHTML = '<p class="error">No pack combination found for the given order quantity.</p>';
            return;
        }

        let html = `
            <p>Order quantity: <strong>${escapeHtml(orderQuantity)}</strong></p>
            <p>Total items to be shipped: <strong>${escapeHtml(totalItems)}</strong></p>
            <p>Total packs: <strong>${escapeHtml(totalPacks)}</strong></p>
            
            <h3>Pack Breakdown:</h3>
            <table>
                <thead>
                    <tr>
                        <th>Pack Size</th>
                        <th>Quantity</th>
                        <th>Total Items</th>
                    </tr>
                </thead>
                <tbody>
        `;

        // Sort packs by size in descending order for better readability
        packs.sort((a, b) => (b.size || 0) - (a.size || 0)).forEach(pack => {
            const size = pack.size || 0;
            const quantity = pack.quantity || 0;
            const totalPackItems = size * quantity;

            html += `
                <tr>
                    <td>${escapeHtml(size)}</td>
                    <td>${escapeHtml(quantity)}</td>
                    <td>${escapeHtml(totalPackItems)}</td>
                </tr>
            `;
        });

        html += `
                </tbody>
            </table>
            
            <div class="summary">
                <p>Excess items: <strong>${escapeHtml(Math.max(0, totalItems - orderQuantity))}</strong></p>
            </div>
        `;

        resultsContainer.innerHTML = html;
    }
});
