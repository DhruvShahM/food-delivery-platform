const fs = require('fs');

function analyzeResults() {
  const results = JSON.parse(fs.readFileSync('benchmark/results.json'));
  const aggregate = results && results.aggregate ? results.aggregate : {};
  const counters = aggregate.counters || {};
  const rates = aggregate.rates || {};
  const summaries = aggregate.summaries || {};

  console.log('🚀 FOOD DELIVERY PERFORMANCE ANALYSIS\n');
  console.log('=' .repeat(50));
  
  // Overall Statistics
  console.log('📊 OVERALL STATISTICS:');
  const totalRequests = counters['http.requests'] || 0;
  const successfulResponses = counters['http.responses'] || 0;
  const requestRate = rates['http.request_rate'] || 0;
  const hasTimes = typeof aggregate.firstMetricAt === 'number' && typeof aggregate.lastMetricAt === 'number';
  const durationSec = hasTimes ? ((aggregate.lastMetricAt - aggregate.firstMetricAt) / 1000).toFixed(1) : '0.0';
  console.log(`Total Requests: ${totalRequests}`);
  console.log(`Successful Responses: ${successfulResponses}`);
  console.log(`Request Rate: ${requestRate} RPS`);
  console.log(`Test Duration: ${durationSec}s`);
  console.log();
  
  // Response Time Analysis
  console.log('⏱️  RESPONSE TIME ANALYSIS:');
  const rt = summaries['http.response_time'] || { mean: 0, p50: 0, p95: 0, p99: 0, min: 0, max: 0 };
  console.log(`Average: ${rt.mean || 0}ms`);
  console.log(`50th percentile: ${rt.p50 || 0}ms`);
  console.log(`95th percentile: ${rt.p95 || 0}ms`);
  console.log(`99th percentile: ${rt.p99 || 0}ms`);
  console.log(`Min/Max: ${rt.min || 0}ms / ${rt.max || 0}ms`);
  console.log();
  
  // Error Analysis
  console.log('❌ ERROR ANALYSIS:');
  const errors = {
    '503 Service Unavailable': counters['http.codes.503'] || 0,
    '401 Unauthorized': counters['http.codes.401'] || 0,
    '500 Server Error': counters['http.codes.500'] || 0,
    'Timeout': counters['errors.ETIMEDOUT'] || 0,
    'Invalid URL': counters['errors.Invalid URL - undefined'] || 0,
    'Connection Refused': counters['errors.ECONNREFUSED'] || 0
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
  const endpointMetrics = Object.keys(counters)
    .filter(key => key.includes('plugins.metrics-by-endpoint'))
    .reduce((acc, key) => {
      const endpoint = key.match(/\/[^.]*/)[0];
      const metric = key.split('.').pop();
      const count = counters[key];
      
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
  const successRate = totalRequests > 0 ? (((successfulResponses || 0) / totalRequests) * 100).toFixed(1) : '0.0';
  console.log(`Success Rate: ${successRate}%`);
  
  if (rt.p95 < 500) {
    console.log('✅ 95th percentile response time: EXCELLENT (< 500ms)');
  } else if (rt.p95 < 1000) {
    console.log('⚠️  95th percentile response time: GOOD (500-1000ms)');
  } else {
    console.log('❌ 95th percentile response time: POOR (> 1000ms)');
  }
  
  if (requestRate > 100) {
    console.log('✅ Request throughput: EXCELLENT (> 100 RPS)');
  } else if (requestRate > 50) {
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
      totalRequests: totalRequests,
      successRate: successRate,
      avgResponseTime: rt.mean || 0,
      p95ResponseTime: rt.p95 || 0,
      requestRate: requestRate
    },
    errors: errors,
    assessment: {
      responseTime: rt.p95 < 500 ? 'EXCELLENT' : rt.p95 < 1000 ? 'GOOD' : 'POOR',
      throughput: requestRate > 100 ? 'EXCELLENT' : requestRate > 50 ? 'GOOD' : 'POOR',
      reliability: parseFloat(successRate) > 95 ? 'EXCELLENT' : parseFloat(successRate) > 90 ? 'ACCEPTABLE' : 'CRITICAL'
    }
  };
  
  fs.writeFileSync('benchmark/analysis.json', JSON.stringify(analysis, null, 2));
  console.log('\n📄 Analysis saved to: benchmark/analysis.json');
}

analyzeResults();
