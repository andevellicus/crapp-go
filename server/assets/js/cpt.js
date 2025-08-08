/**
 * Initializes and runs the Continuous Performance Test (CPT).
 * @param {string} containerId - The ID of the DOM element to render the test into.
 * @param {object} settings - Configuration options for the test from the server.
 * @param {function} onTestEnd - Callback function executed when the test is complete.
 */
function initCPT(containerId, settings, onTestEnd) {
    const container = document.getElementById(containerId);
    if (!container) {
        console.error(`CPT container with ID #${containerId} not found.`);
        return;
    }

    // Default settings, merged with server-provided settings
    const testSettings = {
        testDuration: 120000,
        stimulusDuration: 250,
        interStimulusInterval: 2000,
        targetProbability: 0.7,
        targets: ['X'],
        nonTargets: ['A', 'B', 'C', 'E', 'F', 'H', 'K', 'L'],
        ...settings,
    };

    // Convert string settings from YAML to their correct types
    const testDuration = parseInt(testSettings.testDuration, 10);
    const stimulusDuration = parseInt(testSettings.stimulusDuration, 10);
    const interStimulusInterval = parseInt(testSettings.interStimulusInterval, 10);
    const targetProbability = parseFloat(testSettings.targetProbability);
    const nonTargetArray = testSettings.nonTargets.split(',').map(s => s.trim());

    // --- State Variables ---
    let isRunning = false;
    let remainingTime = testDuration;
    let timerIntervalRef, stimulusTimeoutRef;
    let currentStimulus = null;
    let stimulusStartTime = 0;
    const testData = {
        testStartTime: 0,
        testEndTime: 0,
        stimuliPresented: [],
        responses: [],
        settings: testSettings,
    };

    function formatTime(ms) {
        const minutes = Math.floor(ms / 60000);
        const seconds = ((ms % 60000) / 1000).toFixed(0);
        return `${minutes}:${seconds < 10 ? '0' : ''}${seconds}`;
    }

    function renderActiveTest() {
        container.innerHTML = `
            <div class="text-lg mb-4">Time Remaining: <span id="cpt-timer">${formatTime(remainingTime)}</span></div>
            <div id="cpt-stimulus-display" class="w-full h-48 bg-gray-200 flex items-center justify-center text-6xl font-bold rounded-lg"></div>
            <p class="mt-4 text-secondary">Press the SPACEBAR for '${testSettings.targets[0]}' only.</p>
        `;
    }

    function presentStimulus() {
        if (!isRunning) return;
        const stimulusEl = document.getElementById('cpt-stimulus-display');
        if (!stimulusEl) return; // Stop if the element is gone

        const isTarget = Math.random() < targetProbability;
        const stimulusValue = isTarget ? testSettings.targets[0] : nonTargetArray[Math.floor(Math.random() * nonTargetArray.length)];

        stimulusStartTime = performance.now();
        currentStimulus = { value: stimulusValue, isTarget, responded: false };
        stimulusEl.textContent = stimulusValue;

        testData.stimuliPresented.push({
            value: stimulusValue,
            isTarget: isTarget,
            presentedAt: stimulusStartTime - testData.testStartTime,
        });

        // Hide the stimulus after its duration
        setTimeout(() => {
            if (stimulusEl) stimulusEl.textContent = '';
            currentStimulus = null;
            // Schedule the next stimulus
            stimulusTimeoutRef = setTimeout(presentStimulus, interStimulusInterval);
        }, stimulusDuration);
    }

    function handleKeyPress(e) {
        if (e.code !== 'Space' || !currentStimulus || currentStimulus.responded) return;
        e.preventDefault();
        currentStimulus.responded = true;
        testData.responses.push({
            stimulus: currentStimulus.value,
            isTarget: currentStimulus.isTarget,
            responseTime: performance.now() - stimulusStartTime,
            stimulusIndex: testData.stimuliPresented.length - 1
        });
    }

    function startTest() {
        isRunning = true;
        testData.testStartTime = performance.now();
        renderActiveTest();
        document.addEventListener('keydown', handleKeyPress);

        // Main timer for test duration
        timerIntervalRef = setInterval(() => {
            remainingTime -= 1000;
            const timerEl = document.getElementById('cpt-timer');
            if (timerEl) timerEl.textContent = formatTime(remainingTime);
            if (remainingTime <= 0) endTest();
        }, 1000);
        
        // **THE FIX IS HERE:** Delay the start of the stimulus presentation.
        setTimeout(presentStimulus, interStimulusInterval);
    }

    function endTest() {
        if (!isRunning) return;
        isRunning = false;
        clearInterval(timerIntervalRef);
        clearTimeout(stimulusTimeoutRef);
        document.removeEventListener('keydown', handleKeyPress);
        
        testData.testEndTime = performance.now();
        container.innerHTML = `<p class="text-lg font-semibold">Test complete. Saving results...</p>`;
        
        if (onTestEnd) {
            onTestEnd(testData);
        }
    }

    // Initial render of the start button
    container.innerHTML = `<button id="start-cpt-btn" class="primary-button">Start Test</button>`;
    document.getElementById('start-cpt-btn').addEventListener('click', startTest);
}