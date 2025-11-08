const fs = require('fs');

function analyzeResults() {
  const results = JSON.parse(fs.readFileSync('benchmark/results.json'));
  const aggregate = results.aggregate;
  
  console.log('🚀 FOOD DELIVERY PERFORMANCE ANALYSIS\n');
  console.log('=' .repeat(50));
  
  // Overall Statistics
  console.log('📊 OVERALL STATISTICS:');
  console.log(`Total Requests: ${aggregate.counters['http.requests']}`);
  console.log(`Successful Responses: ${aggregate.counters['http.responses']}`);
  console.log(`Request Rate: ${aggregate.rates['http.request_rate']} RPS`);
  console.log(`Test Duration: ${((aggregate.lastMetricAt - aggregate.firstMetricAt) / 1000).toFixed(1)}s`);
  console.log();
  
  // Response Time Analysis
  console.log('⏱️  RESPONSE TIME ANALYSIS:');
  const rt = aggregate.summaries['http.response_time'];
  console.log(`Average: ${rt.mean}ms`);
  console.log(`50th percentile: ${rt.p50}ms`);
  console.log(`95th percentile: ${rt.p95}ms`);
  console.log(`99th percentile: ${rt.p99}ms`);
  console.log(`Min/Max: ${rt.min}ms / ${rt.max}ms`);
  console.log();
  
  // Error Analysis
  console.log('❌ ERROR ANALYSIS:');
  const totalRequests = aggregate.counters['http.requests'];
  const errors = {
    '503 Service Unavailable': aggregate.counters['http.codes.503'] || 0,
    '401 Unauthorized': aggregate.counters['http.codes.401'] || 0,
    '500 Server Error': aggregate.counters['http.codes.500'] || 0,
    'Timeout': aggregate.counters['errors.ETIMEDOUT'] || 0,
    'Invalid URL': aggregate.counters['errors.Invalid URL - undefined'] || 0
  };
  
  Object.entries(errors).forEach(([error, count]) => {
    if (count > 0) {
      const percentage = ((count / totalRequests) * 100).toFixed(1);
      console.log(`${error}: ${count} (${percentage}%)`);
    }
  });
  console.log();
  
  // Endpoint Performance
  console.log('🎯 ENDPOINT PERFORMANCE:');
  const endpointMetrics = Object.keys(aggregate.counters)
    .filter(key => key.includes('plugins.metrics-by-endpoint'))
    .reduce((acc, key) => {
      const endpoint = key.match(/\/[^.]*/)[0];
      const metric = key.split('.').pop();
      const count = aggregate.counters[key];
      
      if (!acc[endpoint]) acc[endpoint] = {};
      acc[endpoint][metric] = count;
      return acc;
    }, {});
  
  Object.entries(endpointMetrics).forEach(([endpoint, metrics]) => {
    console.log(`\n${endpoint}:`);
    if (metrics['codes.200']) console.log(`  ✅ 200 OK: ${metrics['codes.200']}`);
    if (metrics['codes.401']) console.log(`  🔒 401 Unauthorized: ${metrics['codes.401']}`);
    if (metrics['codes.503']) console.log(`  🚫 503 Unavailable: ${metrics['codes.503']}`);
  });
  
  // Performance Assessment
  console.log('\n🎯 PERFORMANCE ASSESSMENT:');
  const successRate = ((aggregate.counters['http.responses'] || 0) / totalRequests * 100).toFixed(1);
  console.log(`Success Rate: ${successRate}%`);
  
  if (rt.p95 < 500) {
    console.log('✅ 95th percentile response time: EXCELLENT (< 500ms)');
  } else if (rt.p95 < 1000) {
    console.log('⚠️  95th percentile response time: GOOD (500-1000ms)');
  } else {
    console.log('❌ 95th percentile response time: POOR (> 1000ms)');
  }
  
  if (aggregate.rates['http.request_rate'] > 100) {
    console.log('✅ Request throughput: EXCELLENT (> 100 RPS)');
  } else if (aggregate.rates['http.request_rate'] > 50) {
    console.log('⚠️  Request throughput: GOOD (50-100 RPS)');
  } else {
    console.log('❌ Request throughput: POOR (< 50 RPS)');
  }
  
  if (parseFloat(successRate) > 95) {
    console.log('✅ Error rate: EXCELLENT (< 5%)');
  } else if (parseFloat(successRate) > 90) {
    console.log('⚠️  Error rate: ACCEPTABLE (5-10%)');
  } else {
    console.log('❌ Error rate: CRITICAL (> 10%)');
  }
  
  // Save analysis to file
  const analysis = {
    timestamp: new Date().toISOString(),
    summary: {
      totalRequests: aggregate.counters['http.requests'],
      successRate: successRate,
      avgResponseTime: rt.mean,
      p95ResponseTime: rt.p95,
      requestRate: aggregate.rates['http.request_rate']
    },
    errors: errors,
    assessment: {
      responseTime: rt.p95 < 500 ? 'EXCELLENT' : rt.p95 < 1000 ? 'GOOD' : 'POOR',
      throughput: aggregate.rates['http.request_rate'] > 100 ? 'EXCELLENT' : aggregate.rates['http.request_rate'] > 50 ? 'GOOD' : 'POOR',
      reliability: parseFloat(successRate) > 95 ? 'EXCELLENT' : parseFloat(successRate) > 90 ? 'ACCEPTABLE' : 'CRITICAL'
    }
  };
  
  fs.writeFileSync('benchmark/analysis.json', JSON.stringify(analysis, null, 2));
  console.log('\n📄 Analysis saved to: benchmark/analysis.json');
}

analyzeResults();
