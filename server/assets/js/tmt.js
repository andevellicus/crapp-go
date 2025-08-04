/**
 * Initializes and runs the Trail Making Test (TMT).
 * @param {string} containerId - The ID of the DOM element to render the test into.
 * @param {object} settings - Configuration options for the test from the server.
 * @param {function} onTestEnd - Callback function executed when the test is complete.
 */
function initTMT(containerId, settings, onTestEnd) {
    const container = document.getElementById(containerId);
    if (!container) {
        console.error(`TMT container with ID #${containerId} not found.`);
        return;
    }

    const testSettings = {
        partAItems: 15,
        partBItems: 15,
        includePartB: true,
        ...settings,
    };
    
    const partAItems = parseInt(testSettings.partAItems, 10);
    const partBItems = parseInt(testSettings.partBItems, 10);
    const includePartB = testSettings.includePartB === 'true';

    // --- State Variables ---
    let phase = 'idle'; // idle, practice, part_a, part_b
    let currentPart = 'Practice';
    let items = [];
    let currentItemIndex = 0;
    let errors = 0;
    let partStartTime = 0;
    const canvasSize = { width: 600, height: 450 };
    const testData = {
        testStartTime: 0,
        testEndTime: 0,
        partACompletionTime: 0,
        partAErrors: 0,
        partBCompletionTime: 0,
        partBErrors: 0,
        clicks: [],
        settings: testSettings,
    };

    function generateItems() {
        const numItems = currentPart === 'Practice' ? 5 : (currentPart === 'Part A' ? partAItems : partBItems);
        items = [];
        const radius = 20;
        const minDistance = radius * 3.5;

        for (let i = 0; i < numItems; i++) {
            let label;
            if (currentPart === 'Part B') {
                label = (i % 2 === 0) ? (i / 2 + 1).toString() : String.fromCharCode(65 + i / 2);
            } else {
                label = (i + 1).toString();
            }
            
            let x, y, isOverlapping;
            do {
                isOverlapping = false;
                x = radius + Math.random() * (canvasSize.width - radius * 2);
                y = radius + Math.random() * (canvasSize.height - radius * 2);
                for (const item of items) {
                    if (Math.sqrt((x - item.x)**2 + (y - item.y)**2) < minDistance) {
                        isOverlapping = true;
                        break;
                    }
                }
            } while (isOverlapping);

            items.push({ id: i, label, x, y, radius, connected: false });
        }
        currentItemIndex = 0;
    }

    function draw() {
        const canvas = document.getElementById('tmt-canvas');
        if (!canvas) return;
        const ctx = canvas.getContext('2d');
        ctx.clearRect(0, 0, canvas.width, canvas.height);

        // Draw lines for connected items
        ctx.strokeStyle = '#3b82f6';
        ctx.lineWidth = 3;
        for (let i = 1; i < currentItemIndex; i++) {
            ctx.beginPath();
            ctx.moveTo(items[i-1].x, items[i-1].y);
            ctx.lineTo(items[i].x, items[i].y);
            ctx.stroke();
        }

        // Draw circles
        items.forEach(item => {
            ctx.beginPath();
            ctx.arc(item.x, item.y, item.radius, 0, 2 * Math.PI);
            ctx.fillStyle = item.connected ? '#dbeafe' : 'white';
            ctx.fill();
            ctx.strokeStyle = (item.id === currentItemIndex && currentPart === 'Practice') ? '#10b981' : '#9ca3af';
            ctx.lineWidth = (item.id === currentItemIndex && currentPart === 'Practice') ? 4 : 2;
            ctx.stroke();
            ctx.fillStyle = '#1f2937';
            ctx.textAlign = 'center';
            ctx.textBaseline = 'middle';
            ctx.font = 'bold 18px Arial';
            ctx.fillText(item.label, item.x, item.y);
        });
    }

    function handleCanvasClick(e) {
        if (phase === 'idle') return;
        const canvas = e.target;
        const rect = canvas.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;

        const clickedItem = items.find(item => Math.sqrt((x - item.x)**2 + (y - item.y)**2) <= item.radius);
        testData.clicks.push({ x, y, time: performance.now() - testData.testStartTime, part: currentPart });

        if (clickedItem) {
            if (clickedItem.id === currentItemIndex) {
                clickedItem.connected = true;
                currentItemIndex++;
                if (currentItemIndex >= items.length) {
                    endPart();
                }
            } else if (!clickedItem.connected) {
                errors++;
                document.getElementById('tmt-errors').textContent = errors;
            }
        }
        draw();
    }

    function startPart(partName) {
        phase = 'active';
        currentPart = partName;
        errors = 0;
        partStartTime = performance.now();
        
        container.innerHTML = `
            <div class="flex justify-between items-center mb-2">
                <div class="text-xl font-bold">${currentPart}</div>
                <div class="text-lg">Errors: <span id="tmt-errors">0</span></div>
            </div>
            <canvas id="tmt-canvas" width="${canvasSize.width}" height="${canvasSize.height}" class="bg-gray-100 rounded-md border-2 border-gray-300"></canvas>
        `;
        document.getElementById('tmt-canvas').addEventListener('click', handleCanvasClick);
        generateItems();
        draw();
    }

    function endPart() {
        const completionTime = performance.now() - partStartTime;
        if (currentPart === 'Practice') {
            startPart('Part A');
        } else if (currentPart === 'Part A') {
            testData.partACompletionTime = completionTime;
            testData.partAErrors = errors;
            if (includePartB) {
                startPart('Part B');
            } else {
                endTest();
            }
        } else { // Part B
            testData.partBCompletionTime = completionTime;
            testData.partBErrors = errors;
            endTest();
        }
    }

    function endTest() {
        phase = 'idle';
        testData.testEndTime = performance.now();
        container.innerHTML = `<p class="text-lg font-semibold">Test complete. Saving results...</p>`;
        if (onTestEnd) onTestEnd(testData);
    }
    
    container.innerHTML = `<button id="start-tmt-btn" class="primary-button">Start Test</button>`;
    document.getElementById('start-tmt-btn').onclick = () => {
        testData.testStartTime = performance.now();
        startPart('Practice');
    };
}