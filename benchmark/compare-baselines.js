const fs = require('fs');

function compareBaselines() {
  if (!fs.existsSync('benchmark/baseline-results.json')) {
    console.log('❌ No baseline results found. Run tests first and save as baseline.');
    return;
  }
  
  const baseline = JSON.parse(fs.readFileSync('benchmark/baseline-results.json'));
  const current = JSON.parse(fs.readFileSync('benchmark/results.json'));
  
  console.log('📊 PERFORMANCE COMPARISON REPORT\n');
  console.log('=' .repeat(50));
  
  const compare = (label, baselineVal, currentVal, unit = '') => {
    const diff = ((currentVal - baselineVal) / baselineVal * 100).toFixed(1);
    const change = diff > 0 ? `+${diff}%` : `${diff}%`;
    const icon = diff > 5 ? '📈' : diff < -5 ? '📉' : '➡️';
    
    console.log(`${label}:`);
    console.log(`  Baseline: ${baselineVal}${unit}`);
    console.log(`  Current:  ${currentVal}${unit}`);
    console.log(`  Change:   ${icon} ${change}`);
    console.log();
  };
  
  compare('Request Rate', 
    baseline.aggregate.rates['http.request_rate'], 
    current.aggregate.rates['http.request_rate'], ' RPS');
    
  compare('95th Percentile Response Time', 
    baseline.aggregate.summaries['http.response_time'].p95, 
    current.aggregate.summaries['http.response_time'].p95, 'ms');
    
  compare('Average Response Time', 
    baseline.aggregate.summaries['http.response_time'].mean, 
    current.aggregate.summaries['http.response_time'].mean, 'ms');
    
  compare('Success Rate', 
    (baseline.aggregate.counters['http.responses'] / baseline.aggregate.counters['http.requests'] * 100),
    (current.aggregate.counters['http.responses'] / current.aggregate.counters['http.requests'] * 100), '%');
    
  console.log('💡 INTERPRETATION:');
  console.log('• 📈 Improvements: Performance enhancements detected');
  console.log('• 📉 Regressions: Performance degradation detected');
  console.log('• ➡️ Stable: Performance within acceptable range');
}

compareBaselines();
