/**
 * Initializes and runs the Digit Span Test (DST).
 * @param {string} containerId - The ID of the DOM element to render the test into.
 * @param {object} settings - Configuration options for the test from the server.
 * @param {function} onTestEnd - Callback function executed when the test is complete.
 */
function initDST(containerId, settings, onTestEnd) {
    const container = document.getElementById(containerId);
    if (!container) {
        console.error(`DST container with ID #${containerId} not found.`);
        return;
    }

    const testSettings = {
        initialSpan: 3,
        maxSpan: 10,
        displayTimePerDigit: 1000,
        interDigitInterval: 500,
        recallTimeout: 10000,
        trialsPerSpan: 2,
        ...settings,
    };

    // Convert string settings from YAML to their correct types
    const initialSpan = parseInt(testSettings.initialSpan, 10);
    const maxSpan = parseInt(testSettings.maxSpan, 10);
    const displayTimePerDigit = parseInt(testSettings.displayTimePerDigit, 10);
    const interDigitInterval = parseInt(testSettings.interDigitInterval, 10);
    const recallTimeout = parseInt(testSettings.recallTimeout, 10);
    const trialsPerSpan = parseInt(testSettings.trialsPerSpan, 10);
    
    // --- State Variables ---
    let phase = 'idle'; // idle, presenting, recalling
    let currentSpan = initialSpan;
    let trial = 1;
    let currentSequence = [];
    let displayIndex = 0;
    let timerRef = null;
    const testData = {
        testStartTime: 0,
        testEndTime: 0,
        results: [],
        settings: testSettings,
    };

    function generateSequence(length) {
        return Array.from({ length }, () => Math.floor(Math.random() * 9) + 1);
    }

    function startTrial(spanLength) {
        phase = 'presenting';
        currentSequence = generateSequence(spanLength);
        displayIndex = 0;
        if (testData.testStartTime === 0) testData.testStartTime = performance.now();
        presentDigit();
    }
    
    function presentDigit() {
        if (phase !== 'presenting') return;
        
        container.innerHTML = `
            <div class="text-lg mb-4">Span: ${currentSpan}, Trial: ${trial}</div>
            <div id="dst-stimulus-display" class="w-full h-48 bg-gray-200 flex items-center justify-center text-6xl font-bold rounded-lg">
                ${currentSequence[displayIndex]}
            </div>
            <p class="mt-4 text-secondary">Memorize the digit.</p>
        `;

        timerRef = setTimeout(() => {
            container.querySelector('#dst-stimulus-display').textContent = '';
            displayIndex++;
            if (displayIndex < currentSequence.length) {
                timerRef = setTimeout(presentDigit, interDigitInterval);
            } else {
                startRecallPhase();
            }
        }, displayTimePerDigit);
    }

    function startRecallPhase() {
        phase = 'recalling';
        let remaining = recallTimeout;

        container.innerHTML = `
            <div class="text-lg mb-4">Time Remaining: <span id="recall-timer">${(remaining/1000).toFixed(1)}s</span></div>
            <p class="mb-2">Enter the sequence:</p>
            <input id="recall-input" type="text" inputmode="numeric" class="text-input text-2xl text-center" />
            <button id="submit-recall" class="primary-button mt-4">Submit</button>
        `;

        const input = document.getElementById('recall-input');
        input.focus();
        input.addEventListener('keydown', e => e.key === 'Enter' && e.preventDefault() & document.getElementById('submit-recall').click());
        document.getElementById('submit-recall').onclick = () => {
             handleRecallSubmit(input.value);
        };
        
        const recallTimer = setInterval(() => {
            remaining -= 1000;
            const timerEl = document.getElementById('recall-timer');
            if (timerEl) timerEl.textContent = `${(remaining/1000).toFixed(1)}s`;
        }, 1000);

        timerRef = setTimeout(() => {
            handleRecallSubmit(input.value);
        }, recallTimeout);
    }

    function handleRecallSubmit(userInput) {
        if (phase !== 'recalling') return;
        clearTimeout(timerRef);
        phase = 'feedback';

        const correct = userInput === currentSequence.join('');
        testData.results.push({ span: currentSpan, trial, sequence: currentSequence.join(''), input: userInput, correct, timestamp: performance.now() - testData.testStartTime });

        container.innerHTML = `<p class="text-2xl font-bold ${correct ? 'text-green-700' : 'text-red-700'}">${correct ? 'Correct!' : 'Incorrect'}</p>`;
        
        setTimeout(() => {
            if (correct) {
                trial = 1;
                currentSpan++;
                if (currentSpan > maxSpan) return endTest();
                startTrial(currentSpan);
            } else {
                trial++;
                if (trial > trialsPerSpan) return endTest();
                startTrial(currentSpan);
            }
        }, 1500);
    }
    
    function endTest() {
        phase = 'complete';
        testData.testEndTime = performance.now();
        container.innerHTML = `<p class="text-lg font-semibold">Test complete. Saving results...</p>`;
        if (onTestEnd) onTestEnd(testData);
    }

    container.innerHTML = `<button id="start-dst-btn" class="primary-button">Start Test</button>`;
    document.getElementById('start-dst-btn').onclick = () => startTrial(currentSpan);
}