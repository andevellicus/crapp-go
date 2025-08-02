class InteractionTracker {
    constructor() {
        // Basic data storage
        this.movements = [];
        this.interactions = [];
        this.keyboardEvents = [];
        this.currentQuestion = null;
        this.currentTarget = null;
        this.startTime = performance.now();
        
        // Throttling for performance
        this.lastRecordedTime = 0;
        this.throttleInterval = 50; // Record at most every 50ms
        
        // --- NEW: Periodic Sending ---
        this.sendInterval = 30000; // Send data every 30 seconds
        this.intervalId = null;

        // Store event listener references for proper cleanup
        this.mouseMoveListener = this.handleMouseMove.bind(this);
        this.keyDownListener = this.handleKeyDown.bind(this);
        this.keyUpListener = this.handleKeyUp.bind(this);
        
        // Initialize tracking
        this.setupListeners();
    }
    
    setupListeners() {
        document.addEventListener('mousemove', this.mouseMoveListener);
        document.addEventListener('keydown', this.keyDownListener);
        document.addEventListener('keyup', this.keyUpListener);
        
        // --- NEW: Listen for HTMX navigation to send data before swapping content ---
        document.body.addEventListener('htmx:beforeSwap', () => {
            this.sendData();
            this.reset();
        });

        this.mutationObserver = new MutationObserver(() => {
            this.findInteractiveElements();
            this.detectCurrentQuestion();
        });
        
        this.mutationObserver.observe(document.body, { childList: true, subtree: true });
        
        // --- NEW: Start periodic sending ---
        this.intervalId = setInterval(() => this.sendData(), this.sendInterval);
        
        this.findInteractiveElements();
        this.detectCurrentQuestion();
    }
    
    cleanup() {
        document.removeEventListener('mousemove', this.mouseMoveListener);
        document.removeEventListener('keydown', this.keyDownListener);
        document.removeEventListener('keyup', this.keyUpListener);
        
        if (this.mutationObserver) {
            this.mutationObserver.disconnect();
        }
        
        if (this.intersectionObservers) {
            this.intersectionObservers.forEach(observer => observer.disconnect());
        }

        // --- NEW: Stop periodic sending to prevent memory leaks ---
        if (this.intervalId) {
            clearInterval(this.intervalId);
        }
    }
    
    detectCurrentQuestion() {
        if (this.intersectionObservers) {
            this.intersectionObservers.forEach(observer => observer.disconnect());
        }
        this.intersectionObservers = [];
        
        document.querySelectorAll('.form-group').forEach(section => {
            let questionId = section.dataset.questionId;
            
            if (!questionId) {
                const inputElement = section.querySelector('input, textarea, select');
                if (inputElement) {
                    questionId = inputElement.name || 'unknown_question'; // Fallback
                    section.dataset.questionId = questionId;
                }
            }
            
            if (questionId) {
                const observer = new IntersectionObserver((entries) => {
                    entries.forEach(entry => {
                        if (entry.isIntersecting) {
                            this.currentQuestion = questionId;
                        }
                    });
                }, { threshold: 0.5 });

                observer.observe(section);
                this.intersectionObservers.push(observer);
            }
        });
    }
    
    findInteractiveElements() {
        document.querySelectorAll('button, input, textarea, select, label.option-label').forEach(element => {
            if (element.dataset.tracked) return;
            
            element.dataset.tracked = 'true';
            element.dataset.targetId = Math.random().toString(36).substring(2, 10);
            
            let questionId = null;
            let parentForm = element.closest('.form-group');
            if (parentForm) {
                questionId = parentForm.dataset.questionId;
            }
            
            element.dataset.questionId = questionId;
            
            const rect = element.getBoundingClientRect();
            const targetData = {
                id: element.dataset.targetId,
                questionId: questionId,
                x: rect.left + rect.width/2,
                y: rect.top + rect.height/2,
                width: rect.width,
                height: rect.height,
                type: element.tagName.toLowerCase()
            };
            
            element.addEventListener('mouseover', () => this.currentTarget = targetData);
            element.addEventListener('click', (event) => this.handleInteraction(event, targetData));
            
            if (['input', 'textarea', 'select'].includes(element.tagName.toLowerCase())) {
                element.addEventListener('focus', () => this.currentQuestion = questionId);
            }
        });
    }
    
    handleMouseMove(event) {
        const now = performance.now();
        if (now - this.lastRecordedTime < this.throttleInterval) return;
        
        this.lastRecordedTime = now;
        this.movements.push({
            x: event.clientX,
            y: event.clientY,
            timestamp: now - this.startTime,
            targetId: this.currentTarget?.id,
            questionId: this.currentQuestion
        });
    }
    
    handleKeyDown(event) {
        if (['Control', 'Shift', 'Alt', 'Meta'].includes(event.key)) return;
        
        this.keyboardEvents.push({
            type: 'keydown',
            key: event.key,
            isModifier: event.ctrlKey || event.shiftKey || event.altKey || event.metaKey,
            timestamp: performance.now() - this.startTime,
            questionId: this.currentQuestion
        });
    }
    
    handleKeyUp(event) {
        if (['Control', 'Shift', 'Alt', 'Meta'].includes(event.key)) return;
        
        this.keyboardEvents.push({
            type: 'keyup',
            key: event.key,
            isModifier: event.ctrlKey || event.shiftKey || event.altKey || event.metaKey,
            timestamp: performance.now() - this.startTime,
            questionId: this.currentQuestion
        });
    }
    
    handleInteraction(event, targetData) {
        const rect = event.target.getBoundingClientRect();
        const timestamp = performance.now() - this.startTime;
        
        targetData.x = rect.left + rect.width/2;
        targetData.y = rect.top + rect.height/2;
        
        this.interactions.push({
            targetId: targetData.id,
            targetType: targetData.type,
            questionId: targetData.questionId,
            clickX: event.clientX,
            clickY: event.clientY,
            targetX: targetData.x,
            targetY: targetData.y,
            timestamp: timestamp
        });
    }
    
    getData() {
        return {
            movements: this.movements,
            interactions: this.interactions,
            keyboardEvents: this.keyboardEvents,
            startTime: this.startTime
        };
    }
    
    sendData() {
        // --- NEW: Only send if there is data to send ---
        if (this.movements.length === 0 && this.interactions.length === 0 && this.keyboardEvents.length === 0) {
            return;
        }

        const data = this.getData();
        
        // Use `navigator.sendBeacon` for reliability when the page is unloading.
        // Fall back to fetch for other cases.
        const blob = new Blob([JSON.stringify(data)], { type: 'application/json' });
        if (navigator.sendBeacon) {
            navigator.sendBeacon('/metrics', blob);
        } else {
            fetch('/metrics', {
                method: 'POST',
                body: blob,
                keepalive: true
            });
        }

        // --- NEW: Reset after sending ---
        this.reset();
    }
    
    reset() {
        this.movements = [];
        this.interactions = [];
        this.keyboardEvents = [];
        this.startTime = performance.now();
    }
}

// Create a singleton instance
if (!window.interactionTracker) {
    window.interactionTracker = new InteractionTracker();
}

// Send any final data when the user closes the tab or navigates away from the site
window.addEventListener('beforeunload', () => {
    if (window.interactionTracker) {
        window.interactionTracker.sendData();
        window.interactionTracker.cleanup();
    }
});