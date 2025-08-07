// charts.js
(function() {
    // Initialize charts function
    function initializeCharts() {
        // Check if echarts is loaded
        if (typeof echarts === 'undefined') {
            console.log('ECharts not yet loaded, retrying...');
            setTimeout(initializeCharts, 100);
            return;
        }

        // Find all chart containers that haven't been initialized
        var containers = document.querySelectorAll('.chart-container:not([data-initialized])');
        
        containers.forEach(function(container) {
            try {
                var optionsStr = container.dataset.options;
                if (!optionsStr) {
                    console.error('No options found for chart container', container.id);
                    return;
                }
                
                var options = JSON.parse(optionsStr);
                var chart = echarts.init(container);
                chart.setOption(options);
                container.setAttribute('data-initialized', 'true');
                
                // Store for resize
                container._echartsInstance = chart;
                
            } catch (error) {
                console.error('Error initializing chart:', container.id, error);
            }
        });
        
        // Set up resize handler once
        if (!window._chartsResizeHandlerSet) {
            window.addEventListener('resize', function() {
                document.querySelectorAll('.chart-container[data-initialized]').forEach(function(container) {
                    if (container._echartsInstance) {
                        container._echartsInstance.resize();
                    }
                });
            });
            window._chartsResizeHandlerSet = true;
        }
    }

    // Make the function globally available
    window.initializeCharts = initializeCharts;

    // Listen for various events
    document.addEventListener('DOMContentLoaded', initializeCharts);
    document.addEventListener('htmx:afterSwap', function(event) {
        // Add a small delay to ensure scripts are loaded
        setTimeout(initializeCharts, 100);
    });
    document.addEventListener('htmx:afterSettle', function(event) {
        // Try again after settle as well
        initializeCharts();
    });

    // Also try to initialize immediately in case the script loads after content
    if (document.readyState === 'complete' || document.readyState === 'interactive') {
        setTimeout(initializeCharts, 100);
    }
})();